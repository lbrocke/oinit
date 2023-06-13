package main

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net"
	"oinit/src/dnsutil"
	"oinit/src/liboinitca"
	"oinit/src/oidc"
	"oinit/src/sshutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-tty"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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
			LogWarn("\toinit add " + args[0] + " [ca]")
			return
		}

		ca = detected
		LogInfo("Determined CA from DNS: " + ca)
	}

	// Try to contact CA, which returns the host CA public key to be added
	// to the user's known_hosts file.
	if res, err := liboinitca.NewClient(ca).GetHost(host); err != nil {
		LogError("Could not contact CA: " + err.Error())
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
		LogInfo("been modified to invoke oinit when connecting to hosts managed by it.")
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

func promptProviders(providers []string) (string, error) {
	if len(providers) == 0 {
		//lint:ignore ST1005 Error is display to user directly
		return "", errors.New("The server indicated that no OIDC provider is supported")
	}

	// todo: Compare list of supported providers with already configured
	//       or loaded accounts. Unfortunately, liboidcagent's
	//       GetLoadedAccounts and GetConfiguredAccounts only return the
	//       short names and no issuer URLs.

	for i, e := range providers {
		LogTTY(fmt.Sprintf("[%d] %s", i+1, e))
	}

	tty, err := tty.Open()
	if err != nil {
		return "", errors.New("There was an error opening your TTY: " + err.Error())
	}

	PromptTTY(fmt.Sprintf("Please select a provider to use [1-%d]: ", len(providers)))

	sel, err := tty.ReadString()
	tty.Close()

	if err != nil {
		return "", errors.New("There was an error reading from your TTY: " + err.Error())
	}

	selected, err := strconv.Atoi(sel)
	if err != nil || selected < 1 || selected > len(providers) {
		//lint:ignore ST1005 Error is display to user directly
		return "", errors.New("Your selection is invalid.")
	}

	return providers[selected-1], nil
}

// generateEd25519Keys generates a new ED25519 keypair and returns the
// marshalled public key (ssh-ed25519 AAA...) as well as private key.
func generateEd25519Keys() (string, ed25519.PrivateKey, error) {
	pubkey, privkey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", nil, err
	}

	pubkeyInst, err := ssh.NewPublicKey(pubkey)
	if err != nil {
		return "", nil, err
	}

	return strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pubkeyInst)), "\n"), privkey, nil
}

func handleCommandMatch(args []string) {
	if len(args) != 2 {
		os.Exit(1)
	}

	host := args[0]
	port := args[1]
	hostport := net.JoinHostPort(host, port)

	if is, err := sshutil.IsManagedHost(hostport); err != nil || !is {
		// Return non-zero exit code to indicate that host/port do not match
		os.Exit(1)
	}

	// Make sure both ssh-agent and oidc-agent are running
	sshAgentRunning, oidcAgentRunning := sshutil.AgentIsRunning(), oidc.AgentIsRunning()
	if !sshAgentRunning || !oidcAgentRunning {
		for agent, running := range map[string]bool{
			"ssh-agent":  sshAgentRunning,
			"oidc-agent": oidcAgentRunning,
		} {
			if running {
				LogSuccessTTY(agent + " is running.")
			} else {
				LogErrorTTY(agent + " is not running, please start it first.")
			}
		}

		os.Exit(1)
	}

	sshAgent, _ := sshutil.GetAgent()
	if exists, err := sshutil.AgentHasCertificate(sshAgent); err == nil && exists {
		// Agent already holds certificate, therefore do not request a new one
		return
	}

	ca, err := sshutil.GetCA(hostport)
	if err != nil {
		LogErrorTTY("The CA managing '" + host + "' could not be determined.")
		LogErrorTTY("Did you run 'oinit add " + hostport + "' yet?")
		os.Exit(1)
	}

	caClient := liboinitca.NewClient(ca)

	hostRes, err := caClient.GetHost(host)
	if err != nil {
		LogErrorTTY("Contacting the CA failed: " + err.Error())
		os.Exit(1)
	}

	provider, err := promptProviders(hostRes.Providers)
	if err != nil {
		LogErrorTTY(err.Error())
		os.Exit(1)
	}

	token, err := oidc.GetToken(provider)
	if err != nil || token == "" {
		LogErrorTTY("Could not get token: " + err.Error())
		os.Exit(1)
	}

	pubkey, privkey, err := generateEd25519Keys()
	if err != nil {
		LogErrorTTY("There was an error generating a temporary key pair.")
		os.Exit(1)
	}

	res, err := caClient.PostHostCertificate(host, pubkey, token)
	if err != nil {
		LogErrorTTY("CA responded: " + err.Error())
		os.Exit(1)
	}

	certPk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(res.Certificate))
	if err != nil {
		LogErrorTTY("Cannot parse certificate.")
		os.Exit(1)
	}

	cert := certPk.(*ssh.Certificate)
	validUntil := time.Unix(int64(cert.ValidBefore-1), 0)

	if sshAgent.Add(agent.AddedKey{
		PrivateKey:   privkey,
		Certificate:  cert,
		LifetimeSecs: uint32(time.Until(validUntil).Seconds()),
	}) != nil {
		LogErrorTTY("Cannot add private key and certificate to ssh-agent.")
		os.Exit(1)
	} else {
		LogSuccessTTY(fmt.Sprintf("Received a certificate which is valid until %s", validUntil))
	}
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
