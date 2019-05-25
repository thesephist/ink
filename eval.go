package main

import (
	"log"
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
		evalNode(heap, node)
	}

	done <- true
}

func evalNode(heap map[string]Value, node interface{}) interface{} {
	log.Printf("--- Evaluating Node ---")
	switch n := node.(type) {
	case IdentifierNode:
		return heap[n.val]
	case BinaryExprNode:
		// do something
	}

	log.Println(node)
	return nil
}
