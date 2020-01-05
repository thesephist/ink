# Todo items


## Interpreter

- [ ] Reduce memory allocations at runtime, ideally without impacting runtime Ink performance of hot code paths. If done right, this should free up both memory and CPU.
- [ ] Implement go-fuzz to fuzz test the whole toolchain.
    - go-fuzz talk: http://go-talks.appspot.com/github.com/dvyukov/go-fuzz/slides/go-fuzz.slide#1
- [ ] Run Ink on Windows, ensure core functionality + builtins work out of the box and document any non-POSIX interfaces or discrepancies.
- [ ] Make the interpreter a Homebrew brew tap.
- [ ] --no-color option for piping output to another application / for scripting use (e.g. inker).


## Language core

- [ ] Type system? I like the way typescript does it, I think Ink’s type checker should be a completely separate layer in the toolchain from the lex/parse/eval layer. But let’s think about the merits of having type annotations and how we can make it simple to lex/parse out while effective at bringing out the forte’s of Ink’s functional style.
    - It seems helpful to think of it as a constraint system on the source code, instead of as something that’s an attribute of the runtime execution itself.
    - Since Ink has no implicit casts, this seems like it'll be straightforward to infer most variable types from their declaration (what they're bound to in the beginning) and recurse up the tree. So to compute "what's the type of this expression?" the type checker will recursively ask its children for their types, and assuming none of them return an error, we can define a `func (n Node) Type() (Type, error)` that recursively descends to type check an entire AST node from the top. We can expose this behind an `ink -check <file>.ink` flag.
    - To support Ink's way of error handling and signaling (returning null data values), the type system must support sum types, i.e. `number | ()`
    - Enforce mutability restrictions at the type level -- variables are (deeply) immutable by default, must be marked as mutable to allow mutation. This also improves functional ergonomics of the language.
    - Potential type annotation: `myVar<type>` (`myVar` is of type `type`), `myFunc<string, boolean => {number}>` (`myFunc` is of type function mapping `string`, `boolean` to type composite of `number`)
- [ ] Add to [GitHub Linguist](https://github.com/github/linguist)


## Potential examples / projects

- [ ] Markdown parser (or, a reduced subset of Markdown to HTML)
- [ ] The Knuth/McIlroy test -- read a text file (stream), find the top N most frequent words and print it.
