package main

import (
	"fmt"
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

	ctx.LoadFunc(NativeFunctionValue{"string", inkString})
	ctx.LoadFunc(NativeFunctionValue{"number", inkNumber})
	ctx.LoadFunc(NativeFunctionValue{"bytes", inkBytes})
	ctx.LoadFunc(NativeFunctionValue{"boolean", inkBoolean})
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

	// TODO: implement
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
	// TODO
	return NullValue{}, nil
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
	// TODO
	return NullValue{}, nil
}

func inkLn(in []Value) (Value, error) {
	// TODO
	return NullValue{}, nil
}

func inkString(in []Value) (Value, error) {
	if len(in) == 0 {
		// TODO: probably should use the language's native way of handling
		//	errors -- we just haven't decided on one yet.
		return nil, Err{
			ErrRuntime,
			"string() takes exactly one argument, none was provided",
		}
	}

	switch v := in[0].(type) {
	case StringValue:
		return v, nil
	case NumberValue:
		return StringValue{nToS(v.val)}, nil
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
