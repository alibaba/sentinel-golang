package isolation

import "testing"

func TestRuleGetGroupID(t *testing.T) {
	t.Run("configured independently from rule ID", func(t *testing.T) {
		rule := &Rule{ID: "rule-id", GroupID: "group-id"}
		if got := rule.GetGroupID(); got != "group-id" {
			t.Fatalf("GetGroupID() = %q, want %q", got, "group-id")
		}
	})

	t.Run("does not fall back to rule ID", func(t *testing.T) {
		rule := &Rule{ID: "rule-id"}
		if got := rule.GetGroupID(); got != "" {
			t.Fatalf("GetGroupID() with no group = %q, want empty", got)
		}
	})
}
