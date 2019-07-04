` ink fizzbuzz implementation `

log := load('std').log

fb := n => (
	[n % 3, n % 5] :: {
		[0, 0] -> log('FizzBuzz')
		[0, _] -> log('Fizz')
		[_, 0] -> log('Buzz')
		_ -> log(n)
	}
)

fizzbuzzhelper := (n, max) => (
	n :: {
		max -> fb(n)
		_ -> (
			fb(n)
			fizzbuzzhelper(n + 1, max)
		)
	}
)

fizzbuzz := max => fizzbuzzhelper(1, max)

fizzbuzz(100)
