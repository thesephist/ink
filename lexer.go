package main

import (
	"fmt"
	"unicode"
)

type span struct {
	startLine, startCol int
	endLine, endCol     int
}

type Tok struct {
	val string
	span
}

func Tokenize(input <-chan rune) chan<- Tok {
	output := make(chan Tok)

	// parse stuff
	for char := range input {
		fmt.Println(string(char))
	}

	return output
}

func isValidIdentifierChar(char rune) bool {
	if unicode.IsDigit(char) || unicode.IsLetter(char) {
		return true
	}

	switch char {
	case '@', '!', '?':
		return true
	default:
		return false
	}
}

func isKeyword(tok string) bool {
	switch tok {
	case "true", "false", "null":
		return true
	default:
		return false
	}
}
