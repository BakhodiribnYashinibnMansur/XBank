package domain

import (
	"testing"
)

func TestNewRuleGroup_Success(t *testing.T) {
	rg, err := NewRuleGroup("flag-1", "Admin Only", "true", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rg.FlagID != "flag-1" {
		t.Errorf("FlagID expected flag-1, got: %s", rg.FlagID)
	}
	if rg.Name != "Admin Only" {
		t.Errorf("Name expected Admin Only, got: %s", rg.Name)
	}
	if rg.Value != "true" {
		t.Errorf("Value expected true, got: %s", rg.Value)
	}
	if rg.Priority != 1 {
		t.Errorf("Priority expected 1, got: %d", rg.Priority)
	}
}

func TestNewRuleGroup_EmptyName(t *testing.T) {
	_, err := NewRuleGroup("flag-1", "", "true", 1)
	if err != ErrEmptyRuleGroupName {
		t.Errorf("expected ErrEmptyRuleGroupName, got: %v", err)
	}
}

func TestRuleGroup_AddCondition(t *testing.T) {
	rg, _ := NewRuleGroup("flag-1", "Test", "val", 0)
	c := Condition{Attribute: "role", Operator: OpEquals, Value: "ADMIN"}
	rg.AddCondition(c)

	if len(rg.Conditions) != 1 {
		t.Errorf("Conditions length expected 1, got: %d", len(rg.Conditions))
	}
}

func TestRuleGroup_Matches_AllConditions(t *testing.T) {
	rg, _ := NewRuleGroup("flag-1", "Test", "val", 0)
	rg.AddCondition(Condition{Attribute: "role", Operator: OpEquals, Value: "ADMIN"})
	rg.AddCondition(Condition{Attribute: "country", Operator: OpEquals, Value: "UZ"})

	attrs := map[string]string{"role": "ADMIN", "country": "UZ"}
	if !rg.Matches(attrs) {
		t.Error("should match when all conditions pass")
	}
}

func TestRuleGroup_Matches_PartialFail(t *testing.T) {
	rg, _ := NewRuleGroup("flag-1", "Test", "val", 0)
	rg.AddCondition(Condition{Attribute: "role", Operator: OpEquals, Value: "ADMIN"})
	rg.AddCondition(Condition{Attribute: "country", Operator: OpEquals, Value: "US"})

	attrs := map[string]string{"role": "ADMIN", "country": "UZ"}
	if rg.Matches(attrs) {
		t.Error("should not match when any condition fails (AND logic)")
	}
}

func TestRuleGroup_Matches_NoConditions(t *testing.T) {
	rg, _ := NewRuleGroup("flag-1", "Test", "val", 0)

	attrs := map[string]string{"role": "ADMIN"}
	if rg.Matches(attrs) {
		t.Error("rule group with no conditions should not match")
	}
}
