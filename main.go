package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
)

const VERSION = "0.1.2"

const HELP_MSG = `
Ink is a minimal, powerful, functional programming language.
	ink v%s

By default, ink interprets from stdin.
	ink < main.ink
Run Ink programs from source files by passing it to the interpreter.
	ink main.ink other.ink
Start an interactive repl with -repl.
	ink -repl
	> ...
Run from the command line with -eval.
	ink -eval "f := () => out('hi'), f()"

`

func main() {
	flag.Usage = func() {
		fmt.Printf(HELP_MSG, VERSION)
		flag.PrintDefaults()
	}

	// permission flags
	noRead := flag.Bool("no-read", false, "Silently fail all read operations")
	noWrite := flag.Bool("no-write", false, "Silently fail all write operations")
	noNet := flag.Bool("no-net", false, "Silently fail all access to local network")
	isolate := flag.Bool("isolate", false, "Isolate all system operations: read, write, net")

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
	files := flag.Args()

	// if asked for version, disregard everything else
	if *version {
		fmt.Println("ink", VERSION)
		os.Exit(0)
	} else if *help {
		flag.Usage()
		os.Exit(0)
	}

	// execution environment
	eng := Engine{
		FatalError: false,
		Permissions: PermissionsConfig{
			Read:  !*noRead && !*isolate,
			Write: !*noWrite && !*isolate,
			Net:   !*noNet && !*isolate,
		},
		Debug: DebugConfig{
			Lex:   *debugLexer || *verbose,
			Parse: *debugParser || *verbose,
			Dump:  *dump || *verbose,
		},
	}

	if *repl {
		ctx := eng.CreateContext()

		// run interactively in a repl
		reader := bufio.NewReader(os.Stdin)

	replLoop:
		for {
			// green arrow
			fmt.Printf(ANSI_GREEN_BOLD + "> " + ANSI_RESET)
			text, err := reader.ReadString('\n')

			if err != nil {
				logErrf(ErrSystem, "unexpected stop to input:\n\t-> %s", err.Error())
			}

			switch {
			// specialized introspection / observability directives
			//	in repl session
			case strings.HasPrefix(text, "@dump"):
				ctx.Dump()
			case strings.HasPrefix(text, "@clear"):
				fmt.Printf("[2J[H")
			case strings.HasPrefix(text, "@exit"):
				break replLoop

			default:
				input := make(chan rune)
				resolver := ctx.ExecStream(input)

				for _, char := range text {
					input <- char
				}
				close(input)

				// wait for main Context to finish executing before
				//	yielding to the repl loop
				val, err := (<-resolver)()
				if err != nil {
					// err handled by the error stream
				} else if val != nil {
					logInteractive(val.String())
				}
			}
		}

		// no need to wait for eng.Listeners, since this part here
		//	is unreachable
	} else if *eval != "" {
		ctx := eng.CreateContext()
		eng.FatalError = true

		input := make(chan rune)
		ctx.ExecStream(input)

		for _, char := range *eval {
			input <- char
		}
		close(input)

		eng.Listeners.Wait()
	} else if len(files) > 0 {
		// read from file
		for _, filePath := range files {
			// execution context is one-per-file
			ctx := eng.CreateContext()

			// expand out ~ for $HOME, which is not done by shells
			if strings.HasPrefix(filePath, "~/") {
				filePath = os.Getenv("HOME") + "/" + filePath[2:]
			}

			// canonicalize relative paths, but not absolute ones
			if !path.IsAbs(filePath) {
				filePath = path.Join(ctx.Cwd, filePath)
			}

			err := ctx.ExecFile(filePath)
			if err != nil {
				logSafeErr(
					ErrSystem,
					fmt.Sprintf("could not open %s for execution:\n\t-> %s",
						filePath, err),
				)
			}

			eng.Listeners.Wait()
		}
	} else {
		ctx := eng.CreateContext()
		eng.FatalError = true

		// read from stdin
		input := make(chan rune)
		ctx.ExecStream(input)

		inputReader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := inputReader.ReadRune()
			if err != nil {
				break
			}
			input <- char
		}
		close(input)

		eng.Listeners.Wait()
	}
}
