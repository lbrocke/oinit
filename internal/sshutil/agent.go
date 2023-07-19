package sshutil

import (
	"net"
	"os"
	"strings"

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

// agentGetOinitCertificates returns a slice of all certificates in the agent
// that have been issued by oinit for the given host.
//
// The KeyId field, which is set to oinit@<host> by oinit-ca, as well as the
// occurrences of "oinit" in the ValidPrincipals field are used to identify
// certificates issued by oinit.
func agentGetOinitCertificates(agent agent.ExtendedAgent, host string) ([]ssh.Certificate, error) {
	var certificates []ssh.Certificate

	keys, err := agent.List()
	if err != nil {
		return certificates, err
	}

	keyId := PRINCIPAL + "@" + strings.ToLower(host)

	for _, key := range keys {
		pk, err := ssh.ParsePublicKey(key.Blob)
		if err != nil {
			// This should never happen
			continue
		}

		cert := pk.(*ssh.Certificate)

		if cert.KeyId == keyId && slices.Contains(cert.ValidPrincipals, PRINCIPAL) {
			certificates = append(certificates, *cert)
		}
	}

	return certificates, nil
}

// AgentHasCertificate returns a bool indicating whether a certificate issued
// by oinit-ca for the given host is currently present in the agent.
// An error is returned when communication with the agent is not possible, for
// example if it isn't running.
func AgentHasCertificate(agent agent.ExtendedAgent, host string) (bool, error) {
	certificates, err := agentGetOinitCertificates(agent, host)

	return len(certificates) != 0, err
}

// AgentRemoveCertificates removes all certificates issued by oinit-ca for the
// given host from the agent.
// An error is returned when communication with the agent is not possible, for
// example if it isn't running.
func AgentRemoveCertificates(agent agent.ExtendedAgent, host string) error {
	certificates, err := agentGetOinitCertificates(agent, host)
	if err != nil {
		return err
	}

	for _, cert := range certificates {
		agent.Remove(&cert)
	}

	return nil
}
