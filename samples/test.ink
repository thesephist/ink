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

` fibonacci test for concatenated / minified ink code `
fb:=n=>([n%3,n%5]::{[0,0]->log('FizzBuzz'),[0,_]->log('Fizz'),[
_,0]->log('Buzz'),_->log(string(n))}),fizzbuzzhelper:=(n,max)=>
(n::{max->fb(n),_->(fb(n),fizzbuzzhelper(n+1,max))}),(max=>
fizzbuzzhelper(1,max))(18)
