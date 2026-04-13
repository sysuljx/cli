// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

type commentDocRef struct {
	Kind  string
	Token string
}

type resolvedCommentTarget struct {
	DocID      string
	FileToken  string
	FileType   string
	ResolvedBy string
	WikiToken  string
}

type commentReplyElementInput struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	MentionUser string `json:"mention_user"`
	Link        string `json:"link"`
}

type commentMode string

const (
	commentModeLocal commentMode = "local"
	commentModeFull  commentMode = "full"
)

var DriveAddComment = common.Shortcut{
	Service:     "drive",
	Command:     "+add-comment",
	Description: "Add a full-document comment, or a local comment to selected docx text (also supports wiki URL resolving to doc/docx)",
	Risk:        "write",
	Scopes: []string{
		"docx:document:readonly",
		"docs:document.comment:create",
		"docs:document.comment:write_only",
	},
	AuthTypes: []string{"user", "bot"},
	Flags: []common.Flag{
		{Name: "doc", Desc: "document URL/token, or wiki URL that resolves to doc/docx", Required: true},
		{Name: "content", Desc: "reply_elements JSON string", Required: true},
		{Name: "full-comment", Type: "bool", Desc: "create a full-document comment; also the default when no location is provided"},
		{Name: "block-id", Desc: "anchor block ID for local comment"},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		docRef, err := parseCommentDocRef(runtime.Str("doc"))
		if err != nil {
			return err
		}

		if _, err := parseCommentReplyElements(runtime.Str("content")); err != nil {
			return err
		}

		blockID := strings.TrimSpace(runtime.Str("block-id"))
		if runtime.Bool("full-comment") && blockID != "" {
			return output.ErrValidation("--full-comment cannot be used with --block-id")
		}

		mode := resolveCommentMode(runtime.Bool("full-comment"), blockID)
		if mode == commentModeLocal && docRef.Kind == "doc" {
			return output.ErrValidation("local comments only support docx documents; use --full-comment or omit location flags for a whole-document comment")
		}

		return nil
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		docRef, _ := parseCommentDocRef(runtime.Str("doc"))
		replyElements, _ := parseCommentReplyElements(runtime.Str("content"))
		blockID := strings.TrimSpace(runtime.Str("block-id"))
		mode := resolveCommentMode(runtime.Bool("full-comment"), blockID)

		targetToken, targetFileType, resolvedBy := dryRunResolvedCommentTarget(docRef, mode)

		createPath := "/open-apis/drive/v1/files/:file_token/new_comments"
		commentBody := buildCommentCreateV2Request(targetFileType, "", replyElements)
		if mode == commentModeLocal {
			commentBody = buildCommentCreateV2Request(targetFileType, blockID, replyElements)
		}

		dry := common.NewDryRunAPI()
		switch {
		case mode == commentModeFull && resolvedBy == "wiki":
			dry.Desc("2-step orchestration: resolve wiki -> create full comment")
		case mode == commentModeFull:
			dry.Desc("1-step request: create full comment")
		case resolvedBy == "wiki":
			dry.Desc("2-step orchestration: resolve wiki -> create local comment")
		default:
			dry.Desc("1-step request: create local comment with explicit block ID")
		}

		if resolvedBy == "wiki" {
			dry.GET("/open-apis/wiki/v2/spaces/get_node").
				Desc("[1] Resolve wiki node to target document").
				Params(map[string]interface{}{"token": docRef.Token})
		}

		step := "[1]"
		createDesc := "Create full comment"
		if mode == commentModeLocal {
			createDesc = "Create local comment"
			if resolvedBy == "wiki" {
				step = "[2]"
			}
		} else if resolvedBy == "wiki" {
			step = "[2]"
		}

		return dry.POST(createPath).
			Desc(step+" "+createDesc).
			Body(commentBody).
			Set("file_token", targetToken)
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		blockID := strings.TrimSpace(runtime.Str("block-id"))
		mode := resolveCommentMode(runtime.Bool("full-comment"), blockID)

		target, err := resolveCommentTarget(ctx, runtime, runtime.Str("doc"), mode)
		if err != nil {
			return err
		}

		replyElements, err := parseCommentReplyElements(runtime.Str("content"))
		if err != nil {
			return err
		}

		if mode == commentModeLocal {
			fmt.Fprintf(runtime.IO().ErrOut, "Using explicit block ID: %s\n", blockID)
		}

		requestPath := fmt.Sprintf("/open-apis/drive/v1/files/%s/new_comments", validate.EncodePathSegment(target.FileToken))
		requestBody := buildCommentCreateV2Request(target.FileType, "", replyElements)
		if mode == commentModeLocal {
			requestBody = buildCommentCreateV2Request(target.FileType, blockID, replyElements)
		}

		if mode == commentModeLocal {
			fmt.Fprintf(runtime.IO().ErrOut, "Creating local comment in %s\n", common.MaskToken(target.FileToken))
		} else {
			fmt.Fprintf(runtime.IO().ErrOut, "Creating full comment in %s\n", common.MaskToken(target.FileToken))
		}

		data, err := runtime.CallAPI(
			"POST",
			requestPath,
			nil,
			requestBody,
		)
		if err != nil {
			return err
		}

		out := map[string]interface{}{
			"comment_id":   data["comment_id"],
			"doc_id":       target.DocID,
			"file_token":   target.FileToken,
			"file_type":    target.FileType,
			"resolved_by":  target.ResolvedBy,
			"comment_mode": string(mode),
		}
		if createdAt := firstPresentValue(data, "created_at", "create_time"); createdAt != nil {
			out["created_at"] = createdAt
		}
		if target.WikiToken != "" {
			out["wiki_token"] = target.WikiToken
		}
		if mode == commentModeLocal {
			out["anchor_block_id"] = blockID
		} else if isWhole, ok := data["is_whole"]; ok {
			out["is_whole"] = isWhole
		}

		runtime.Out(out, nil)
		return nil
	},
}

func resolveCommentMode(explicitFullComment bool, blockID string) commentMode {
	if explicitFullComment {
		return commentModeFull
	}
	if strings.TrimSpace(blockID) == "" {
		return commentModeFull
	}
	return commentModeLocal
}

func parseCommentDocRef(input string) (commentDocRef, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return commentDocRef{}, output.ErrValidation("--doc cannot be empty")
	}

	if token, ok := extractURLToken(raw, "/wiki/"); ok {
		return commentDocRef{Kind: "wiki", Token: token}, nil
	}
	if token, ok := extractURLToken(raw, "/docx/"); ok {
		return commentDocRef{Kind: "docx", Token: token}, nil
	}
	if token, ok := extractURLToken(raw, "/doc/"); ok {
		return commentDocRef{Kind: "doc", Token: token}, nil
	}
	if strings.Contains(raw, "://") {
		return commentDocRef{}, output.ErrValidation("unsupported --doc input %q: use a doc/docx URL, a docx token, or a wiki URL that resolves to doc/docx", raw)
	}
	if strings.ContainsAny(raw, "/?#") {
		return commentDocRef{}, output.ErrValidation("unsupported --doc input %q: use a docx token or a wiki URL", raw)
	}

	return commentDocRef{Kind: "docx", Token: raw}, nil
}

func dryRunResolvedCommentTarget(docRef commentDocRef, mode commentMode) (token, fileType, resolvedBy string) {
	switch docRef.Kind {
	case "docx":
		return docRef.Token, "docx", "docx"
	case "doc":
		return docRef.Token, "doc", "doc"
	case "wiki":
		if mode == commentModeFull {
			return "<resolved_file_token>", "<resolved_file_type>", "wiki"
		}
		return "<resolved_docx_token>", "docx", "wiki"
	default:
		return "<resolved_docx_token>", "docx", "docx"
	}
}

func resolveCommentTarget(ctx context.Context, runtime *common.RuntimeContext, input string, mode commentMode) (resolvedCommentTarget, error) {
	docRef, err := parseCommentDocRef(input)
	if err != nil {
		return resolvedCommentTarget{}, err
	}

	if docRef.Kind == "docx" || docRef.Kind == "doc" {
		if mode == commentModeLocal && docRef.Kind != "docx" {
			return resolvedCommentTarget{}, output.ErrValidation("local comments only support docx documents")
		}
		return resolvedCommentTarget{
			DocID:      docRef.Token,
			FileToken:  docRef.Token,
			FileType:   docRef.Kind,
			ResolvedBy: docRef.Kind,
		}, nil
	}

	fmt.Fprintf(runtime.IO().ErrOut, "Resolving wiki node: %s\n", common.MaskToken(docRef.Token))
	data, err := runtime.CallAPI(
		"GET",
		"/open-apis/wiki/v2/spaces/get_node",
		map[string]interface{}{"token": docRef.Token},
		nil,
	)
	if err != nil {
		return resolvedCommentTarget{}, err
	}

	node := common.GetMap(data, "node")
	objType := common.GetString(node, "obj_type")
	objToken := common.GetString(node, "obj_token")
	if objType == "" || objToken == "" {
		return resolvedCommentTarget{}, output.Errorf(output.ExitAPI, "api_error", "wiki get_node returned incomplete node data")
	}
	if mode == commentModeLocal && objType != "docx" {
		return resolvedCommentTarget{}, output.ErrValidation("wiki resolved to %q, but local comments currently only support docx documents", objType)
	}
	if mode == commentModeFull && objType != "docx" && objType != "doc" {
		return resolvedCommentTarget{}, output.ErrValidation("wiki resolved to %q, but full comments only support doc/docx documents", objType)
	}

	fmt.Fprintf(runtime.IO().ErrOut, "Resolved wiki to %s: %s\n", objType, common.MaskToken(objToken))
	return resolvedCommentTarget{
		DocID:      objToken,
		FileToken:  objToken,
		FileType:   objType,
		ResolvedBy: "wiki",
		WikiToken:  docRef.Token,
	}, nil
}

func parseCommentReplyElements(raw string) ([]map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, output.ErrValidation("--content cannot be empty")
	}

	var inputs []commentReplyElementInput
	if err := json.Unmarshal([]byte(raw), &inputs); err != nil {
		return nil, output.ErrValidation("--content is not valid JSON: %s\nexample: --content '[{\"type\":\"text\",\"text\":\"文本信息\"}]'", err)
	}
	if len(inputs) == 0 {
		return nil, output.ErrValidation("--content must contain at least one reply element")
	}

	replyElements := make([]map[string]interface{}, 0, len(inputs))
	for i, input := range inputs {
		index := i + 1
		elementType := strings.TrimSpace(input.Type)
		switch elementType {
		case "text":
			if strings.TrimSpace(input.Text) == "" {
				return nil, output.ErrValidation("--content element #%d type=text requires non-empty text", index)
			}
			if utf8.RuneCountInString(input.Text) > 1000 {
				return nil, output.ErrValidation("--content element #%d text exceeds 1000 characters", index)
			}
			replyElements = append(replyElements, map[string]interface{}{
				"type": "text",
				"text": input.Text,
			})
		case "mention_user":
			mentionUser := firstNonEmptyString(input.MentionUser, input.Text)
			if mentionUser == "" {
				return nil, output.ErrValidation("--content element #%d type=mention_user requires text or mention_user", index)
			}
			replyElements = append(replyElements, map[string]interface{}{
				"type":         "mention_user",
				"mention_user": mentionUser,
			})
		case "link":
			link := firstNonEmptyString(input.Link, input.Text)
			if link == "" {
				return nil, output.ErrValidation("--content element #%d type=link requires text or link", index)
			}
			replyElements = append(replyElements, map[string]interface{}{
				"type": "link",
				"link": link,
			})
		default:
			return nil, output.ErrValidation("--content element #%d has unsupported type %q; allowed values: text, mention_user, link", index, input.Type)
		}
	}

	return replyElements, nil
}

func buildCommentCreateV2Request(fileType, blockID string, replyElements []map[string]interface{}) map[string]interface{} {
	body := map[string]interface{}{
		"file_type":      fileType,
		"reply_elements": replyElements,
	}
	if strings.TrimSpace(blockID) != "" {
		body["anchor"] = map[string]interface{}{
			"block_id": blockID,
		}
	}
	return body
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func firstPresentValue(m map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := m[key]; ok && value != nil {
			return value
		}
	}
	return nil
}

func extractURLToken(raw, marker string) (string, bool) {
	idx := strings.Index(raw, marker)
	if idx < 0 {
		return "", false
	}
	token := raw[idx+len(marker):]
	if end := strings.IndexAny(token, "/?#"); end >= 0 {
		token = token[:end]
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}
