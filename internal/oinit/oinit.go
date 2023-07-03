package oinit

import (
	"bufio"
	"errors"
	"net"
	"os"
	"strings"

	"github.com/lbrocke/oinit/internal/sshutil"
	"github.com/lbrocke/oinit/internal/util"
)

// AddHostUser adds the given host/port and CA to the user's hosts file.
func AddHostUser(hostport, ca string) error {
	hostport = strings.ToLower(hostport)
	ca = strings.ToLower(ca)

	paths, err := sshutil.PathsHosts()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(paths.User, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	if _, err = f.Write([]byte(hostport + " " + ca + "\n")); err != nil {
		f.Close()
		return err
	}

	f.Close()

	return nil
}

// DeleteHostUser deletes a given host/port from the user's hosts file.
// An error is returned in case of file-handling related errors. nil is
// returned when the host/port was successfully deleted or if it wasn't found
// in the user's file.
func DeleteHostUser(hostport string) (bool, error) {
	hostport = strings.ToLower(hostport)

	paths, err := sshutil.PathsHosts()
	if err != nil {
		return false, err
	}

	// Use port 22 if not specified
	_, _, err = net.SplitHostPort(hostport)
	if err != nil {
		hostport = net.JoinHostPort(hostport, "22")
	}

	content, err := os.ReadFile(paths.User)
	if err != nil {
		return false, err
	}

	foundIndex := -1

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		managedHostport, _, found := strings.Cut(line, " ")
		if !found {
			// malformed line, ignore
			continue
		}

		if managedHostport == hostport {
			foundIndex = i
			break
		}
	}

	if foundIndex == -1 {
		return false, nil
	}

	// Remove index i and write back to file
	lines = append(lines[:foundIndex], lines[foundIndex+1:]...)
	if err := os.WriteFile(paths.User, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		return true, err
	}

	return true, nil
}

// IsManagedHost checks whether the given host/port is a managed ssh server,
// meaning that it is present in either the system or user managed hosts file.
func IsManagedHost(hostport string) (bool, error) {
	ca, err := GetCA(hostport)

	return ca != "", err
}

// GetCA returns the CA stored in the user's hosts file for a given host/port.
func GetCA(hostport string) (string, error) {
	hostport = strings.ToLower(hostport)

	managedHosts, err := GetManagedHosts()
	if err != nil {
		return "", err
	}

	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return "", err
	}

	for managedHostport, ca := range managedHosts {
		managedHost, managedPort, err := net.SplitHostPort(managedHostport)
		if err != nil {
			continue
		}

		if util.MatchesHost(host, port, managedHost, managedPort) {
			return ca, nil
		}
	}

	return "", nil
}

// GetManagedHosts returns all managed hosts (keys) and their respective CAs
// (values) as a map.
func GetManagedHosts() (map[string]string, error) {
	paths, err := sshutil.PathsHosts()
	if err != nil {
		return nil, err
	}

	hosts := make(map[string]string)

	for _, path := range []string{paths.User, paths.System} {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			hostport, ca, found := strings.Cut(scanner.Text(), " ")
			if !found || strings.Contains(ca, " ") {
				return nil, errors.New("malformed hosts file")
			}

			if _, exists := hosts[hostport]; exists {
				// do not overwrite existing keys
				continue
			}

			hosts[hostport] = ca
		}
	}

	return hosts, nil
}
