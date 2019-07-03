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
- Ink doesn't need to be particularly concise, though we try to avoid verbosity when we can

The rest of this README is a gentle introduction to the Ink language and documentation about the project and its Go interpreter implementation. For more information and formal documentation about the Ink language itself, please see [SPEC.md](SPEC.md).

## Getting started

You can run Ink in three main ways:

1. The Ink binary `ink` defaults to executing whatever comes through standard input. So you can pipe any Ink script to the binary to execute it.
```
$ cat <file>.ink | ink
    # or
$ ink < <file>.ink
```
2. Use `ink -input <file>.ink` to execute an ink script file. You may pass the flag multiple times to execute multiple scripts, like `ink -input a.ink -input b.ink`.
3. Invoke `ink -repl` to start an interactive repl session, and start typing ink code. You can run files in this context by executing `@load <file>.ink` in the repl prompt.

## Introduction

Here's an implementation of FizzBuzz in the Ink language.

```ink
` ink fizzbuzz implementation `

fb := n => (
    [n % 3, n % 5] :: {
        [0, 0] -> log('FizzBuzz')
        [0, _] -> log('Fizz')
        [_, 0] -> log('Buzz')
        _ -> log(string(n))
    }
)

helper := (n, max) => (
    n :: {
        max -> fb(n)
        _ -> (
            fb(n)
            helper(n + 1, max)
        )
    }
)

fizzbuzz := max => helper(1, max)
fizzbuzz(100)
```

If you're looking for more realistic and complex examples, check out [quicksort.ink](samples/quicksort.ink), [newton.ink](samples/newton.ink), [the standard library](samples/std.ink), and [this file I/O demo](samples.io.ink).

You'll notice a few characteristic things about Ink:

- Functions are defined using arrows (`=>`) _a la_ JavaScript arrow functions
- Ink does not have a looping primitive (no `for` or `while`), and instead defaults to tail-optimized recursion. Loops may be possible to have in syntax with macros in the near future.
- Rather than using `if`/`else`, Ink uses pattern matching using the match (`::`) operator. Match expressions in Ink allows for very expressive definition of complex flow control.
- Ink does not have explicit return statements. Instead, everything is an expression that evaluates to a value, and function bodies are a list of expressions whose last-evaluated expression value becomes the "return value" of the function.
- As a general convention, Ink tries not to use to many English keywords in favor of a small set of short symbols. In fact, the only keyword using the English alphabet in the language is `is`, for reference equality checks.

You can find more sample code in the `samples/` directory and run them with `ink -input <sample>.ink`.

## Why?

I started the Ink project to become more familiar with how interpreters work, and to try my hand at designing a language that fit my preferences for the balance between elegance, simplicity, practicality, and expressiveness. The first part -- to learn about programming languages and interpreters -- is straightforward, so I want to expand on the second part.

My language of choice at work is currently JavaScript. JavaScript is expressive, very fast (for a dynamic language), and has an approach to concurrency that I really like, using a combination of closures with event loops and message passing to communicate between separate threads of execution. But JavaScript has grown increasingly large in its size and complexity, and also carries a lot of old cruft for sake of backwards compatibility. I've also been increasingly interested in composing programs from functional components, and there are features in the functional PL world that haven't yet made their way into JavaScript like expressive pattern matching and guaranteed tail recursion optimizations (the former has been in TC39 limbo for several years, and the latter is only supported by recent versions of WebKit/JavaScriptCore).

So Ink as a language is my attempt to build a language in the functional paradigm that doesn't sacrifice the concurrency benefits or expressiveness of JavaScript, while being minimal and self-consistent in syntax and semantics. I sometimes think about Ink as what JavaScript would be if it were rewritten by a Lisp programmer. Given this motivation, Ink tries to be a small language with little noise in the syntax, few special tokens, and a few essential builtins, that becomes expressive and powerful by being extremely composable and extensible. While modern dynamic languages routinely have over 100 syntactic forms, Ink has just 10 syntactic forms, from which everything else is derived. Ink deliberately avoids adding features into the language for sake of building a feature-rich language; whenever something can be achieved idiomatically within the constraints and patterns of the existing language or core libraries, that's preferred over adding new features into the language itself. This is how Ink remains tiny and self-consistent.

I'm also very interested in Elixir's approach towards language development, where there is a finite set of features planned to be added to the language itself, and the language is designed to become "complete" at some point in its lifetime, after which further growth happens through extending the language with macros and the ecosystem. Since simplicity and minimalism is a core goal of Ink, this perspective really appeals to me, and you can expect Ink to become "complete" at some finite point in the future. In fact, the feature set documented in this repository today is probably 85-90% of the total language features Ink will get eventually.

## Development

Ink is currently a single go package. Run `go run .` to run from source, and `go build -ldflags="-s -w"` to build the release binary.

The `ink` binary takes in scripts from standard input, unless at least one `-input` flag is provided, in which case it reads from the filesystem.

### Go API

As the canonical interpreter is currently written in Go, if you want to embed Ink within your own application, you can use the Go APIs from this package to do so.

The APIs are still in development / in flux, but you can check out `main.go` and `eval.go` for the Go channels-based concurrent lexer/parser/evaler APIs. As the APIs are finalized, I'll put more information here directly.

### IDE support

Ink currently has a vim syntax definition file, under `utils/ink.vim`. I'm also hoping to support Monaco / VSCode's language definition format soon.
