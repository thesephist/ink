package ink

import (
	"fmt"
	"os"
	"strings"
)

const (
	// ANSI terminal escape codes for color output
	AnsiReset      = "[0;0m"
	AnsiBlue       = "[34;22m"
	AnsiGreen      = "[32;22m"
	ansiYellow     = "[33;22m"
	AnsiRed        = "[31;22m"
	AnsiBlueBold   = "[34;1m"
	AnsiGreenBold  = "[32;1m"
	AnsiYellowBold = "[33;1m"
	AnsiRedBold    = "[31;1m"
)

func LogDebug(args ...string) {
	fmt.Println(AnsiBlueBold + "debug: " + AnsiBlue + strings.Join(args, " ") + AnsiReset)
}

func LogDebugf(s string, args ...interface{}) {
	LogDebug(fmt.Sprintf(s, args...))
}

func LogInteractive(args ...string) {
	fmt.Println(AnsiGreen + strings.Join(args, " ") + AnsiReset)
}

func LogInteractivef(s string, args ...interface{}) {
	LogInteractive(fmt.Sprintf(s, args...))
}

// LogSafeErr is like LogErr, but does not immediately exit the interpreter
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
	fmt.Fprintln(os.Stderr, AnsiRedBold+errStr+": "+AnsiRed+strings.Join(args, " ")+AnsiReset)
}

func LogErr(reason int, args ...string) {
	LogSafeErr(reason, args...)
	os.Exit(reason)
}

func LogErrf(reason int, s string, args ...interface{}) {
	LogErr(reason, fmt.Sprintf(s, args...))
}
