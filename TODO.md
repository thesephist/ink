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
- [ ] Move language specification things into SPEC.md, language documentation into docs/


## Language

- [ ] A procedural macro system to extend the language (syntax?). Study Rust's procedural macro system and Sweet.js, build an AST-aware macro system into the standard library.
- [ ] Can it be possible to do class-based / object oriented or inheritance based programming in ink? How would that work? What would be the "finished" version of the ink language? What about an interface-based programming like Rust (traits) or Go (interfaces)? I think this is better, but this requires a great type system. I think prototype is actually fine if we can do inheritance and dev ergonomics is good. So let's lean towards that, but make this kind of programming model (OO style) possible if it can be done elegantly, because it's a hugely useful mental model.
- [ ] Settle on a way to handle exception conditions / runtime errors


## Standard library / utilities

- [ ] Finish builtin functions in `runtime.go`.
- [ ] Implement event loop as a default multitasking model.
- [ ] JSON serde system
- [ ] Impl streams / channels / reactive-across-time primitives for programming in the standard library


## Bugs

- [ ] Implement the bytes literal (or constructor?) and values
- [ ] Add bitwise binary operators, `&`, `|`, `>>`, `<<`
- [ ] There are parts in the parser where I just move `idx++` through the tokens stream without checking that the token I passed through is the one I assumed it was (usually closing delimiters). Fix these and check all cases.
