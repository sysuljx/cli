// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package draft

import (
	"strings"
	"testing"
)

// fixtureSendSeparatelyDraft is a minimal draft used by every send-
// separately patch test below. Keeping it inline (vs a shared helper)
// keeps each test self-contained.
const fixtureSendSeparatelyDraft = `Subject: Test
From: Alice <alice@example.com>
To: Bob <bob@example.com>
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

hello
`

// TestApplySetSendSeparatelyTrueAddsHeader verifies the "true" value
// inserts X-Lms-Send-Separately: 1 once.
func TestApplySetSendSeparatelyTrueAddsHeader(t *testing.T) {
	snapshot := mustParseFixtureDraft(t, fixtureSendSeparatelyDraft)
	err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
		Ops: []PatchOp{{Op: "set_send_separately", Value: "true"}},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	got := headerValue(snapshot.Headers, "X-Lms-Send-Separately")
	if got != "1" {
		t.Fatalf("X-Lms-Send-Separately = %q, want %q", got, "1")
	}
}

// TestApplySetSendSeparatelyOneAddsHeader covers the "1" alias for true.
func TestApplySetSendSeparatelyOneAddsHeader(t *testing.T) {
	snapshot := mustParseFixtureDraft(t, fixtureSendSeparatelyDraft)
	err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
		Ops: []PatchOp{{Op: "set_send_separately", Value: "1"}},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "1" {
		t.Fatalf("X-Lms-Send-Separately = %q, want %q", got, "1")
	}
}

// TestApplySetSendSeparatelyFalseRemovesHeader verifies "false" deletes
// any existing X-Lms-Send-Separately header.
func TestApplySetSendSeparatelyFalseRemovesHeader(t *testing.T) {
	snapshot := mustParseFixtureDraft(t, `Subject: Test
From: Alice <alice@example.com>
To: Bob <bob@example.com>
X-Lms-Send-Separately: 1
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

hello
`)
	if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "1" {
		t.Fatalf("precondition: header missing; got %q", got)
	}
	err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
		Ops: []PatchOp{{Op: "set_send_separately", Value: "false"}},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "" {
		t.Fatalf("X-Lms-Send-Separately still present: %q", got)
	}
}

// TestApplySetSendSeparatelyZeroRemovesHeader covers the "0" alias for
// false.
func TestApplySetSendSeparatelyZeroRemovesHeader(t *testing.T) {
	snapshot := mustParseFixtureDraft(t, `Subject: Test
From: Alice <alice@example.com>
To: Bob <bob@example.com>
X-Lms-Send-Separately: 1
MIME-Version: 1.0
Content-Type: text/plain; charset=UTF-8

hello
`)
	err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
		Ops: []PatchOp{{Op: "set_send_separately", Value: "0"}},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "" {
		t.Fatalf("X-Lms-Send-Separately still present: %q", got)
	}
}

// TestApplySetSendSeparatelyCaseInsensitive verifies the value parsing
// tolerates mixed case (mirroring the data-access caseEqualFold check).
func TestApplySetSendSeparatelyCaseInsensitive(t *testing.T) {
	for _, v := range []string{"TRUE", "True", "tRuE"} {
		t.Run(v, func(t *testing.T) {
			snapshot := mustParseFixtureDraft(t, fixtureSendSeparatelyDraft)
			err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
				Ops: []PatchOp{{Op: "set_send_separately", Value: v}},
			})
			if err != nil {
				t.Fatalf("Apply(%q) error = %v", v, err)
			}
			if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "1" {
				t.Fatalf("X-Lms-Send-Separately = %q, want %q", got, "1")
			}
		})
	}
}

// TestApplySetSendSeparatelyInvalidValueRejected verifies values outside
// {true,false,1,0} are rejected at Validate time (before applyOp runs),
// so the snapshot is left untouched.
func TestApplySetSendSeparatelyInvalidValueRejected(t *testing.T) {
	for _, v := range []string{"yes", "no", "on", "", "2", "maybe"} {
		t.Run("value="+v, func(t *testing.T) {
			snapshot := mustParseFixtureDraft(t, fixtureSendSeparatelyDraft)
			err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
				Ops: []PatchOp{{Op: "set_send_separately", Value: v}},
			})
			if err == nil {
				t.Fatalf("expected error for invalid value %q, got nil", v)
			}
			if !strings.Contains(err.Error(), "set_send_separately") {
				t.Fatalf("error %v does not mention op name; got %v", err, err)
			}
			if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "" {
				t.Fatalf("snapshot mutated on validation failure; X-Lms-Send-Separately = %q", got)
			}
		})
	}
}

// TestPatchOpValidateSendSeparately exercises the PatchOp.Validate()
// shape directly (the same gate that Patch.Validate() invokes for every
// op).
func TestPatchOpValidateSendSeparately(t *testing.T) {
	for _, v := range []string{"true", "false", "1", "0", "TRUE", "False"} {
		if err := (PatchOp{Op: "set_send_separately", Value: v}).Validate(); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", v, err)
		}
	}
	for _, v := range []string{"yes", "no", "", "2"} {
		if err := (PatchOp{Op: "set_send_separately", Value: v}).Validate(); err == nil {
			t.Errorf("Validate(%q) = nil, want error", v)
		}
	}
}

// TestApplySetSendSeparatelyTrueThenFalse verifies the toggle semantics
// (set, then unset) end up with no header — exercising the upsert /
// remove path together in a single test.
func TestApplySetSendSeparatelyTrueThenFalse(t *testing.T) {
	snapshot := mustParseFixtureDraft(t, fixtureSendSeparatelyDraft)
	err := Apply(&DraftCtx{FIO: testFIO}, snapshot, Patch{
		Ops: []PatchOp{
			{Op: "set_send_separately", Value: "true"},
			{Op: "set_send_separately", Value: "false"},
		},
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if got := headerValue(snapshot.Headers, "X-Lms-Send-Separately"); got != "" {
		t.Fatalf("X-Lms-Send-Separately should be cleared after true→false; got %q", got)
	}
}
