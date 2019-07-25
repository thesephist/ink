package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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
	ctx.LoadFunc("dir", inkDir)
	ctx.LoadFunc("make", inkMake)
	ctx.LoadFunc("stat", inkStat)
	ctx.LoadFunc("read", inkRead)
	ctx.LoadFunc("write", inkWrite)
	ctx.LoadFunc("delete", inkDelete)
	ctx.LoadFunc("listen", inkListen)
	ctx.LoadFunc("req", inkReq)
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
	ctx.Frame.Set(name, NativeFunctionValue{
		name,
		exec,
		ctx,
	})
}

// Create and return a standard error callback response with the given message
func errMsg(message string) CompositeValue {
	return CompositeValue{
		"type":    StringValue("error"),
		"message": StringValue(message),
	}
}

func inkLoad(ctx *Context, in []Value) (Value, error) {
	if len(in) == 1 {
		if givenPath, ok := in[0].(StringValue); ok && len(givenPath) > 0 {
			// imports via load() are assumed to be relative
			importPath := path.Join(ctx.Cwd, string(givenPath)+".ink")

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

			return CompositeValue(childCtx.Frame.vt), nil
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

	cbErr := func(err error) {
		ctx.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("error in callback to in(), %s", err.Error()),
		})
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
				"type": StringValue("data"),
				"data": StringValue(str),
			})
			if err != nil {
				cbErr(err)
				return
			}

			if boolValue, isBool := rv.(BooleanValue); isBool {
				if !boolValue {
					break
				}
			} else {
				ctx.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("callback to in() should return a boolean, but got %s", rv),
				})
				return
			}
		}

		_, err := evalInkFunction(in[0], false, CompositeValue{
			"type": StringValue("end"),
		})
		if err != nil {
			cbErr(err)
			return
		}
	})

	return NullValue{}, nil
}

func inkOut(ctx *Context, in []Value) (Value, error) {
	if len(in) == 1 {
		if output, ok := in[0].(StringValue); ok {
			os.Stdout.Write([]byte(output))
			return NullValue{}, nil
		}
	}

	return nil, Err{
		ErrRuntime,
		"out() takes one string argument",
	}
}

func inkDir(ctx *Context, in []Value) (Value, error) {
	if (len(in)) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("dir() expects two arguments: path and callback, but got %d", len(in)),
		}
	}

	dirPath, isDirPathString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)
	if !isDirPathString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in dir()",
		}
	}

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to dir(), %s", err.Error()),
			})
		}
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Read {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("data"),
					"data": CompositeValue{},
				})
				cbMaybeErr(err)
			})
			return
		}

		fileInfos, err := ioutil.ReadDir(string(dirPath))
		if err != nil {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, errMsg(
					fmt.Sprintf("error listing directory contents in dir(), %s", err.Error()),
				))
				cbMaybeErr(err)
			})
			return
		}

		fileList := CompositeValue{}
		for i, fi := range fileInfos {
			fileList[strconv.Itoa(i)] = CompositeValue{
				"name": StringValue(fi.Name()),
				"len":  NumberValue(fi.Size()),
				"dir":  BooleanValue(fi.IsDir()),
			}
		}

		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("data"),
				"data": fileList,
			})
			cbMaybeErr(err)
		})
	}()

	return NullValue{}, nil
}

func inkMake(ctx *Context, in []Value) (Value, error) {
	if (len(in)) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("make() expects two arguments: path and callback, but got %d", len(in)),
		}
	}

	dirPath, isDirPathString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)
	if !isDirPathString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in make()",
		}
	}

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to make(), %s", err.Error()),
			})
		}
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Write {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("end"),
				})
				cbMaybeErr(err)
			})
			return
		}

		// mkdir
		err := os.Mkdir(string(dirPath), 0755)
		if err != nil {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, errMsg(
					fmt.Sprintf("error making a new directory in make(), %s", err.Error()),
				))
				cbMaybeErr(err)
			})
			return
		}

		ctx.ExecListener(func() {
			_, err = evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("end"),
			})
			cbMaybeErr(err)
		})
	}()

	return NullValue{}, nil
}

func inkStat(ctx *Context, in []Value) (Value, error) {
	if (len(in)) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("stat() expects two arguments: path and callback, but got %d", len(in)),
		}
	}

	statPath, isStatPathString := in[0].(StringValue)
	cb, isCbFunction := in[1].(FunctionValue)
	if !isStatPathString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in stat()",
		}
	}

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to stat(), %s", err.Error()),
			})
		}
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Read {
			statPathBase := make([]byte, 0, len(statPath))
			statPathCopy := StringValue(append(statPathBase, statPath...))
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("data"),
					"data": CompositeValue{
						"name": statPathCopy,
						"len":  NumberValue(0),
						"dir":  BooleanValue(false),
					},
				})
				cbMaybeErr(err)
			})
			return
		}

		fi, err := os.Stat(string(statPath))
		if err != nil {
			if os.IsNotExist(err) {
				ctx.ExecListener(func() {
					_, err := evalInkFunction(cb, false, CompositeValue{
						"type": StringValue("data"),
						"data": NullValue{},
					})
					cbMaybeErr(err)
				})
			} else {
				ctx.ExecListener(func() {
					_, err := evalInkFunction(cb, false, errMsg(
						fmt.Sprintf("error getting file data in stat(), %s", err.Error()),
					))
					cbMaybeErr(err)
				})
			}
			return
		}

		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("data"),
				"data": CompositeValue{
					"name": StringValue(fi.Name()),
					"len":  NumberValue(fi.Size()),
					"dir":  BooleanValue(fi.IsDir()),
				},
			})
			cbMaybeErr(err)
		})
	}()

	return NullValue{}, nil
}

func inkRead(ctx *Context, in []Value) (Value, error) {
	if len(in) != 4 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("read() expects four arguments: path, offset, length, and callback, but got %d", len(in)),
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

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to read(), %s", err.Error()),
			})
		}
	}

	sendErr := func(msg string) {
		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, errMsg(msg))
			cbMaybeErr(err)
		})
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Read {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("data"),
					"data": StringValue{},
				})
				cbMaybeErr(err)
			})
			return
		}

		// open
		file, err := os.OpenFile(string(filePath), os.O_RDONLY, 0644)
		defer file.Close()
		if err != nil {
			sendErr(fmt.Sprintf("error opening requested file in read(), %s", err.Error()))
			return
		}

		// seek
		ofs := int64(offset)
		if ofs != 0 {
			_, err := file.Seek(ofs, 0) // 0 means relative to start of file
			if err != nil {
				sendErr(fmt.Sprintf("error seeking requested file in read(), %s", err.Error()))
				return
			}
		}

		// read
		buf := make([]byte, int64(length))
		count, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				// if first read returns EOF, it's an empty file
				ctx.ExecListener(func() {
					_, err = evalInkFunction(cb, false, CompositeValue{
						"type": StringValue("data"),
						"data": StringValue{},
					})
					cbMaybeErr(err)
				})
			} else {
				sendErr(fmt.Sprintf("error reading requested file in read(), %s", err.Error()))
			}
			return
		}

		ctx.ExecListener(func() {
			_, err = evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("data"),
				"data": StringValue(buf[:count]),
			})
			cbMaybeErr(err)
		})
	}()

	return NullValue{}, nil
}

func inkWrite(ctx *Context, in []Value) (Value, error) {
	if len(in) != 4 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("write() expects four arguments: path, offset, length, and callback, but got %d", len(in)),
		}
	}

	filePath, isFilePathString := in[0].(StringValue)
	offset, isOffsetNumber := in[1].(NumberValue)
	buf, isString := in[2].(StringValue)
	cb, isCbFunction := in[3].(FunctionValue)
	if !isFilePathString || !isOffsetNumber || !isString || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in write()",
		}
	}

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to write(), %s", err.Error()),
			})
		}
	}

	sendErr := func(msg string) {
		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, errMsg(msg))
			cbMaybeErr(err)
		})
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Write {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("end"),
				})
				cbMaybeErr(err)
			})
			return
		}

		// open
		var flag int
		if offset == -1 {
			// -1 offset is append
			flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
		} else {
			// all other offsets are writing
			flag = os.O_CREATE | os.O_WRONLY
		}
		file, err := os.OpenFile(string(filePath), flag, 0644)
		defer file.Close()
		if err != nil {
			sendErr(fmt.Sprintf("error opening requested file in write(), %s", err.Error()))
			return
		}

		// seek
		if offset != -1 {
			ofs := int64(offset)
			_, err := file.Seek(ofs, 0) // 0 means relative to start of file
			if err != nil {
				sendErr(fmt.Sprintf("error seeking requested file in write(), %s", err.Error()))
				return
			}
		}

		// write
		_, err = file.Write(buf)
		if err != nil {
			sendErr(fmt.Sprintf("error writing to requested file in write(), %s", err.Error()))
			return
		}

		ctx.ExecListener(func() {
			_, err = evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("end"),
			})
			cbMaybeErr(err)
		})
	}()

	return NullValue{}, nil
}

func inkDelete(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("delete() expects two arguments: path and callback, but got %d", len(in)),
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

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to delete(), %s", err.Error()),
			})
		}
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		if !ctx.Engine.Permissions.Write {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, CompositeValue{
					"type": StringValue("end"),
				})
				cbMaybeErr(err)
			})
			return
		}

		// delete
		err := os.Remove(string(filePath))
		if err != nil {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, errMsg(
					fmt.Sprintf("error removing requested file in delete(), %s", err.Error()),
				))
				cbMaybeErr(err)
			})
			return
		}

		ctx.ExecListener(func() {
			_, err = evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("end"),
			})
			cbMaybeErr(err)
		})
	}()

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

	cbMaybeErr := func(err error) {
		if err != nil {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("error in callback to listen(), %s", err.Error()),
			})
		}
	}

	// unmarshal request
	method := r.Method
	url := r.URL.String()
	headers := CompositeValue{}
	for key, values := range r.Header {
		headers[key] = StringValue(strings.Join(values, ","))
	}
	var body Value
	if r.ContentLength == 0 {
		body = NullValue{}
	} else {
		bodyBuf, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ctx.ExecListener(func() {
				_, err := evalInkFunction(cb, false, errMsg(
					fmt.Sprintf("error reading request in listen(), %s", err.Error()),
				))
				cbMaybeErr(err)
			})
			return
		}
		body = StringValue(bodyBuf)
	}

	// construct request object to pass to Ink, and call handler
	responseEnded := false
	responses := make(chan Value, 1)
	// this is what Ink's callback calls to send a response
	endHandler := func(ctx *Context, in []Value) (Value, error) {
		if len(in) != 1 {
			ctx.LogErr(Err{
				ErrRuntime,
				"end() callback to listen() must have one argument",
			})
		}
		if responseEnded {
			ctx.LogErr(Err{
				ErrRuntime,
				"end() callback to listen() was called more than once",
			})
		}
		responseEnded = true
		responses <- in[0]

		return NullValue{}, nil
	}

	ctx.ExecListener(func() {
		_, err := evalInkFunction(cb, false, CompositeValue{
			"type": StringValue("req"),
			"data": CompositeValue{
				"method":  StringValue(method),
				"url":     StringValue(url),
				"headers": headers,
				"body":    body,
			},
			"end": NativeFunctionValue{
				name: "end",
				exec: endHandler,
				ctx:  ctx,
			},
		})
		cbMaybeErr(err)
	})

	// validate response from Ink callback
	resp := <-responses
	rsp, isComposite := resp.(CompositeValue)
	if !isComposite {
		ctx.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("callback to listen() should return a response, got %s", resp),
		})
	}

	// unmarshal response from the return value
	// response = {status, headers, body}
	statusVal, okStatus := rsp["status"]
	headersVal, okHeaders := rsp["headers"]
	bodyVal, okBody := rsp["body"]

	resStatus, okStatus := statusVal.(NumberValue)
	resHeaders, okHeaders := headersVal.(CompositeValue)
	resBody, okBody := bodyVal.(StringValue)

	if !okStatus || !okHeaders || !okBody {
		ctx.LogErr(Err{
			ErrRuntime,
			fmt.Sprintf("callback to listen() returned malformed response, %s", rsp),
		})
	}

	// write values to response
	// Content-Length is automatically set for us by Go
	for k, v := range resHeaders {
		if str, isStr := v.(StringValue); isStr {
			w.Header().Set(k, string(str))
		} else {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("could not set response header, value %s was not a string", v),
			})
		}
	}
	// must follow all other header writes
	w.WriteHeader(int(resStatus))
	_, err := w.Write(resBody)
	if err != nil {
		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, errMsg(
				fmt.Sprintf("error writing request body in listen(), %s", err.Error()),
			))
			cbMaybeErr(err)
		})
		return
	}
}

func inkListen(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("listen() expects two arguments: host and handler, but got %d", len(in)),
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

	sendErr := func(msg string) {
		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, errMsg(msg))
			if err != nil {
				ctx.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to listen(), %s", err.Error()),
				})
			}
		})
	}

	server := &http.Server{
		Addr: string(host),
		Handler: inkHTTPHandler{
			ctx:         ctx,
			inkCallback: cb,
		},
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			sendErr(fmt.Sprintf("error starting http server in listen(), %s", err.Error()))
		}
	}()

	closer := func(ctx *Context, in []Value) (Value, error) {
		// attempt graceful shutdown, concurrently, without
		//	blocking Ink evaluation thread
		ctx.Engine.Listeners.Add(1)
		go func() {
			defer ctx.Engine.Listeners.Done()

			err := server.Shutdown(context.Background())
			if err != nil {
				sendErr(fmt.Sprintf("error closing server in listen(), %s", err.Error()))
			}
		}()

		return NullValue{}, nil
	}

	return NativeFunctionValue{
		name: "close",
		exec: closer,
		ctx:  ctx,
	}, nil
}

func inkReq(ctx *Context, in []Value) (Value, error) {
	if len(in) != 2 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("req() expects two arguments: data and callback, but got %d", len(in)),
		}
	}

	data, isComposite := in[0].(CompositeValue)
	cb, isCbFunction := in[1].(FunctionValue)

	if !isComposite || !isCbFunction {
		return nil, Err{
			ErrRuntime,
			"unsupported combination of argument types in req()",
		}
	}

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

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// do not follow redirects
			return http.ErrUseLastResponse
		},
	}
	reqContext, reqCancel := context.WithCancel(context.Background())

	closer := func(ctx *Context, in []Value) (Value, error) {
		reqCancel()
		return NullValue{}, nil
	}

	sendErr := func(msg string) {
		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, errMsg(msg))
			if err != nil {
				ctx.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to req(), %s", err.Error()),
				})
			}
		})
	}

	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		// unmarshal request contents
		methodVal, okMethod := data["method"]
		urlVal, okURL := data["url"]
		headersVal, okHeaders := data["headers"]
		bodyVal, okBody := data["body"]

		reqMethod, okMethod := methodVal.(StringValue)
		reqURL, okURL := urlVal.(StringValue)
		reqHeaders, okHeaders := headersVal.(CompositeValue)
		reqBody, okBody := bodyVal.(StringValue)

		if !okMethod || !okURL || !okHeaders || !okBody {
			ctx.LogErr(Err{
				ErrRuntime,
				fmt.Sprintf("request in req() is malformed, %s", data),
			})
		}

		req, err := http.NewRequest(
			string(reqMethod),
			string(reqURL),
			bytes.NewReader(reqBody),
		)
		if err != nil {
			sendErr(fmt.Sprintf("error creating request in req(), %s", err.Error()))
			return
		}

		req = req.WithContext(reqContext)

		// construct headers
		// Content-Length is automatically set for us by Go
		req.Header.Set("User-Agent", "") // remove Go's default user agent header
		for k, v := range reqHeaders {
			if str, isStr := v.(StringValue); isStr {
				req.Header.Set(k, string(str))
			} else {
				ctx.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("could not set request header, value %s was not a string", v),
				})
			}
		}

		// send request
		resp, err := client.Do(req)
		if err != nil {
			sendErr(fmt.Sprintf("error processing request in req(), %s", err.Error()))
			return
		}

		resStatus := NumberValue(resp.StatusCode)
		resHeaders := CompositeValue{}
		for key, values := range resp.Header {
			resHeaders[key] = StringValue(strings.Join(values, ","))
		}

		var resBody Value
		if resp.ContentLength == 0 {
			resBody = NullValue{}
		} else {
			bodyBuf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				sendErr(fmt.Sprintf("error reading response in req(), %s", err.Error()))
				return
			}
			resBody = StringValue(bodyBuf)
		}

		ctx.ExecListener(func() {
			_, err := evalInkFunction(cb, false, CompositeValue{
				"type": StringValue("resp"),
				"data": CompositeValue{
					"status":  resStatus,
					"headers": resHeaders,
					"body":    resBody,
				},
			})
			if err != nil {
				ctx.LogErr(Err{
					ErrRuntime,
					fmt.Sprintf("error in callback to req(), %s", err.Error()),
				})
			}
		})
	}()

	return NativeFunctionValue{
		name: "close",
		exec: closer,
		ctx:  ctx,
	}, nil
}

func inkRand(ctx *Context, in []Value) (Value, error) {
	return NumberValue(rand.Float64()), nil
}

func inkTime(ctx *Context, in []Value) (Value, error) {
	unixSeconds := float64(time.Now().UnixNano()) / 1e9
	return NumberValue(unixSeconds), nil
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
			fmt.Sprintf("first argument to wait() should be a number, but got %s", in[0]),
		}
	}

	// This is a bit tricky, since we don't want wait() to hold the evalLock
	//	on the Context while we're waiting for the timeout, but do want to hold
	//	the main goroutine from completing with sync.WaitGroup.
	ctx.Engine.Listeners.Add(1)
	go func() {
		defer ctx.Engine.Listeners.Done()

		time.Sleep(time.Duration(
			int64(float64(secs) * float64(time.Second)),
		))

		ctx.ExecListener(func() {
			_, err := evalInkFunction(in[1], false)
			if err != nil {
				if e, isErr := err.(Err); isErr {
					ctx.LogErr(e)
				} else {
					logErrf(ErrAssert, "Eval of an Ink node returned error not of type Err")
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
			fmt.Sprintf("sin() takes a number argument, got %s", in[0]),
		}
	}

	return NumberValue(math.Sin(float64(inNum))), nil
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
			fmt.Sprintf("cos() takes a number argument, got %s", in[0]),
		}
	}

	return NumberValue(math.Cos(float64(inNum))), nil
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
		if base == 0 && exp == 0 {
			return nil, Err{
				ErrRuntime,
				"math error, pow(0, 0) is not defined",
			}
		} else if base < 0 && !isIntable(exp) {
			return nil, Err{
				ErrRuntime,
				"math error, fractional power of negative number",
			}
		}
		return NumberValue(math.Pow(float64(base), float64(exp))), nil
	} else {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("pow() takes exactly 2 number arguments, but got %s, %s", in[0], in[1]),
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
			fmt.Sprintf("ln() takes exactly one number argument, but got %s", in[0]),
		}
	}

	if n <= 0 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("cannot take natural logarithm of non-positive number %s", nvToS(n)),
		}
	}

	return NumberValue(math.Log(float64(n))), nil
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
			fmt.Sprintf("floor() takes exactly one number argument, but got %s", in[0]),
		}
	}

	return NumberValue(math.Trunc(float64(n))), nil
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
		return StringValue(nvToS(v)), nil
	case BooleanValue:
		if v {
			return StringValue("true"), nil
		} else {
			return StringValue("false"), nil
		}
	case NullValue:
		return StringValue("()"), nil
	case CompositeValue:
		return StringValue(v.String()), nil
	case FunctionValue, NativeFunctionValue:
		return StringValue("(function)"), nil
	default:
		return StringValue(""), nil
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
		f, err := strconv.ParseFloat(string(v), 64)
		if err != nil {
			return NullValue{}, nil
		} else {
			return NumberValue(f), nil
		}
	case NumberValue:
		return v, nil
	case BooleanValue:
		if v {
			return NumberValue(1), nil
		} else {
			return NumberValue(0), nil
		}
	default:
		return NumberValue(0), nil
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
	if !isString || len(str) < 1 {
		return nil, Err{
			ErrRuntime,
			fmt.Sprintf("point() takes a string argument of length at least 1, got %s", in[0]),
		}
	}

	return NumberValue(str[0]), nil
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
			fmt.Sprintf("char() takes a number argument, got %s", in[0]),
		}
	}

	return StringValue([]byte{byte(cp)}), nil
}

func inkType(ctx *Context, in []Value) (Value, error) {
	if len(in) != 1 {
		return nil, Err{
			ErrRuntime,
			"type() takes exactly one argument",
		}
	}

	switch in[0].(type) {
	case StringValue:
		return StringValue("string"), nil
	case NumberValue:
		return StringValue("number"), nil
	case BooleanValue:
		return StringValue("boolean"), nil
	case NullValue:
		return StringValue("()"), nil
	case CompositeValue:
		return StringValue("composite"), nil
	case FunctionValue, NativeFunctionValue:
		return StringValue("function"), nil
	default:
		return StringValue(""), nil
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
		return NumberValue(len(list)), nil
	} else if str, isString := in[0].(StringValue); isString {
		return NumberValue(len(str)), nil
	}

	return nil, Err{
		ErrRuntime,
		fmt.Sprintf("len() takes a string or composite value, but got %s", in[0]),
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
			fmt.Sprintf("keys() takes a composite value, but got %s", in[0]),
		}
	}

	cv := CompositeValue{}

	i := 0
	for k, _ := range obj {
		cv[strconv.Itoa(i)] = StringValue(k)
		i++
	}

	return cv, nil
}
