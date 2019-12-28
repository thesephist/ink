# Todo items


## Interpreter

- [ ] Potential room for optimizations:
    - Reducing memory allocations. Specifically, pooling `StackFrame`, `FunctionCallThunkValue`, and other function call-related data structures to reduce allocations.
- [ ] Implement go-fuzz to fuzz test the whole toolchain
    - go-fuzz talk: http://go-talks.appspot.com/github.com/dvyukov/go-fuzz/slides/go-fuzz.slide#1
- [ ] Make the interpreter a Homebrew brew tap
- [ ] Run Ink on Windows, ensure core functionality + builtins work out of the box and document any non-POSIX interfaces or discrepancies
- [ ] --no-color option for piping output to another application / for scripting use (e.g. inker).


## Language core

- [ ] Type system? I like the way typescript does it, I think Ink’s type checker should be a completely separate layer in the toolchain from the lex/parse/eval layer. But let’s think about the merits of having type annotations and how we can make it simple to lex/parse out while effective at bringing out the forte’s of Ink’s functional style.
    - It seems helpful to think of it as a constraint system on the runtime, instead of as something that’s an attribute of the runtime execution itself. I like the Haskell approach of having definition / type declarations separate from the imperative flow definition/syntax itself. Also look at how Haskell / Erlang / Clojure / other functional languages do type annotation, and keep in mind that while Ink is functional flavor it should still feel great in imperative / OO style.
    - Since Ink has no implicit casts, this seems like it'll be straightforward to infer most variable types from their declaration (what they're bound to in the beginning) and recurse up the tree. So to compute "what's the type of this expression?" the type checker will recursively ask its children for their types, and assuming none of them return an error, we can define a `func (n Node) TypeCheck() (Type, error)` that recursively descends to type check an entire AST node from the top. We can expose this behind an `ink -type-check <file>.ink` flag.
    - To support Ink's way of error handling and signaling (returning null data values), the type system must support sum types, i.e. `number | ()`
- [ ] Implement the concurrency system (`send()`, `receive()`, `create()` builtins) as described in the language spec.
- [ ] Should study event systems / event loop models like libuv and Tokio more
- [ ] Add to [GitHub Linguist](https://github.com/github/linguist)


## Potential examples / projects

- [ ] Markdown parser (or, a reduced subset of Markdown to HTML)
- [ ] The Knuth/McIlroy test -- read a text file (stream), find the top N most frequent words and print it.
