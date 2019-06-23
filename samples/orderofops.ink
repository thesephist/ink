compare := (num, exp) => (
    log(string(num) + ' should be ' + string(exp))
)

compare(1 + 2 - 3 + 5 - 3, 2)

compare(1 + 2 * 3 + 5 - 3, 9)

compare(10 - 2 * 16/4 + 3, 5)

compare(3 + (10 - 2) * 4, 35)

compare(1 + 2 + (4 - 2) * 3 - (~1), 10)

compare(1 - ~(10 - 3 * 3), 2)

compare(10 - 2 * 24 % 20 / 8 - 1 + 5 + 10/10, 14)
