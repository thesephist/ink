` test suite for logical and bitwise operators `

log := load('std').log

allpass := [true]

log('Logical operator test failures:')

test := (result, expected, msg) => (
	result :: {
		expected -> ()
		_ -> (
			log('ERROR: ' + msg)
			allpass.0 := false
		)
	}
)

` and `
test(1 & 4, 0, 'num & num')
test(2 & 3, 2, 'num & num')

test(true & true, true, 't & t')
test(true & false, false, 't & f')
test(false & true, false, 'f & t')
test(false & false, false, 'f & f')

` or `
test(1 | 4, 5, 'num | num')
test(2 | 3, 3, 'num | num')

test(true | true, true, 't | t')
test(true | false, true, 't | f')
test(false | true, true, 'f | t')
test(false | false, false, 'f | f')

` xor `
test(2 ^ 7, 5, 'num ^ num')
test(2 ^ 3, 1, 'num ^ num')

test(true ^ true, false, 't ^ t')
test(true ^ false, true, 't ^ f')
test(false ^ true, true, 'f ^ t')
test(false ^ false, false, 'f ^ f')

allpass.0 :: {true -> log('all passed!')}

log('')
