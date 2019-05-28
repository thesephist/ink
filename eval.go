package main

import (
	"fmt"
	"log"
)

type Value interface {
	String() string
}

type NumberValue struct {
	val float64
}

type StringValue struct {
	val string
}

type BooleanValue struct {
	val bool
}

type NullValue struct{}

type CompositeValue struct {
	val map[string]*Value
}

// XXX: for now, our GC heuristic is simply to dump/free
//	heaps from functions that are no longer referenced in the
//	main isolate's heap, and keep all other heaps, recursively descending.
// This is conservative and inefficient, but will get us started.
type FunctionValue struct {
	node       Node
	parentHeap map[string]*Value
	heap       map[string]*Value
}

type Node interface {
	String() string
	Eval(map[string]*Value) *Value
}

// TODO: we should pull node definitions and these impls out
//	to node.go

func (n *IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

func (n *IdentifierNode) Eval(heap map[string]*Value) *Value {
	return heap[n.val]
}

type Isolate struct {
	Heap map[string]*Value
}

func (iso *Isolate) Eval(nodes <-chan interface{}, done chan<- bool) {
	for node := range nodes {
		evalNode(iso.Heap, node)
	}

	done <- true
}

func evalNode(heap map[string]*Value, node interface{}) interface{} {
	log.Printf("--- Evaluating Node ---")
	switch n := node.(type) {
	case Node:
		return n.Eval(heap)
	default:
		log.Println(n)
	}

	return nil
}
