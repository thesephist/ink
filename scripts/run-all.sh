#!/bin/sh

# for stdin / scan() test
echo 'Linus' | go run . -input samples/stdlib.ink \
    -input samples/fizzbuzz.ink \
    -input samples/graph.ink \
    -input samples/basic.ink \
    -input samples/dict.ink \
    -input samples/kv.ink \
    -input samples/fib.ink \
    -input samples/newton.ink \
    -input samples/callback.ink \
    -input samples/pi.ink \
    -input samples/prime.ink \
    -input samples/quicksort.ink \
    -input samples/mapfilterreduce.ink \
    -input samples/undefinedme.ink \
    -input samples/error.ink \
    -input samples/prompt.ink \
