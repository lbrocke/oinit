package oidc

import (
	"net"
	"os"
)

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
