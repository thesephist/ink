package main

import (
	"bufio"
	"os"
)

func main() {
	input := make(chan rune)
	inputReader := bufio.NewReader(os.Stdin)

	tokens := make(chan Tok)
	nodes := make(chan interface{})

	done := make(chan bool, 3)
	go Tokenize(input, tokens, done)
	go Parse(tokens, nodes, done)
	go Eval(nodes, done)

	for {
		char, _, err := inputReader.ReadRune()
		if err != nil {
			break
		}
		input <- char
	}
	close(input)

	for i := 0; i < 3; i++ {
		<-done
	}
}
