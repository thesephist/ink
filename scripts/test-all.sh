#!/bin/sh

go run . < samples/test.ink

echo 'Should say 14:'
go run . -eval "f := n => () => out('say hi ' + string(n)), f(14)()"
echo ''
