#!/bin/sh

# for stdin / scan() test
# std.ink is tested through other samples that consume it
#   so no need to specifically test it
echo 'Linus' | go run -race . \
    samples/fizzbuzz.ink \
    samples/graph.ink \
    samples/basic.ink \
    samples/dict.ink \
    samples/kv.ink \
    samples/fib.ink \
    samples/newton.ink \
    samples/callback.ink \
    samples/pi.ink \
    samples/prime.ink \
    samples/quicksort.ink \
    samples/mapfilterreduce.ink \
    samples/pingpong.ink \
    samples/undefinedme.ink \
    samples/error.ink \
    samples/prompt.ink
