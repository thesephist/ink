fb := n => (
    [n % 3, n % 5] :: {
        [0, 0] -> out('FizzBuzz')
        [0, _] -> out('Fizz')
        [_, 0] -> out('Buzz')
        _ -> out(string(n))
    }
)

fizzbuzzhelper := (n, max) => (
    n :: {
        max -> fb(n)
        _ -> (
            fb(n)
            fizzbuzzhelp(n + 1, max)
        )
    }
)

fizzbuzz := max => fizbuzzhelper(1, max)

fizzbuzz(100)
