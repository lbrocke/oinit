package main

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/lbrocke/oinit/pkg/log"
	"github.com/lbrocke/oinit/pkg/passwd"
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
// exist, getUser exits immediately with code 1.
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

// getLocaleEnvs returns a list of all environment variables (from os.Environ())
// that start with LC_ or lc_.
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

// This program must be invoked using "oinit-switch <target>".
// It will, as the user <target>, spawn an interactive login shell or execute
// the command given in the SSH_ORIGINAL_COMMAND environment variable, if set.
//
// For this, root permissions are required. Therefore this program must be
// either executed as root or (this is the intended way) by having the setuid
// bit set on the executable file.
//
// Only the user "oinit" as well as root are allowed to switch to all other
// users (except to root).
// Normal users are only allowed to switch to themselves.
func main() {
	// Verify that this program is executed with root permissions.
	if os.Geteuid() != 0 {
		log.LogFatal(ERR_MISSING_PERMISSION)
	}

	if len(os.Args) != 2 {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	target := os.Args[1]
	targetUid, targetGid := getUser(target)

	// Switching to root is not allowed for anybody.
	if targetUid == 0 || targetGid == 0 {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	oinitUid, _ := getUser(OINIT_USER)

	if os.Getuid() == 0 || os.Getuid() == oinitUid || os.Getuid() == targetUid {
		// Only the user "oinit" as well as root are allowed to switch to all
		// other users (except to root).
		// Normal users are only allowed to switch to themselves.
		dropPrivileges(targetUid, targetGid)
		goto UNPRIVILEGED
	}

	// Exit if no jump to UNPRIVILEGED occurred.
	log.LogFatal(ERR_NOT_ALLOWED)

UNPRIVILEGED:

	// Double check b/c paranoia
	if os.Getuid() != targetUid || os.Getgid() != targetGid {
		log.LogFatal(ERR_DROP_PRIVILEGES)
	}

	// The raw value from /etc/passwd, e.g. '/bin/bash'
	shell := passwd.Shell(target, DEFAULT_SHELL)

	// syscall.Exec() requires full path
	argv0, err := exec.LookPath(shell)
	if err != nil {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	// Shell name without path, e.g. 'bash'
	shellName := filepath.Base(argv0)

	var argv []string

	// Check if SSH_ORIGINAL_COMMAND is set, if yes run command instead of
	// spawning an interactive shell.
	if sshCmd, ok := os.LookupEnv("SSH_ORIGINAL_COMMAND"); ok {
		// Use user's shell to execute command so that nice-to-have's like glob
		// expansions and built-in functions work.
		argv = []string{shellName, "-c", sshCmd}
	} else {
		// Invoke login shell by prepending "-"
		argv = []string{"-" + shellName}
	}

	targetUser, _ := user.Lookup(target)

	// Only set certain required or useful environment variables similar to
	// login(1), in addition to locale (LC_*) and two OpenSSH-defined variables:
	//   The environment variable values for $HOME, $USER, $SHELL, $PATH,
	//   $LOGNAME, and $MAIL are set according to the appropriate fields in the
	//   password entry. $PATH defaults to [..] for normal users, and to [..]
	//   for root, if not otherwise configured.
	envv := append([]string{
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

	if err := syscall.Exec(argv0, argv, envv); err != nil {
		os.Exit(1)
	}
}
