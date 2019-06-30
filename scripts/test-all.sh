#!/bin/sh

go run . -input samples/stdlib.ink \
    -input samples/test.ink \
    -input samples/orderofops.ink \
    -input samples/logictest.ink \

echo 'Should say 14:'
go run . -eval "f := n => () => out('say hi ' + string(n)), f(14)()"
echo ''
