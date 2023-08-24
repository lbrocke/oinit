package oidc

import (
	"sort"

	"github.com/indigo-dc/liboidcagent-go"
	"golang.org/x/exp/maps"
)

const (
	APP_HINT = "oinit"
)

type socket struct {
	AddressEnvVar string
	Type          string
}

// GetConfiguredAccounts returns a map of configured accounts. The map key is
// the issuer URL, while the corresponding value is a list of oidc-agent
// account short names.
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

func GetToken(issuer string, scopes []string) (string, error) {
	req := liboidcagent.TokenRequest{
		IssuerURL:       issuer,
		Scopes:          scopes,
		ApplicationHint: APP_HINT,
	}

	res, err := liboidcagent.GetTokenResponse(req)
	if err != nil {
		return "", err
	}

	return res.Token, nil
}
