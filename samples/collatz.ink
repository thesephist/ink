` finding long collatz sequences `

std := load('std')

log := std.log
f := std.format
sList := std.stringList

next := n => n % 2 :: {
	0 -> n / 2
	1 -> 3 * n + 1
}

sequence := start => start < 1 :: {
	true -> []
	false -> (sub := (n, acc) => n :: {
		1 -> acc
		_ -> sub(nx := next(n), acc.len(acc) := nx)
	})(start, [start])
}

longestSequenceUnder := max => (sub := (n, prevMax, prevSeq) => n :: {
	max -> prevSeq
	_ -> (
		currSeq := sequence(n)
		currMax := len(currSeq)
		currMax > prevMax :: {
			true -> sub(n + 1, currMax, currSeq)
			false -> sub(n + 1, prevMax, prevSeq)
		}
	)
})(1, 0, [])

` run a search for longest collatz sequence under Max `
Max := 1000
longest := longestSequenceUnder(Max)
log(f('Longest collatz seq under {{ max }} is {{ len }} items', {
	max: Max
	len: len(longest)
}))
log(sList(longest))
