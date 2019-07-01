` prime sieve `

` is a single number prime? `
isPrime := n => (
	ip := (p, acc) => p :: {
		1 -> acc
		_ -> ip(p - 1, acc & n % p > 0)
	}
	ip(floor(pow(n, 0.5)), true)
)

` build a list of consecutive integers from 2 .. max `
buildConsecutive := max => (
	bc := (i, acc) => (
		i :: {
			(max + 1) -> acc
			_ -> (
				acc.(i - 2) := i
				bc(i + 1, acc)
			)
		}
	)
	bc(2, [])
)

` primes under N are numbers 2 .. N, filtered by isPrime `
getPrimesUnder := n => filter(buildConsecutive(n), isPrime)

ps := getPrimesUnder(1000)
log(stringList(ps))
log('Total number of primes under 1000: ' + string(len(ps)))
