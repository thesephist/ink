#!/bin/sh

# run the standard test suite
go run -race . samples/mangled.ink samples/test.ink

# test file IO sample
go run -race .  samples/io.ink

echo 'Should say hi 14:'
go run . -eval "f := n => () => out('say hi ' + string(n)), f(14)()"
echo ''
