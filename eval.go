package main

import (
	"fmt"
)

const (
	NumberType = iota
	StringType
	BooleanType
	NullType
	CompositeType
	FunctionType
)

type Value struct {
	val  interface{}
	kind int
}

func Eval(nodes <-chan interface{}, done chan<- bool) {
	heap := make(map[string]Value)
	for node := range nodes {
		evalNode(&heap, node)
	}

	done <- true
}

func evalNode(hp *map[string]Value, node interface{}) {
	switch node.(type) {
	default:
		fmt.Println(node)
	}
}
