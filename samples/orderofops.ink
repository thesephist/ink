compare := (num, exp) => (
    out(string(num))
    log(' should be ' + string(exp))
)

compare(1 + 2 - 3 + 5 - 3, 2)

compare(1 + 2 * 3 + 5 - 3, 9)

compare(10 - 2 * 16/4 + 3, 5)

compare((10 - 2) * 4, 32)
