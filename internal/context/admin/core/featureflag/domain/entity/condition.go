package entity

import "strings"

// Operator defines comparison operators for conditions.
type Operator string

const (
	OpEquals      Operator = "eq"
	OpNotEquals   Operator = "neq"
	OpContains    Operator = "contains"
	OpStartsWith  Operator = "starts_with"
	OpEndsWith    Operator = "ends_with"
	OpIn          Operator = "in"      // comma-separated values
	OpNotIn       Operator = "not_in"
)

// Condition is a single attribute-operator-value check.
// Example: attribute="role", operator="eq", value="ADMIN"
type Condition struct {
	ID          string
	RuleGroupID string
	Attribute   string   // the context key to check (e.g. "role", "country")
	Operator    Operator
	Value       string   // expected value (or comma-separated for in/not_in)
}

// Matches checks if this condition matches against the given attributes map.
func (c Condition) Matches(attributes map[string]string) bool {
	actual, ok := attributes[c.Attribute]
	if !ok {
		return false
	}

	switch c.Operator {
	case OpEquals:
		return actual == c.Value
	case OpNotEquals:
		return actual != c.Value
	case OpContains:
		return strings.Contains(actual, c.Value)
	case OpStartsWith:
		return strings.HasPrefix(actual, c.Value)
	case OpEndsWith:
		return strings.HasSuffix(actual, c.Value)
	case OpIn:
		for _, v := range strings.Split(c.Value, ",") {
			if strings.TrimSpace(v) == actual {
				return true
			}
		}
		return false
	case OpNotIn:
		for _, v := range strings.Split(c.Value, ",") {
			if strings.TrimSpace(v) == actual {
				return false
			}
		}
		return true
	default:
		return false
	}
}
