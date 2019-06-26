# Todo items

## Codebase / Golang

- [ ] Go through it with Effective Go with a fine toothed comb
- [ ] As we get time also make a ink -> JavaScript in JS and/or an Ink interpreter in JS and maybe ship it as a javascript compiler? Great for
    1. correctness checking against Go implementation
    2. writing web code
    3. having a second independent implementation of the language
- [ ] Write tests for parser/lexer/evaler separately to catch regressions more easily
- [ ] `func (n Node) prettyString() string` to pretty-print AST
- [ ] Add godoc.


## Language

- [ ] A procedural macro system to extend the language (syntax?). Study Rust's procedural macro system and Sweet.js, build an AST-aware macro system into the standard library.
- [ ] Can it be possible to do class-based / object oriented or inheritance based programming in ink? How would that work? What would be the "finished" version of the ink language? What about an interface-based programming like Rust (traits) or Go (interfaces)? I think this is better, but this requires a great type system. I think prototype is actually fine if we can do inheritance and dev ergonomics is good. So let's lean towards that, but make this kind of programming model (OO style) possible if it can be done elegantly, because it's a hugely useful mental model.
- [ ] Settle on a way to handle exception conditions / runtime errors
- [ ] We need to settle on a default way of converting the `number` type to the `string` type. This happens all the time when defining composite values and accessing properties.
    - One candidate is to parse integers into `%d` and all other numbers into `%f`.
    - Encapsulate in `ntos(float64) string`


## Standard library / utilities

- [ ] Finish builtin functions in `runtime.go`.
- [ ] Implement event loop as a default multitasking model.
- [ ] JSON serde system


## Bugs

- [ ] Implement the bytes literal (or constructor?) and values
- [ ] Add bitwise binary operators, `&`, `|`, `>>`, `<<`
