package sshutil

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/exp/slices"
)

type socket struct {
	AddressEnvVar string
	Type          string
}

const (
	PRINCIPAL = "oinit"
)

func AgentIsRunning() bool {
	sockets := []socket{
		{
			"SSH_AUTH_SOCK",
			"unix",
		},
	}

	for _, sock := range sockets {
		if _, err := net.Dial(sock.Type, os.Getenv(sock.AddressEnvVar)); err == nil {
			return true
		}
	}

	return false
}

func GetAgent() (agent.ExtendedAgent, error) {
	sshAgentSock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	return agent.NewClient(sshAgentSock), err
}

func AgentHasCertificate(agent agent.ExtendedAgent, host string) (bool, error) {
	keys, err := agent.List()
	if err != nil {
		return false, err
	}

	// Certificates issued by oinit-ca will have the KeyId field set to "user@host"
	findKeyId := PRINCIPAL + "@" + host

	for _, key := range keys {
		pk, err := ssh.ParsePublicKey(key.Blob)
		if err != nil {
			// This case should never happen
			return false, err
		}

		cert := pk.(*ssh.Certificate)

		// A certificate listing the correct key id and principal is used
		// as criteria for a matching certificate.
		if cert.KeyId == findKeyId &&
			slices.Contains(cert.ValidPrincipals, PRINCIPAL) {
			return true, nil
		}
	}

	return false, nil
}
