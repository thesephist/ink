package main

import (
	"fmt"
	"log"
)

type NativeFunctionValue struct {
	name string
	exec func([]Value) Value
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
	iso.Heap.setValue(nf.name, nf)
}

func inkIn(_ []Value) Value {
	// TODO: probably take in one char of input at a time?
	//	this is probably unix/posix specific. hm.
	fmt.Println("Returning some input: TBD")
	return StringValue{"input"}
}

func inkOut(in []Value) Value {
	if len(in) == 1 {
		output, ok := in[0].(StringValue)
		if ok {
			fmt.Printf(output.val)
			return NullValue{}
		}
	}

	fmt.Println("runtime error: out() takes one string parameter")
	return NullValue{}
}

func inkRead(in []Value) Value {
	// TODO: once BufferValue gets written, write this
	return NullValue{}
}

func inkWrite(in []Value) Value {
	// TODO: once BufferValue gets written, write this
	return NullValue{}
}

func inkTime(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkSin(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkCos(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkLn(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkString(in []Value) Value {
	if len(in) == 0 {
		log.Fatal("string() takes exactly one argument, none was provided")
	}

	switch v := in[0].(type) {
	case StringValue:
		return v
	case NumberValue:
		// XXX: not the most reliable check for int because of int64 range
		//	limitations, but works for now until we nail down Ink's number
		//	spec.
		if v.val == float64(int64(v.val)) {
			return StringValue{val: fmt.Sprintf("%d", int64(v.val))}
		} else {
			return StringValue{val: fmt.Sprintf("%f", v.val)}
		}
	default:
		// TODO
		return NullValue{}
	}
}

func inkNumber(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkBytes(in []Value) Value {
	// TODO
	return NullValue{}
}

func inkBoolean(in []Value) Value {
	// TODO
	return NullValue{}
}
