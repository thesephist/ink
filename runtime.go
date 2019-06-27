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

func (iso *Isolate) LoadEnvironment() {
	iso.LoadFunc(NativeFunctionValue{"in", inkIn})
	iso.LoadFunc(NativeFunctionValue{"out", inkOut})
	iso.LoadFunc(NativeFunctionValue{"read", inkRead})
	iso.LoadFunc(NativeFunctionValue{"write", inkWrite})
	iso.LoadFunc(NativeFunctionValue{"time", inkTime})

	iso.LoadFunc(NativeFunctionValue{"sin", inkSin})
	iso.LoadFunc(NativeFunctionValue{"cos", inkCos})
	iso.LoadFunc(NativeFunctionValue{"ln", inkLn})

	iso.LoadFunc(NativeFunctionValue{"string", inkString})
	iso.LoadFunc(NativeFunctionValue{"number", inkNumber})
	iso.LoadFunc(NativeFunctionValue{"bytes", inkBytes})
	iso.LoadFunc(NativeFunctionValue{"boolean", inkBoolean})
}

func (iso *Isolate) LoadFunc(nf NativeFunctionValue) {
	iso.Frame.setValue(nf.name, nf)
}

func inkIn(_ []Value) (Value, error) {
	// TODO: probably take in one char of input at a time?
	//	this is probably unix/posix specific. hm.
	fmt.Println("Returning some input: TBD")
	return StringValue{"input"}, nil
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
		"out() takes one string parameter",
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
