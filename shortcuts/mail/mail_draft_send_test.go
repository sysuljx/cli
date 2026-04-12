// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"strings"
	"testing"

	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

// runtimeForMailDraftSendTest creates a RuntimeContext with the given flag values.
func runtimeForMailDraftSendTest(t *testing.T, values map[string]string) *common.RuntimeContext {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	for _, fl := range MailDraftSend.Flags {
		switch fl.Type {
		case "bool":
			cmd.Flags().Bool(fl.Name, fl.Default == "true", "")
		case "int":
			cmd.Flags().Int(fl.Name, 0, "")
		default:
			cmd.Flags().String(fl.Name, fl.Default, "")
		}
	}
	// Set values after registering flags
	for k, v := range values {
		if err := cmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set flag --%s failed: %v", k, err)
		}
	}
	return common.TestNewRuntimeContext(cmd, nil)
}

func TestMailDraftSend_Validate_Success(t *testing.T) {
	rt := runtimeForMailDraftSendTest(t, map[string]string{
		"draft-id": "DR_test123",
		"mailbox":  "me",
	})
	err := MailDraftSend.Validate(context.Background(), rt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailDraftSend_Validate_MissingDraftID(t *testing.T) {
	rt := runtimeForMailDraftSendTest(t, map[string]string{
		"mailbox": "me",
	})
	err := MailDraftSend.Validate(context.Background(), rt)
	if err == nil {
		t.Fatal("expected error for missing --draft-id")
	}
	if !strings.Contains(err.Error(), "--draft-id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailDraftSend_Validate_InvalidSendTime(t *testing.T) {
	// send-time that is in the past should fail
	rt := runtimeForMailDraftSendTest(t, map[string]string{
		"draft-id":  "DR_test123",
		"mailbox":   "me",
		"send-time": "1000000", // way in the past
	})
	err := MailDraftSend.Validate(context.Background(), rt)
	if err == nil {
		t.Fatal("expected error for past send-time")
	}
	if !strings.Contains(err.Error(), "at least 5 minutes from now") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailDraftSend_Validate_InvalidSendAfter(t *testing.T) {
	// send-after 30s should fail (less than 5 minutes)
	rt := runtimeForMailDraftSendTest(t, map[string]string{
		"draft-id":   "DR_test123",
		"mailbox":    "me",
		"send-after": "30s",
	})
	err := MailDraftSend.Validate(context.Background(), rt)
	if err == nil {
		t.Fatal("expected error for --send-after 30s")
	}
	if !strings.Contains(err.Error(), "at least 5 minutes from now") {
		t.Fatalf("unexpected error: %v", err)
	}
}
