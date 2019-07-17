` stress tests for tail call optimization `

log := load('std').log

` recursion tests need some very large stack size goal.
	Minimum stack size to cause Goroutine overflow:
	Golang has max stack size of 1GB, which means
	this leaves < 10 bytes per stack frame if not TCO `
RECLIMIT := 100 * 1000 * 1000

success := () => log('-> success! we are tail call optimized')

` we defina a fast function that can be TC optimized.
	this also uses a form of TCO in match expressions `
log('testing simple recursion ... (test 1/3)')
testTCO := n => n :: {
	0 -> success()
	_ -> testTCO(n - 1)
}
testTCO(RECLIMIT)

` second form of TCO -- in expression lists `
log('testing expr list recursion ... (test 2/3)')
testTCO2 := n => n :: {
	0 -> success()
	_ -> (
		1
		testTCO(n - 1)
	)
}
testTCO2(RECLIMIT)

` third form of TCO -- mutual recursion `
log('testing mutual recursion ... (test 3/3)')
testA := n => n :: {
	0 -> success()
	_ -> testB(n - 1)
}
testB := m => m :: {
	0 -> success()
	_ -> (
		3
		testA(m - 1)
	)
}
testA(RECLIMIT)
