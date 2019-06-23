package main

import (
	"bufio"
	"flag"
	"os"
)

// TODO: By default, read from stdin but if passed flag -i --input,
// 	read from input file(s) including imports if multipel inputs.
func main() {
	// flags
	debugLexer := flag.Bool("debug-lex", false, "Log lexer output")
	debugParser := flag.Bool("debug-parse", false, "Log parser output")

	flag.Parse()

	input := make(chan rune)
	done := make(chan bool, 3)

	iso := Isolate{}
	tokens := make(chan Tok)
	nodes := make(chan Node)
	go Tokenize(input, tokens, *debugLexer, done)
	go Parse(tokens, nodes, *debugParser, done)
	go iso.Eval(nodes, done)

	inputReader := bufio.NewReader(os.Stdin)
	for {
		char, _, err := inputReader.ReadRune()
		if err != nil {
			break
		}
		input <- char
	}
	close(input)

	// wait for evals on other threads to finish
	for i := 0; i < 3; i++ {
		<-done
	}
}
