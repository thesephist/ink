#!/bin/sh

# std.ink is tested through other samples that consume it
#   so no need to specifically test it
go run -race . \
    samples/test.ink \
    samples/orderofops.ink \
    samples/logictest.ink

go run -race .  samples/io.ink

echo 'Should say hi 14:'
go run . -eval "f := n => () => out('say hi ' + string(n)), f(14)()"
echo ''
