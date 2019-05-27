fb := n => {
    [n % 3, n % 5] :: {
        [0, 0] -> out('FizzBuzz')
        [0, _] -> out('Fizz')
        [_, 0] -> out('Buzz')
        _ -> out(string(n))
    }
}
fizzbuzzhelp := (n, max) => {
    (n = max) :: {
        true -> fb(n)
        false -> {
            fb(n)
            fizzbuzzhelp(n + 1, max)
        }
    }
}
fizzbuzz := max => {
    fizzbuzzhelp(1, max)
}
fizzbuzz(100)
