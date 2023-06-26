package main

import (
	"fmt"
	"oinit/pkg/libmotleycue"
	"oinit/pkg/passwd"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"oinit/pkg/log"
)

const (
	OINIT_USER    = "oinit"
	DEFAULT_SHELL = "/bin/sh"

	// These messages should never be seen by normal users, when the system is
	// set up correctly. Curious users may however execute this program manually.
	ERR_WRONG_USER   = "Program must be executed by oinit user."
	ERR_NO_ROOT_PERM = "Program must be executed with root permissions."
	ERR_INVALID_ARGS = "Invalid arguments."

	// These messages could be displayed to the user.
	ERR_DEPLOY_FAILED = "Deploying your user account failed, the server responded: %s"
	ERR_SUSPENDED     = "Your user account is not deployed or suspended."
	ERR_DROP_PRIVS    = "Could not switch to your user account."
)

// assertUser makes sure that this program is executed by OINIT_USER but with
// root permissions. Otherwise it fails with exit code 1.
func assertUser() {
	// Program must be executed with root permissions
	if os.Geteuid() != 0 {
		log.LogFatal(ERR_NO_ROOT_PERM)
	}

	user, err := user.Lookup(OINIT_USER)
	if err != nil {
		log.LogFatal(ERR_WRONG_USER)
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		log.LogFatal(ERR_WRONG_USER)
	}

	// Program must be executed by OINIT_USER (or root)
	if os.Getuid() != uid || os.Getuid() != 0 {
		log.LogFatal(ERR_WRONG_USER)
	}
}

// dropPrivileges uses setuid/setgid to drop all root privileges. Any error
// will immediately exit this program.
// Returns the user that was switched to.
func dropPrivileges(username string) *user.User {
	su, err := user.Lookup(username)
	if err != nil {
		log.LogFatal(ERR_DROP_PRIVS)
	}

	var uid, gid int
	if uid, err = strconv.Atoi(su.Uid); err != nil {
		log.LogFatal(ERR_DROP_PRIVS)
	}
	if gid, err = strconv.Atoi(su.Gid); err != nil {
		log.LogFatal(ERR_DROP_PRIVS)
	}

	if err := syscall.Setgroups([]int{}); err != nil {
		log.LogFatal(ERR_DROP_PRIVS)
	}

	if err := syscall.Setgid(gid); err != nil || uid == 0 {
		log.LogFatal(ERR_DROP_PRIVS)
	}

	if err := syscall.Setuid(uid); err != nil || gid == 0 {
		log.LogFatal(ERR_DROP_PRIVS)
	}

	os.Chdir(su.HomeDir)

	return su
}

// Usage: "oinit-switch <token> <url>" with
//   - token: JWT access token
//   - url:   HTTPS URL of motley_cue
func main() {
	// Make sure that this program is executed by correct user (uid) but with
	// root permissions (euid) due to sticky bit being set
	assertUser()

	if len(os.Args) != 3 {
		log.LogFatal(ERR_INVALID_ARGS)
	}

	// The token is not verified any further. The signer key is unknown anyway,
	// therefore decoding without verification would only give an information
	// about whether the given string is a valid JWT. An expiry check is done
	// by motley_cue anyway.
	// Also, parsing the JWT (which runs unknown code) is dangerous when this
	// program is executed with root permissions.
	token := os.Args[1]
	url := os.Args[2]

	res, err := libmotleycue.NewClient(url).GetUserDeploy(token)
	if err != nil {
		log.LogFatal(fmt.Sprintf(ERR_DEPLOY_FAILED, err))
	} else if res.State != libmotleycue.StateDeployed {
		log.LogFatal(ERR_SUSPENDED)
	}

	// Drop root privileges using setgid/setgid system calls.
	user := dropPrivileges(res.Credentials.SSHUser)
	shell := passwd.Shell(res.Credentials.SSHUser, DEFAULT_SHELL)

	sshCmd := os.Getenv("SSH_ORIGINAL_COMMAND")
	hasSshCmd := sshCmd != ""

	var cmd *exec.Cmd

	// Check if SSH_ORIGINAL_COMMAND is set, if yes run command instead of
	// spawning an interactive shell.
	if hasSshCmd {
		// Use user's shell to execute command so that nice-to-have's like glob
		// expansions work.
		cmd = exec.Command(shell, "-c", sshCmd)
	} else {
		cmd = exec.Command(shell, "-il")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Only set stdin if interactive login shell
	if !hasSshCmd {
		cmd.Stdin = os.Stdin
	}

	// Override environment variables as setuid/setgid and exec.Command() don't
	// do this automatically. Only certain required or useful variables
	// inspired by login(1) are set or copied from this process.
	// From login(1):
	//   The environment variable values for $HOME, $USER, $SHELL, $PATH,
	//   $LOGNAME, and $MAIL are set according to the appropriate fields in the
	//   password entry. $PATH defaults to [..] for normal users, and to [..]
	//   for root, if not otherwise configured.

	cmd.Env = []string{
		"USER=" + user.Username,
		"LOGNAME=" + user.Username,
		"HOME=" + user.HomeDir,
		"PWD=" + user.HomeDir,
		"SHELL=" + shell,
		"TERM=" + os.Getenv("TERM"),
		// Set by OpenSSH:
		"SSH_CONNECTION=" + os.Getenv("SSH_CONNECTION"),
		"SSH_CLIENT=" + os.Getenv("SSH_CLIENT"),
	}

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}

		os.Exit(1)
	}
}
