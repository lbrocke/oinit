package main

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lbrocke/oinit/internal/dnsutil"
	"github.com/lbrocke/oinit/internal/liboinitca"
	"github.com/lbrocke/oinit/internal/oidc"
	"github.com/lbrocke/oinit/internal/oinit"
	"github.com/lbrocke/oinit/internal/sshutil"
	"github.com/lbrocke/oinit/internal/util"
	"github.com/lbrocke/oinit/pkg/log"

	"github.com/mattn/go-tty"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/exp/slices"
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

// handleCommandAdd handles the 'add' command to add a host managed by oinit.
// It takes the host and optional CA as arguments.
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
	if found, err := oinit.IsManagedHost(hostport); err != nil {
		log.LogError("Could not read hosts file: " + err.Error())
		return
	} else if found {
		log.LogInfo("This host was already added.")
		return
	}

	// Determine CA from DNS if not given on command line.
	var ca string
	if len(args) >= 2 {
		ca = args[1]
	} else {
		detected, err := dnsutil.LookupCA(host)
		if err != nil {
			log.LogWarn("The CA for this host could not be determined from DNS.")
			log.LogWarn("You can manually specify the CA by running:")
			log.LogWarn("")
			log.LogWarn("\toinit add " + args[0] + " [ca]")
			return
		}

		ca = detected
		log.LogInfo("Determined CA from DNS: " + ca)
	}

	// Try to contact CA, which returns the host CA public key to be added
	// to the user's known_hosts file.
	if res, err := liboinitca.NewClient(ca).GetHost(host); err != nil {
		log.LogError("Could not contact CA: " + err.Error())
		return
	} else {
		if err := sshutil.AddSSHKnownHost(host, port, res.PublicKey); err != nil {
			log.LogWarn("Could not add public key to your known_hosts file.")

			if newLine, err := sshutil.GenerateKnownHosts(host, port, res.PublicKey); err == nil {
				log.LogWarn("Please add the following line by yourself:")
				log.LogWarn("\t" + newLine)
			}
		}
	}

	// Add to users' hosts file.
	if err := oinit.AddHostUser(hostport, ca); err != nil {
		log.LogError("Could not add host: " + err.Error())
		return
	} else {
		log.LogSuccess(hostport + " was added.")
	}

	// Check if 'Match exec ...' block is present, and if not try to add it.
	if added, err := sshutil.AddSSHMatchBlock(); err != nil {
		log.LogWarn("Could not read or modify your OpenSSH config file.")
		log.LogWarn("Please verify it contains the following lines:")
		log.LogWarn("")

		for _, line := range strings.Split(sshutil.GenerateMatchBlock(), "\n") {
			log.LogWarn("\t" + line)
		}
	} else if added {
		log.LogInfo("As this is your first time running oinit, your OpenSSH config file has")
		log.LogInfo("been modified to invoke oinit when connecting to hosts managed by it.")
	}
}

// handleCommandDelete handles the 'delete' command to delete a host.
// It takes the host as an argument.
func handleCommandDelete(args []string) {
	if len(args) < 1 {
		fmt.Print(USAGE)
		return
	}

	hostport := args[0]

	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		host = strings.TrimSpace(hostport)
	}

	found, err := oinit.DeleteHostUser(hostport)
	if err != nil {
		log.LogFatal("Could not delete host: " + err.Error())
	}
	if !found {
		log.LogFatal(hostport + " is either not managed by oinit, or configured system-wide.")
	}

	if sshutil.AgentIsRunning() {
		sshAgent, _ := sshutil.GetAgent()

		sshutil.AgentRemoveCertificates(sshAgent, host)
	}

	log.LogSuccess(hostport + " was deleted.")
}

// handleCommandList handles the 'list' command to list all hosts managed by oinit.
func handleCommandList() {
	all, err := oinit.GetManagedHosts()
	if err != nil {
		log.LogError("Could not load hosts: " + err.Error())
		return
	}

	hosts := make([]string, 0, len(all))
	for hostport := range all {
		hosts = append(hosts, hostport)
	}
	sort.Strings(hosts)

	log.LogInfo("The following hosts are managed by oinit:")

	for _, host := range hosts {
		fmt.Println("\t" + host)
	}
}

// getTokenFromOidcAgent prompts the user to select a supported OIDC issuer
// and then requests an access token via oidc-agent. It takes the CA client
// and host as arguments and returns the access token.
func getTokenFromOidcAgent(caClient liboinitca.Client, host string) string {
	if !oidc.AgentIsRunning() {
		log.LogFatalTTY("oidc-agent is not running, please start it first.")
	}

	hostRes, err := caClient.GetHost(host)
	if err != nil {
		log.LogFatalTTY("Contacting the CA failed: " + err.Error())
	}

	// Put provider URLs into slice to be able to sort them
	providers := make([]string, len(hostRes.Providers))
	for i, info := range hostRes.Providers {
		providers[i] = info.URL
	}
	sort.Strings(providers)

	provider, err := promptProviders(providers)
	if err != nil {
		log.LogFatalTTY(err.Error())
	}

	// Get scopes for selected provider
	var scopes []string
	for _, info := range hostRes.Providers {
		if info.URL != provider {
			continue
		}

		scopes = info.Scopes
	}

	token, err := oidc.GetToken(provider, scopes)
	if err != nil {
		log.LogFatalTTY("Could not get token from oidc-agent: " + err.Error())
	}
	if token == "" {
		log.LogFatalTTY("Received an empty token from oidc-agent.")
	}

	return token
}

// promptProviders prompts the user to select an OIDC provider from the list
// of available providers. It takes a list of provider URLs as arguments and
// returns the selected provider URL.
func promptProviders(providers []string) (string, error) {
	if len(providers) == 0 {
		//lint:ignore ST1005 Error is display to user directly
		return "", errors.New("The server indicated that no OIDC provider is supported")
	}

	accs := oidc.GetConfiguredAccounts()

	// Check if user pre-selected an account
	if account := os.Getenv("OIDC_AGENT_ACCOUNT"); account != "" {
		for issuer, accounts := range accs {
			if slices.Contains(accounts, account) {
				return issuer, nil
			}
		}
	}

	// Check if user pre-selected an issuer
	if issuer := util.Getenvs("OIDC_ISS", "OIDC_ISSUER"); issuer != "" {
		if _, ok := accs[issuer]; ok {
			return issuer, nil
		}
	}

	for i, issuer := range providers {
		str := issuer

		if accounts, ok := accs[issuer]; ok && len(accounts) > 0 {
			str += " (Accounts: " + strings.Join(accs[issuer], ", ") + ")"
		}

		log.LogTTY(fmt.Sprintf("[%d] %s", i+1, str))
	}

	tty, err := tty.Open()
	if err != nil {
		return "", errors.New("There was an error opening your TTY: " + err.Error())
	}

	log.PromptTTY(fmt.Sprintf("Please select a provider to use [1-%d]: ", len(providers)))

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

// generateEd25519Keys generates a new ED25519 key pair and returns the
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

// handleCommandMatch handles the 'match' command to match a host managed by oinit.
// It takes the host and port as arguments.
func handleCommandMatch(args []string) {
	if len(args) != 2 {
		os.Exit(1)
	}

	host := strings.ToLower(args[0])
	port := args[1]
	hostport := strings.ToLower(net.JoinHostPort(host, port))

	if is, err := oinit.IsManagedHost(hostport); err != nil || !is {
		// Return non-zero exit code to indicate that host/port do not match
		os.Exit(1)
	}

	ca, err := oinit.GetCA(hostport)
	if err != nil {
		log.LogFatalTTY("The CA managing '" + host + "' could not be determined.\n" +
			"Did you run 'oinit add " + hostport + "' yet?")
	}

	caClient := liboinitca.NewClient(ca)

	// Verify that ssh-agent is running, which is required in any case
	if !sshutil.AgentIsRunning() {
		log.LogFatalTTY("ssh-agent is not running, please start it first.")
	}

	sshAgent, _ := sshutil.GetAgent()

	if exists, err := sshutil.AgentHasCertificate(sshAgent, host); err == nil && exists {
		// Agent already holds certificate, therefore do not request a new one
		return
	}

	// Try to get token from environment variable
	token := util.Getenvs("ACCESS_TOKEN", "OIDC", "OS_ACCESS_TOKEN",
		"OIDC_ACCESS_TOKEN", "WATTS_TOKEN", "WATTSON_TOKEN")

	if token == "" {
		// Use oidc-agent to get token.
		// getTokenFromOidcAgent() exits with -1 for any errors.
		token = getTokenFromOidcAgent(caClient, host)
	}

	pubkey, privkey, err := generateEd25519Keys()
	if err != nil {
		log.LogFatalTTY("There was an error generating a temporary key pair.")
	}

	res, err := caClient.PostHostCertificate(host, pubkey, token)
	if err != nil {
		log.LogFatalTTY("CA responded: " + err.Error())
	}

	certPk, _, _, _, err := ssh.ParseAuthorizedKey([]byte(res.Certificate))
	if err != nil {
		log.LogFatalTTY("Cannot parse certificate.")
	}

	cert := certPk.(*ssh.Certificate)
	validUntil := time.Unix(int64(cert.ValidBefore-1), 0)

	if sshAgent.Add(agent.AddedKey{
		PrivateKey:   privkey,
		Certificate:  cert,
		LifetimeSecs: uint32(time.Until(validUntil).Seconds()),
	}) != nil {
		log.LogFatalTTY("Cannot add private key and certificate to ssh-agent.")
	} else {
		log.LogSuccessTTY(fmt.Sprintf("Received a certificate which is valid until %s", validUntil))
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
