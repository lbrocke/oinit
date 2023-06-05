package main

import (
	"fmt"
	"net"
	"oinit/src/dnsutil"
	"oinit/src/liboinitca"
	"oinit/src/sshutil"
	"os"
	"sort"
	"strings"
)

const (
	COMMAND_ADD    = "add"
	COMMAND_DELETE = "delete"
	COMMAND_LIST   = "list"
	COMMAND_MATCH  = "match"

	USAGE = "Usage:\n" +
		"\toinit add    <host>[:port] [ca]\tAdd a host managed by oinit.\n" +
		"\toinit delete <host>[:port]\tDelete a host.\n" +
		"\toinit list\t\t\tList all hosts managed by oinit.\n"
)

func handleCommandAdd(args []string) {
	if len(args) < 1 {
		fmt.Print(USAGE)
		return
	}

	hostport := args[0]

	// Split into host and port, as in some cases special handling is required
	// (such as in the known hosts file).
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		// Assume that hostport string given to net.SplitHostPort() doesn't contain a port
		// https://groups.google.com/g/golang-nuts/c/KA41Tj9Aabg/m/1NUcxQcoUjwJ
		host = strings.TrimSpace(hostport)
		port = "22"

		hostport = net.JoinHostPort(host, port)
	}

	// Check if host was already added before. This also includes the system-wide
	// configuration, therefore hosts that were added by the system admin
	// won't be added to the user config again.
	if found, err := sshutil.IsManagedHost(hostport); err != nil {
		LogError("Could not read hosts file: " + err.Error())
		return
	} else if found {
		LogInfo("This host was already added.")
		return
	}

	// Determine CA from DNS if not given on command line.
	var ca string
	if len(args) >= 2 {
		ca = args[1]
	} else {
		detected, err := dnsutil.LookupCA(host)
		if err != nil {
			LogWarn("The CA for this host could not be determined from DNS.")
			LogWarn("You can manually specify the CA by running:")
			LogWarn("")
			LogWarn("\toinit add " + hostport + " [ca]")
			return
		}

		ca = detected
	}

	// Try to contact CA, which returns the host CA public key to be added
	// to the user's known_hosts file.
	if res, err := liboinitca.NewClient(ca).GetHost(host); err != nil {
		LogError("Could not contact CA, " + err.Error())
		return
	} else {
		if err := sshutil.AddSSHKnownHost(host, port, res.PublicKey); err != nil {
			LogWarn("Could not add public key to your known_hosts file.")

			if newLine, err := sshutil.GenerateKnownHosts(host, port, res.PublicKey); err == nil {
				LogWarn("Please add the following line by yourself:")
				LogWarn("\t" + newLine)
			}
		}
	}

	// Add to users' hosts file.
	if err := sshutil.AddHostUser(hostport, ca); err != nil {
		LogError("Could not add host: " + err.Error())
		return
	} else {
		LogSuccess(hostport + " was added.")
	}

	// Check if 'Match exec ...' block is present, and if not try to add it.
	if added, err := sshutil.AddSSHMatchBlock(); err != nil {
		LogWarn("Could not read or modify your OpenSSH config file.")
		LogWarn("Please verify it contains the following lines:")
		LogWarn("")

		for _, line := range strings.Split(sshutil.GenerateMatchBlock(), "\n") {
			LogWarn("\t" + line)
		}
	} else if added {
		LogInfo("As this is your first time running oinit, your OpenSSH config file has")
		LogInfo("been modified to invoke oinit when ssh'ing to hosts managed by it.")
	}
}

func handleCommandDelete(args []string) {
	if len(args) < 1 {
		fmt.Print(USAGE)
		return
	}

	hostport := args[0]

	found, err := sshutil.DeleteHostUser(hostport)
	if err != nil {
		LogError("Could not delete host: " + err.Error())
		return
	}

	if !found {
		LogError(hostport + " is either not managed by oinit, or configured system-wide.")
	} else {
		LogSuccess(hostport + " was deleted.")
	}
}

func handleCommandList() {
	all, err := sshutil.GetManagedHosts()
	if err != nil {
		LogError("Could not load hosts: " + err.Error())
		return
	}

	hosts := make([]string, 0, len(all))
	for hostport := range all {
		hosts = append(hosts, hostport)
	}
	sort.Strings(hosts)

	LogInfo("The following hosts are managed by oinit:")

	for _, host := range hosts {
		fmt.Println("\t" + host)
	}
}

func handleCommandMatch(args []string) {
	// todo
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Print(USAGE)
		return
	}

	switch args[0] {
	case COMMAND_ADD:
		handleCommandAdd(args[1:])
	case COMMAND_DELETE:
		handleCommandDelete(args[1:])
	case COMMAND_LIST:
		handleCommandList()
	case COMMAND_MATCH:
		handleCommandMatch(args[1:])
	default:
		fmt.Print(USAGE)
	}
}
