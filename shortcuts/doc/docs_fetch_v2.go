// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"fmt"
	"io"

	"github.com/larksuite/cli/shortcuts/common"
)

// v2FetchFlags returns the flag definitions for the v2 (OpenAPI) fetch path.
func v2FetchFlags() []common.Flag {
	return []common.Flag{
		{Name: "doc-format", Desc: "content format", Hidden: true, Default: "xml", Enum: []string{"xml", "markdown", "text"}},
		{Name: "detail", Desc: "export detail level: simple (read-only) | with-ids (block IDs for cross-referencing) | full (all attrs for editing)", Hidden: true, Default: "simple", Enum: []string{"simple", "with-ids", "full"}},
		{Name: "revision-id", Desc: "document revision (-1 = latest)", Hidden: true, Type: "int", Default: "-1"},
	}
}

func dryRunFetchV2(_ context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
	ref, err := parseDocumentRef(runtime.Str("doc"))
	if err != nil {
		return common.NewDryRunAPI().Desc(fmt.Sprintf("error: %v", err))
	}
	body := buildFetchBody(runtime)
	apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s/fetch", ref.Token)
	return common.NewDryRunAPI().
		POST(apiPath).
		Desc("OpenAPI: fetch document").
		Body(body).
		Set("document_id", ref.Token)
}

func executeFetchV2(_ context.Context, runtime *common.RuntimeContext) error {
	ref, err := parseDocumentRef(runtime.Str("doc"))
	if err != nil {
		return err
	}

	apiPath := fmt.Sprintf("/open-apis/docs_ai/v1/documents/%s/fetch", ref.Token)
	body := buildFetchBody(runtime)

	data, err := doDocAPI(runtime, "POST", apiPath, body)
	if err != nil {
		return err
	}

	runtime.OutFormatRaw(data, nil, func(w io.Writer) {
		if doc, ok := data["document"].(map[string]interface{}); ok {
			if content, ok := doc["content"].(string); ok {
				fmt.Fprintln(w, content)
			}
		}
	})
	return nil
}

func buildFetchBody(runtime *common.RuntimeContext) map[string]interface{} {
	body := map[string]interface{}{
		"format": runtime.Str("doc-format"),
	}
	if v := runtime.Int("revision-id"); v > 0 {
		body["revision_id"] = v
	}

	detail := runtime.Str("detail")
	switch detail {
	case "", "simple":
		body["export_option"] = map[string]interface{}{
			"export_block_id":        false,
			"export_style_attrs":     false,
			"export_cite_extra_data": false,
		}
	case "with-ids":
		body["export_option"] = map[string]interface{}{
			"export_block_id": true,
		}
	case "full":
		body["export_option"] = map[string]interface{}{
			"export_block_id":        true,
			"export_style_attrs":     true,
			"export_cite_extra_data": true,
		}
	}

	return body
}
