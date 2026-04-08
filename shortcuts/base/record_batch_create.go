// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"context"

	"github.com/larksuite/cli/shortcuts/common"
)

var BaseRecordBatchCreate = common.Shortcut{
	Service:     "base",
	Command:     "+record-batch-create",
	Description: "Batch create records",
	Risk:        "write",
	Scopes:      []string{"base:record:create"},
	AuthTypes:   authTypes(),
	Flags: []common.Flag{
		baseTokenFlag(true),
		tableRefFlag(true),
		{Name: "json", Desc: "batch create JSON object, e.g. {\"fields\":[],\"rows\":[]}", Required: true},
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return validateRecordJSON(runtime)
	},
	DryRun: dryRunRecordBatchCreate,
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeRecordBatchCreate(runtime)
	},
}
