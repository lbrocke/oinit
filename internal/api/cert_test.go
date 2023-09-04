package api

import (
	"crypto/ed25519"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func TestGenerateUserCertificate(t *testing.T) {
	host := "example.com"

	pk, _, _ := ed25519.GenerateKey(nil)
	pubkey, _ := ssh.NewPublicKey(pk)

	username := "testuser"
	duration := uint64(3600)

	certificate := generateUserCertificate(host, pubkey, username, duration)

	if certificate.Serial != 0 {
		t.Error("Expected Serial to be 0")
	}

	if certificate.CertType != ssh.UserCert {
		t.Error("Expected CertType to be ssh.UserCert")
	}

	expectedKeyId := PRINCIPAL + "@" + host
	if certificate.KeyId != expectedKeyId {
		t.Errorf("Expected KeyId to be %s, but got %s", expectedKeyId, certificate.KeyId)
	}

	expectedValidPrincipals := []string{PRINCIPAL, username}
	if !stringSlicesEqual(certificate.ValidPrincipals, expectedValidPrincipals) {
		t.Errorf("Expected ValidPrincipals to be %v, but got %v", expectedValidPrincipals, certificate.ValidPrincipals)
	}

	currentTime := uint64(time.Now().Unix())
	if !(certificate.ValidAfter <= currentTime && currentTime < certificate.ValidBefore) {
		t.Error("Invalid certificate validity period")
	}

	expectedForceCommand := FORCE_COMMAND + " " + username
	if certificate.Permissions.CriticalOptions["force-command"] != expectedForceCommand {
		t.Errorf("Expected force-command to be %s, but got %s", expectedForceCommand, certificate.Permissions.CriticalOptions["force-command"])
	}

	if _, ok := certificate.Permissions.Extensions["permit-agent-forwarding"]; !ok {
		t.Error("Expected permit-agent-forwarding extension to be present")
	}

	if _, ok := certificate.Permissions.Extensions["permit-pty"]; !ok {
		t.Error("Expected permit-pty extension to be present")
	}
}

func stringSlicesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}
	return true
}
