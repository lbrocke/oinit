package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/lbrocke/oinit/pkg/log"
)

func main() {
	// This programm will be invoked by OpenSSH as
	//   oinit-shell -c 'oinit-switch <name>'
	// Make sure that no other command can be run and that no interactive shell
	// is provided.

	if len(os.Args) != 3 || os.Args[1] != "-c" ||
		!strings.HasPrefix(os.Args[2], "oinit-switch ") {
		// Make sure that shell was invoked using "oinit-shell -c 'oinit-switch ...'",
		// otherwise exit which will terminate the SSH connection.
		log.LogFatal("This user does not provide interactive access.")
	}

	// Make sure 'oinit-switch' exists, then run it and return exit code.
	if _, err := exec.LookPath("oinit-switch"); err != nil {
		os.Exit(1)
	}

	// os.Args[2] contains "oinit-switch ...", split it at whitespace
	// and pass it into exec.Command()
	args := strings.Fields(os.Args[2])

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}

		os.Exit(1)
	}
}
