package kafka

import "strings"

// matchSuffix checks if a topic name ends with the given suffix.
// Works with dot-separated topics like "xbank.accounts.opened".
func matchSuffix(topic, suffix string) bool {
	return strings.HasSuffix(topic, "."+suffix) || topic == suffix
}
