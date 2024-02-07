package main

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/lbrocke/oinit/pkg/log"
	"github.com/mattn/go-isatty"
)

const (
	SU_COMMAND  = "su"
	OINIT_USER  = "oinit"
	SYS_UID_MAX = 999

	ERR_NOT_ALLOWED = "This is not allowed."
	ERR_INTERNAL    = "Internal error. oinit might not be set up correctly."
)

// getUser returns the uid for the given username. If the user doesn't exist,
// an error is returned.
func getUid(name string) (int, error) {
	user, err := user.Lookup(name)
	if err != nil {
		return -1, err
	}

	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return -1, err
	}

	return uid, nil
}

func main() {
	if len(os.Args) != 2 {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	target := os.Args[1]

	// Make sure target user is not a system user. This is not strictly
	// necessary because (a) oinit-ca would never issue a certificate
	// containing a force-command to switch to a system user and (b) all proper
	// system users (except root) have their shell so to /bin/nologin (or
	// similar), however this check doesn't hurt and increases security.
	targetUid, err := getUid(target)
	if err != nil || targetUid < SYS_UID_MAX {
		log.LogFatal(ERR_NOT_ALLOWED)
	}

	curUser, err := user.Current()
	if err != nil {
		log.LogFatal(ERR_INTERNAL)
	}
	curUid, err := strconv.Atoi(curUser.Uid)
	if err != nil {
		log.LogFatal(ERR_INTERNAL)
	}

	// Allow user to switch to themself. This is necessary because
	// issued certificates also contain the target username as principal,
	// allowing the user to connect as themself directly without using
	// the oinit user.
	// Note that it requires the user to have set a password (which isn't set
	// by motley_cue by default) and the user still has to enter his/her own
	// password, because 'su' requires this.
	if targetUid != curUid {
		// In all other cases, make sure the program is executed by the oinit
		// user. This is not strictly necessary, because the 'su' command would
		// just prompt for a password in case the user executing this program
		// isn't oinit.

		oinitUid, err := getUid(OINIT_USER)
		if err != nil {
			log.LogFatal(ERR_INTERNAL)
		}

		if curUid != oinitUid {
			log.LogFatal(ERR_NOT_ALLOWED)
		}
	}

	// syscall.Exec() requires full path
	argv0, err := exec.LookPath(SU_COMMAND)
	if err != nil {
		log.LogFatal(ERR_INTERNAL)
	}

	var argv []string
	if sshCmd, ok := os.LookupEnv("SSH_ORIGINAL_COMMAND"); ok {
		// In case a command was given to ssh, execute this command instead of
		// starting an interactive shell session.

		// By default, ssh does not request a tty when a command is given.
		// Using the '-P' option for 'su' (which is recommended to prevent
		// TIOCSTI ioctl terminal injection) however would create a pseudo tty
		// anyways. This results in problems for other programs using ssh (like
		// git and rsync), as they expect no tty to be created. Therefore,
		// do not use the '-P' option here.
		// To prevent users abusing this security hole, for example by forcing
		// tty allocation anyway ("ssh -tt example.org /bin/bash"), make sure
		// this program does not run ssh commands when a tty is present.

		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			log.LogFatal(ERR_NOT_ALLOWED)
		}

		argv = []string{SU_COMMAND, "-", target, "-c", sshCmd}
	} else {
		argv = []string{SU_COMMAND, "-", target, "-P"}
	}

	// Use syscall.Exec (which calls execve) instead of exec.Command (which does fork + evecve)
	// to prevent unnecessary resource hogging and hide this script in htop
	if err := syscall.Exec(argv0, argv, os.Environ()); err != nil {
		os.Exit(1)
	}
}
