package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	input := make(chan rune)
	inputReader := bufio.NewReader(os.Stdin)

	tokens := make(chan Tok)

	go Tokenize(input, tokens)
	// temp
	go func() {
		for tok := range tokens {
			fmt.Println("Token: " + tokKindToName(tok.kind) + " | " + tok.stringVal())
		}
	}()

	for {
		char, _, err := inputReader.ReadRune()
		if err != nil {
			break
		}
		input <- char
	}
}
