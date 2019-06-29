` the ink standard library `

log := str => (
    out(str + '
')
)

scan := () => (
    in()
)

` TODO: JSON serde system `

` TODO: slice(composite, start, end)
        join(composite, composite) (append)
        -> impl for lists `

` TODO: clone(composite) function`

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
