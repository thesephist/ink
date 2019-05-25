package main

import (
	"fmt"
)

const (
	Literal = iota
	Ident
	Atom
	UnaryOp
	BinaryOp
	TernaryOp
)

const (
	NumberType = iota
	StringType
	BooleanType
	NullType
	CompositeType
	FunctionType
)

type srcslice struct {
	startLine, startCol int
	endLine, endCol     int
}

type Tok struct {
	val string
	srcslice
}

func Tokenize(input <-chan rune) chan<- Tok {
	output := make(chan Tok)

	// parse stuff
	for char := range input {
		fmt.Println(string(char))
	}

	return output
}
