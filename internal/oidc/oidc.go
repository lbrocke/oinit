package oidc

import (
	"net"
	"os"
	"sort"

	"github.com/indigo-dc/liboidcagent-go"
	"golang.org/x/exp/maps"
)

type socket struct {
	AddressEnvVar string
	Type          string
}

func AgentIsRunning() bool {
	// From https://github.com/indigo-dc/liboidcagent-go/blob/master/comm.go
	sockets := []socket{
		{
			"OIDC_SOCK",
			"unix",
		},
		{
			"OIDC_REMOTE_SOCK",
			"tcp",
		},
	}

	for _, sock := range sockets {
		if _, err := net.Dial(sock.Type, os.Getenv(sock.AddressEnvVar)); err == nil {
			return true
		}
	}

	return false
}

func GetConfiguredAccounts() map[string][]string {
	// liboidcagent provides GetConfiguredAccounts() which however only returns
	// the short names of accounts, no issuer URLs.
	// Make use of GetAccountInfos() to build a map of issuers with existing
	// accounts.

	accounts := make(map[string][]string)

	infos, err := liboidcagent.GetAccountInfos()
	if err != nil {
		// This may also happen if oidc-agent 5 is not installed, as previous
		// versions do not support this call.
		return accounts
	}

	for issuer, info := range infos {
		accs := maps.Keys(info.Accounts)
		sort.Strings(accs)

		accounts[issuer] = accs
	}

	return accounts
}

func GetToken(issuer string) (string, error) {
	req := liboidcagent.TokenRequest{
		IssuerURL:       issuer,
		Scopes:          []string{"openid", "profile", "offline_access"},
		ApplicationHint: "oinit",
	}

	res, err := liboidcagent.GetTokenResponse(req)
	if err != nil {
		return "", err
	}

	return res.Token, err
}
