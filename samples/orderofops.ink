` testing that Ink respects order of operations `

log := load('std').log

allpass := [true]

log('Order of operations tests:')

compare := (num, exp) => (
	num :: {
		exp -> ()
		_ -> (
			log('ERROR: ' + string(num) + ' should be ' + string(exp))
			allpass.0 := false
		)
	}
)

compare(1 + 2 - 3 + 5 - 3, 2)

compare(1 + 2 * 3 + 5 - 3, 9)

compare(10 - 2 * 16/4 + 3, 5)

compare(3 + (10 - 2) * 4, 35)

compare(1 + 2 + (4 - 2) * 3 - (~1), 10)

compare(1 - ~(10 - 3 * 3), 2)

compare(10 - 2 * 24 % 20 / 8 - 1 + 5 + 10/10, 14)

compare(1 & 5 | 4 ^ 1, (1 & 5) | (4 ^ 1))

compare(1 + 1 & 5 % 3 * 10, (1 + 1) & ((5 % 3) * 10))

` wrap up tests `
allpass.0 :: {true -> log('all passed!')}

log('')
