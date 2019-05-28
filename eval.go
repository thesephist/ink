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

func (v NumberValue) String() string {
	return "NumberValue"
}

type StringValue struct {
	val string
}

func (v StringValue) String() string {
	return "StringValue"
}

type BooleanValue struct {
	val bool
}

func (v BooleanValue) String() string {
	return "BooleanValue"
}

type NullValue struct{}

func (v NullValue) String() string {
	return "NullValue"
}

type CompositeValue struct {
	val map[string]Value
}

func (v CompositeValue) String() string {
	return "CompositeValue"
}

// XXX: for now, our GC heuristic is simply to dump/free
//	heaps from functions that are no longer referenced in the
//	main isolate's heap, and keep all other heaps, recursively descending.
// This is conservative and inefficient, but will get us started.
type FunctionValue struct {
	node       Node
	parentHeap map[string]Value
	heap       map[string]Value
}

func (v FunctionValue) String() string {
	return "FunctionValue"
}

type Node interface {
	String() string
	Eval(map[string]Value) Value
}

// TODO: we should pull node definitions and these impls out
//	to node.go

func (n UnaryExprNode) String() string {
	return fmt.Sprintf("Unary %s (%s)", n.operator.String(), n.operand)
}

func (n UnaryExprNode) Eval(heap map[string]Value) Value {
	switch n.operator.kind {
	case NegationOp:
		operand := heap[n.operand.Eval(heap).String()]
		switch o := operand.(type) {
		case NumberValue:
			return &NumberValue{-o.val}
		default:
			log.Fatal("Cannot negate non-number value %s", o.String())
			return nil
		}
	}

	return nil
}

func (n BinaryExprNode) String() string {
	return "BinaryExprNode"
}

func (n BinaryExprNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n FunctionCallNode) String() string {
	return "FunctionCallNode"
}

func (n FunctionCallNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n MatchClauseNode) String() string {
	return "MatchClauseNode"
}

func (n MatchClauseNode) Eval(heap map[string]Value) Value {
	log.Fatal("Cannot evaluate a MatchClauseNode")
	return nil
}

func (n MatchExprNode) String() string {
	return "MatchExprNode"
}

func (n MatchExprNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n EmptyIdentifierNode) String() string {
	return "EmptyIdentifierNode"
}

func (n EmptyIdentifierNode) Eval(heap map[string]Value) Value {
	log.Fatal("Cannot evaluate a EmptyIdentifierNode")
	return nil
}

func (n IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

func (n IdentifierNode) Eval(heap map[string]Value) Value {
	return heap[n.val]
}

func (n NumberLiteralNode) String() string {
	return "NumberLiteralNode"
}

func (n NumberLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n StringLiteralNode) String() string {
	return "StringLiteralNode"
}

func (n StringLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n BooleanLiteralNode) String() string {
	return "BooleanLiteralNode"
}

func (n BooleanLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n NullLiteralNode) String() string {
	return "NullLiteralNode"
}

func (n NullLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n ObjectLiteralNode) String() string {
	return "ObjectLiteralNode"
}

func (n ObjectLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n ObjectEntryNode) String() string {
	return "ObjectEntryNode"
}

func (n ObjectEntryNode) Eval(heap map[string]Value) Value {
	log.Fatal("Cannot evaluate ObjectEntryNode")
	return nil
}

func (n ListLiteralNode) String() string {
	return "ListLiteralNode"
}

func (n ListLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n FunctionLiteralNode) String() string {
	return "FunctionLiteralNode"
}

func (n FunctionLiteralNode) Eval(heap map[string]Value) Value {
	return nil
}

type Isolate struct {
	Heap map[string]Value
}

func (iso *Isolate) Eval(nodes <-chan Node, done chan<- bool) {
	for node := range nodes {
		evalNode(iso.Heap, node)
	}

	done <- true
}

func evalNode(heap map[string]Value, node Node) Value {
	log.Printf("Evaluating Node: %s", node.String())
	switch n := node.(type) {
	case Node:
		return n.Eval(heap)
	default:
		log.Println(n)
	}

	return nil
}
