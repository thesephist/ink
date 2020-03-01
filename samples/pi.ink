` Monte-Carlo estimation of pi using random number generator `

log := load('std').log

` take count from CLI, defaulting to 250k `
Count := (c := number(args().2) :: {
	0 -> 250000
	_ -> c
})

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

	` log progress at 100k intervals `
	iterCount :: {
		Count -> ()
		_ -> iterCount % 100000 :: {
			0 -> log(string(iterCount) + ' runs left, Pi at ' +
				string(4 * state.inCount / (Count - iterCount)))
		}
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
repeatableIteration(Count) `` do Count times

log('Estimate of Pi after ' + string(Count) + ' runs: ' +
	string(4 * state.inCount / Count))
