` mangled script for syntax / parser testing `

` test weird line breaks `
log :=
(str => (
	out(str)

	out('
')
))
1 :: ` line break after match `
{1 -> 'hi', 2 ->
	'thing'}
() =>
	() ` line break after arrow `
log2 :=
	(str =>
	(

	out(str)
		out('
')
	'log2 text'
))('hilog')
log(log2)

` test automatic Separator insertion `
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

` redefine log for a more concise fizzbuzz `
log := s => out(string(s) + ' ')

` fibonacci test for concatenated / minified ink code `
fb:=n=>([n%3,n%5]::{[0,0]->log('FizzBuzz'),[0,_]->log('Fizz'),[
_,0]->log('Buzz'),_->log(n)}),fizzbuzzhelper:=(n,max)=>
(n::{max->fb(n),_->(fb(n),fizzbuzzhelper(n+1,max))}),(max=>
fizzbuzzhelper(1,max))(25)

` exit with newline `
out('
')
