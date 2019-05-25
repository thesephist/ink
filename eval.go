package main

const (
	NumberType = iota
	StringType
	BooleanType
	NullType
	CompositeType
	FunctionType
)

type Value struct {
	val  interface{}
	kind int
}

func Eval(node interface{}) {
	heap := make(map[string]Value)
	evalNode(&heap, node)
}

func evalNode(hp *map[string]Value, node interface{}) {
	switch node.(type) {
	default:
	}
}
