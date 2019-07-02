` prime sieve `

` is a single number prime? `
isPrime := n => (
	` is n coprime with nums < p? `
	max := floor(pow(n, 0.5)) + 1
	(ip := p => p :: {
		max -> true
		_ -> n % p :: {
			0 -> false
			_ -> ip(p + 1)
		}
	})(2) ` start with smaller # = more efficient `
)

` build a list of consecutive integers from 2 .. max `
buildConsecutive := max => (
	peak := max + 1
	acc := []
	(bc := i => i :: {
		peak -> ()
		_ -> (
			acc.(i - 2) := i
			bc(i + 1)
		)
	})(2)
	acc
)

` utility function for printing things `
log := s => out(s + '
')

` primes under N are numbers 2 .. N, filtered by isPrime `
getPrimesUnder := n => filter(buildConsecutive(n), isPrime)

ps := getPrimesUnder(5000)
log(stringList(ps))
log('Total number of primes under 5000: ' + string(len(ps)))
