# Ink programming language 🖋

Ink is a minimal programming language inspired by modern JavaScript and Go, with functional style.

Ink has a few goals. In order, they are

- Ink should have a simple, minimal syntax
- Ink should be easy to learn regardless of skill level
- Ink should be quickly readable and expressive
- Ink should feel productive to use
- Ink should have a great, fully featured, and modular standard library
- Ink should have an ergonomic interpreter and runtime API

Design is always a game of tradeoffs. Ink's goals for minimalism and readability / expressivity means the language deliberately does not aim to be best in other ways:

- Ink doesn't need to be highly efficient or fast, especially compared to compiled languages
- Ink doesn't need to be particularly concise, though we try to avoid verbosity when we can

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

## Overview

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

## Syntax

Ink's syntax is inspired by JavaScript and Go, but strives to be minimal. This is not a comprehensive grammar, but expresses the high level structure.

```
Program: Expression*

Expression: (Atom | BinaryExpr | MatchExpr) ','
ExpressionList: '(' Expression* ')'


Atom: UnaryExpr | EmptyIdentifier
        | Identifier | FunctionCall
        | Literal | ExpressionList

UnaryExpr: UnaryOp Atom

EmptyIdentifier: '_'
Identifier: (A-Za-z@!?)[A-Za-z0-9@!?]*

FunctionCall: Atom ExpressionList

Literal: NumberLiteral | StringLiteral
        | BooleanLiteral | FunctionLiteral
        | ObjectLiteral | ListLiteral

NumberLiteral: (0-9)+ ['.' (0-9)+]
StringLiteral: '\'' (.*) '\''

BooleanLiteral: 'true' | 'false'
FunctionLiteral: (Identifier | '(' (Identifier ',')* ')')
        '=>' ( Expression | ExpressionList )

ObjectLiteral: '{' ObjectEntry* '}'
ObjectEntry: Expression ':' Expression
ListLiteral: '[' Expression* ']'


BinaryExpr: (Atom | BinaryExpr) BinaryOp (Atom | BinaryExpr)


MatchExpr: (Atom | BinaryExpr) '::' '{' MatchClause* '}'
MatchClause: Atom '->' Expression


UnaryOp: (
    '~' // negation
)
BinaryOp: (
    '+' | '-' | '*' | '/' | '%' // arithmetic
    '&' | '|' | '^' // logical and bitwise
    | '>' | '<' // arithmetic comparisons
    | '=' // value comparison operator
    | 'is' // reference comparison operator
    | ':=' // assignment operator
    | '.' // property accessor
)
```

A few quirks of this syntax:

- All variables use lexical binding and scope, and are bound to the most local ExpressionList (execution block)
- Commas (`Separator` tokens) are always required where they are marked in the formal grammar, but the tokenizer inserts commas on newlines if it can be inserted, except after unary and binary operators and after opening delimiters, so few are required after expressions, before closing delimiters, and before the ':' in an Object literal. Here, they are auto-inserted during tokenization.
    - This allows for "minification" of Ink code the same way JavaScript source can be minified. Minified Ink code can be more compact, because in Ink, almost all whitespace is unnecessary (except those wrapping the `is` operator).
- String literals cannot contain comments. Backticks inside string literals are counted as a part of the string literal. String literals are also multiline.
    - This also allows the programmer to comment out a block with an explanation, simply like this:
    ```
    realCode()
    ` this block is commented out for testing reasons
    someOtherCode()
    `
    moreRealCode()
    ```
- List and object property/element access have the same syntax, which is the reference to the list/object followed by the `.` (property access) operator. This means we access array indexes with `arr.1`, `arr.(index + 1)`, etc. and object property with `obj.propName`, `obj.(computed + propName)`, etc.
- Object (dictionary) keys can be arbitrary expressions, including variable names. If the key is a single identifier, the identifier's name will be used as a key in the dict, and if it's not an identifier (a literal, function call, etc.) the value of the expression will be computed and used as the key. This seems like it may cause trouble conceptually, but turns out to be intuitive in practice.
- Assignment is always (re)declaration of a variable in its local scope; this means, for the moment, there is no way to mutate a variable from a parents scope (it'll just shadow the variable in the local scope). I think this is fine, since it forbids a class of potentially confusing state mutations, but I might change my mind in the future and add an assignment-that-isn't-declare. Note that this doesn't affect composite values -- you can mutate objects from a parents scope.
- Ink allows boolean algebra with both logical/bitwise (`&|^`) and algebraic (`+*~`) operators, and which one is used depends on context.

## Types

Ink is strongly but dynamically typed, and has seven non-extendable types.

- Number
- String
- Boolean
- Null
- Composite (including both Objects (dictionaries) and Lists, like Lua tables)
- Function

Composite and Function types are reference-typed, which means assigning a composite to a variable just assigns a reference to the same composite or function value. All other types are value-typed, which means assigning these values to variables will create new copies of those values. i.e.

```
` for simple values `
a := 3, b := a
a := 42

b = 42 `` false, since assignment of values are all copies


` for composite values `
list := [1, 2, 3]
twin := list
clone := clone(list) `` makes a shallow clone

list.(len(list)) := 4 `` append 4 to list
list.(len(list)) := 5 `` append 5 to list

len(list) = 5 `` true
len(twin) = 5 `` true, since it keeps the same reference
len(clone) = 5 `` false, since it keeps a copy of the value instead
```

These are tested in `samples/test.ink`.

## Builtins

### System interfaces

- `in() => string`: Read from stdin or until ENTER key (might change later)
- `out(string)`: Print to stdout
- `read(string, number, number) => list`: Read from given file descriptor from some offset for some bytes
- `write(string, number, list)`: Write to given file descriptor at some offset
- `rand() => number`: a pseudorandom floating point number in interval `[0, 1)`
- `time() => number`: number of seconds in floating point in UNIX epoch

### Math

- `sin(number) => number`: sine
- `cos(number) => number`: cosine
- `pow(number, number) => number`: power, also stands in for finding roots with exponent < 1
- `ln(number) => number`: natural log
- `floor(number) => number`: floor / truncation

### Type casts and utilities (implemented as native functions)

- `string(any) => string`
- `number(any) => number`
- `boolean(any) => boolean`
- `len(composite) => number`: length of a list or list-like composite value

## Development

Ink is currently a single go package. Run `go run .` to run from source, and `go build -ldflags="-s -w"` to build the release binary.

The `ink` binary takes in scripts from standard input, unless at least one `-input` flag is provided, in which case it reads from the filesystem.

### IDE support

Ink currently has a vim syntax definition file, under `utils/ink.vim`. I'm also hoping to support Monaco / VSCode's language definition format soon.
