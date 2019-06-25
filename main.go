package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
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

	var files inkFiles
	flag.Var(&files, "input", "Source code to execute, can be invoked multiple times")

	flag.Parse()

	// if asked for version, disregard everything else
	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	} else if *help {
		flag.Usage()
	}

	// rep(l)
	input := make(chan rune)
	done := make(chan bool, 3)

	iso := Isolate{}
	tokens := make(chan Tok)
	nodes := make(chan Node)
	go Tokenize(input, tokens, *debugLexer || *verbose, done)
	go Parse(tokens, nodes, *debugParser || *verbose, done)
	go iso.Eval(nodes, *dump || *verbose, done)

	if len(files) > 0 {
		for _, path := range files {
			file, err := os.Open(path)
			if err != nil {
				log.Fatalf("Could not open %s for execution: %s", path, err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				for _, char := range scanner.Text() {
					input <- char
				}
				input <- '\n'
			}
		}
	} else {
		inputReader := bufio.NewReader(os.Stdin)
		for {
			char, _, err := inputReader.ReadRune()
			if err != nil {
				break
			}
			input <- char
		}
	}
	close(input)

	// wait for evals on other threads to finish
	for i := 0; i < 3; i++ {
		<-done
	}
}
