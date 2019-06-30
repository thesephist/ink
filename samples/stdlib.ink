` the ink standard library `

log := str => (
    out(str + '
')
)

scan := () => (
    in()
)

` TODO: slice(composite, start, end)
        join(composite, composite) (append)
        -> impl for lists `

` TODO: clone(composite) function`
clone := comp => (
    reduce(keys(comp), (acc, k) => (
        acc.(k) := comp.(k)
        acc
    ), {})
)

` tail recursive reversing a list `
reverse := list => (
    state := [len(list) - 1]
    reduce(list, (acc, item) => (
        acc.(state.0) := item
        state.0 := state.0 - 1
        acc
    ), {})
)

` tail recursive map `
map := (list, f) => (
    reduce(list, (l, item) => (
        l.(len(l)) := f(item)
        l
    ), {})
)

` tail recursive filter `
filter := (list, f) => (
    reduce(list, (l, item) => (
        f(item) :: {
            true -> l.(len(l)) := item
        }
        l
    ), {})
)

` tail recursive reduce `
reduce := (list, f, acc) => (
    (reducesub := (idx, acc) => (
        idx :: {
            len(list) -> acc
            _ -> reducesub(
                idx + 1
                f(acc, list.(idx))
            )
        }
    )
    )(0, acc)
)
