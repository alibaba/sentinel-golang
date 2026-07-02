package base

import "testing"

func TestMockRuleGetGroupID(t *testing.T) {
	rule := &MockRule{Id: "rule-id", GroupID: "group-id"}
	if got := rule.GetGroupID(); got != "group-id" {
		t.Fatalf("GetGroupID() = %q, want %q", got, "group-id")
	}

	rule.GroupID = ""
	if got := rule.GetGroupID(); got != "" {
		t.Fatalf("GetGroupID() with no group = %q, want empty", got)
	}
}
