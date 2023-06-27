package main

import (
	"oinit/pkg/passwd"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"oinit/pkg/log"
)

const (
	OINIT_USER    = "oinit"
	DEFAULT_SHELL = "/bin/sh"

	ERR_MISSING_PERMISSION = "Missing permissions."
	ERR_NOT_ALLOWED        = "This is not allowed."
	ERR_DROP_PRIVILEGES    = "Could not switch user."
)

// dropPrivileges uses setuid/setgid to drop all root privileges. Any error
// will immediately exit this program.
func dropPrivileges(uid, gid int) {
	if err := syscall.Setgroups([]int{}); err != nil {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	if err := syscall.Setgid(gid); err != nil || uid == 0 {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	if err := syscall.Setuid(uid); err != nil || gid == 0 {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}
}

// getUser returns the uid and gid for the given username. If the user doesn't
// exist, getUser exists immediately with code 1.
func getUser(name string) (int, int) {
	u, err := user.Lookup(name)
	if err != nil {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	var uid, gid int
	if uid, err = strconv.Atoi(u.Uid); err != nil {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}
	if gid, err = strconv.Atoi(u.Gid); err != nil {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	return uid, gid
}

// Returns a list of all environment variables (from os.Environ()) that start
// with LC_ or lc_
func getLocaleEnvs() []string {
	var envs []string

	for _, env := range os.Environ() {
		name, _, ok := strings.Cut(env, "=")

		if ok && strings.HasPrefix(strings.ToUpper(name), "LC_") {
			envs = append(envs, env)
		}
	}

	return envs
}

// Usage: "oinit-switch <target>" with target begin the name of user to switch
// to.
func main() {
	// root permissions are required in any case.
	// Either the process owner is root or process was started with root
	// permissions (via setuid bit on executable).
	if os.Geteuid() != 0 {
		log.LogFatal(ERR_MISSING_PERMISSION)
	}

	if len(os.Args) != 2 {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	target := os.Args[1]

	targetUid, targetGid := getUser(target)

	// Switching to root is not allowed.
	if targetUid == 0 || targetGid == 0 {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	oinitUid, _ := getUser(OINIT_USER)

	if os.Getuid() == 0 || os.Getuid() == oinitUid || os.Getuid() == targetUid {
		// Only the "oinit" user is allowed to switch to anyone (except to root).
		// Also allow a non-root user to switch to himself/herself.
		dropPrivileges(targetUid, targetGid)
		goto UNPRIVILEGED
	}

	log.LogFatal(ERR_NOT_ALLOWED)

UNPRIVILEGED:

	// Double check b/c paranoia
	if os.Getuid() != targetUid || os.Getgid() != targetGid {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	shell := passwd.Shell(target, DEFAULT_SHELL)

	var cmd *exec.Cmd

	// Check if SSH_ORIGINAL_COMMAND is set, if yes run command instead of
	// spawning an interactive shell.
	if sshCmd, ok := os.LookupEnv("SSH_ORIGINAL_COMMAND"); ok {
		// Use user's shell to execute command so that nice-to-have's like glob
		// expansions work.
		cmd = exec.Command(shell, "-c", sshCmd)
	} else {
		cmd = exec.Command(shell, "-il")
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Override environment variables as setuid/setgid and exec.Command() don't
	// do this automatically. Only certain required or useful variables
	// inspired by login(1) are set or copied from this process.
	// From login(1):
	//   The environment variable values for $HOME, $USER, $SHELL, $PATH,
	//   $LOGNAME, and $MAIL are set according to the appropriate fields in the
	//   password entry. $PATH defaults to [..] for normal users, and to [..]
	//   for root, if not otherwise configured.

	targetUser, _ := user.Lookup(target)

	cmd.Env = append([]string{
		"USER=" + targetUser.Username,
		"LOGNAME=" + targetUser.Username,
		"HOME=" + targetUser.HomeDir,
		"PWD=" + targetUser.HomeDir,
		"SHELL=" + shell,
		"TERM=" + os.Getenv("TERM"),
		// Set by OpenSSH:
		"SSH_CONNECTION=" + os.Getenv("SSH_CONNECTION"),
		"SSH_CLIENT=" + os.Getenv("SSH_CLIENT"),
	}, getLocaleEnvs()...)

	// Required, as setting PWD isn't enough
	os.Chdir(targetUser.HomeDir)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}

		os.Exit(1)
	}
}
