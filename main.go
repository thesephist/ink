package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
)

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
	// cli arguments
	verbose := flag.Bool("verbose", false, "Log all interpreter steps")
	debugLexer := flag.Bool("debug-lex", false, "Log lexer output")
	debugParser := flag.Bool("debug-parse", false, "Log parser output")

	var files inkFiles
	flag.Var(&files, "input", "Source code to execute")

	flag.Parse()

	// rep(l)
	input := make(chan rune)
	done := make(chan bool, 3)

	iso := Isolate{}
	tokens := make(chan Tok)
	nodes := make(chan Node)
	go Tokenize(input, tokens, *debugLexer || *verbose, done)
	go Parse(tokens, nodes, *debugParser || *verbose, done)
	go iso.Eval(nodes, done)

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
