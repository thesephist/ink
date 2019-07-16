` ink standard test suite tools `

std := load('std')

` borrow from std `
log := std.log
f := std.format

` suite constructor `
suite := label => (
	` suite data store `
	s := {
		all: 0
		passed: 0
		msgs: []
	}

	` mark sections of a test suite with human labels `
	mark := label => s.msgs.(len(s.msgs)) := '- ' + label

	` signal end of test suite, print out results `
	end := () => (
		log(f('suite: {{ label }}', {label: label}))

		max := len(s.msgs)
		(sub := i => i :: {
			max -> ()
			_ -> (
				log('  ' + s.msgs.(i))
				sub(i + 1)
			)
		})(0)

		log('')
		s.passed :: {
			(s.all) -> log(f('ALL {{ passed }} / {{ all }} PASSED', s))
			_ -> log(f('{{ passed }} / {{ all }} PASSED', s))
		}
	)

	` log a passed test `
	onSuccess := () => (
		s.all := s.all + 1
		s.passed := s.passed + 1
	)

	` log a failed test `
	onFail := msg => (
		s.all := s.all + 1
		s.msgs.(len(s.msgs)) := msg
	)

	` perform a new test case `
	test := (result, expected) => result :: {
		expected -> onSuccess()
		_ -> (
			msg := f('  + got {{ result }}' + '
  ' + '  ' + '  exp {{ expected }}', {expected: expected, result: result})
			onFail(msg)
		)
	}
	
	` expose API functions `
	{
		mark: mark
		test: test
		end: end
	}
)
