# Todo items


## Interpreter

- [ ] Set up Travis CI for Ink: for now, just run `make run && make test`, and pass/fail on the exit value.
- [ ] Potential room for optimizations:
    - Reducing memory allocations. Specifically, pooling `StackFrame` and other function call-related data structures to reduce allocations.
    - Optimized tail recursion unwrapping that doesn't require cost of wrapping and then immediatley unwrapping thunks.
    - Cache locality for data structures (stack-local variables?)
- [ ] Implement go-fuzz to fuzz test the whole toolchain
    - go-fuzz talk: http://go-talks.appspot.com/github.com/dvyukov/go-fuzz/slides/go-fuzz.slide#1
- [ ] Implement the concurrency system (`send()`, `receive()`, `create()` builtins) as described in the language spec.
- [ ] Make the interpreter a Homebrew brew tap
- [ ] Ink -> JavaScript in JS and/or an Ink interpreter in JS and maybe ship it as a javascript compiler? Great for
    1. correctness checking against Go implementation
    2. writing web code
    3. having a second independent implementation of the language
    - If we have ink in go and JS we can fuzz both together which allows us to also test for correctness better
- [ ] --no-color option for piping output to another application / for scripting use (e.g. inker).


## Language + standard library

- [ ] Type system? I like the way typescript does it, I think ink’s type checker should be a completely separate layer in the toolchain from the lex/parse/eval layer. But let’s think about the merits of having type annotations and how we can make it simple to lex/parse out while effective at bringing out the forte’s of ink’s functional style.
    - It seems helpful to think of it as a constraint system on the runtime, instead of as something that’s an attribute of the runtime execution itself. I like the Haskell approach of having definition / type declarations separate from the imperative flow definition/syntax itself. Also look at how Haskell / Erlang / Clojure / other functional languages do type annotation, and keep in mind that while Ink is functional flavor it should still feel great in imperative / OO style.
    - Since Ink has no implicit casts, this seems like it'll be straightforward to infer most variable types from their declaration (what they're bound to in the beginning) and recurse up the tree. So to compute "what's the type of this expression?" the type checker will recursively ask its children for their types, and assuming none of them return an error, we can define a `func (n Node) TypeCheck() (Type, error)` that recursively descends to type check an entire AST node from the top. We can expose this behind an `ink -type-check <file>.ink` flag.
    - To support Ink's way of error handling and signaling (returning null data values), the type system must support sum types, i.e. `number | ()`
- [ ] Should study event systems / event loop models like libuv and Tokio more, especially in light of Golang's strange Erlangy processes / green threads model.


## Potential examples / projects

- [ ] Markdown parser (or, a reduced subset of Markdown to HTML)
- [ ] The Knuth/McIlroy test -- read a text file (stream), find the top N most frequent words and print it.
- [ ] Math guessing game, but where we show you a product of two 2x2 or 2x3 digit numbers and you have to guess the output. It should ask you to retry until you get it.
