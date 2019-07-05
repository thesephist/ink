# Todo items

## Codebase / Golang

- [ ] Go through it with Effective Go with a fine toothed comb
- [ ] As we get time also make a ink -> JavaScript in JS and/or an Ink interpreter in JS and maybe ship it as a javascript compiler? Great for
    1. correctness checking against Go implementation
    2. writing web code
    3. having a second independent implementation of the language
    - If we have ink in go and JS we can fuzz both together which allows us to also test for correctness better
- [ ] Implement go-fuzz to fuzz test the whole toolchain
    - go-fuzz talk: http://go-talks.appspot.com/github.com/dvyukov/go-fuzz/slides/go-fuzz.slide#1
- [ ] Set up travis ci for Ink, and for now make it run run-all and then the test-all script. If zero exit value, it worked.
- [ ] implement various data structures and algorithms in Ink/samples
    - Binary Search Tree
    - Computing and rendering the Mendelbrot set.
    - Project Euler solutions?
- [ ] Document the Go API / bindings


## Interpreter

- Make a note of permissions isolation in README (not SPEC since it's not a language feature), and explain why it's open by default and not secure by default: (1) we don't need weblike security of default-everything-sandbox imo because we don't have web problems, and (2) Most other interpreters are default everything by access with no off switch, and I think this is a happy medium that won't be bothersome.
- [ ] `func (n Node) prettyString() string` to pretty-print AST, use this to implement `ink -fmt <file>.ink`
- [ ] Start benchmarking Ink against JavaScript and Python and keep a progress history. A suite of tests across different aspects of the interpreter, like calling stack frames vs allocating lots of objects etc.
    - `quicksort.ink` implementation with 50k/100k elements seems like a good starting point for a benchmark. Let's measure that every commit, and pit that against JavaScript?
    - `prime.ink` is also a good candidate.
- [ ] --no-color option for piping output to another application / for scripting use (e.g. inker).
- [ ] Errors should print a stacktrace (elided if tail recursed)


## Language

- [ ] A procedural macro system to extend the language (syntax?). Study Rust's procedural macro system and Sweet.js, build an AST-aware macro system into the standard library.
- [ ] Type system? I like the way typescript does it, I think ink’s type checker should be a completely separate layer in the toolchain from the lex/parse/eval layer. But let’s think about the merits of having type annotations and how we can make it simple to lex/parse out while effective at bringing out the forte’s of ink’s functional style.
    - It seems helpful to think of it as a constraint system on the runtime, instead of as something that’s an attribute of the runtime execution itself. Maybe take the Haskell approach of having definition / type declarations separate from the imperative flow definition/syntax itself? Also look at how Haskell / Erlang / Clojure / other functional languages do type annotation, and keep in mind that while Ink is functional flavor it should still feel great in imperative / OO style.

## Standard library / utilities

- [ ] Finish builtin function `listen()` with `--no-net` respected. Let's start with http, for practicality's sake.
- [ ] JSON serde system
- [ ] Impl streams / channels / reactive-across-time primitives for programming in the standard library, building on events / input primitives.
- [ ] Promises / futures should be in the standard library in Ink, composed of callback primitives.
- [ ] We should study event systems / event loop models like libuv and Tokio more, especially in light of Golang's strange Erlangy processes / green threads model.


## Bugs

- [ ] There are parts in the parser where I just move `idx++` through the tokens stream without checking that the token I passed through is the one I assumed it was (usually closing delimiters). Fix these and check all cases.
- [ ] `encode()` and `decode()` do not currently support the full Unicode plane -- we might revisit this in the future.
- [ ] `EqRefOp` (comparison with `is`) is buggy, and I think we won't really use it in idiomatic Ink. We may remove it in the future.


## Potential exampels / projects

- [ ] Path tracer
- [ ] Markdown parser (or, a reduced subset of Markdown to HTML)
