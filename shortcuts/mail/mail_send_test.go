// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
)

// TestMailSend_SendSeparatelyHeaderInjected verifies that running `+send`
// with --send-separately injects X-Lms-Send-Separately: 1 into the
// base64url-encoded raw EML POSTed to drafts.create. Mirrors the
// inspection style used by TestMailDraftCreate_WithCalendarEventFlags
// (decode CapturedBody → raw → base64url decode → assert substring).
func TestMailSend_SendSeparatelyHeaderInjected(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"primary_email_address": "me@example.com"},
		},
	})
	draftsStub := &httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"draft_id": "draft_send_ss_001"},
		},
	}
	reg.Register(draftsStub)

	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "alice@example.com,bob@example.com",
		"--subject", "send-separately",
		"--body", "<p>hi</p>",
		"--send-separately",
	}, f, stdout)
	if err != nil {
		t.Fatalf("+send with --send-separately failed: %v", err)
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(draftsStub.CapturedBody, &reqBody); err != nil {
		t.Fatalf("unmarshal captured request body: %v", err)
	}
	raw, _ := reqBody["raw"].(string)
	decoded, decErr := base64.URLEncoding.DecodeString(raw)
	if decErr != nil {
		t.Fatalf("base64url decode raw: %v", decErr)
	}
	eml := string(decoded)
	if !strings.Contains(eml, "X-Lms-Send-Separately: 1") {
		t.Errorf("expected X-Lms-Send-Separately: 1 in EML, got:\n%s", eml)
	}
	if c := strings.Count(eml, "X-Lms-Send-Separately"); c != 1 {
		t.Errorf("expected X-Lms-Send-Separately to appear exactly 1 time, got %d", c)
	}
}

// TestMailSend_SendSeparatelyAbsentByDefault verifies that omitting the
// flag results in no X-Lms-Send-Separately header.
func TestMailSend_SendSeparatelyAbsentByDefault(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"primary_email_address": "me@example.com"},
		},
	})
	draftsStub := &httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"draft_id": "draft_send_ss_002"},
		},
	}
	reg.Register(draftsStub)

	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "alice@example.com",
		"--subject", "no separately",
		"--body", "<p>hi</p>",
	}, f, stdout)
	if err != nil {
		t.Fatalf("+send without --send-separately failed: %v", err)
	}

	var reqBody map[string]interface{}
	if err := json.Unmarshal(draftsStub.CapturedBody, &reqBody); err != nil {
		t.Fatalf("unmarshal captured request body: %v", err)
	}
	raw, _ := reqBody["raw"].(string)
	decoded, decErr := base64.URLEncoding.DecodeString(raw)
	if decErr != nil {
		t.Fatalf("base64url decode raw: %v", decErr)
	}
	eml := string(decoded)
	if strings.Contains(eml, "X-Lms-Send-Separately") {
		t.Errorf("expected no X-Lms-Send-Separately header by default, got:\n%s", eml)
	}
}

// TestMailSend_SendSeparatelyDryRunPreview verifies that the dry-run
// preview map surfaces the send_separately key, mirroring +draft-create.
func TestMailSend_SendSeparatelyDryRunPreview(t *testing.T) {
	f, stdout, _, _ := mailShortcutTestFactory(t)

	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "alice@example.com,bob@example.com",
		"--subject", "preview",
		"--body", "<p>hello</p>",
		"--send-separately",
		"--dry-run",
	}, f, stdout)
	if err != nil {
		t.Fatalf("dry-run failed: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, `"send_separately"`) {
		t.Fatalf("expected dry-run preview to surface send_separately, got: %s", out)
	}
	if !strings.Contains(out, "true") {
		t.Fatalf("expected dry-run preview send_separately=true, got: %s", out)
	}
}

// TestMailSend_SendSeparatelySingleRecipientWarn checks the warning is
// emitted to stderr when there is only one recipient and --send-separately
// is set; the command must NOT fail.
func TestMailSend_SendSeparatelySingleRecipientWarn(t *testing.T) {
	f, stdout, stderr, reg := mailShortcutTestFactory(t)

	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"primary_email_address": "me@example.com"},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"draft_id": "draft_send_ss_warn"},
		},
	})

	err := runMountedMailShortcut(t, MailSend, []string{
		"+send",
		"--to", "alice@example.com",
		"--subject", "single",
		"--body", "<p>hi</p>",
		"--send-separately",
	}, f, stdout)
	if err != nil {
		t.Fatalf("expected +send to succeed with single recipient + send-separately, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "--send-separately has no observable effect with only 1 recipient") {
		t.Errorf("expected single-recipient warning on stderr, got: %s", stderr.String())
	}
}
