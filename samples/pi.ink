` Monte-Carlo estimation of pi using random number generator `

COUNT := 50000

` pick a random point in [0, 1) in x and y `
randCoord := () => [rand(), rand()]

sqrt := n => pow(n, 0.5)
inCircle := coordPair => (
	` is a given point in a quarter-circle at the origin? `
	x := coordPair.0
	y := coordPair.1

	sqrt(x * x + y * y) < 1
)

` a single iteration of the Monte Carlo simulation `
iteration := iterCount => (
	inCircle(randCoord()) :: {
		true -> state.inCount := state.inCount + 1
	}

	iterCount % 5000 :: {
		1 -> log(string(iterCount) + ' runs left, Pi at ' +
			string(4 * state.inCount / (COUNT - iterCount)))
	}
)

` composable higher order function for looping `
loop := f => (
	iter := n => n :: {
		0 -> ()
		_ -> (
			f(n)
			iter(n - 1)
		)
	}
)

` initial state `
state := {
	inCount: 0
}

` estimation routine `
repeatableIteration := loop(iteration)
repeatableIteration(COUNT) `` do COUNT times

log('Estimate of Pi after ' + string(COUNT) + ' runs: ' +
	string(4 * state.inCount / COUNT))
