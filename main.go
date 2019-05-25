package main

import (
	"bufio"
	"os"
)

func main() {
	input := make(chan rune)
	inputReader := bufio.NewReader(os.Stdin)

	go Tokenize(input)

	for {
		char, _, err := inputReader.ReadRune()
		if err != nil {
			break
		}
		input <- char
	}
}
