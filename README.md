# Ink

Ink is a functional programming language inspired by modern JavaScript.

Ink has a few goals. In order, they are

- Simple, minimal syntax
- High readability and expressiveness
- Small efficient interpreter and runtime API
- Performance (within reason)

## Setup and introduction

(This section is to be written as the project matures.)

## Syntax

Ink's syntax is inspired by JavaScript and Go, but much more minimal.

Comments are delimited on both sides with the backtick `\`` character.

```
Program: Expression*

Expression: (Atom | BinaryExpr | MatchExpr) ','

UnaryExpr: UnaryOp Atom
BinaryExpr: Atom BinaryOp Atom
MatchExpr: Atom '::' '{' MatchClause* '}'

MatchClause: Atom '->' Expression


Atom: UnaryExpr | EmptyIdentifier | Identifier | Literal
        | FunctionCall | '(' Expression* ')'

EmptyIdentifier: '_'
Identifier: (A-Za-z@!?)[A-Za-z0-9@!?]* | _

FunctionCall: (Identifier
        | FunctionLiteral
        | FunctionCall
        | '(' Expression* ')') '(' Expression* ')'

Literal: NumberLiteral | StringLiteral
        | BooleanLiteral | NullLiteral
        | ObjectLiteral | ListLiteral | FunctionLiteral

NumberLiteral: (0-9)+ ['.' (0-9)*]
StringLiteral: '\'' (.*) '\''

BooleanLiteral: 'true' | 'false'
NullLiteral: 'null'

ObjectLiteral: '{' ObjectEntry* '}'
ObjectEntry: Expression ':' Expression
ListLiteral: '[' Expression* ']'
FunctionLiteral: (Identifier | '(' (Identifier ',')* ')')
        '=>' Expression 

UnaryOp: (
    '~' // negation
)
BinaryOp: (
    '+' | '-' | '*' | '/' | '%' // arithmetic
    | '>' | '<' // arithmetic comparisons
    | '=' // value comparison operator
    | 'is' // reference comparison operator
    | ':=' // assignment operator
    | '.' // property accessor
)
```

A few quirks of this syntax:

- All variables use lexical binding and scope, and are bound to the block
- Commas (`Separator` tokens) are always required where they are marked in the formal grammar, but the tokenizer inserts commas on newlines if it can be inserted, except after binary operators, so few are required after expressions, before delimiters, and before the ':' in an Object literal. Here, they are auto-inserted during tokenization.
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

## Types

Ink is strongly and statically typed, and has seven non-extendable types.

- Number
- String
- Bytes
- Boolean
- Null
- Composite (including Objects and Lists)
- Function

## Builtins

### Constants

- `pi`: Millisecond timestamp. By convention, global constants begin with `@`.

### Functions

- `in() => string`: Read from stdin or until ENTER key (might change later)
- `out(string)`: Print to stdout
- `read(string, number, number) => bytes`: Read from given file descriptor from some offset for some bytes
- `write(string, number, bytes)`: Write to given file descriptor at some offset
- `time() => number`: Current millisecond (since UNIX epoch) timestamp

### Math

- `sin(number) => number`: sine
- `cos(number) => number`: cosine
- `ln(number) => number`: natural log

### Type casts (implemented as functions)

- `string(any) => string`
- `number(any) => number`
- `bytes(any) => bytes`
- `boolean(any) => boolean`

## Samples

You can find up-to-date code samples in `samples/`.

