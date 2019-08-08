package ink

import (
	"fmt"
	"os"
	"strings"
)

const (
	ANSI_RESET       = "[0;0m"
	ANSI_BLUE        = "[34;22m"
	ANSI_GREEN       = "[32;22m"
	ANSI_YELLOW      = "[33;22m"
	ANSI_RED         = "[31;22m"
	ANSI_BLUE_BOLD   = "[34;1m"
	ANSI_GREEN_BOLD  = "[32;1m"
	ANSI_YELLOW_BOLD = "[33;1m"
	ANSI_RED_BOLD    = "[31;1m"
)

func LogDebug(args ...string) {
	fmt.Println(ANSI_BLUE_BOLD + "debug: " + ANSI_BLUE + strings.Join(args, " ") + ANSI_RESET)
}

func LogDebugf(s string, args ...interface{}) {
	LogDebug(fmt.Sprintf(s, args...))
}

func LogInteractive(args ...string) {
	fmt.Println(ANSI_GREEN + strings.Join(args, " ") + ANSI_RESET)
}

func LogInteractivef(s string, args ...interface{}) {
	LogInteractive(fmt.Sprintf(s, args...))
}

func LogSafeErr(reason int, args ...string) {
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
	default:
		errStr = "error"
	}
	fmt.Fprintln(os.Stderr, ANSI_RED_BOLD+errStr+": "+ANSI_RED+strings.Join(args, " ")+ANSI_RESET)
}

func LogErr(reason int, args ...string) {
	LogSafeErr(reason, args...)
	os.Exit(reason)
}

func LogErrf(reason int, s string, args ...interface{}) {
	LogErr(reason, fmt.Sprintf(s, args...))
}
