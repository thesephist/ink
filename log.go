package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	ANSI_RESET       = "[0;0m"
	ANSI_BLUE        = "[34;22m"
	ANSI_YELLOW      = "[33;22m"
	ANSI_RED         = "[31;22m"
	ANSI_BLUE_BOLD   = "[34;1m"
	ANSI_YELLOW_BOLD = "[33;1m"
	ANSI_RED_BOLD    = "[31;1m"
)

const (
	ErrSyntax  = 1
	ErrRuntime = 2
	ErrSystem  = 40
	ErrAssert  = 100
)

func logDebug(args ...string) {
	fmt.Println(ANSI_BLUE_BOLD + "debug: " + ANSI_BLUE + strings.Join(args, " ") + ANSI_RESET)
}

func logDebugf(s string, args ...interface{}) {
	logDebug(fmt.Sprintf(s, args...))
}

func logWarn(args ...string) {
	fmt.Println(ANSI_YELLOW_BOLD + "warn: " + ANSI_YELLOW + strings.Join(args, " ") + ANSI_RESET)
}

func logWarnf(s string, args ...interface{}) {
	logWarn(fmt.Sprintf(s, args...))
}

func logErr(reason int, args ...string) {
	errStr := "error"
	switch reason {
	case ErrSyntax:
		errStr = "syntax error"
	case ErrRuntime:
		errStr = "runtime error"
	case ErrSystem:
		errStr = "system error"
	case ErrAssert:
		errStr = "invariant violation"
	}
	fmt.Println(ANSI_RED_BOLD + errStr + ": " + ANSI_RED + strings.Join(args, " ") + ANSI_RESET)
	os.Exit(reason)
}

func logErrf(reason int, s string, args ...interface{}) {
	logErr(reason, fmt.Sprintf(s, args...))
}
