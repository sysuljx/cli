// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"errors"
	"fmt"

	"github.com/larksuite/cli/shortcuts/common"
)

// CancelScheduledSend cancels a scheduled draft send. The message must be in
// SCHEDULED state (DeliveryState == SCHEDULED). After cancellation, the message
// is returned to DRAFT state.
var CancelScheduledSend = common.Shortcut{
	Service:     "mail",
	Command:     "+cancel-scheduled-send",
	Description: "Cancel a scheduled email that has not been sent yet. The message will be moved back to drafts.",
	Risk:        "write",
	Scopes:      []string{"mail:user_mailbox.message:send", "mail:user_mailbox:readonly"},
	AuthTypes:   []string{"user"},
	Flags: []common.Flag{
		{Name: "mailbox", Desc: "Optional. Mailbox email address (default: me). Use 'me' for the current user's mailbox."},
		{Name: "message-id", Desc: "Required. The message ID (messageBizID) of the scheduled message to cancel.", Required: true},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		messageID := runtime.Str("message-id")
		if messageID == "" {
			return errors.New("--message-id is required")
		}
		return nil
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		mailboxID := resolveMailboxID(runtime)
		messageID := runtime.Str("message-id")

		_, err := runtime.CallAPI("POST", mailboxPath(mailboxID, "messages", messageID, "cancel_scheduled_send"), nil, nil)
		if err != nil {
			return fmt.Errorf("failed to cancel scheduled send: %w", err)
		}

		runtime.Out(map[string]interface{}{
			"message_id": messageID,
			"state":      "DRAFT",
		}, nil)
		return nil
	},
}
