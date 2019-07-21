# Ink programming language 🖋

Ink is a minimal programming language inspired by modern JavaScript and Go, with functional style.

Ink has a few goals. In order, they are

- Ink should have a simple, minimal syntax and feature set
- Ink should be easy to learn regardless of skill level
- Ink should be quickly readable yet expressive
- Ink should have a great, fully featured, and modular standard library
- Ink should have an ergonomic interpreter and runtime API

Design is always a game of tradeoffs. Ink's goals for minimalism and readability / expressivity means the language deliberately does not aim to be best in other ways:

- Ink doesn't need to be highly efficient or fast, especially compared to compiled languages
    - However, within the constraints of the interpreter design, I try not to leave performance on the table, both in execution speed and in memory footprint. Efficiently composed Ink programs are between 2-4x slower than equivalent Python programs, in my experience. Small programs can run on as little as 3MB of memory, while the interpreter can stably scale up to gigabytes of memory for data-heavy tasks.
- Ink doesn't need to be particularly concise, though we try to avoid verbosity when we can
- Ink doesn't value platform portability as much as some other languages in this realm, like Lua or JavaScript -- not running on every piece of hardware available is okay, as long as it runs on most of the popular platforms

The rest of this README is a light introduction to the Ink language and documentation about the project and its interpreter, written in Go. For more information and formal specification about the Ink language itself, please see [SPEC.md](SPEC.md).

## Introduction

Here's an implementation of FizzBuzz in Ink.

```ink
` ink fizzbuzz implementation `

std := load('std')

log := std.log
range := std.range
each := std.each

fizzbuzz := n => each(
	range(1, n + 1, 1)
	n => [n % 3, n % 5] :: {
		[0, 0] -> log('FizzBuzz')
		[0, _] -> log('Fizz')
		[_, 0] -> log('Buzz')
		_ -> log(n)
	}
)

fizzbuzz(100)
```

Here's a simple Hello World HTTP server program.

```ink
std := load('std')

log := std.log
encode := std.encode

close := listen('0.0.0.0:8080', evt => (
	evt.type :: {
		'error' -> log('Error: ' + evt.message)
		'req' -> (evt.end)({
			status: 200
			headers: {'Content-Type': 'text/plain'}
			body: encode('Hello, World!')
		})
	}
))

` shut down server after 5 seconds `
wait(5, close)
```

If you're looking for more realistic and complex examples, check out...

- [the standard library](samples/std.ink)
- [quicksort](samples/quicksort.ink)
- [the standard test suite](samples/test.ink)
- [Newton's root finding algorithm](samples/newton.ink)
- [a small static file server](samples/fileserver.ink)
- [Mandelbrot set renderer](samples/mandelbrot.ink)

You'll notice a few characteristic things about Ink:

- Functions are defined using arrows (`=>`) _a la_ JavaScript arrow functions
- Ink does not have a looping primitive (no `for` or `while`), and instead defaults to tail-optimized recursion. Loops may be possible to have in syntax with macros in the near future.
- Rather than using `if`/`else`, Ink uses pattern matching using the match (`::`) operator. Match expressions in Ink allows for very expressive definition of complex flow control.
- Ink does not have explicit return statements. Instead, everything is an expression that evaluates to a value, and function bodies are a list of expressions whose last-evaluated expression value becomes the "return value" of the function.
- As a general principle, Ink tries not to use English keywords in favor of a small set of short symbols.

You can find more sample code in the `samples/` directory and run them with `ink samples/<file>.ink`.

## Getting started

You can run Ink in three main ways:

1. The Ink binary `ink` defaults to executing whatever comes through standard input. So you can pipe any Ink script to the binary to execute it.
```
$ cat <file>.ink | ink
	# or
$ ink < <file>.ink
```
2. Use `ink <file>.ink` to execute an Ink script file. You may pass multiple files to execute multiple scripts, like `ink a.ink b.ink`.
3. Invoke `ink -repl` to start an interactive repl session, and start typing Ink code. You can run files in this context by loading Ink files into the context using the `load(<filename>)` function in the repl.

Additionally, you can also invoke an Ink script with a [shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)). Mark the _first line_ of your Ink program file with this directive, which tells the operating system to run the program file with `ink`, which will then accept this file and run it for you when you execute the file.

```ink
#!/usr/bin/env ink

... the rest of your program
```

You can find an example of this in `samples/fileserver.ink`, which you can start by simply running `./samples/fileserver.ink` (without having to specifically call `ink samples/fileserver.ink`).

To summarize, ink's input priority is, from highest to lowest, `-repl` -> `-eval` -> files -> `stdin`. Note that command line flags to `ink` should _precede_ any program files given as arguments. If you need to pass a file name that begins with a dash, use `--`.

## Why?

I started the Ink project to become more familiar with how interpreters work, and to try my hand at designing a language that fit my preferences for the balance between elegance, simplicity, practicality, and expressiveness. The first part -- to learn about programming languages and interpreters -- is straightforward, so I want to expand on the second part.

My language of choice at work is currently JavaScript. JavaScript is expressive, very fast (for a dynamic language), and has an approach to concurrency that I really like, using a combination of closures with event loops and message passing to communicate between separate threads of execution. But JavaScript has grown increasingly large in its size and complexity, and also carries a lot of old cruft for sake of backwards compatibility. I've also been increasingly interested in composing programs from functional components, and there are features in the functional PL world that haven't yet made their way into JavaScript like expressive pattern matching and guaranteed tail recursion optimizations (the former has been in TC39 limbo for several years, and the latter is only supported by recent versions of WebKit/JavaScriptCore).

So Ink as a language is my attempt to build a language in the functional paradigm that doesn't sacrifice the concurrency benefits or expressiveness of JavaScript, while being minimal and self-consistent in syntax and semantics. I sometimes think about Ink as what JavaScript would be if it were rewritten by a Lisp programmer. Given this motivation, Ink tries to be a small language with little noise in the syntax, few special tokens, and a few essential builtins, that becomes expressive and powerful by being extremely composable and extensible. While modern dynamic languages routinely have over 100 syntactic forms, Ink has just 10 syntactic forms, from which everything else is derived. Ink deliberately avoids adding features into the language for sake of building a feature-rich language; whenever something can be achieved idiomatically within the constraints and patterns of the existing language or core libraries, that's preferred over adding new features into the language itself. This is how Ink remains tiny and self-consistent.

I'm also very interested in Elixir's approach towards language development, where there is a finite set of features planned to be added to the language itself, and the language is designed to become "complete" at some point in its lifetime, after which further growth happens through extending the language with macros and the ecosystem. Since simplicity and minimalism is a core goal of Ink, this perspective really appeals to me, and you can expect Ink to become "complete" at some finite point in the future. In fact, the feature set documented in this repository today is probably 85-90% of the total language features Ink will get eventually.

## Isolation and permissions model

Ink has a very small surface area to interface with the rest of the interpreter and runtime, which is through the list of builtin functions defined in `runtime.go`. In an effort to make it safe and easy to run potentially untrusted scripts, the Ink interpreter provides a few flags that determines whether the running Ink program may interface with the operating system in certain ways. Rather than simply fail or error on any restricted interface calls, the runtime will silently ignore the requested action and potentially return empty but valid data.

- `--no-read`: When enabled, the builtin `read()` function will simply return an empty read, as if the file being read was of size 0. `--no-read` also blocks directory traversals.
- `--no-write`: When enabled, the builtins `write()`, `delete()`, and `make()` will pretend to have written the requested data or finished the requested filesystem operationg safely, but cause no change.
- `--no-net`: When enabled, the builtin `listen()` function will pretend to have bound to a local network socket, but will not actually bind. The builtin `req()` will also pretend to have sent a valid request, but will do nothing.

To run an Ink program completely untrusted, run `ink -isolate` (with the "isolate" flag), which will revoke all revokable permissions from the running script.

## Development

Ink is currently a single go package. Run `go run .` to run from source, and `go build -ldflags="-s -w"` to build the release binary.

The `ink` binary takes in scripts from standard input, unless at least one input file is provided, in which case it reads from the filesystem.

### Make

Ink uses [GNU Make](https://www.gnu.org/software/make/manual/make.html) to manage build and development processes:

- `make test` runs the full test suite, including filesystem and syntax/parser tests
- `make run` runs the _extra_ set of tests, which are at the moment just the full suite of samples in the repository
- `make build-<OS>` builds the Ink interpreter for a given operating system target. For example, `make build-linux` will build Ink for Linux to `ink-linux`.
- `make-build` by itself builds all release targets. We currently build for 4 OS targets: Windows, macOS, Linux, and OpenBSD
- `make install` installs Ink to your system
- `make precommit` will performa any pre-commit checks for commiting changes to the development tree. Currently it lints and formats the Go code.
- `make clean` cleans any files that may have been generated by running make scripts or sample Ink programs

### Go API

As the canonical interpreter is currently written in Go, if you want to embed Ink within your own application, you can use the Go APIs from this package to do so.

The APIs are still in development / in flux, but you can check out `main.go` and `eval.go` for the Go channels-based concurrent lexer/parser/evaler APIs. As the APIs are finalized, I'll put more information here directly.

### IDE support

Ink currently has a vim syntax definition file, under `utils/ink.vim`. I'm also hoping to support Monaco / VSCode's language definition format soon.
