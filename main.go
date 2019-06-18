package main

import (
	"bufio"
	"os"
)

func main() {
	input := make(chan rune)
	inputReader := bufio.NewReader(os.Stdin)

	tokens := make(chan Tok)
	nodes := make(chan Node)

	done := make(chan bool, 3)

	// Ideally, we don't have to use this to do async because
	// 	all the info are included in the channels... TODO: fix.
	iso := Isolate{}
	go Tokenize(input, tokens, done)
	go Parse(tokens, nodes, done)
	go iso.Eval(nodes, done)

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
