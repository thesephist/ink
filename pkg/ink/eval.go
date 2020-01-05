package ink

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

const maxPrintLen = 120

// Value represents any value in the Ink programming language.
// Each value corresponds to some primitive or object value created
// during the execution of an Ink program.
type Value interface {
	String() string
	// Equals reports whether the given value is deep-equal to the
	// receiving value. It does not compare references.
	Equals(Value) bool
}

func isIntable(n NumberValue) bool {
	// XXX: not the most reliable check for int because of int64 range
	// limitations, but works for now until we nail down Ink's number
	// spec more rigorously
	return n == NumberValue(int64(n))
}

// Utility func to get a consistent, language spec-compliant
// string representation of numbers
func nToS(f float64) string {
	if i := int64(f); f == float64(i) {
		return strconv.FormatInt(i, 10)
	}

	return strconv.FormatFloat(f, 'f', 8, 64)
}

// nToS for NumberValue type
func nvToS(v NumberValue) string {
	return nToS(float64(v))
}

// EmptyValue is the value of the empty identifier.
// it is globally unique and matches everything in equality.
type EmptyValue struct{}

func (v EmptyValue) String() string {
	return "_"
}

func (v EmptyValue) Equals(other Value) bool {
	return true
}

// NumberValue represents the number type (integer and floating point)
// in the Ink language.
type NumberValue float64

func (v NumberValue) String() string {
	return nvToS(v)
}

func (v NumberValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(NumberValue); ok {
		return v == ov
	}

	return false
}

// StringValue represents all characters and strings in Ink
type StringValue []byte

func (v StringValue) String() string {
	return "'" + strings.ReplaceAll(
		strings.ReplaceAll(string(v), "\\", "\\\\"),
		"'", "\\'") + "'"
}

func (v StringValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(StringValue); ok {
		return bytes.Equal(v, ov)
	}

	return false
}

// BooleanValue is either `true` or `false`
type BooleanValue bool

func (v BooleanValue) String() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}

func (v BooleanValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(BooleanValue); ok {
		return v == ov
	}

	return false
}

// NullValue is a value that only exists at the type level,
// and is represented by the empty expression list `()`.
type NullValue struct{}

func (v NullValue) String() string {
	return "()"
}

func (v NullValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	_, ok := other.(NullValue)
	return ok
}

// CompositeValue includes all objects and list values
type CompositeValue ValueTable

func (v CompositeValue) String() string {
	entries := make([]string, 0, len(v))
	for key, val := range v {
		entries = append(entries, fmt.Sprintf("%s: %s", key, val))
	}
	return "{" + strings.Join(entries, ", ") + "}"
}

func (v CompositeValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(CompositeValue); ok {
		if len(v) != len(ov) {
			return false
		}

		for key, val := range v {
			otherVal, prs := ov[key]
			if !prs || prs && !val.Equals(otherVal) {
				return false
			}
		}
		return true
	}

	return false
}

// FunctionValue is the value of any variables referencing functions
// defined in an Ink program.
type FunctionValue struct {
	defn        *FunctionLiteralNode
	parentFrame *StackFrame
}

func (v FunctionValue) String() string {
	// ellipsize function body at a reasonable length,
	// so as not to be too verbose in repl environments
	fstr := v.defn.String()
	if len(fstr) > maxPrintLen {
		fstr = fstr[:maxPrintLen] + ".."
	}
	return fstr
}

func (v FunctionValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(FunctionValue); ok {
		// to compare structs containing slices, we really want
		// a pointer comparison, not a value comparison
		return v.defn == ov.defn
	}

	return false
}

// FunctionCallThunkValue is an internal representation of a lazy
// function evaluation used to implement tail call optimization.
type FunctionCallThunkValue struct {
	vt       ValueTable
	function FunctionValue
}

func (v FunctionCallThunkValue) String() string {
	return fmt.Sprintf("Tail Call Thunk of (%s)", v.function)
}

func (v FunctionCallThunkValue) Equals(other Value) bool {
	if _, isEmpty := other.(EmptyValue); isEmpty {
		return true
	}

	if ov, ok := other.(FunctionCallThunkValue); ok {
		// to compare structs containing slices, we really want
		// a pointer comparison, not a value comparison
		return &v.vt == &ov.vt && &v.function == &ov.function
	}

	return false
}

// unwrapThunk expands out a recursive structure of thunks
// 	into a flat for loop control structure
func unwrapThunk(thunk FunctionCallThunkValue) (v Value, err error) {
	isThunk := true
	for isThunk {
		frame := &StackFrame{
			parent: thunk.function.parentFrame,
			vt:     thunk.vt,
		}
		v, err = thunk.function.defn.body.Eval(frame, true)
		if err != nil {
			return
		}
		thunk, isThunk = v.(FunctionCallThunkValue)
	}

	return
}

func (n UnaryExprNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	switch n.operator {
	case NegationOp:
		operand, err := n.operand.Eval(frame, false)
		if err != nil {
			return nil, err
		}

		switch o := operand.(type) {
		case NumberValue:
			return -o, nil
		case BooleanValue:
			return BooleanValue(!o), nil
		default:
			return nil, Err{
				ErrRuntime,
				fmt.Sprintf("cannot negate non-boolean and non-number value %s [%s]",
					o, poss(n.operand)),
			}
		}
	}

	LogErrf(ErrAssert, "unrecognized unary operator %s", n)
	return nil, nil
}

func operandToStringKey(rightOperand Node, frame *StackFrame) (string, error) {
	switch right := rightOperand.(type) {
	case IdentifierNode:
		return right.val, nil

	case StringLiteralNode:
		return right.val, nil

	case NumberLiteralNode:
		return nToS(right.val), nil

	default:
		rightEvaluatedValue, err := rightOperand.Eval(frame, false)
		if err != nil {
			return "", err
		}

		switch rv := rightEvaluatedValue.(type) {
		case StringValue:
			return string(rv), nil
		case NumberValue:
			return nvToS(rv), nil
		default:
			return "", Err{
				ErrRuntime,
				fmt.Sprintf("cannot access invalid property name %s of a composite value [%s]",
					rightEvaluatedValue, poss(rightOperand)),
			}
		}
	}
}

func (n BinaryExprNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	if n.operator == DefineOp {
		if leftIdent, okIdent := n.leftOperand.(IdentifierNode); okIdent {
			if _, isEmpty := n.rightOperand.(EmptyIdentifierNode); isEmpty {
				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot assign an empty identifier value to %s [%s]",
						leftIdent, poss(n.leftOperand)),
				}
			}

			rightValue, err := n.rightOperand.Eval(frame, false)
			if err != nil {
				return nil, err
			}

			frame.Set(leftIdent.val, rightValue)
			return rightValue, nil
		} else if leftAccess, okAccess := n.leftOperand.(BinaryExprNode); okAccess &&
			leftAccess.operator == AccessorOp {

			leftValue, err := leftAccess.leftOperand.Eval(frame, false)
			if err != nil {
				return nil, err
			}

			leftKey, err := operandToStringKey(leftAccess.rightOperand, frame)
			if err != nil {
				return nil, err
			}

			if leftValueComposite, isComposite := leftValue.(CompositeValue); isComposite {
				rightValue, err := n.rightOperand.Eval(frame, false)
				if err != nil {
					return nil, err
				}

				leftValueComposite[leftKey] = rightValue
				return leftValueComposite, nil
			} else if leftString, isString := leftValue.(StringValue); isString {
				leftIdent, isLeftIdent := leftAccess.leftOperand.(IdentifierNode)
				if !isLeftIdent {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("cannot set string %s at index because string is not an identifier",
							leftString),
					}
				}

				rightValue, err := n.rightOperand.Eval(frame, false)
				if err != nil {
					return nil, err
				}

				rightString, isString := rightValue.(StringValue)
				if !isString {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("cannot set part of string to a non-character %s", rightValue),
					}
				}

				rightNum, err := strconv.ParseInt(leftKey, 10, 64)
				if err != nil {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("while accessing string %s at an index, found non-integer index %s [%s]",
							leftString, leftKey, poss(leftAccess.rightOperand)),
					}
				}

				rn := int(rightNum)
				if -1 < rn && rn < len(leftString) {
					for i, r := range rightString {
						if rn+i < len(leftString) {
							leftString[rn+i] = r
						} else {
							leftString = append(leftString, r)
						}
					}
					frame.Up(leftIdent.val, leftString)
					return leftString, nil
				} else if rn == len(leftString) {
					leftString = append(leftString, rightString...)
					frame.Up(leftIdent.val, leftString)
					return leftString, nil
				} else {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("tried to modify string %s at out of bounds index %s [%s]",
							leftString, leftKey, poss(leftAccess.rightOperand)),
					}
				}
			} else {
				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot set property of a non-composite value %s [%s]",
						leftValue, poss(leftAccess.leftOperand)),
				}
			}
		} else {
			left, err := n.leftOperand.Eval(frame, false)
			if err != nil {
				return nil, err
			}

			return nil, Err{
				ErrRuntime,
				fmt.Sprintf("cannot assign value to non-identifier %s [%s]",
					left, poss(n.leftOperand)),
			}
		}
	} else if n.operator == AccessorOp {
		leftValue, err := n.leftOperand.Eval(frame, false)
		if err != nil {
			return nil, err
		}

		rightValueStr, err := operandToStringKey(n.rightOperand, frame)
		if err != nil {
			return nil, err
		}

		if leftValueComposite, isComposite := leftValue.(CompositeValue); isComposite {
			v, prs := leftValueComposite[rightValueStr]
			if prs {
				return v, nil
			}

			return NullValue{}, nil
		} else if leftString, isString := leftValue.(StringValue); isString {
			rightNum, err := strconv.ParseInt(rightValueStr, 10, 64)
			if err != nil {
				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("while accessing string %s at an index, found non-integer index %s [%s]",
						leftString, rightValueStr, poss(n.rightOperand)),
				}
			}

			rn := int(rightNum)
			if -1 < rn && rn < len(leftString) {
				return StringValue([]byte{leftString[rn]}), nil
			}

			return NullValue{}, nil
		} else {
			return nil, Err{
				ErrRuntime,
				fmt.Sprintf("cannot access property %s of a non-composite value %s [%s]",
					n.rightOperand, leftValue, poss(n.rightOperand)),
			}
		}
	}

	leftValue, err := n.leftOperand.Eval(frame, false)
	if err != nil {
		return nil, err
	}
	rightValue, err := n.rightOperand.Eval(frame, false)
	if err != nil {
		return nil, err
	}

	switch n.operator {
	case AddOp:
		switch left := leftValue.(type) {
		case NumberValue:
			if right, ok := rightValue.(NumberValue); ok {
				return left + right, nil
			}
		case StringValue:
			if right, ok := rightValue.(StringValue); ok {
				// In this context, strings are immutable. i.e. concatenating
				// strings should produce a completely new string whose modifications
				// won't be observable by the original strings.
				base := make([]byte, 0, len(left)+len(right))
				base = append(base, left...)
				return StringValue(append(base, right...)), nil
			}
		case BooleanValue:
			if right, ok := rightValue.(BooleanValue); ok {
				return BooleanValue(left || right), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support addition [%s]",
				leftValue, rightValue, poss(n)),
		}
	case SubtractOp:
		switch left := leftValue.(type) {
		case NumberValue:
			if right, ok := rightValue.(NumberValue); ok {
				return left - right, nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support subtraction [%s]",
				leftValue, rightValue, poss(n)),
		}
	case MultiplyOp:
		switch left := leftValue.(type) {
		case NumberValue:
			if right, ok := rightValue.(NumberValue); ok {
				return left * right, nil
			}
		case BooleanValue:
			if right, ok := rightValue.(BooleanValue); ok {
				return BooleanValue(left && right), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support multiplication [%s]",
				leftValue, rightValue, poss(n)),
		}
	case DivideOp:
		if leftNum, isNum := leftValue.(NumberValue); isNum {
			if right, ok := rightValue.(NumberValue); ok {
				if right == 0 {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("division by zero error [%s]", poss(n.rightOperand)),
					}
				}

				return leftNum / right, nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support division [%s]",
				leftValue, rightValue, poss(n)),
		}
	case ModulusOp:
		if leftNum, isNum := leftValue.(NumberValue); isNum {
			if right, ok := rightValue.(NumberValue); ok {
				if right == 0 {
					return nil, Err{
						ErrRuntime,
						fmt.Sprintf("division by zero error in modulus [%s]", poss(n.rightOperand)),
					}
				}

				if isIntable(right) {
					return NumberValue(int(leftNum) % int(right)), nil
				}

				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot take modulus of non-integer value %s [%s]",
						nvToS(right), poss(n.leftOperand)),
				}
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support modulus [%s]",
				leftValue, rightValue, poss(n)),
		}
	case LogicalAndOp:
		if leftNum, isNum := leftValue.(NumberValue); isNum {
			if rightNum, ok := rightValue.(NumberValue); ok {
				if isIntable(leftNum) && isIntable(rightNum) {
					return NumberValue(int64(leftNum) & int64(rightNum)), nil
				}

				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot take bitwise & of non-integer values %s, %s [%s]",
						nvToS(rightNum), nvToS(leftNum), poss(n)),
				}
			}
		} else if leftBool, isBool := leftValue.(BooleanValue); isBool {
			if rightBool, ok := rightValue.(BooleanValue); ok {
				return BooleanValue(leftBool && rightBool), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support bitwise or logical & [%s]",
				leftValue, rightValue, poss(n)),
		}
	case LogicalOrOp:
		if leftNum, isNum := leftValue.(NumberValue); isNum {
			if rightNum, ok := rightValue.(NumberValue); ok {
				if isIntable(leftNum) && isIntable(rightNum) {
					return NumberValue(int64(leftNum) | int64(rightNum)), nil
				}

				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot take bitwise | of non-integer values %s, %s [%s]",
						nvToS(rightNum), nvToS(leftNum), poss(n)),
				}
			}
		} else if leftBool, isBool := leftValue.(BooleanValue); isBool {
			if rightBool, ok := rightValue.(BooleanValue); ok {
				return BooleanValue(leftBool || rightBool), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support bitwise or logical | [%s]",
				leftValue, rightValue, poss(n)),
		}
	case LogicalXorOp:
		if leftNum, isNum := leftValue.(NumberValue); isNum {
			if rightNum, ok := rightValue.(NumberValue); ok {
				if isIntable(leftNum) && isIntable(rightNum) {
					return NumberValue(int64(leftNum) ^ int64(rightNum)), nil
				}

				return nil, Err{
					ErrRuntime,
					fmt.Sprintf("cannot take logical & of non-integer values %s, %s [%s]",
						nvToS(rightNum), nvToS(leftNum), poss(n)),
				}
			}
		} else if leftBool, isBool := leftValue.(BooleanValue); isBool {
			if rightBool, ok := rightValue.(BooleanValue); ok {
				return BooleanValue(leftBool != rightBool), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support bitwise or logical ^ [%s]",
				leftValue, rightValue, poss(n)),
		}
	case GreaterThanOp:
		switch left := leftValue.(type) {
		case NumberValue:
			if right, ok := rightValue.(NumberValue); ok {
				return BooleanValue(left > right), nil
			}
		case StringValue:
			if right, ok := rightValue.(StringValue); ok {
				return BooleanValue(bytes.Compare(left, right) > 0), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support comparison [%s]",
				leftValue, rightValue, poss(n)),
		}
	case LessThanOp:
		switch left := leftValue.(type) {
		case NumberValue:
			if right, ok := rightValue.(NumberValue); ok {
				return BooleanValue(left < right), nil
			}
		case StringValue:
			if right, ok := rightValue.(StringValue); ok {
				return BooleanValue(bytes.Compare(left, right) < 0), nil
			}
		}
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("values %s and %s do not support comparison [%s]",
				leftValue, rightValue, poss(n)),
		}
	case EqualOp:
		return BooleanValue(leftValue.Equals(rightValue)), nil
	}

	LogErrf(ErrAssert, "unknown binary operator %s", n.String())
	return nil, err
}

func (n FunctionCallNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	fn, err := n.function.Eval(frame, false)
	if err != nil {
		return nil, err
	}

	argResults := make([]Value, len(n.arguments))
	for i, arg := range n.arguments {
		argResults[i], err = arg.Eval(frame, false)
		if err != nil {
			return nil, err
		}
	}
	return evalInkFunction(fn, allowThunk, argResults...)
}

// call into an Ink callback function synchronously
func evalInkFunction(fn Value, allowThunk bool, args ...Value) (Value, error) {
	if fnt, isFunc := fn.(FunctionValue); isFunc {
		argValueTable := ValueTable{}
		for i, argNode := range fnt.defn.arguments {
			if i < len(args) {
				if identNode, isIdent := argNode.(IdentifierNode); isIdent {
					argValueTable[identNode.val] = args[i]
				}
			}
		}

		// TCO: used for evaluating expressions that may be in tail positions
		// at the end of Nodes whose evaluation allocates another StackFrame
		// like ExpressionList and FunctionLiteral's body
		returnThunk := FunctionCallThunkValue{
			vt:       argValueTable,
			function: fnt,
		}

		if allowThunk {
			return returnThunk, nil
		}
		return unwrapThunk(returnThunk)
	} else if fnt, isNativeFunc := fn.(NativeFunctionValue); isNativeFunc {
		return fnt.exec(fnt.ctx, args)
	} else {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("attempted to call a non-function value %s", fn),
		}
	}
}

func (n MatchClauseNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	LogErrf(ErrAssert, "cannot Eval a MatchClauseNode")
	return nil, nil
}

func (n MatchExprNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	conditionVal, err := n.condition.Eval(frame, false)
	if err != nil {
		return nil, err
	}

	for _, cl := range n.clauses {
		targetVal, err := cl.target.Eval(frame, false)
		if err != nil {
			return nil, err
		}

		if conditionVal.Equals(targetVal) {
			rv, err := cl.expression.Eval(frame, allowThunk)
			if err != nil {
				return nil, err
			}

			// match expression clauses are tail call optimized,
			// so return a maybe ThunkValue
			return rv, nil
		}
	}

	return NullValue{}, nil
}

func (n ExpressionListNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	length := len(n.expressions)

	if length == 0 {
		return NullValue{}, nil
	}

	callFrame := &StackFrame{
		parent: frame,
		vt:     ValueTable{},
	}
	for _, expr := range n.expressions[:length-1] {
		_, err := expr.Eval(callFrame, false)
		if err != nil {
			return nil, err
		}
	}

	// return values of expression lists are tail call optimized,
	// so return a maybe ThunkValue
	return n.expressions[length-1].Eval(callFrame, allowThunk)
}

func (n EmptyIdentifierNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	return EmptyValue{}, nil
}

func (n IdentifierNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	val, prs := frame.Get(n.val)
	if !prs {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("%s is not defined [%s]", n.val, poss(n)),
		}
	}
	return val, nil
}

func (n NumberLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	return NumberValue(n.val), nil
}

func (n StringLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	return StringValue(n.val), nil
}

func (n BooleanLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	return BooleanValue(n.val), nil
}

func (n ObjectLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	obj := CompositeValue{}
	for _, entry := range n.entries {
		keyStr, err := operandToStringKey(entry.key, frame)
		if err != nil {
			return nil, err
		}

		obj[keyStr], err = entry.val.Eval(frame, false)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func (n ObjectEntryNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	LogErrf(ErrAssert, "cannot Eval an ObjectEntryNode")
	return nil, nil
}

func (n ListLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	listVal := CompositeValue{}
	for i, n := range n.vals {
		var err error
		listVal[strconv.Itoa(i)], err = n.Eval(frame, false)
		if err != nil {
			return nil, err
		}
	}

	return listVal, nil
}

func (n FunctionLiteralNode) Eval(frame *StackFrame, allowThunk bool) (Value, error) {
	return FunctionValue{
		defn:        &n,
		parentFrame: frame,
	}, nil
}

// ValueTable is used anytime a map of names/labels to Ink Values is needed,
// and is notably used to represent stack frames / heaps and CompositeValue dictionaries.
type ValueTable map[string]Value

// StackFrame represents the heap of variables local to a particular function call frame,
// and recursively references other parent StackFrames internally.
type StackFrame struct {
	parent *StackFrame
	vt     ValueTable
}

// Get a value from the stack frame chain
func (frame *StackFrame) Get(name string) (Value, bool) {
	for frame != nil {
		val, ok := frame.vt[name]
		if ok {
			return val, true
		}

		frame = frame.parent
	}

	return NullValue{}, false
}

// Set a value to the most recent call stack frame
func (frame *StackFrame) Set(name string, val Value) {
	frame.vt[name] = val
}

// Up updates a value in the stack frame chain
func (frame *StackFrame) Up(name string, val Value) {
	for frame != nil {
		_, ok := frame.vt[name]
		if ok {
			frame.vt[name] = val
			return
		}

		frame = frame.parent
	}

	LogErrf(
		ErrAssert,
		fmt.Sprintf("StackFrame.Up expected to find variable '%s' in frame but did not",
			name),
	)
}

func (frame *StackFrame) String() string {
	entries := make([]string, 0, len(frame.vt))
	for k, v := range frame.vt {
		vstr := v.String()
		if len(vstr) > maxPrintLen {
			vstr = vstr[:maxPrintLen] + ".."
		}
		entries = append(entries, fmt.Sprintf("%s -> %s", k, vstr))
	}

	return fmt.Sprintf("{\n\t%s\n} -prnt-> %s", strings.Join(entries, "\n\t"), frame.parent)
}

// Engine is a single global context of Ink program execution.
//
// A single thread of execution may run within an Engine at any given moment,
// and this is ensured by an internal execution lock. An execution's Engine
// also holds all permission and debugging flags.
//
// Within an Engine, there may exist multiple Contexts that each contain different
// execution environments, running concurrently under a single lock.
type Engine struct {
	// Listeners keeps track of the concurrent threads of execution running
	// in the Engine. Call `Engine.Listeners.Wait()` to block until all concurrent
	// execution threads finish on an Engine.
	Listeners sync.WaitGroup

	// If FatalError is true, an error will halt the interpreter
	FatalError  bool
	Permissions PermissionsConfig
	Debug       DebugConfig

	// Ink de-duplicates imported source files here, where
	// Contexts from imports are deduplicated keyed by the
	// canonicalized import path. This prevents recursive
	// imports from crashing the interpreter and allows other
	// nice functionality.
	Contexts map[string]*Context

	// Only a single function may write to the stack frames
	// at any moment.
	evalLock sync.Mutex
}

// CreateContext creates and initializes a new Context tied to a given Engine.
func (eng *Engine) CreateContext() *Context {
	ctx := &Context{
		Engine: eng,
		Frame: &StackFrame{
			parent: nil,
			vt:     ValueTable{},
		},
	}

	// If first time creating Context in this Engine,
	// initialize the Contexts map
	if eng.Contexts == nil {
		eng.Contexts = make(map[string]*Context)
	}

	ctx.resetWd()
	ctx.LoadEnvironment()

	return ctx
}

// Context represents a single, isolated execution context with its global heap,
// imports, call stack, and cwd (working directory).
type Context struct {
	// Cwd is an always-absolute path to current working dir (of module system)
	Cwd string
	// currently executing file's path, if any
	File   string
	Engine *Engine
	// Frame represents the Context's global heap
	Frame *StackFrame
}

// LogErr logs an Err (interpreter error) according to the configurations
// specified in the Context's Engine.
func (ctx *Context) LogErr(e Err) {
	msg := e.message
	if ctx.File != "" {
		msg = e.message + " in " + ctx.File
	}

	if ctx.Engine.FatalError {
		LogErr(e.reason, msg)
	} else {
		LogSafeErr(e.reason, msg)
	}
}

// PermissionsConfig defines Context's permissions to
// operating system interfaces
type PermissionsConfig struct {
	Read  bool
	Write bool
	Net   bool
	Exec  bool
}

// DebugConfig defines any debugging flags referenced at runtime
type DebugConfig struct {
	Lex   bool
	Parse bool
	Dump  bool
}

// Dump prints the current state of the Context's global heap
func (ctx *Context) Dump() {
	LogDebug("frame dump", ctx.Frame.String())
}

func (ctx *Context) resetWd() {
	var err error
	ctx.Cwd, err = os.Getwd()
	if err != nil {
		LogErrf(
			ErrSystem,
			"could not identify current working directory\n\t-> %s", err,
		)
	}
}

// Eval takes a channel of Nodes to evaluate, and executes the Ink programs defined
// in the syntax tree. Eval returns the last value of the last expression in the AST,
// or an error if there was a runtime error.
func (ctx *Context) Eval(nodes <-chan Node, dumpFrame bool) (val Value, err error) {
	ctx.Engine.evalLock.Lock()
	defer ctx.Engine.evalLock.Unlock()

	for node := range nodes {
		val, err = node.Eval(ctx.Frame, false)
		if err != nil {
			if e, isErr := err.(Err); isErr {
				ctx.LogErr(e)
			}
			break
		}
	}

	if dumpFrame {
		ctx.Dump()
	}

	return
}

// ExecListener queues an asynchronous callback task to the Engine behind the Context.
// Callbacks registered this way will also run with the Engine's execution lock.
func (ctx *Context) ExecListener(callback func()) {
	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		ctx.Engine.evalLock.Lock()
		defer ctx.Engine.evalLock.Unlock()

		callback()
	}()
}

// Exec runs an Ink program defined by an io.Reader.
// This is the main way to invoke Ink programs from Go.
// Exec blocks until the Ink program exits.
func (ctx *Context) Exec(input io.Reader) (Value, error) {
	eng := ctx.Engine

	tokens := make(chan Tok)
	nodes := make(chan Node)
	go Tokenize(input, tokens, eng.FatalError, eng.Debug.Lex)
	go Parse(tokens, nodes, eng.FatalError, eng.Debug.Parse)

	return ctx.Eval(nodes, eng.Debug.Dump)
}

// ExecPath is a convenience function to Exec() a program file in a given Context.
func (ctx *Context) ExecPath(filePath string) {
	if !path.IsAbs(filePath) {
		LogErrf(
			ErrAssert,
			"Context.ExecPath expected an absolute path, got something else",
		)
	}

	// update Cwd for any potential load() calls this file will make
	ctx.Cwd = path.Dir(filePath)
	ctx.File = filePath

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		LogSafeErr(
			ErrSystem,
			fmt.Sprintf("could not open %s for execution:\n\t-> %s", filePath, err),
		)
	}

	ctx.Exec(file)
}
