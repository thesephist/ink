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
out('		-> ')
` can't just do comp.list.2.what because
	2.what is not a valid identifier.
	these are some other recommended ways `
log(comp.list.(2).what)
log('again...')
out('		-> ')
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
log('')

` comment syntaxes ` section()
`` log('this should NEVER be seen')
`` log('neither should this')
log('this line should be seen.') `` but here's still a comment
log(`right` '... and this line')

` more complex pattern matching ` section()
log('expect: true true false false')
log(string([_, [2, _], 6] = [10, [2, 7], 6]))
log(string({
	hi: 'hello'
	bye: {
		good: 'goodbye'
	}
} = {
	hi: _
	bye: {
		good: _
	}
}))
log(string([_, [2, _], 6, _] = [10, [2, 7], 6]))
log(string({6: 9} = {6: _, 7: _}))
log('')

` object keys / list ` section()
log('expect: dict, then keys, then modified and clone')
(
	obj := {
		first: 1
		second: 2
		third: 3
	}
	list := ['red', 'green', 'blue']
	log(string(obj))
	log(string(list))

	log('keys --')
	log(string(keys(obj)))

	cobj := clone(obj)
	obj.fourth := 4
	out('modified: ')
	log(string(obj))
	log(string(cobj))
	log('')
)

` pass by reference / mutation check ` section()
log('checking pass by reference / mutation:')
(
	obj := [1, 2, 3]
	twin := obj
	clone := clone(obj)

	obj.len(obj) := 4
	obj.len(obj) := 5
	obj.len(obj) := 6

	[len(obj), len(twin), len(clone)] :: {
		[6, 6, 3] -> log('passed!')
		_ -> log('ERROR: there is a problem with copying references to objects later modified')
	}
	log('')
)

` number and string conversion tests ` section()
log('checking number and string conversions')
(
	allpassed := [true]
	test := (result, expect) => result = expect :: {
		false -> (
			allpassed.0 := false
			log('expected ' + string(result) + ' to be ' + string(expect))
		)
	}

	test(string(3.14), '3.14000000')
	test(string(42), '42')
	test(string(true), 'true')
	test(string(false), 'false')
	test(string('hello'), 'hello')
	test(string([0]), '{0: 0}')

	test(number('3.14'), 3.14)
	test(number('-42'), ~42)
	test(number(true), 1)
	test(number(false), 0)

	test(stringList([5, 4, 3, 2, 1]), '[5, 4, 3, 2, 1]')
	test(string({a: 3.14}), '{a: 3.14000000}')

	allpassed.0 :: {true -> (
		log('all passed!')
	)}
	log('')
)

section()
log('checking function / composite equality checks correctness')
(
	fn1 := () => (3 + 4, 'hello')
	fn2 := () => (3 + 4, 'hello')

	allpassed := [true]
	test := (result, expect) => result = expect :: {
		false -> (
			allpassed.0 := false
			log('expected ' + string(result) + ' to be ' + string(expect))
		)
	}

	test(fn1 = fn2, false)
	test(fn1 = fn1, true)

	comp1 := {1: 2, hi: '4'}
	comp2 := {1: 2, hi: '4'}

	list1 := [1, 2, 3, 4, 5]
	list2 := [1, 2, 3, 4, 5]

	test(comp1 = comp2, true)
	test(list1 = list2, true)

	allpassed.0 :: {true -> (
		log('all passed!')
	)}
	log('')
)

section()
log('type() builtin')
(
	allpassed := [true]
	test := (result, expect) => result = expect :: {
		false -> (
			allpassed.0 := false
			log('expected ' + string(result) + ' to be ' + string(expect))
		)
	}

	test(type('hi'), 'string')
	test(type(3.14), 'number')
	test(type([0, 1, 2]), 'composite')
	test(type({hi: 'what'}), 'composite')
	test(type(() => 'hi'), 'function')
	test(type(()), '()')

	allpassed.0 :: {true -> (
		log('all passed!')
	)}
	log('')
)

section()
log('stdlib slice functions and stringList')
(
	allpassed := [true]
	test := (result, expect) => result = expect :: {
		false -> (
			allpassed.0 := false
			log('expected ' + string(result) + ' to be ' + string(expect))
		)
	}

	sl := (l, s, e) => stringList(sliceList(l, s, e))
	list := [10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0]
	str := 'abracadabra'

	test(sl(list, 0, 5), '[10, 9, 8, 7, 6]')
	test(sl(list, ~5, 2), '[10, 9]')
	test(sl(list, 7, 20), '[3, 2, 1, 0]')
	test(sl(list, 20, 1), '[]')

	test(slice(str, 0, 5), 'abrac')
	test(slice(str, ~5, 2), 'ab')
	test(slice(str, 7, 20), 'abra')
	test(slice(str, 20, 1), '')

	allpassed.0 :: {true -> (
		log('all passed!')
	)}
	log('')
)
