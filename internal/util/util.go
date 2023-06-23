package util

import "strings"

// matchesHost determines whether given host and port match host2 and port2.
// host2 may be a wildcard domain in the form of
//
//	*.example.com
//
// which matches any subdomain of example.com, but not example.com itself.
func MatchesHost(host string, port string, host2 string, port2 string) bool {
	if strings.HasPrefix(host2, "*.") {
		root, _ := strings.CutPrefix(host2, "*.")

		return strings.HasSuffix(host, root) && host != root && port == port2
	} else {
		return host == host2 && port == port2
	}
}
