` let's graph the sine / cosine functions in Ink! `

log := load('std').log

` repeat a string n times `
repeat := (s, n) => n :: {
	0 -> ''
	_ -> s + repeat(s, n - 1)
}

` loop a function F, N times `
loop := (f, n) => n > 0 :: {
	true -> (
		f()
		loop(f, n - 1)
	)
}

` graph a single point `
draw := (x, func, symbol) => (
	n := func(x / 2) + 1
	` some fuzzy math here to make spacing look decent `
	log(repeat(' ', floor(20 * n)) + symbol)
)

` recursively draw from a single value `
drawRec := max => max :: {
	0 -> ()
	_ -> (
		draw(max, sin, '+')
		draw(max, x => cos(x) + 0.7, 'o')

		drawRec(max - 1)
	)
}

` actually draw `
drawRec(30)
