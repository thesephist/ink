package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type NativeFunctionValue struct {
	name string
	exec func(*Context, []Value) (Value, error)
	ctx  *Context // runtime context to dispatch async errors
}

func (v NativeFunctionValue) String() string {
	return fmt.Sprintf("Native Function (%s)", v.name)
}

func (v NativeFunctionValue) Equals(other Value) bool {
	if ov, ok := other.(NativeFunctionValue); ok {
		return v.name == ov.name
	} else {
		return false
	}
}

func (ctx *Context) LoadEnvironment() {
	ctx.LoadFunc("load", inkLoad)

	ctx.LoadFunc("in", inkIn)
	ctx.LoadFunc("out", inkOut)
	ctx.LoadFunc("read", inkRead)
	ctx.LoadFunc("write", inkWrite)
	ctx.LoadFunc("delete", inkDelete)
	ctx.LoadFunc("listen", inkListen)
	ctx.LoadFunc("rand", inkRand)
	ctx.LoadFunc("time", inkTime)
	ctx.LoadFunc("wait", inkWait)

	ctx.LoadFunc("sin", inkSin)
	ctx.LoadFunc("cos", inkCos)
	ctx.LoadFunc("pow", inkPow)
	ctx.LoadFunc("ln", inkLn)
	ctx.LoadFunc("floor", inkFloor)

	ctx.LoadFunc("string", inkString)
	ctx.LoadFunc("number", inkNumber)
	ctx.LoadFunc("point", inkPoint)
	ctx.LoadFunc("char", inkChar)

	ctx.LoadFunc("type", inkType)
	ctx.LoadFunc("len", inkLen)
	ctx.LoadFunc("keys", inkKeys)

	// side effects
	rand.Seed(time.Now().UTC().UnixNano())
}

func (ctx *Context) LoadFunc(
	name string,
	exec func(*Context, []Value) (Value, error),
) {
	ctx.Frame.setValue(name, NativeFunctionValue{
		name,
		exec,
		ctx,
	})
}

func inkIn(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"in() takes one callback argument",
		}
	}

	ctx.ExecListener(func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			// XXX: currently reads after every newline / return
			//	but should ideally read every character input / keystroke
			//	that would also require stdlib/scan() to change.
			str, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			rv, err := evalInkFunction(in[0], false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"data"},
					"data": StringValue{str},
				},
			})
			if err != nil {
				ctx.ErrorStream <- Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to in()\n\t-> %s",
						err.Error()),
				}
				return
			}

			if boolValue, isBool := rv.(BooleanValue); isBool {
				if !boolValue.val {
					break
				}
			} else {
				ctx.ErrorStream <- Err{
					ErrRuntime,
					fmt.Sprintf("callback to in() should return a boolean, but got %s",
						rv.String()),
				}
				return
			}
		}

		_, err := evalInkFunction(in[0], false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"end"},
			},
		})
		if err != nil {
			ctx.ErrorStream <- Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to in()\n\t-> %s",
					err.Error()),
			}
			return
		}
	})

	return NullValue{}, nil
}

func inkLoad(ctx *Context, in []Value) (Value, error) {
	if len(in) == 1 {
		if path, ok := in[0].(StringValue); ok {
			inner := Context{}
			inner.Init()
			inner.DebugOpts = ctx.DebugOpts

			err := inner.ExecFile(path.val + ".ink")
			if err != nil {
				return NullValue{}, err
			}

			return CompositeValue{
				entries: inner.Frame.vt,
			}, nil
		}
	}

	return nil, Err{
		ErrRuntime,
		"load() takes one string argument",
	}
}

func inkOut(ctx *Context, in []Value) (Value, error) {
	if len(in) == 1 {
		if output, ok := in[0].(StringValue); ok {
			fmt.Printf(output.val)
			return NullValue{}, nil
		}
	}

	return nil, Err{
		ErrRuntime,
		"out() takes one string argument",
	}
}

func inkRead(ctx *Context, in []Value) (Value, error) {
	if len(in) != 4 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("read() expects four arguments: path, offset, length, and callback, but got %d",
				len(in)),
		}
	}

	path, isPathString := in[0].(StringValue)
	offset, isOffsetNumber := in[1].(NumberValue)
	length, isLengthNumber := in[2].(NumberValue)
	cb, isCbFunction := in[3].(FunctionValue)
	if !isPathString || !isOffsetNumber || !isLengthNumber || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in read()",
		}
	}

	sendErr := func(msg string) {
		evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type":    StringValue{"error"},
				"message": StringValue{msg},
			},
		})
	}

	ctx.ExecListener(func() {
		// open
		file, err := os.OpenFile(path.val, os.O_RDONLY, 0644)
		defer file.Close()
		if err != nil {
			sendErr(fmt.Sprintf(
				"error opening requested file in read(), %s", err.Error(),
			))
			return
		}

		// seek
		ofs := int64(offset.val)
		if ofs != 0 {
			_, err := file.Seek(ofs, 0) // 0 means relative to start of file
			if err != nil {
				sendErr(fmt.Sprintf(
					"error seeking requested file in read(), %s", err.Error(),
				))
				return
			}
		}

		// read
		buf := make([]byte, int64(length.val))
		count, err := file.Read(buf)
		if err != nil {
			sendErr(fmt.Sprintf(
				"error reading requested file in read(), %s", err.Error(),
			))
		}

		// marshal
		vt := ValueTable{}
		for i, b := range buf[:count] {
			vt[nToS(float64(i))] = NumberValue{float64(b)}
		}
		list := CompositeValue{entries: vt}

		// callback
		_, err = evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"data"},
				"data": list,
			},
		})
		if err != nil {
			ctx.ErrorStream <- Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to read()\n\t-> %s",
					err.Error()),
			}
			return
		}
	})

	return NullValue{}, nil
}

func inkWrite(ctx *Context, in []Value) (Value, error) {
	if len(in) != 4 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("write() expects four arguments: path, offset, length, and callback, but got %d",
				len(in)),
		}
	}

	path, isPathString := in[0].(StringValue)
	offset, isOffsetNumber := in[1].(NumberValue)
	encoded, isComposite := in[2].(CompositeValue)
	cb, isCbFunction := in[3].(FunctionValue)
	if !isPathString || !isOffsetNumber || !isComposite || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in write()",
		}
	}

	sendErr := func(msg string) {
		evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type":    StringValue{"error"},
				"message": StringValue{msg},
			},
		})
	}

	ctx.ExecListener(func() {
		// open
		var flag int
		if offset.val == -1 {
			// -1 offset is append
			flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
		} else {
			// all other offsets are writing
			flag = os.O_CREATE | os.O_WRONLY
		}
		file, err := os.OpenFile(path.val, flag, 0644)
		if err != nil {
			sendErr(fmt.Sprintf(
				"error opening requested file in write(), %s", err.Error(),
			))
			return
		}
		defer file.Close()

		// seek
		if offset.val != -1 {
			ofs := int64(offset.val)
			_, err := file.Seek(ofs, 0) // 0 means relative to start of file
			if err != nil {
				sendErr(fmt.Sprintf(
					"error seeking requested file in write(), %s", err.Error(),
				))
				return
			}
		}

		// unmarshal
		buf := make([]byte, getLength(encoded))
		for i, v := range encoded.entries {
			idx, err := strconv.Atoi(i)
			if err != nil {
				sendErr(fmt.Sprintf(
					"error unmarshaling data in write(), %s", err.Error(),
				))
			}

			if num, isNum := v.(NumberValue); isNum {
				buf[idx] = byte(num.val)
			} else {
				sendErr("error unmarshaling data in write(), byte value is not number")
			}
		}

		// write
		_, err = file.Write(buf)
		if err != nil {
			sendErr(fmt.Sprintf(
				"error writing to requested file in write(), %s", err.Error(),
			))
		}

		// callback
		_, err = evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"end"},
			},
		})
		if err != nil {
			ctx.ErrorStream <- Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to write()\n\t-> %s",
					err.Error()),
			}
			return
		}
	})

	return NullValue{}, nil
}

func inkDelete(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("delete() expects two arguments: path and callback, but got %d",
				len(in)),
		}
	}

	path, isPathString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)
	if !isPathString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in delete()",
		}
	}

	ctx.ExecListener(func() {
		// delete
		err := os.Remove(path.val)
		if err != nil {
			evalInkFunction(cb, false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"error"},
					"message": StringValue{
						fmt.Sprintf("error removing requested file in delete(), %s", err.Error()),
					},
				},
			})
			return
		}

		// callback
		_, err = evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"end"},
			},
		})
		if err != nil {
			ctx.ErrorStream <- Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to write()\n\t-> %s",
					err.Error()),
			}
			return
		}
	})

	return NullValue{}, nil
}

func inkListen(ctx *Context, in []Value) (Value, error) {
	return NullValue{}, nil
}

func inkRand(ctx *Context, in []Value) (Value, error) {
	return NumberValue{rand.Float64()}, nil
}

func inkTime(ctx *Context, in []Value) (Value, error) {
	unixSeconds := float64(time.Now().UnixNano()) / 1e9
	return NumberValue{unixSeconds}, nil
}

func inkWait(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			"wait() takes exactly two arguments",
		}
	}

	secs, isNum := in[0].(NumberValue)
	if !isNum {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("first argument to wait() should be a number, but got %s",
				in[0].String()),
		}
	}

	go func() {
		// This is a bit tricky, since we don't want wait() to hold the evalLock
		//	on the Context while we're waiting for the timeout, but do want to hold
		//	the main goroutine from completing with sync.WaitGroup.
		ctx.Listeners.Add(1)
		defer ctx.Listeners.Done()

		time.Sleep(time.Duration(
			int64(secs.val * float64(time.Second)),
		))

		ctx.ExecListener(func() {
			_, err := evalInkFunction(in[1], false)
			if err != nil {
				if e, isErr := err.(Err); isErr {
					ctx.ErrorStream <- e
				}
			}
		})
	}()

	return NullValue{}, nil
}

func inkSin(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"sin() takes exactly one number argument",
		}
	}
	inNum, isNum := in[0].(NumberValue)
	if !isNum {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("sin() takes a number argument, got %s", in[0].String()),
		}
	}

	return NumberValue{math.Sin(inNum.val)}, nil
}

func inkCos(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"cos() takes exactly one number argument",
		}
	}
	inNum, isNum := in[0].(NumberValue)
	if !isNum {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("cos() takes a number argument, got %s", in[0].String()),
		}
	}

	return NumberValue{math.Cos(inNum.val)}, nil
}

func inkPow(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			"pow() takes exactly 2 number arguments",
		}
	}

	base, baseIsNum := in[0].(NumberValue)
	exp, expIsNum := in[1].(NumberValue)
	if baseIsNum && expIsNum {
		if base.val == 0 && exp.val == 0 {
			return nil, Err{
				ErrRuntime,
				"math error, pow(0, 0) is not defined",
			}
		} else if base.val < 0 && !isIntable(exp.val) {
			return nil, Err{
				ErrRuntime,
				"math error, fractional power of negative number",
			}
		}
		return NumberValue{math.Pow(base.val, exp.val)}, nil
	} else {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("pow() takes exactly 2 number arguments, but got %s, %s",
				in[0].String(), in[1].String()),
		}
	}
}

func inkLn(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"ln() takes exactly one argument",
		}
	}

	n, isNumber := in[0].(NumberValue)
	if !isNumber {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("ln() takes exactly one number argument, but got %s",
				in[0].String()),
		}
	}

	if n.val <= 0 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("cannot take natural logarithm of non-positive number %s",
				nToS(n.val)),
		}
	}

	return NumberValue{math.Log(n.val)}, nil
}

func inkFloor(ctx *Context, in []Value) (Value, error) {
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

func inkString(ctx *Context, in []Value) (Value, error) {
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
	case CompositeValue:
		return StringValue{v.String()}, nil
	default:
		return StringValue{""}, nil
	}
}

func inkNumber(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"number() takes exactly one argument",
		}
	}

	switch v := in[0].(type) {
	case StringValue:
		f, err := strconv.ParseFloat(v.val, 64)
		if err != nil {
			return NullValue{}, nil
		} else {
			return NumberValue{f}, nil
		}
	case NumberValue:
		return v, nil
	case BooleanValue:
		if v.val {
			return NumberValue{1.0}, nil
		} else {
			return NumberValue{0.0}, nil
		}
	default:
		return NumberValue{0.0}, nil
	}
}

func inkPoint(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"point() takes exactly one argument",
		}
	}
	str, isString := in[0].(StringValue)
	if !isString {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("point() takes a string argument, got %s", in[0].String()),
		}
	}

	// Ink treats all characters as ASCII byte chars, and
	// 	transparently ignores unicode and surrogate pairs.
	return NumberValue{float64(str.val[0])}, nil
}

func inkChar(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"char() takes exactly one argument",
		}
	}
	cp, isNumber := in[0].(NumberValue)
	if !isNumber {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("char() takes a number argument, got %s", in[0].String()),
		}
	}

	// lol this type conversion disaster
	return StringValue{string(rune(int64(cp.val)))}, nil
}

func inkType(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"type() takes exactly one argument",
		}
	}

	rv := ""
	switch in[0].(type) {
	case StringValue:
		rv = "string"
	case NumberValue:
		rv = "number"
	case BooleanValue:
		rv = "boolean"
	case NullValue:
		rv = "()"
	case CompositeValue:
		rv = "composite"
	case FunctionValue:
		rv = "function"
	}

	return StringValue{rv}, nil
}

func getLength(list CompositeValue) int64 {
	// count up from 0 index until we find an index that doesn't
	//	contain a value.
	for idx := 0.0; ; idx++ {
		_, prs := list.entries[nToS(idx)]
		if !prs {
			return int64(idx)
		}
	}
}

func inkLen(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"len() takes exactly one argument",
		}
	}

	if list, isComposite := in[0].(CompositeValue); isComposite {
		return NumberValue{float64(getLength(list))}, nil
	} else if str, isString := in[0].(StringValue); isString {
		return NumberValue{float64(len(str.val))}, nil
	}

	return nil, Err{
		ErrRuntime,
		fmt.Sprintf("len() takes a string or composite value, but got %s",
			in[0].String()),
	}
}

func inkKeys(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"keys() takes exactly one argument",
		}
	}

	obj, isObj := in[0].(CompositeValue)
	if !isObj {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("keys() takes a composite value, but got %s",
				in[0].String()),
		}
	}

	vt := ValueTable{}

	var i float64 = 0
	for k, _ := range obj.entries {
		vt[nToS(i)] = StringValue{k}
		i++
	}
	return CompositeValue{vt}, nil
}
