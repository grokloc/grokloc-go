// Package security provides crypto and hashing support
package security

import "strings"

// SafeStr makes sure a string is safe for storage
func SafeStr(s string) bool {
	return len(s) != 0 && !strings.ContainsAny(s, "'\"`")
}
