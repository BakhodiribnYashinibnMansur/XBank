package domain

import (
	"testing"
)

func TestCondition_Matches(t *testing.T) {
	attrs := map[string]string{
		"role":    "ADMIN",
		"country": "UZ",
		"email":   "admin@xbank.uz",
		"plan":    "premium",
	}

	tests := []struct {
		name      string
		condition Condition
		want      bool
	}{
		{
			name:      "equals match",
			condition: Condition{Attribute: "role", Operator: OpEquals, Value: "ADMIN"},
			want:      true,
		},
		{
			name:      "equals no match",
			condition: Condition{Attribute: "role", Operator: OpEquals, Value: "USER"},
			want:      false,
		},
		{
			name:      "not equals match",
			condition: Condition{Attribute: "role", Operator: OpNotEquals, Value: "USER"},
			want:      true,
		},
		{
			name:      "not equals no match",
			condition: Condition{Attribute: "role", Operator: OpNotEquals, Value: "ADMIN"},
			want:      false,
		},
		{
			name:      "contains match",
			condition: Condition{Attribute: "email", Operator: OpContains, Value: "xbank"},
			want:      true,
		},
		{
			name:      "contains no match",
			condition: Condition{Attribute: "email", Operator: OpContains, Value: "gmail"},
			want:      false,
		},
		{
			name:      "starts_with match",
			condition: Condition{Attribute: "email", Operator: OpStartsWith, Value: "admin"},
			want:      true,
		},
		{
			name:      "starts_with no match",
			condition: Condition{Attribute: "email", Operator: OpStartsWith, Value: "user"},
			want:      false,
		},
		{
			name:      "ends_with match",
			condition: Condition{Attribute: "email", Operator: OpEndsWith, Value: ".uz"},
			want:      true,
		},
		{
			name:      "ends_with no match",
			condition: Condition{Attribute: "email", Operator: OpEndsWith, Value: ".com"},
			want:      false,
		},
		{
			name:      "in match",
			condition: Condition{Attribute: "country", Operator: OpIn, Value: "UZ,KZ,RU"},
			want:      true,
		},
		{
			name:      "in no match",
			condition: Condition{Attribute: "country", Operator: OpIn, Value: "US,UK,DE"},
			want:      false,
		},
		{
			name:      "not_in match (not in list)",
			condition: Condition{Attribute: "country", Operator: OpNotIn, Value: "US,UK,DE"},
			want:      true,
		},
		{
			name:      "not_in no match (in list)",
			condition: Condition{Attribute: "country", Operator: OpNotIn, Value: "UZ,KZ,RU"},
			want:      false,
		},
		{
			name:      "missing attribute",
			condition: Condition{Attribute: "nonexistent", Operator: OpEquals, Value: "val"},
			want:      false,
		},
		{
			name:      "unknown operator",
			condition: Condition{Attribute: "role", Operator: "unknown_op", Value: "ADMIN"},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition.Matches(attrs)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCondition_In_WithSpaces(t *testing.T) {
	attrs := map[string]string{"role": "ADMIN"}
	c := Condition{Attribute: "role", Operator: OpIn, Value: "USER, ADMIN, MANAGER"}
	if !c.Matches(attrs) {
		t.Error("OpIn should trim spaces in comma-separated values")
	}
}
