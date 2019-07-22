` JSON serde `

std := load('std')

map := std.map
cat := std.cat

` composite to JSON string `
ser := c => type(c) :: {
    '()' -> 'null'
    'string' -> '"' + c + '"'
    'number' -> string(c)
    'boolean' -> string(boolean)
    ` do not serialize functions `
    'function' -> 'null'
    'composite' -> '{' + cat(map(keys(c), k => '"' + k + '":' + ser(c.(k))), ',') + '}'
}

` is this character a numeral digit or .? `
num? := c => c :: {
    '.' -> true
    _ -> 47 < point(c) & point(c) < 58
}
` is the char a whitespace? `
ws? := c => point(c) :: {
    ` hard tab `
    9 -> true
    ` newline `
    10 -> true
    ` carriage return `
    13 -> true
    ` space `
    32 -> true
    _ -> false
}

` reader implementation with internal state for deserialization `
reader := s => (
    state := {
        idx: 0
        ` has there been a parse error? `
        err?: false
    }

    {
        next: () => (
            state.idx := state.idx + 1
            s.(state.idx - 1)
        )
        peek: () => s.(state.idx)
        done?: () => ~(state.idx < len(s))
        err: () => state.err? := true
    }
)

` JSON string to composite `
de := s => (
    ` deserialize null `
    deNull := r => (
        n := r.next
        n() + n() + n() + n() :: {
            'null' -> ()
            _ -> (r.err)()
        }
    )

    ` deserialize string `
    deString := r => (
        n := r.next
        p := r.peek

        ` known to be a '"' `
        n()

        (sub := acc => p() :: {
            () -> (r.err)()
            '\\' -> sub(acc + n())
            '"' -> (
                n()
                acc
            )
            _ -> sub(acc + n())
        })('')
    )

    ` deserialize number `
    deNumber := r => (
        n := r.next
        p := r.peek
        state := {
            ` have we seen a '.' yet? `
            decimal?: false
        }

        result := (sub := acc => num?(p()) :: {
            true -> p() :: {
                '.' -> state.decimal? :: {
                    true -> (r.err)()
                    false -> (
                        state.decimal? := true
                        sub(acc + n())
                    )
                }
                _ -> sub(acc + n())
            }
            false -> (r.err)()
        })('')

        number(result)
    )

    ` deserialize boolean `
    deTrue := r => (
        n := r.next
        n() + n() + n() + n() :: {
            'true' -> true
            _ -> (r.err)()
        }
    )
    deFalse := r => (
        n := r.next
        n() + n() + n() + n() + n() :: {
            'false' -> false
            _ -> (r.err)()
        }
    )

    ` deserialize list `
    deList := r => (
        n := r.next
        p := r.peek

        ` known to be a '[' `
        n()

        ` TODO: check for errors in child parse `
        result := (sub := acc => ws?(p()) :: {
            true -> (
                r.next()
                sub(acc)
            )
            false -> p() :: {
                ']' -> (
                    n()
                    acc
                )
                _ -> (
                    acc.len(acc) := de(r)
                    sub(acc)
                )
            }
        })([])
    )

    ` TODO: deserialize composite `
    deComp := r => (
        n := r.next
        p := r.peek
    )

    ` process next char, not ignoring whitespace `
    next := c => c :: {
        '{' -> (
            state.which := 1
            state.idx := state.idx + 1
            raw.(state.idx) :: {
                '"' -> ()
                _ -> brk()
            }
        )
        _ -> brk()
    }

    ` process next char, ignoring whitespace `
    nextWithWS := () => (
        c := raw.(state.idx)

        ws?(c) :: {
            true -> (
                state.idx := state.idx + 1
                nextWithWs()
            )
            false -> next(c)
        }
    )

    ` trim preceding whitespace `
    s := (sub := s => ws?(s.0) :: {
        true -> sub(slice(s, 1, len(s) - 1))
        false -> s
    })(s)

    ` create a reader and hand off parsing to recursive descent `
    r := reader(s)
    (r.peek)() :: {
        'n' -> deNull(r)
        '"' -> deString(r)
        't' -> deTrue(r)
        'f' -> deFalse(r)
        '[' -> deList(r)
        '{' -> deComp(r)
        _ -> deNumber(r)
    }
)

` tests - TODO: move to test.ink `
obj := {
    a: 'hi'
    b: {c: 'what'}
    d: ['hello', 5, 36.3]
    e: () ` null `
    23: 25.3
}
log := load('std').log
result := ser(obj)
log('serialized ->')
log(result)
log('deserialized ->')
log(de(result))

