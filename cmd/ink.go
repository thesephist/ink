package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/thesephist/ink/pkg/ink"
)

const cliVersion = "0.1.7"

const helpMessage = `
Ink is a minimal, powerful, functional programming language.
	ink v%s

By default, ink interprets from stdin.
	ink < main.ink
Run Ink programs from source files by passing it to the interpreter.
	ink main.ink other.ink
Start an interactive repl.
	ink
	>
Run from the command line with -eval.
	ink -eval "f := () => out('hi'), f()"

`

func main() {
	flag.Usage = func() {
		fmt.Printf(helpMessage, cliVersion)
		flag.PrintDefaults()
	}

	// permission flags
	noRead := flag.Bool("no-read", false, "Silently fail all read operations")
	noWrite := flag.Bool("no-write", false, "Silently fail all write operations")
	noNet := flag.Bool("no-net", false, "Silently fail all access to local network")
	noExec := flag.Bool("no-exec", false, "Silently fail all exec calls")
	isolate := flag.Bool("isolate", false, "Isolate all system operations: read, write, net, exec")

	// cli arguments
	verbose := flag.Bool("verbose", false, "Log all interpreter debug information")
	debugLexer := flag.Bool("debug-lex", false, "Log lexer output")
	debugParser := flag.Bool("debug-parse", false, "Log parser output")
	dump := flag.Bool("dump", false, "Dump global frame after eval")

	version := flag.Bool("version", false, "Print version string and exit")
	help := flag.Bool("help", false, "Print help message and exit")

	repl := flag.Bool("repl", false, "Run as an interactive repl")
	eval := flag.String("eval", "", "Evaluate argument as an Ink program")

	flag.Parse()

	// collect all other non-parsed arguments from the CLI as files to be run
	args := flag.Args()

	// if asked for version, disregard everything else
	if *version {
		fmt.Printf("ink v%s\n", cliVersion)
		os.Exit(0)
	} else if *help {
		flag.Usage()
		os.Exit(0)
	}

	// if no files given and no stdin, default to repl
	stdinStat, _ := os.Stdin.Stat()
	if len(args) == 0 && (stdinStat.Mode()&os.ModeCharDevice) != 0 && *eval == "" {
		*repl = true
	}

	// execution environment
	eng := ink.Engine{
		FatalError: false,
		Permissions: ink.PermissionsConfig{
			Read:  !*noRead && !*isolate,
			Write: !*noWrite && !*isolate,
			Net:   !*noNet && !*isolate,
			Exec:  !*noExec && !*isolate,
		},
		Debug: ink.DebugConfig{
			Lex:   *debugLexer || *verbose,
			Parse: *debugParser || *verbose,
			Dump:  *dump || *verbose,
		},
	}

	if *repl {
		ctx := eng.CreateContext()

		// add repl-specific builtins
		ctx.LoadFunc("clear", func(ctx *ink.Context, in []ink.Value) (ink.Value, error) {
			fmt.Printf("[2J[H")
			return ink.NullValue{}, nil
		})
		ctx.LoadFunc("dump", func(ctx *ink.Context, in []ink.Value) (ink.Value, error) {
			ctx.Dump()
			return ink.NullValue{}, nil
		})

		// run interactively in a repl
		reader := bufio.NewReader(os.Stdin)

		for {
			// green arrow
			fmt.Printf(ink.AnsiGreenBold + "> " + ink.AnsiReset)
			text, err := reader.ReadString('\n')

			if err == io.EOF {
				break
			} else if err != nil {
				ink.LogErrf(
					ink.ErrSystem,
					"unexpected end of input:\n\t-> %s", err.Error(),
				)
			}

			// we don't really care if expressions fail to eval
			// at the top level, user will see regardless, so drop err
			val, _ := ctx.Exec(strings.NewReader(text))
			if val != nil {
				ink.LogInteractive(val.String())
			}
		}

		// no need to wait for eng.Listeners, since this part here
		//	is unreachable
	} else if *eval != "" {
		ctx := eng.CreateContext()
		eng.FatalError = true

		ctx.Exec(strings.NewReader(*eval))
		eng.Listeners.Wait()
	} else if len(args) > 0 {
		filePath := args[0]
		ctx := eng.CreateContext()

		ctx.ExecPath(filePath)
		eng.Listeners.Wait()
	} else {
		ctx := eng.CreateContext()
		eng.FatalError = true

		ctx.Exec(os.Stdin)
		eng.Listeners.Wait()
	}
}
