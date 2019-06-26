` fibonacci sequence generator `

` naive implementation `
fib := n => (
	n :: {
		0 -> 0
		1 -> 1
		_ -> fib(n - 1) + fib(n - 2)
	}
)

` memoized / dynamic programming implementation `
memo := {
    0: 0,
    1: 1,
}
fibMemo := n => (
    (memo.(n)) :: {
        null -> (
            memo.(n) := fibMemo(n - 1) + fibMemo(n - 2)
        )
        _ -> memo.(n)
    }
)

log('fib(10) is 55:')
out('Naive solution: '), log(string(fib(10)))
out('Dynamic solution: '), log(string(fibMemo(10)))
