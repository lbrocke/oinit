//go:build windows
// +build windows

package sshutil

import "os"

// Returns the file path for system-wide OpenSSH configuration.
// On Windows, this is the content of $PROGRAMDATA
func getSystemSSHPath() string {
	return os.Getenv("PROGRAMDATA")
}
