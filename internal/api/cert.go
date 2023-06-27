package api

import (
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
)

const (
	PRINCIPAL     = "oinit"
	FORCE_COMMAND = "oinit-switch"
)

// generateUserCertificate generates a new OpenSSH certificate based on the
// given public key.
func generateUserCertificate(host string, pubkey ssh.PublicKey, username string, duration uint64) ssh.Certificate {
	validAfter := uint64(time.Now().Unix())
	validBefore := validAfter + duration

	return ssh.Certificate{
		Key: pubkey,
		// From OpenSSH PROTOCOL.certkeys:
		//   serial is an optional certificate serial number set by the CA to
		//   provide an abbreviated way to refer to certificates from that CA.
		//   If a CA does not wish to number its certificates it must set this
		//   field to zero.
		Serial:   0,
		CertType: ssh.UserCert,
		// From OpenSSH PROTOCOL.certkeys:
		//   key id is a free-form text field that is filled in by the CA at
		//   the time of signing; the intention is that the contents of this
		//   field are used to identify the identity principal in log messages.
		//
		// Set KeyId to "user@host" which can be used by the client to check
		// which host this certificate was issued for.
		KeyId:           PRINCIPAL + "@" + host,
		ValidPrincipals: []string{PRINCIPAL, username},
		// From OpenSSH PROTOCOL.certkeys:
		//   "valid after" and "valid before" specify a validity period for the
		//   certificate. Each represents a time in seconds since 1970-01-01
		//   00:00:00. A certificate is considered valid if:
		//     valid after <= current time < valid before
		ValidAfter:  validAfter - 10, // account for slight clock differences
		ValidBefore: validBefore,
		Permissions: ssh.Permissions{
			CriticalOptions: map[string]string{
				"force-command": FORCE_COMMAND + " " + username,
			},
			Extensions: map[string]string{
				"permit-agent-forwarding": "",
				"permit-pty":              "",
			},
		},
	}
}

func hasPrincipal(validPrincipals []string) bool {
	return slices.Contains(validPrincipals, PRINCIPAL)
}

func hasForceCommand(criticalOptions map[string]string) bool {
	for k, v := range criticalOptions {
		if k == "force-command" && strings.HasPrefix(v, FORCE_COMMAND) {
			return true
		}
	}

	return false
}

func validateUserCertificate(cert ssh.Certificate, caPubkey ssh.PublicKey) bool {
	var certChecker ssh.CertChecker

	certChecker.IsUserAuthority = func(auth ssh.PublicKey) bool {
		return auth == caPubkey
	}

	return certChecker.CheckCert(PRINCIPAL, &cert) != nil &&
		hasForceCommand(cert.CriticalOptions) &&
		hasPrincipal(cert.ValidPrincipals)
}
