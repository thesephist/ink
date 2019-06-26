` basic key-value storage library
    built on composite values`

makeGet := store => (
    key => store.(key)
)

makeSet := store => (
    (key, val) => store.(key) := val
)

makeDelete := store => (
    key => store.(key) := ()
)

create := () => (
    store := {}

    {
        type: 'kv-store',
        store: store,
        get: makeGet(store),
        set: makeSet(store),
        delete: makeDelete(store),
    }
)

` test `
s := create()
(s.set)('hi', 'value')
out('expect: value --> ')
log((s.get)('hi'))

(s.delete)('hi')
out('expect: null --> ')
(s.get)('hi') :: {
    () ->
        log('null')
    _ ->
        log('not null... it\'s broken')
}
