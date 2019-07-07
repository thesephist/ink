#!/bin/sh

go run -race . -input samples/std.ink \
    -input samples/test.ink \
    -input samples/orderofops.ink \
    -input samples/logictest.ink \

go run -race . -input samples/io.ink

echo 'Should say hi 14:'
go run . -eval "f := n => () => out('say hi ' + string(n)), f(14)()"
echo ''
