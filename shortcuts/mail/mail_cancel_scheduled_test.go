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

// runtimeForCancelScheduledSendTest creates a RuntimeContext with the given flag values.
func runtimeForCancelScheduledSendTest(t *testing.T, values map[string]string) *common.RuntimeContext {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	for _, fl := range CancelScheduledSend.Flags {
		switch fl.Type {
		case "bool":
			cmd.Flags().Bool(fl.Name, fl.Default == "true", "")
		case "int":
			cmd.Flags().Int(fl.Name, 0, "")
		default:
			cmd.Flags().String(fl.Name, fl.Default, "")
		}
	}
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatalf("parse flags failed: %v", err)
	}
	for k, v := range values {
		if err := cmd.Flags().Set(k, v); err != nil {
			t.Fatalf("set flag --%s failed: %v", k, err)
		}
	}
	return common.TestNewRuntimeContext(cmd, nil)
}

func TestCancelScheduledSend_Validate_Success(t *testing.T) {
	rt := runtimeForCancelScheduledSendTest(t, map[string]string{
		"message-id": "MSG_abc123",
		"mailbox":    "me",
	})

	err := CancelScheduledSend.Validate(context.Background(), rt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelScheduledSend_Validate_MissingMessageID(t *testing.T) {
	rt := runtimeForCancelScheduledSendTest(t, map[string]string{
		"mailbox": "me",
	})

	err := CancelScheduledSend.Validate(context.Background(), rt)
	if err == nil {
		t.Fatal("expected error for missing --message-id")
	}
	if !strings.Contains(err.Error(), "--message-id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelScheduledSend_Validate_EmptyMessageID(t *testing.T) {
	rt := runtimeForCancelScheduledSendTest(t, map[string]string{
		"message-id": "",
		"mailbox":    "me",
	})

	err := CancelScheduledSend.Validate(context.Background(), rt)
	if err == nil {
		t.Fatal("expected error for empty --message-id")
	}
	if !strings.Contains(err.Error(), "--message-id is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
