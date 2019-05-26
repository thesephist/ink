fizzbuzzhelp := (n, max) => {
    (n = max) :: {
        true -> fb(n)
        false -> {
            fb(n)
            fizzbuzzhelp(n + 1, max)
        }
    }
}
