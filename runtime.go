package main

import (
	"fmt"
)

// The runtime defines any builtin functions and constants

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
	heap := iso.Heap

	heap["in"] = NativeFunctionValue{"in", inkIn}
	heap["out"] = NativeFunctionValue{"out", inkOut}
	heap["read"] = NativeFunctionValue{"read", inkRead}
	heap["write"] = NativeFunctionValue{"write", inkWrite}
	heap["time"] = NativeFunctionValue{"time", inkTime}

	heap["sin"] = NativeFunctionValue{"sin", inkSin}
	heap["cos"] = NativeFunctionValue{"cos", inkCos}
	heap["ln"] = NativeFunctionValue{"ln", inkLn}

	heap["string"] = NativeFunctionValue{"string", inkString}
	heap["number"] = NativeFunctionValue{"number", inkString}
	heap["bytes"] = NativeFunctionValue{"bytes", inkString}
	heap["boolean"] = NativeFunctionValue{"boolean", inkString}
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
	// TODO
	return NullValue{}
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
