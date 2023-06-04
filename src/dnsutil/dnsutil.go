package dnsutil

import (
	"errors"
	"net"
	"strings"
)

const (
	TXT_PREFIX = "_oinit-ca."
)

// Try to find the CA for a given ssh server host name by
// querying the domain name system.
//
// For a given ssh server login.example.com, a TXT record
// for either _oinit-ca.login.example.com or _oinit-ca.example.com
// is expected.
// Wildcard domains such as *.login.example.com are supported and
// will result in similar lookups of _oinit-ca.login.example.com and
// _oinit-ca.example.com
func LookupCA(host string) (string, error) {
	lookup1, _ := strings.CutPrefix(host, "*.")

	records, err := net.LookupTXT(TXT_PREFIX + lookup1)
	if err == nil && len(records) > 0 {
		return records[0], nil
	}

	// remove subdomain and try again
	_, lookup2, found := strings.Cut(lookup1, ".")
	if !found || strings.Count(lookup2, ".") == 0 {
		return "", errors.New("could not find DNS record")
	}

	records, err = net.LookupTXT(TXT_PREFIX + lookup2)
	if err == nil && len(records) > 0 {
		return records[0], nil
	}

	return "", errors.New("could not find DNS record")
}
