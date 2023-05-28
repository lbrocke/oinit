package main

import (
	"fmt"
	"os"
)

type Color string

const (
	SYMBOL_ERROR   = "✘"
	SYMBOL_SUCCESS = "✔"
	SYMBOL_WARN    = "!"
	SYMBOL_INFO    = "i"

	COLOR_RED    = "\u001b[31m"
	COLOR_GREEN  = "\u001b[32m"
	COLOR_YELLOW = "\u001b[33m"
	COLOR_BLUE   = "\u001b[34m"
	COLOR_RESET  = "\u001b[0m"
)

func log(msg string, color string, symbol string, tty bool) {
	var str string

	// See https://no-color.org
	if os.Getenv("NO_COLOR") != "" {
		str = fmt.Sprint(symbol, " ", msg, "\n")
	} else {
		str = fmt.Sprint(color, symbol, COLOR_RESET, " ", msg, "\n")
	}

	if tty {
		os.WriteFile("/dev/tty", []byte(str), 0644)
	} else {
		fmt.Print(str)
	}
}

func LogError(msg string) {
	log(msg, COLOR_RED, SYMBOL_ERROR, false)
}

func LogErrorTTY(msg string) {
	log(msg, COLOR_RED, SYMBOL_ERROR, true)
}

func LogSuccess(msg string) {
	log(msg, COLOR_GREEN, SYMBOL_SUCCESS, false)
}

func LogSuccessTTY(msg string) {
	log(msg, COLOR_GREEN, SYMBOL_SUCCESS, true)
}

func LogWarn(msg string) {
	log(msg, COLOR_YELLOW, SYMBOL_WARN, false)
}

func LogWarnTTY(msg string) {
	log(msg, COLOR_YELLOW, SYMBOL_WARN, true)
}

func LogInfo(msg string) {
	log(msg, COLOR_BLUE, SYMBOL_INFO, false)
}

func LogInfoTTY(msg string) {
	log(msg, COLOR_BLUE, SYMBOL_INFO, true)
}
