section := () => out('
section - -
')

` test weird line breaks ` section()
log :=
(str => (
    out(str)

    out('
')
))
log2 :=
    (str => (

    out(str)
        out('
')
    'log2 text'
))('hilog')
log('what wow')
log(log2)

` test automatic Separator insertion ` section()
kl := [
    5
    4
    3
    2
    1
].2
ol := {
    ('te' +
        '-st'): 'magic'
}.('te-st')
log('should be magic: ' + ol)
log('should be 3: ' + string(kl))

` fibonacci test for concatenated / minified ink code ` section()
fb:=n=>([n%3,n%5]::{[0,0]->log('FizzBuzz'),[0,_]->log('Fizz'),[
_,0]->log('Buzz'),_->log(string(n))}),fizzbuzzhelper:=(n,max)=>
(n::{max->fb(n),_->(fb(n),fizzbuzzhelper(n+1,max))}),(max=>
fizzbuzzhelper(1,max))(18)

` test composite value access` section()
obj := {}
` when calling a function that's a prop of a composite,
    we need to remember that AccessorOp is just a binary op
    and the function call precedes it in priority `
obj.xyz := log, (obj.xyz)('this statement sho' +
    'uld be logged.')
` another case of something similar `
obj:=({a:()=>{xyz:n => (log(n),log(n))}}.a)()
(obj.xyz)('this should be printed twice!')

` composite inside composite inside composite ` section()
comp := {
    list: ['hi', 'hello', {what: 'thing to be printed'}]
}
log('should log thing to be printed:')
out('        -> ')
` can't just do comp.list.2.what because
    2.what is not a valid identifier.
    these are some other recommended ways `
log(comp.list.(2).what)
log('again...')
out('        -> ')
log((comp.list.2).what)

` binary and other complex expressions in match expression ` section()
log('should log hello mac:')
log(
    'what ' + string(1 + 2 + 3 + 4) :: {
        'what 10' -> 'hello mac'
        'what 1234' -> 'wrong answer!'
    }
)

` accessing properties strangely, accessing nonexistent properties ` section()
{}.1
[].1
{}.'1'
[].(1)
log('should say hi:')
log(
    string(
        {1: 'hi'}.(1.0)
    )
)
log('should say hi again:')
log(
    string(
        {1: 'hi again'}.('1')
    )
)
log('should print 4200 here:')
log(string({test: 4200}.('te' + 'st')))
dashed := {
    'test-key': 14
}
log('expect 14:')
log(string(dashed.('test-key')))

` calling functions with mismatched argument length ` section()
tooLong := (a, b, c, d, e) => a + b
log('should be 3:')
log(string(
    tooLong(1, 2)
))
tooShort := (a, b) => a + b
log('should be 5, then 17:')
log(string(
    tooShort(9, 8, 7, 6, 5, 4, log('5'))
))

` EmptyIdentifier in arguments list of functions ` section()
emptySingle := _ => out('snowman ')
emptyMultiple := (_, a, _, b) => out(a + b)
out('should print snowman, then rainbow
-> ')
emptySingle()
emptyMultiple('bright', 'rain', 'sky', 'bow')

section()
`` log('this should NEVER be seen')
`` log('neither should this')
log('this line should be seen.') `` but here's still a comment
log(`right` '... and this line')
