fn1 := n => out('Hello, World!
')

fn2 := () => (
    out('Hello, World 2!
')
)

out('Hello test
')

(
    fn1()
    fn2(1, 2, false)
)
