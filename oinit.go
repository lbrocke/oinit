package main

import (
	"fmt"
	"os"
)

const (
	COMMAND_ADD    = "add"
	COMMAND_DELETE = "delete"
	COMMAND_LIST   = "list"
	COMMAND_MATCH  = "match"

	USAGE = "Usage:\n" +
		"\toinit add    <host>[:port] [ca]\tAdd a host managed by oinit.\n" +
		"\toinit delete <host> [port]\tDelete a host.\n" +
		"\toinit match  <host> [port]\tCheck whether host is managed by oinit. Used in your ssh config.\n"
)

func handleCommandAdd(args []string) {
	// todo
}

func handleCommandDelete(args []string) {
	// todo
}

func handleCommandList() {
	// todo
}

func handleCommandMatch(args []string) {
	// todo

	os.Exit(1) // no match
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
