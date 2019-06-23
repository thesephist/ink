fib := n => (
	n :: {
		0 -> 0
		1 -> 1
		_ -> fib(n - 1) + fib(n - 2)
	}
)

out('Will log 10: ')
log(string(10))

log(string(fib(12)))
