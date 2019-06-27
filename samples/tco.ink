` test for tail call optimization `

` simple test of whether thunks get resolved correctly `
fig := str => out(str)
fig2 := n => (
    fig('hi ')
    fig('hello ' + string(n) + '
')
)
out('should print "hi hello <n>" 3 times:
')
(
    fig2(9)
    28
    (
        fig2(18)
    )
    fig2(27)
)

out('
')

` we defina a fast function that can be TC optimized `
out('testing by pushing stack size ... (test 1/2)
')
testTCO := n => n :: {
    0 -> 'success! we are tail call optimized'
    _ -> testTCO(n - 1)
}

` call it with some very large stack size goal `
` minimum stack size to cause Goroutine overflow
    Golang has max stack size of 1GB, which means
    this leaves <10 bytes per stack if not TCO `
out(testTCO(100000000) + '
')

` second form of TCO -- in ExpressionList `
out('testing by pushing stack size ... (test 2/2)
')
testTCO := n => n :: {
    0 -> 'success! we are tail call optimized'
    _ -> (
        1
        testTCO(n - 1)
    )
}
out(testTCO(100000000) + '
')
