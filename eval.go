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
	entries map[string]Value
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
	return fmt.Sprintf("Unary %s (%s)", n.operator.String(), n.operand.String())
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
	var op = "??"
	switch n.operator.kind {
	case AddOp:
		op = "+"
	case SubtractOp:
		op = "-"
	case MultiplyOp:
		op = "*"
	case DivideOp:
		op = "/"
	case ModulusOp:
		op = "%"
	case GreaterThanOp:
		op = ">"
	case LessThanOp:
		op = "<"
	case EqualOp:
		op = "="
	case EqRefOp:
		op = "is"
	case DefineOp:
		op = ":="
	case AccessorOp:
		op = "."
	}
	return fmt.Sprintf("Binary (%s) %s (%s)",
		n.leftOperand.String(),
		op,
		n.rightOperand.String())
}

func (n BinaryExprNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n FunctionCallNode) String() string {
	if len(n.arguments) == 0 {
		return fmt.Sprintf("Call (%s) on ()", n.function)
	} else {
		args := n.arguments[0].String()
		for _, a := range n.arguments[1:] {
			args += ", " + a.String()
		}
		return fmt.Sprintf("Call (%s) on (%s)", n.function, args)
	}
}

func (n FunctionCallNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n MatchClauseNode) String() string {
	return "MatchClauseNode"
	if len(n.expressions) == 0 {
		return fmt.Sprintf("Clause (%s) -> noop", n.target.String())
	} else {
		exprs := n.expressions[0].String()
		for _, e := range n.expressions[1:] {
			exprs += ", " + e.String()
		}
		return fmt.Sprintf("Clause (%s) -> {%s}", n.target.String(), exprs)
	}
}

func (n MatchClauseNode) Eval(heap map[string]Value) Value {
	log.Fatal("Cannot evaluate a MatchClauseNode")
	return nil
}

func (n MatchExprNode) String() string {
	if len(n.clauses) == 0 {
		return fmt.Sprintf("Match on (%s) to {}", n.condition.String())
	} else {
		clauses := n.clauses[0].String()
		for _, c := range n.clauses[1:] {
			clauses += ", " + c.String()
		}
		return fmt.Sprintf("Match on (%s) to {%s}",
			n.condition.String(),
			clauses)
	}
}

func (n MatchExprNode) Eval(heap map[string]Value) Value {
	return nil
}

func (n EmptyIdentifierNode) String() string {
	return "Empty Identifier"
}

func (n EmptyIdentifierNode) Eval(heap map[string]Value) Value {
	log.Fatal("Cannot evaluate an EmptyIdentifierNode")
	return nil
}

func (n IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

func (n IdentifierNode) Eval(heap map[string]Value) Value {
	val, prs := heap[n.val]
	if !prs {
		log.Fatal("%s is not defined", n.val)
	}
	return val
}

func (n NumberLiteralNode) String() string {
	return fmt.Sprintf("Number %f", n.val)
}

func (n NumberLiteralNode) Eval(heap map[string]Value) Value {
	return NumberValue{n.val}
}

func (n StringLiteralNode) String() string {
	return fmt.Sprintf("String %s", n.val)
}

func (n StringLiteralNode) Eval(heap map[string]Value) Value {
	return StringValue{n.val}
}

func (n BooleanLiteralNode) String() string {
	return fmt.Sprintf("Boolean %t", n.val)
}

func (n BooleanLiteralNode) Eval(heap map[string]Value) Value {
	return BooleanValue{n.val}
}

func (n NullLiteralNode) String() string {
	return "Null"
}

func (n NullLiteralNode) Eval(heap map[string]Value) Value {
	return NullValue{}
}

func (n ObjectLiteralNode) String() string {
	if len(n.entries) == 0 {
		return fmt.Sprintf("Object {}")
	} else {
		entries := n.entries[0].String()
		for _, e := range n.entries[1:] {
			entries += ", " + e.String()
		}
		return fmt.Sprintf("Object {%s}", entries)
	}
}

func (n ObjectLiteralNode) Eval(heap map[string]Value) Value {
	obj := CompositeValue{}
	es := obj.entries
	for _, entry := range n.entries {
		k, ok := entry.key.(IdentifierNode)
		if ok {
			es[k.val] = entry.val.Eval(heap)
		} else {
			key := entry.key.Eval(heap)
			keyStrVal, ok := key.(StringValue)
			if ok {
				es[keyStrVal.val] = entry.val.Eval(heap)
			} else {
				log.Fatal("Cannot access non-string property %s of object",
					key.String())
			}
		}
	}
	return obj
}

func (n ObjectEntryNode) String() string {
	return fmt.Sprintf("Object Entry (%s): (%s)", n.key.String(), n.val.String())
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
