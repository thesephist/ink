package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const VERSION = "0.1"

const HELP_MSG = `
Ink is a minimal, functional programming language.
	ink v%s

By default, ink interprets from stdin. Run an ink script with -input
	ink -input main.ink

`

// input files flag parsing
type inkFiles []string

func (i *inkFiles) Set(val string) error {
	*i = append(*i, val)
	return nil
}

func (i *inkFiles) String() string {
	return strings.Join(*i, ", ")
}

func main() {
	flag.Usage = func() {
		fmt.Printf(HELP_MSG, VERSION)
		flag.PrintDefaults()
	}

	// cli arguments
	verbose := flag.Bool("verbose", false, "Log all interpreter debug information")
	debugLexer := flag.Bool("debug-lex", false, "Log lexer output")
	debugParser := flag.Bool("debug-parse", false, "Log parser output")
	dump := flag.Bool("dump", false, "Dump heap after eval")

	version := flag.Bool("version", false, "Print version string and exit")
	help := flag.Bool("help", false, "Print help message and exit")

	repl := flag.Bool("repl", false, "Run as an interactive repl")

	var files inkFiles
	flag.Var(&files, "input", "Source code to execute, can be invoked multiple times")

	flag.Parse()

	// if asked for version, disregard everything else
	if *version {
		fmt.Println("ink", VERSION)
		os.Exit(0)
	} else if *help {
		flag.Usage()
		os.Exit(0)
	}

	// execution context
	iso := Isolate{}
	iso.Init()

	execFile := func(path string) {
		// read from file
		input, values := iso.ExecStream(
			*debugLexer || *verbose,
			*debugParser || *verbose,
			*dump || *verbose,
		)

		file, err := os.Open(path)
		if err != nil {
			logErrf(ErrSystem, "could not open %s for execution:\n\t-> %s", path, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			for _, char := range scanner.Text() {
				input <- char
			}
			input <- '\n'
		}
		close(input)

		for _ = range values {
		}
	}

	if *repl {
		// run interactively in a repl
		logDebug("started ink repl")
		reader := bufio.NewReader(os.Stdin)

		shouldExit := false
		for !shouldExit {
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
				iso.Dump()
			case strings.HasPrefix(text, "@clear"):
				fmt.Printf("[2J[H")
			case strings.HasPrefix(text, "@exit"):
				shouldExit = true
			case strings.HasPrefix(text, "@load "):
				path := strings.Trim(text[5:], " \t\n")
				execFile(path)
				logDebugf("loaded file:\n\t-> %s", path)

			default:
				input, values := iso.ExecStream(
					*debugLexer || *verbose,
					*debugParser || *verbose,
					*dump || *verbose,
				)

				for _, char := range text {
					input <- char
				}
				close(input)

				for v := range values {
					logInteractive(v.String())
				}
			}
		}

		logDebug("exited ink repl")
		os.Exit(0)
	} else if len(files) > 0 {
		// read from file
		for _, path := range files {
			execFile(path)
		}
	} else {
		// read from stdin
		input, values := iso.ExecStream(
			*debugLexer || *verbose,
			*debugParser || *verbose,
			*dump || *verbose,
		)

		inputReader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := inputReader.ReadRune()
			if err != nil {
				break
			}
			input <- char
		}
		close(input)

		for _ = range values {
		}
	}
}
