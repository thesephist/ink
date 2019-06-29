package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type NativeFunctionValue struct {
	name string
	exec func([]Value) (Value, error)
}

func (v NativeFunctionValue) String() string {
	return fmt.Sprintf("Native Function (%s)", v.name)
}

func (v NativeFunctionValue) Equals(other Value) bool {
	ov, ok := other.(NativeFunctionValue)
	if ok {
		return v.name == ov.name
	} else {
		return false
	}
}

func (ctx *Context) LoadEnvironment() {
	ctx.LoadFunc(NativeFunctionValue{"in", inkIn})
	ctx.LoadFunc(NativeFunctionValue{"out", inkOut})
	ctx.LoadFunc(NativeFunctionValue{"read", inkRead})
	ctx.LoadFunc(NativeFunctionValue{"write", inkWrite})
	ctx.LoadFunc(NativeFunctionValue{"rand", inkRand})
	ctx.LoadFunc(NativeFunctionValue{"time", inkTime})

	ctx.LoadFunc(NativeFunctionValue{"sin", inkSin})
	ctx.LoadFunc(NativeFunctionValue{"cos", inkCos})
	ctx.LoadFunc(NativeFunctionValue{"pow", inkPow})
	ctx.LoadFunc(NativeFunctionValue{"ln", inkLn})
	ctx.LoadFunc(NativeFunctionValue{"floor", inkFloor})

	ctx.LoadFunc(NativeFunctionValue{"string", inkString})
	ctx.LoadFunc(NativeFunctionValue{"number", inkNumber})
	ctx.LoadFunc(NativeFunctionValue{"bytes", inkBytes})
	ctx.LoadFunc(NativeFunctionValue{"boolean", inkBoolean})

	ctx.LoadFunc(NativeFunctionValue{"len", inkLen})

	// side effects
	rand.Seed(time.Now().UTC().UnixNano())
}

func (ctx *Context) LoadFunc(nf NativeFunctionValue) {
	ctx.Frame.setValue(nf.name, nf)
}

func evalInkFunction(fn Value, args ...Value) (Value, error) {
	if fnt, isFunc := fn.(FunctionValue); isFunc {
		argValueTable := ValueTable{}
		for i, argNode := range fnt.defNode.arguments {
			if i < len(args) {
				if identNode, isIdent := argNode.(IdentifierNode); isIdent {
					argValueTable[identNode.val] = args[i]
				}
			}
		}

		callFrame := &StackFrame{
			parent: fnt.parentFrame,
			vt:     argValueTable,
		}
		rv, err := fnt.defNode.body.Eval(callFrame, false)
		if err != nil {
			return nil, err
		}
		return rv, nil
	} else if fnt, isNativeFunc := fn.(NativeFunctionValue); isNativeFunc {
		return fnt.exec(args)
	} else {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("attempted to call a non-function value %s", fn.String()),
		}
	}
}

func inkIn(in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"in() takes one callback argument",
		}
	}

	// TODO: implement as a character-by-character
	//	getter, since scan() in stdlib gets by line.
	_, err := evalInkFunction(in[0])
	if err != nil {
		return nil, err
	}

	return NullValue{}, nil
}

func inkOut(in []Value) (Value, error) {
	if len(in) == 1 {
		output, ok := in[0].(StringValue)
		if ok {
			fmt.Printf(output.val)
			return NullValue{}, nil
		}
	}

	return nil, Err{
		ErrRuntime,
		"out() takes one string argument",
	}
}

func inkRead(in []Value) (Value, error) {
	// TODO: once BufferValue gets written, write this
	return NullValue{}, nil
}

func inkWrite(in []Value) (Value, error) {
	// TODO: once BufferValue gets written, write this
	return NullValue{}, nil
}

func inkRand(in []Value) (Value, error) {
	return NumberValue{rand.Float64()}, nil
}

func inkTime(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkSin(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkCos(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkPow(in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			"pow() takes exactly 2 number arguments",
		}
	}

	base, baseIsNum := in[0].(NumberValue)
	exp, expIsNum := in[1].(NumberValue)
	if baseIsNum && expIsNum {
		return NumberValue{math.Pow(base.val, exp.val)}, nil
	} else {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("pow() takes exactly 2 number arguments, but got %s, %s",
				in[0].String(), in[1].String()),
		}
	}
}

func inkLn(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkFloor(in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"floor() takes exactly one argument",
		}
	}

	n, isNumber := in[0].(NumberValue)
	if !isNumber {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("floor() takes exactly one number argument, but got %s",
				in[0].String()),
		}
	}

	return NumberValue{math.Trunc(n.val)}, nil
}

func inkString(in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"string() takes exactly one argument",
		}
	}

	switch v := in[0].(type) {
	case StringValue:
		return v, nil
	case NumberValue:
		return StringValue{nToS(v.val)}, nil
	case BooleanValue:
		if v.val {
			return StringValue{"true"}, nil
		} else {
			return StringValue{"false"}, nil
		}
	case NullValue:
		return StringValue{"()"}, nil
	default:
		// TODO
		return NullValue{}, nil
	}
}

func inkNumber(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkBytes(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkBoolean(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkLen(in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"len() takes exactly one argument",
		}
	}

	list, isComposite := in[0].(CompositeValue)
	if !isComposite {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("len() takes a composite value, but got %s",
				in[0].String()),
		}
	}

	// count up from 0 index until we find an index that doesn't
	//	contain a value.
	for idx := 0.0; ; idx++ {
		_, prs := list.entries[nToS(idx)]
		if !prs {
			return NumberValue{idx}, nil
		}
	}
}
