// Package sshutil handles all kinds to things related to ssh files like
// system/user configuration files and known_hosts.

package sshutil

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	SSH_CONFIG_USER        = "config"
	SSH_CONFIG_SYSTEM      = "ssh_config"
	SSH_KNOWN_HOSTS_USER   = "known_hosts"
	SSH_KNOWN_HOSTS_SYSTEM = "ssh_known_hosts"
	HOSTS_USER             = "oinit_hosts"
	HOSTS_SYSTEM           = "ssh_oinit_hosts"
)

type FilePaths struct {
	User   string
	System string
}

// findPaths returns a FilePaths with the given file names for user and system
// level OpenSSH files, such as 'config' / 'ssh_config' and 'known_hosts' /
// 'ssh_known_hosts'.
//
// The base paths for user and system are determined based on GOOS.
func findPaths(userFile string, systemFile string) (FilePaths, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return FilePaths{}, err
	}

	var system = filepath.Join(string(filepath.Separator), "etc", "ssh", systemFile)
	if runtime.GOOS == "windows" {
		system = filepath.Join(os.Getenv("PROGRAMDATA"), "ssh", systemFile)
	}

	return FilePaths{
		System: system,
		User:   filepath.Join(userHome, ".ssh", userFile),
	}, nil
}

// pathsConfig returns the user and system config file path for OpenSSH.
//
// On Unix or macOS, it returns:
//
//	user:   $HOME/.ssh/config
//	system: /etc/ssh/ssh_config
//
// On Windows, it returns:
//
//	user:   %userprofile%/.ssh/config
//	system: %programdata%/ssh/ssh_config
func pathsSSHConfig() (FilePaths, error) {
	return findPaths(SSH_CONFIG_USER, SSH_CONFIG_SYSTEM)
}

// pathsKnownHosts returns the user and system known_hosts file path for OpenSSH.
//
// On Unix or macOS, it returns:
//
//	user:   $HOME/.ssh/known_hosts
//	system: /etc/ssh/ssh_known_hosts
//
// On Windows, it returns:
//
//	user:   %userprofile%/.ssh/known_hosts
//	system: %programdata%/ssh/ssh_known_hosts
func pathsSSHKnownHosts() (FilePaths, error) {
	return findPaths(SSH_KNOWN_HOSTS_USER, SSH_KNOWN_HOSTS_SYSTEM)
}

// pathsHosts returns the user and system managed hosts file path.
//
// On Unix or macOS, it returns:
//
//	user:   $HOME/.ssh/oinit_hosts
//	system: /etc/ssh/oinit_hosts
//
// On Windows, it returns:
//
//	user:   %userprofile%/.ssh/oinit_hosts
//	system: %programdata%/ssh/oinit_hosts
func pathsHosts() (FilePaths, error) {
	return findPaths(HOSTS_USER, HOSTS_SYSTEM)
}
