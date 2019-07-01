` implementation of Newton's method to square root `

` higher order function that returns a root finder
	with the given degree of precision threshold `
makeNewtonRoot := threshold => (
	` tail call optimized root finder `
	find := (n, previous) => (
		g := guess(n, previous)
		offset := g * g - n
		offset < threshold :: {
			true -> g
			false -> find(n, g)
		}
	)

	` initial guess is n / 2 `
	n => find(n, n / 2)
)

guess := (target, n) => (n + target / n) / 2

` eight degrees of precision chosen arbitrarily, because
	ink prints numbers to 8 decimal digits`
root := makeNewtonRoot(0.00000001)

out('root of 2 (~1.4142): ' + string(root(2)) + '
')
out('root of 1000 (~31.6): ' + string(root(1000)) + '
')
