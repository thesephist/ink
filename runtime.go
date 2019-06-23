package main

import (
	"fmt"
	"log"
)

// The runtime defines any builtin functions and constants

type NativeFunctionValue struct {
	// TODO: get rid of this, we don't ever use it
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
	heap := iso.Heap

	heap.setValue("in", NativeFunctionValue{"in", inkIn})
	heap.setValue("out", NativeFunctionValue{"out", inkOut})
	heap.setValue("log", NativeFunctionValue{"log", inkLog})
	heap.setValue("read", NativeFunctionValue{"read", inkRead})
	heap.setValue("write", NativeFunctionValue{"write", inkWrite})
	heap.setValue("time", NativeFunctionValue{"time", inkTime})

	heap.setValue("sin", NativeFunctionValue{"sin", inkSin})
	heap.setValue("cos", NativeFunctionValue{"cos", inkCos})
	heap.setValue("ln", NativeFunctionValue{"ln", inkLn})

	heap.setValue("string", NativeFunctionValue{"string", inkString})
	heap.setValue("number", NativeFunctionValue{"number", inkNumber})
	heap.setValue("bytes", NativeFunctionValue{"bytes", inkBytes})
	heap.setValue("boolean", NativeFunctionValue{"boolean", inkBoolean})
}

func inkIn(_ []Value) Value {
	// TODO
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

func inkLog(in []Value) Value {
	rv := inkOut(in)
	fmt.Printf("\n")
	return rv
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
		return StringValue{val: fmt.Sprintf("%f", v.val)}
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
