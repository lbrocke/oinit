//go:build !windows
// +build !windows

package sshutil

import "path/filepath"

// Returns the file path for system-wide OpenSSH configuration.
// On *nix, this is '/etc'
func getSystemSSHPath() string {
	return string(filepath.Separator) + "etc"
}
