# Ink

Ink is a functional programming language inspired by modern JavaScript.

Ink has a few goals. In order, they are

- Simple, minimal syntax
- High readability and expressiveness
- Small interpreter and runtime API
- Performance (within reason)

## Setup and introduction

## Syntax

Ink's syntax is inspired by JavaScript, but much more minimal.

As a side note here, all values and references are immutable.

```
Program: ExpressionList

Block: '{' ExpressionList '}' | Expression

ExpressionList: (Expression [','])*

Expression: (
    Atom
    | UnaryExpr
    | BinaryExpr
    | FunctionCallExpr
    | MatchExpr
) ','

UnaryExpr: UnaryOp Atom
BinaryExpr: Atom BinaryOp Atom
FunctionCallExpr: Atom '(' ExpressionList ')'
MatchExpr: Atom '::' '{' MatchClause* '}'

MatchClause: Atom '->' Block [',']


Atom: Identifier | Literal

Identifier: (A-Za-z@!?)[A-Za-z0-9@!?]* | _

Literal: NumberLiteral | StringLiteral
        | BooleanLiteral | NullLiteral
        | ObjectLiteral | ListLiteral | FunctionLiteral

NumberLiteral: (0-9)+ ['.' (0-9)*]
StringLiteral: '\'' (Escaped unicode bytes) '\''

BooleanLiteral: 'true' | 'false'
NullLiteral: 'null'

ObjectLiteral: '{' (Identifier ':' Atom ',')* '}'
ListLitereal: '[' (Atom ',')* ']'
FunctionLiteral: Identifier* '=>' Block


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

## Types

Ink is strongly and statically typed, and has six non-extendable types.

- Number
- String
- Boolean
- Null
- Composite (including Objects and Lists)
- Function

## Builtins

### Constants

- `@time`: Millisecond timestamp. By convention, global constants begin with `@`.

### Functions

- `in()`: Read until ENTER key (might change later)
- `out()`
- `read()`: Read from given file descriptor
- `write()`: Write to given file descriptor
- `string()`
- `number()`
- `boolean()`

## Samples

// Fibonacci
```ink
fib := n => {
    n :: {
        0 -> 0
        1 -> 1
        _ -> fib(n - 1) + fib(n - 2)
    }
}
```

// fizzbuzz
```ink
fb := n => {
    [n % 3, n % 5] :: {
        [0, 0] -> out('FizzBuzz')
        [0, _] -> out('Fizz')
        [_, 0] -> out('Buzz')
        _ -> out(string(n))
    }
}
fizzbuzzhelp := n, max => {
    n = max :: {
        true -> fb(n)
        false -> {
            fb(n)
            fizzbuzzhelp(n + 1, max)
        }
    }
}
fizzbuzz := max => {
    fizzbuzzhelp(1, max)
}
fizzbuzz(100)
```

// Reading input
```ink
out('What\'s your name?')
username := in()
out('Your name is ' + username)
```

