package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/lbrocke/oinit/pkg/log"
)

const (
	FORCE_COMMAND = "oinit-switch"

	ERR_PROHIBITED = "This user does not provide interactive access."
	ERR_INTERNAL   = "An error occurred."
)

// This program will be invoked by OpenSSH as
//
//	oinit-shell -c 'oinit-switch <target>'
//
// Ensure that only FORCE_COMMAND can be run and no interactive login shell is
// provided.
func main() {
	if len(os.Args) != 3 || os.Args[1] != "-c" {
		log.LogFatal(ERR_PROHIBITED)
	}

	command := os.Args[2]
	argv := strings.Fields(command)

	if !strings.HasPrefix(command, FORCE_COMMAND) || len(argv) != 2 {
		log.LogFatal(ERR_PROHIBITED)
	}

	// syscall.Exec() requires full path
	path, err := exec.LookPath(argv[0])
	if err != nil {
		log.LogFatal(ERR_INTERNAL)
	}

	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		log.LogFatal(ERR_INTERNAL)
	}
}
