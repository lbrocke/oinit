package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-tty"
)

type Color string

const (
	SYMBOL_ERROR   = "✘"
	SYMBOL_SUCCESS = "✔"
	SYMBOL_WARN    = "!"
	SYMBOL_INFO    = "i"
	SYMBOL_NONE    = ""

	COLOR_RED    Color = "\u001b[31m"
	COLOR_GREEN  Color = "\u001b[32m"
	COLOR_YELLOW Color = "\u001b[33m"
	COLOR_BLUE   Color = "\u001b[34m"
	COLOR_RESET  Color = "\u001b[0m"
)

func log(msg string, color Color, symbol string, useTTY bool, lb bool) {
	// See https://no-color.org
	if os.Getenv("NO_COLOR") != "" {
		color = ""
	}

	var prefix string
	if symbol != "" {
		prefix = fmt.Sprintf("%s%s%s ", color, symbol, COLOR_RESET)
	}

	str := prefix + msg
	if lb {
		str = str + "\n"
	}

	if useTTY {
		if tty, err := tty.Open(); err == nil {
			defer tty.Close()
			tty.Output().WriteString(str)
		}
	} else {
		fmt.Print(str)
	}
}

func LogError(msg string) {
	log(msg, COLOR_RED, SYMBOL_ERROR, false, true)
}

func LogErrorTTY(msg string) {
	log(msg, COLOR_RED, SYMBOL_ERROR, true, true)
}

func LogSuccess(msg string) {
	log(msg, COLOR_GREEN, SYMBOL_SUCCESS, false, true)
}

func LogSuccessTTY(msg string) {
	log(msg, COLOR_GREEN, SYMBOL_SUCCESS, true, true)
}

func LogWarn(msg string) {
	log(msg, COLOR_YELLOW, SYMBOL_WARN, false, true)
}

func LogWarnTTY(msg string) {
	log(msg, COLOR_YELLOW, SYMBOL_WARN, true, true)
}

func LogInfo(msg string) {
	log(msg, COLOR_BLUE, SYMBOL_INFO, false, true)
}

func LogInfoTTY(msg string) {
	log(msg, COLOR_BLUE, SYMBOL_INFO, true, true)
}

func Log(msg string) {
	log(msg, COLOR_RESET, SYMBOL_NONE, false, true)
}

func LogTTY(msg string) {
	log(msg, COLOR_RESET, SYMBOL_NONE, true, true)
}

func PromptTTY(msg string) {
	log(msg, COLOR_RESET, SYMBOL_NONE, true, false)
}
