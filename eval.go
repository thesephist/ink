package main

import (
	"fmt"
	"strings"
)

type Value interface {
	String() string
	Equals(Value) bool // deep, value equality
}

// TODO: implement bytes literal and values, and make
//	file read/write APIs on that, rather than text

// TODO: implement exception handling / error values
//	let's do L3-style Result types that are composite values
//	with an error value returned with the return value.

// The EmptyValue is the value of the empty identifier.
//	it is globally unique and matches everything in equality.
type EmptyValue struct{}

func (v EmptyValue) String() string {
	return "_"
}

func (v EmptyValue) Equals(other Value) bool {
	return true
}

type NumberValue struct {
	val float64
}

func (v NumberValue) String() string {
	return fmt.Sprintf("%f", v.val)
}

func (v NumberValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	ov, ok := other.(NumberValue)
	if ok {
		return v.val == ov.val
	} else {
		return false
	}
}

type StringValue struct {
	val string
}

func (v StringValue) String() string {
	return fmt.Sprintf("%s", v.val)
}

func (v StringValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	ov, ok := other.(StringValue)
	if ok {
		return v.val == ov.val
	} else {
		return false
	}
}

type BooleanValue struct {
	val bool
}

func (v BooleanValue) String() string {
	if v.val {
		return "true"
	} else {
		return "false"
	}
}

func (v BooleanValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	ov, ok := other.(BooleanValue)
	if ok {
		return v.val == ov.val
	} else {
		return false
	}
}

type NullValue struct{}

func (v NullValue) String() string {
	return "null"
}

func (v NullValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	_, ok := other.(NullValue)
	return ok
}

type CompositeValue struct {
	entries ValueTable
}

func (v CompositeValue) String() string {
	if len(v.entries) == 0 {
		return "{}"
	} else {
		entries := make([]string, 0)
		for key, ent := range v.entries {
			entries = append(
				entries,
				fmt.Sprintf("%s: %s", key, ent.String()),
			)
		}
		return fmt.Sprintf("{%s}", strings.Join(entries, ", "))
	}
}

func (v CompositeValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	ov, ok := other.(CompositeValue)
	if ok {
		if len(v.entries) != len(ov.entries) {
			return false
		}

		for key, val := range v.entries {
			otherVal, prs := ov.entries[key]
			if prs && !val.Equals(otherVal) {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

// XXX: for now, our GC heuristic is simply to dump/free
//	heaps from functions that are no longer referenced in the
//	main isolate's heap, and keep all other heaps, recursively descending.
// This is conservative and inefficient, but will get us started.
type FunctionValue struct {
	defNode    FunctionLiteralNode
	parentHeap *StackHeap
}

func (v FunctionValue) String() string {
	// XXX: improve this notation
	return v.defNode.String()
}

func (v FunctionValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	ov, ok := other.(FunctionValue)
	if ok {
		// to compare structs containing slices, we really want
		//	a pointer comparison, not a value comparison
		return &v.defNode == &ov.defNode
	} else {
		return false
	}
}

type Node interface {
	String() string
	Eval(*StackHeap) Value
	// TODO: func (n Node) prettyString() string - pretty-print AST
}

func (n UnaryExprNode) String() string {
	return fmt.Sprintf("Unary %s (%s)", n.operator.String(), n.operand.String())
}

func (n UnaryExprNode) Eval(heap *StackHeap) Value {
	switch n.operator.kind {
	case NegationOp:
		operand := n.operand.Eval(heap)
		switch o := operand.(type) {
		case NumberValue:
			return NumberValue{-o.val}
		case BooleanValue:
			return BooleanValue{!o.val}
		default:
			logErrf(ErrRuntime, "cannot negate non-number value %s", o.String())
			return NullValue{}
		}
	}
	logErrf(ErrAssert, "unrecognized unary operator %s", n)
	return NullValue{}
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

func operandToStringKey(rightOperand Node, heap *StackHeap) string {
	var rightValue string
	switch right := rightOperand.(type) {
	case IdentifierNode:
		rightValue = right.val
	case NumberLiteralNode:
		rightValue = fmt.Sprintf("%f", right.val)
	default:
		rightEvaluatedValue := rightOperand.Eval(heap)
		rv, ok := rightEvaluatedValue.(StringValue)
		if ok {
			rightValue = rv.val
		} else {
			logErrf(ErrRuntime, "cannot access property %s of an object",
				rightEvaluatedValue.String())
		}
	}
	return rightValue
}

func (n BinaryExprNode) Eval(heap *StackHeap) Value {
	if n.operator.kind == DefineOp {
		leftIdent, okIdent := n.leftOperand.(IdentifierNode)
		leftAccess, okAccess := n.leftOperand.(BinaryExprNode)
		if okIdent {
			rightValue := n.rightOperand.Eval(heap)
			heap.setValue(leftIdent.val, rightValue)
			return rightValue
		} else if okAccess && leftAccess.operator.kind == AccessorOp {
			leftObject := leftAccess.leftOperand.Eval(heap)
			leftKey := operandToStringKey(leftAccess.rightOperand, heap)
			leftObjectComposite, ok := leftObject.(CompositeValue)
			if ok {
				rightValue := n.rightOperand.Eval(heap)
				leftObjectComposite.entries[leftKey] = rightValue
				return rightValue
			} else {
				logErrf(ErrRuntime, "cannot set property of a non-composite value %s",
					leftObject)
			}
		} else {
			logErrf(ErrRuntime, "cannot assign value to non-identifier %s",
				n.leftOperand.Eval(heap).String())
			return nil
		}
	} else if n.operator.kind == AccessorOp {
		leftValue := n.leftOperand.Eval(heap)
		rightValue := operandToStringKey(n.rightOperand, heap)
		leftValueComposite, ok := leftValue.(CompositeValue)
		if ok {
			return leftValueComposite.entries[rightValue]
		} else {
			logErrf(ErrRuntime, "cannot access property of a non-object %s",
				leftValue)
		}
	}

	leftValue := n.leftOperand.Eval(heap)
	rightValue := n.rightOperand.Eval(heap)
	switch n.operator.kind {
	case AddOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return NumberValue{left.val + right.val}
			}
		case StringValue:
			right, ok := rightValue.(StringValue)
			if ok {
				return StringValue{left.val + right.val}
			}
		case BooleanValue:
			right, ok := rightValue.(BooleanValue)
			if ok {
				return BooleanValue{left.val || right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support addition",
			leftValue, rightValue)
	case SubtractOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return NumberValue{left.val - right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support subtraction",
			leftValue, rightValue)
	case MultiplyOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return NumberValue{left.val * right.val}
			}
		case BooleanValue:
			right, ok := rightValue.(BooleanValue)
			if ok {
				return BooleanValue{left.val && right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support multiplication",
			leftValue, rightValue)
	case DivideOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return NumberValue{left.val / right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support division",
			leftValue, rightValue)
	case ModulusOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				// XXX: warn if not integers
				return NumberValue{float64(
					int(left.val) % int(right.val),
				)}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support modulus",
			leftValue, rightValue)
	case GreaterThanOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return BooleanValue{left.val > right.val}
			}
		case StringValue:
			right, ok := rightValue.(StringValue)
			if ok {
				return BooleanValue{left.val > right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support comparison",
			leftValue, rightValue)
	case LessThanOp:
		switch left := leftValue.(type) {
		case NumberValue:
			right, ok := rightValue.(NumberValue)
			if ok {
				return BooleanValue{left.val < right.val}
			}
		case StringValue:
			right, ok := rightValue.(StringValue)
			if ok {
				return BooleanValue{left.val < right.val}
			}
		}
		logErrf(ErrRuntime, "values %s and %s do not support comparison",
			leftValue, rightValue)
	case EqualOp:
		return BooleanValue{leftValue.Equals(rightValue)}
	case EqRefOp:
		// XXX: this is probably not 100% true. To make a 100% faithful
		//	implementation would require us to roll our own
		//	name table, which isn't a short-term todo item.
		return BooleanValue{leftValue == rightValue}
	}
	logErrf(ErrAssert, "unknown binary operator %s", n.String())
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

func (n FunctionCallNode) Eval(heap *StackHeap) Value {
	fn := n.function.Eval(heap)

	if fn == nil {
		// improve error message
		logErrf(ErrRuntime, "attempted to call an unknown function")
	}

	switch fnt := fn.(type) {
	case FunctionValue:
		argResults := make([]Value, len(n.arguments))
		for i, arg := range n.arguments {
			argResults[i] = arg.Eval(heap)
		}

		callHeap := &StackHeap{
			parent: fnt.parentHeap,
			vt:     ValueTable{},
		}
		for i, identNode := range fnt.defNode.arguments {
			if len(argResults) > i {
				callHeap.vt[identNode.val] = argResults[i]
			}
		}

		return fnt.defNode.body.Eval(callHeap)
	case NativeFunctionValue:
		// eval all arguments
		argResults := make([]Value, len(n.arguments))
		for i, arg := range n.arguments {
			argResults[i] = arg.Eval(heap)
		}
		return fnt.exec(argResults)
	default:
		logErrf(ErrRuntime, "attempted to call a non-function value %s",
			fnt.String())
		return NullValue{}
	}
}

func (n MatchClauseNode) String() string {
	return fmt.Sprintf("Clause (%s) -> (%s)",
		n.target.String(),
		n.expression.String())
}

func (n MatchClauseNode) Eval(heap *StackHeap) Value {
	logErrf(ErrAssert, "cannot Eval a MatchClauseNode")
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

func (n MatchExprNode) Eval(heap *StackHeap) Value {
	conditionVal := n.condition.Eval(heap)
	for _, cl := range n.clauses {
		if conditionVal.Equals(cl.target.Eval(heap)) {
			rv := cl.expression.Eval(heap)
			return rv
		}
	}
	return NullValue{}
}

func (n ExpressionListNode) String() string {
	if len(n.expressions) == 0 {
		return "Expression List ()"
	} else {
		exprs := n.expressions[0].String()
		for _, expr := range n.expressions[1:] {
			exprs += ", " + expr.String()
		}
		return fmt.Sprintf("Expression List (%s)", exprs)
	}
}

func (n ExpressionListNode) Eval(heap *StackHeap) Value {
	callHeap := &StackHeap{
		parent: heap,
		vt:     ValueTable{},
	}
	for _, expr := range n.expressions[:len(n.expressions)-1] {
		expr.Eval(callHeap)
	}
	return n.expressions[len(n.expressions)-1].Eval(callHeap)
}

func (n EmptyIdentifierNode) String() string {
	return "Empty Identifier"
}

func (n EmptyIdentifierNode) Eval(heap *StackHeap) Value {
	return EmptyValue{}
}

func (n IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

func (n IdentifierNode) Eval(heap *StackHeap) Value {
	val, prs := heap.getValue(n.val)
	if !prs {
		logErrf(ErrRuntime, "%s is not defined", n.val)
	}
	return val
}

func (n NumberLiteralNode) String() string {
	return fmt.Sprintf("Number %f", n.val)
}

func (n NumberLiteralNode) Eval(heap *StackHeap) Value {
	return NumberValue{n.val}
}

func (n StringLiteralNode) String() string {
	return fmt.Sprintf("String %s", n.val)
}

func (n StringLiteralNode) Eval(heap *StackHeap) Value {
	return StringValue{n.val}
}

func (n BooleanLiteralNode) String() string {
	return fmt.Sprintf("Boolean %t", n.val)
}

func (n BooleanLiteralNode) Eval(heap *StackHeap) Value {
	return BooleanValue{n.val}
}

func (n NullLiteralNode) String() string {
	return "Null"
}

func (n NullLiteralNode) Eval(heap *StackHeap) Value {
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

func (n ObjectLiteralNode) Eval(heap *StackHeap) Value {
	obj := CompositeValue{
		entries: make(ValueTable),
	}
	es := obj.entries
	for _, entry := range n.entries {
		k, ok := entry.key.(IdentifierNode)
		if ok {
			es[k.val] = entry.val.Eval(heap)
		} else {
			key := entry.key.Eval(heap)
			keyStrVal, sok := key.(StringValue)
			keyNumVal, nok := key.(NumberValue)
			if sok {
				es[keyStrVal.val] = entry.val.Eval(heap)
			} else if nok {
				es[fmt.Sprintf("%f", keyNumVal.val)] = entry.val.Eval(heap)
			} else {
				logErrf(ErrRuntime, "cannot access non-string property %s of object",
					key.String())
			}
		}
	}
	return obj
}

func (n ObjectEntryNode) String() string {
	return fmt.Sprintf("Object Entry (%s): (%s)", n.key.String(), n.val.String())
}

func (n ObjectEntryNode) Eval(heap *StackHeap) Value {
	logErrf(ErrAssert, "cannot Eval an ObjectEntryNode")
	return nil
}

func (n ListLiteralNode) String() string {
	if len(n.vals) == 0 {
		return fmt.Sprintf("List []")
	} else {
		vals := n.vals[0].String()
		for _, v := range n.vals[1:] {
			vals += ", " + v.String()
		}
		return fmt.Sprintf("List [%s]", vals)
	}
}

func (n ListLiteralNode) Eval(heap *StackHeap) Value {
	listVal := CompositeValue{
		entries: ValueTable{},
	}
	for i, n := range n.vals {
		listVal.entries[fmt.Sprintf("%f", float64(i))] = n.Eval(heap)
	}
	return listVal
}

func (n FunctionLiteralNode) String() string {
	if len(n.arguments) == 0 {
		return fmt.Sprintf("Function () => (%s)", n.body.String())
	} else {
		args := n.arguments[0].String()
		for _, a := range n.arguments[1:] {
			args += ", " + a.String()
		}
		return fmt.Sprintf("Function (%s) => (%s)", args, n.body.String())
	}
}

func (n FunctionLiteralNode) Eval(heap *StackHeap) Value {
	return FunctionValue{
		defNode:    n,
		parentHeap: heap,
	}
}

type ValueTable map[string]Value

type StackHeap struct {
	parent *StackHeap
	vt     ValueTable
}

func (sh *StackHeap) getValue(name string) (Value, bool) {
	val, ok := sh.vt[name]
	if ok {
		return val, true
	} else if sh.parent != nil {
		return sh.parent.getValue(name)
	} else {
		return NullValue{}, false
	}
}

func (sh *StackHeap) setValue(name string, val Value) {
	sh.vt[name] = val
}

func (sh *StackHeap) String() string {
	return fmt.Sprintf("heap: %s --prnt-> (%s)", sh.vt, sh.parent)
}

type Isolate struct {
	Heap *StackHeap
}

func (iso *Isolate) Dump() {
	logDebug("heap dump ->", iso.Heap.String())
}

func (iso *Isolate) Init() {
	if iso.Heap == nil {
		iso.Heap = &StackHeap{
			parent: nil,
			vt:     ValueTable{},
		}
	}
	iso.LoadEnvironment()
}

func (iso *Isolate) Eval(nodes <-chan Node, dumpHeap bool, done chan<- bool) {
	for node := range nodes {
		evalNode(iso.Heap, node)
	}
	if dumpHeap {
		iso.Dump()
	}

	done <- true
}

func (iso *Isolate) ExecInputStream(input <-chan rune, debugLex, debugParse, dump bool) func() {
	tokens := make(chan Tok)
	nodes := make(chan Node)
	done := make(chan bool, 3)

	go Tokenize(input, tokens, debugLex, done)
	go Parse(tokens, nodes, debugParse, done)
	go iso.Eval(nodes, dump, done)

	return func() {
		for i := 0; i < 3; i++ {
			<-done
		}
	}
}

func evalNode(heap *StackHeap, node Node) Value {
	switch n := node.(type) {
	case Node:
		return n.Eval(heap)
	default:
		logErrf(ErrAssert, "expected AST node during evaluation, got something else")
	}

	return nil
}
