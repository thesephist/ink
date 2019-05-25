package main

import (
	"bufio"
	"os"
)

func main() {
	input := make(chan rune)
	inputReader := bufio.NewReader(os.Stdin)

	tokens := make(chan Tok)

	go Tokenize(input, tokens)
	go Parse(tokens)

	for {
		char, _, err := inputReader.ReadRune()
		if err != nil {
			break
		}
		input <- char
	}
	input <- '\n' // final line separator, hacky but does the job
}
