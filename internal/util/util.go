// Package util provides utility functions for various common tasks.
package util

import (
	"os"
	"strings"
)

// MatchesHost determines whether the given host and port match host2 and port2.
// The host2 parameter may contain wildcard domains in the form of "*.example.com",
// which match any subdomain of example.com, but not example.com itself.
//
// Example usage:
//
//	result := MatchesHost("example.com", "22", "example.com", "22")
//	// result will be true
//
//	result := MatchesHost("sub.example.com", "22", "*.example.com", "22")
//	// result will be true
//
//	result := MatchesHost("example.com", "22", "*.example.com", "22")
//	// result will be false
func MatchesHost(host string, port string, host2 string, port2 string) bool {
	if strings.HasPrefix(host2, "*.") {
		root, _ := strings.CutPrefix(host2, "*.")

		return strings.HasSuffix(host, root) && host != root && port == port2
	} else {
		return host == host2 && port == port2
	}
}

// Getenvs retrieves environment variable values for multiple keys and returns
// the first non-empty value found, or an empty string if none of the keys are set.
//
// Example usage:
//
//	value := Getenvs("MY_VAR", "ANOTHER_VAR", "YET_ANOTHER_VAR")
//	// value will contain the value of the first set environment variable among "MY_VAR", "ANOTHER_VAR", and "YET_ANOTHER_VAR",
//	// or an empty string if none of them are set.
func Getenvs(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}

	return ""
}
