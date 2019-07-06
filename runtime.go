package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// NativeFunctionValue represents a function whose implementation is written
//	in Go and built-into the runtime.
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

// LoadEnvironment loads all builtins (functions and constants) to a given Context.
func (ctx *Context) LoadEnvironment() {
	ctx.LoadFunc("load", inkLoad)

	// system interfaces
	ctx.LoadFunc("in", inkIn)
	ctx.LoadFunc("out", inkOut)
	ctx.LoadFunc("read", inkRead)
	ctx.LoadFunc("write", inkWrite)
	ctx.LoadFunc("delete", inkDelete)
	ctx.LoadFunc("listen", inkListen)
	ctx.LoadFunc("rand", inkRand)
	ctx.LoadFunc("time", inkTime)
	ctx.LoadFunc("wait", inkWait)

	// math
	ctx.LoadFunc("sin", inkSin)
	ctx.LoadFunc("cos", inkCos)
	ctx.LoadFunc("pow", inkPow)
	ctx.LoadFunc("ln", inkLn)
	ctx.LoadFunc("floor", inkFloor)

	// type conversions
	ctx.LoadFunc("string", inkString)
	ctx.LoadFunc("number", inkNumber)
	ctx.LoadFunc("point", inkPoint)
	ctx.LoadFunc("char", inkChar)

	// introspection
	ctx.LoadFunc("type", inkType)
	ctx.LoadFunc("len", inkLen)
	ctx.LoadFunc("keys", inkKeys)

	// side effects
	rand.Seed(time.Now().UTC().UnixNano())
}

// LoadFunc loads a single Go-implemented function into a Context.
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

func inkLoad(ctx *Context, in []Value) (Value, error) {
	if len(in) == 1 {
		if givenPath, ok := in[0].(StringValue); ok && givenPath.val != "" {
			importPath := path.Join(ctx.Cwd, givenPath.val+".ink")

			// swap out fields
			childCtx := ctx.Engine.CreateContext()

			ctx.Engine.evalLock.Unlock()
			defer ctx.Engine.evalLock.Lock()

			err := childCtx.ExecFile(importPath)
			// Lock() blocks until file is eval'd
			if err != nil {
				return nil, Err{
					ErrSystem,
					fmt.Sprintf("error while executing file, %s", err.Error()),
				}
			}

			return CompositeValue{
				entries: childCtx.Frame.vt,
			}, nil
		}
	}

	return nil, Err{
		ErrRuntime,
		"load() takes one string argument, without the .ink suffix",
	}
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
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to in()\n\t-> %s",
						err.Error()),
				})
				return
			}

			if boolValue, isBool := rv.(BooleanValue); isBool {
				if !boolValue.val {
					break
				}
			} else {
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("callback to in() should return a boolean, but got %s",
						rv.String()),
				})
				return
			}
		}

		_, err := evalInkFunction(in[0], false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"end"},
			},
		})
		if err != nil {
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to in()\n\t-> %s",
					err.Error()),
			})
			return
		}
	})

	return NullValue{}, nil
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
			fmt.Sprintf("read() expects four arguments: filePath, offset, length, and callback, but got %d",
				len(in)),
		}
	}

	filePath, isFilePathString := in[0].(StringValue)
	offset, isOffsetNumber := in[1].(NumberValue)
	length, isLengthNumber := in[2].(NumberValue)
	cb, isCbFunction := in[3].(FunctionValue)
	if !isFilePathString || !isOffsetNumber || !isLengthNumber || !isCbFunction {
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
		// short-circuit out if no read permission
		if !ctx.Engine.Permissions.Read {
			_, err := evalInkFunction(cb, false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"data"},
					"data": CompositeValue{
						entries: ValueTable{},
					},
				},
			})
			if err != nil {
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to read()\n\t-> %s",
						err.Error()),
				})
			}
			return
		}

		// open
		file, err := os.OpenFile(filePath.val, os.O_RDONLY, 0644)
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
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to read()\n\t-> %s",
					err.Error()),
			})
			return
		}
	})

	return NullValue{}, nil
}

func inkWrite(ctx *Context, in []Value) (Value, error) {
	if len(in) != 4 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("write() expects four arguments: filePath, offset, length, and callback, but got %d",
				len(in)),
		}
	}

	filePath, isFilePathString := in[0].(StringValue)
	offset, isOffsetNumber := in[1].(NumberValue)
	encoded, isComposite := in[2].(CompositeValue)
	cb, isCbFunction := in[3].(FunctionValue)
	if !isFilePathString || !isOffsetNumber || !isComposite || !isCbFunction {
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
		if !ctx.Engine.Permissions.Write {
			_, err := evalInkFunction(cb, false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"end"},
				},
			})
			if err != nil {
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to write()\n\t-> %s",
						err.Error()),
				})
			}
			return
		}

		// open
		var flag int
		if offset.val == -1 {
			// -1 offset is append
			flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
		} else {
			// all other offsets are writing
			flag = os.O_CREATE | os.O_WRONLY
		}
		file, err := os.OpenFile(filePath.val, flag, 0644)
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
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to write()\n\t-> %s",
					err.Error()),
			})
			return
		}
	})

	return NullValue{}, nil
}

func inkDelete(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("delete() expects two arguments: filePath and callback, but got %d",
				len(in)),
		}
	}

	filePath, isFilePathString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)
	if !isFilePathString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in delete()",
		}
	}

	ctx.ExecListener(func() {
		if !ctx.Engine.Permissions.Write {
			_, err := evalInkFunction(cb, false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"end"},
				},
			})
			if err != nil {
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to delete()\n\t-> %s",
						err.Error()),
				})
			}
			return
		}

		// delete
		err := os.Remove(filePath.val)
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
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to delete()\n\t-> %s",
					err.Error()),
			})
			return
		}
	})

	return NullValue{}, nil
}

// inkHTTPHandler fulfills the Handler interface for inkListen() to work
type inkHTTPHandler struct {
	ctx         *Context
	inkCallback FunctionValue
}

func (h inkHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := h.ctx
	cb := h.inkCallback

	callbackErr := func(err error) {
		ctx.Engine.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("error in callback to listen()\n\t-> %s",
				err.Error()),
		})
		return
	}

	// This is a bit tricky -- We can't use Context.ExecListener here
	//	because ServeHTTP already runs in a goroutine and has to operate
	//	on the http.ResponseWriter synchronously. So we ad-hoc build
	//	an Engine-locked time block here to do that, instead of leaning on
	//	Context.ExecListener.
	ctx.Engine.Listeners.Add(1)
	defer ctx.Engine.Listeners.Done()

	ctx.Engine.evalLock.Lock()
	defer ctx.Engine.evalLock.Unlock()

	// unmarshal request
	method := r.Method
	url := r.URL.String()
	headers := ValueTable{}
	for key, values := range r.Header {
		headers[key] = StringValue{strings.Join(values, ",")}
	}
	var body Value
	if r.ContentLength == 0 {
		body = NullValue{}
	} else {
		bodyEncoded := ValueTable{}
		// XXX: in the future, we may move to streams
		bodyBuf := make([]byte, r.ContentLength)
		count, err := r.Body.Read(bodyBuf)
		if err != nil {
			_, err := evalInkFunction(cb, false, CompositeValue{
				entries: ValueTable{
					"type": StringValue{"error"},
					"message": StringValue{fmt.Sprintf(
						"error reading request in listen(), %s", err.Error(),
					)},
				},
			})
			if err != nil {
				callbackErr(err)
				return
			}
		}

		for i, b := range bodyBuf[:count] {
			bodyEncoded[nToS(float64(i))] = NumberValue{float64(b)}
		}

		body = CompositeValue{
			entries: bodyEncoded,
		}
	}

	// construct request object to pass to Ink, and call handler
	responseEnded := false
	responses := make(chan Value, 1)
	// this is what Ink's callback calls to send a response
	endHandler := func(ctx *Context, in []Value) (Value, error) {
		if len(in) != 1 {
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				"end() callback to listen() must have one argument",
			})
		}
		if responseEnded {
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				"end() callback to listen() was called more than once",
			})
		}
		responseEnded = true
		responses <- in[0]

		return NullValue{}, nil
	}

	_, err := evalInkFunction(cb, false, CompositeValue{
		entries: ValueTable{
			"type": StringValue{"req"},
			"data": CompositeValue{
				entries: ValueTable{
					"method":  StringValue{method},
					"url":     StringValue{url},
					"headers": CompositeValue{entries: headers},
					"body":    body,
				},
			},
			"end": NativeFunctionValue{
				name: "end",
				exec: endHandler,
				ctx:  ctx,
			},
		},
	})
	if err != nil {
		callbackErr(err)
		return
	}

	// validate response from Ink callback
	resp := <-responses
	rsp, isComposite := resp.(CompositeValue)
	if !isComposite {
		ctx.Engine.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("callback to listen() should return a response, got %s",
				resp.String()),
		})
	}

	// unmarshal response from the return value
	// response = {status, headers, body}
	statusVal, okStatus := rsp.entries["status"]
	headersVal, okHeaders := rsp.entries["headers"]
	bodyVal, okBody := rsp.entries["body"]

	resStatus, okStatus := statusVal.(NumberValue)
	resHeaders, okHeaders := headersVal.(CompositeValue)
	resBody, okBody := bodyVal.(CompositeValue)

	if !okStatus || !okHeaders || !okBody {
		ctx.Engine.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("callback to listen() returned malformed response\n\t-> %s",
				rsp.String()),
		})
	}

	// marshal response object
	writeBuf := make([]byte, len(resBody.entries))
	for i, v := range resBody.entries {
		idx, err := strconv.Atoi(i)
		if err != nil {
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("response body in listen() is malformed, %s", err.Error()),
			})
		}

		if num, isNum := v.(NumberValue); isNum {
			writeBuf[idx] = byte(num.val)
		} else {
			ctx.Engine.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf(
					"response body in listen() is malformed, byte value is not a number",
				),
			})
		}
	}

	// write values to response
	w.WriteHeader(int(resStatus.val))
	wHeaders := w.Header()
	for k, v := range resHeaders.entries {
		if str, isStr := v.(StringValue); isStr {
			wHeaders.Set(k, str.val)
		}
		// blech. silently fail here, it's ok
	}
	_, err = w.Write(writeBuf)
	if err != nil {
		_, err := evalInkFunction(cb, false, CompositeValue{
			entries: ValueTable{
				"type": StringValue{"error"},
				"message": StringValue{fmt.Sprintf(
					"error writing request body in listen(), %s", err.Error(),
				)},
			},
		})
		if err != nil {
			callbackErr(err)
			return
		}
	}
}

func inkListen(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("listen() expects two arguments: host and handler, but got %d",
				len(in)),
		}
	}

	host, isString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)

	if !isString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in listen()",
		}
	}

	// short-circuit out if no read permission
	if !ctx.Engine.Permissions.Net {
		return NativeFunctionValue{
			name: "close",
			exec: func(ctx *Context, in []Value) (Value, error) {
				// fake close callback
				return NullValue{}, nil
			},
			ctx: ctx,
		}, nil
	}

	server := &http.Server{
		Addr: host.val,
		Handler: inkHTTPHandler{
			ctx:         ctx,
			inkCallback: cb,
		},
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					entries: ValueTable{
						"type": StringValue{"error"},
						"message": StringValue{fmt.Sprintf(
							"error starting http server in listen(), %s", err.Error(),
						)},
					},
				})
				if err != nil {
					ctx.Engine.LogErr(Err{
						ErrRuntime,
						fmt.Sprintf("error in callback to listen(), %s",
							err.Error()),
					})
				}
			})
		}
	}()

	return NativeFunctionValue{
		name: "close",
		exec: func(ctx *Context, in []Value) (Value, error) {
			err := server.Close()
			if err != nil {
				_, err = evalInkFunction(cb, false, CompositeValue{
					entries: ValueTable{
						"type": StringValue{"error"},
						"message": StringValue{fmt.Sprintf(
							"error closing server in listen(), %s", err.Error(),
						)},
					},
				})
			}
			if err != nil {
				ctx.Engine.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to listen(), %s",
						err.Error()),
				})
			}
			return NullValue{}, nil
		},
		ctx: ctx,
	}, nil
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

	// This is a bit tricky, since we don't want wait() to hold the evalLock
	//	on the Context while we're waiting for the timeout, but do want to hold
	//	the main goroutine from completing with sync.WaitGroup.
	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		time.Sleep(time.Duration(
			int64(secs.val * float64(time.Second)),
		))

		ctx.ExecListener(func() {
			_, err := evalInkFunction(in[1], false)
			if err != nil {
				if e, isErr := err.(Err); isErr {
					ctx.Engine.LogErr(e)
				} else {
					// should never happen
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
