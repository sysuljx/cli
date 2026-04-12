// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"context"
	"fmt"
	"strconv"

	"github.com/larksuite/cli/shortcuts/common"
	draftpkg "github.com/larksuite/cli/shortcuts/mail/draft"
)

var MailDraftSend = common.Shortcut{
	Service:     "mail",
	Command:     "+draft-send",
	Description: "Send an existing draft by draft ID. Use --send-time or --send-after to schedule the send for a specific time.",
	Risk:        "write",
	Scopes:      []string{"mail:user_mailbox.message:send", "mail:user_mailbox:readonly"},
	AuthTypes:   []string{"user"},
	Flags: []common.Flag{
		{Name: "mailbox", Desc: "Optional. Mailbox email address (default: me). Use 'me' for the current user's mailbox."},
		{Name: "draft-id", Desc: "Required. The draft ID to send.", Required: true},
		{Name: "send-time", Desc: "Optional. Unix timestamp in seconds for scheduled sending. If empty or 0, the draft is sent immediately. Must be at least 5 minutes from now."},
		{Name: "send-after", Desc: "Optional. Duration from now until send time, e.g. '30m', '2h', '1d'. Cannot be less than 5 minutes. Cannot be used together with --send-time."},
	},
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		mailboxID := resolveMailboxID(runtime)
		draftID := runtime.Str("draft-id")
		sendTime, _ := resolveScheduledSendTime(runtime)
		api := common.NewDryRunAPI().
			Desc("Send draft (scheduled)").
			POST(mailboxPath(mailboxID, "drafts", draftID, "send"))
		if sendTime > 0 {
			api.Body(map[string]interface{}{
				"send_time": strconv.FormatInt(sendTime, 10),
			})
		}
		return api
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		draftID := runtime.Str("draft-id")
		if draftID == "" {
			return fmt.Errorf("--draft-id is required")
		}
		if _, err := resolveScheduledSendTime(runtime); err != nil {
			return err
		}
		return nil
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		mailboxID := resolveMailboxID(runtime)
		draftID := runtime.Str("draft-id")
		sendTime, _ := resolveScheduledSendTime(runtime)

		resData, err := draftpkg.SendWithSendTime(runtime, mailboxID, draftID, sendTime)
		if err != nil {
			return fmt.Errorf("failed to send draft: %w", err)
		}

		if sendTime > 0 {
			runtime.Out(map[string]interface{}{
				"message_id": resData["message_id"],
				"thread_id":  resData["thread_id"],
				"state":     "SCHEDULED",
				"send_time": sendTime,
			}, nil)
		} else {
			runtime.Out(map[string]interface{}{
				"message_id": resData["message_id"],
				"thread_id":  resData["thread_id"],
			}, nil)
		}
		return nil
	},
}
