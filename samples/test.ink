` ink language test suite,
	built on the suite library for testing `

s := (load('suite').suite)(
	'Ink language and standard library'
)

` short helper functions on the suite `
m := s.mark
t := s.test

m('composite value access')
(
	obj := {
		39: 'clues'
		('ex' + 'pr'): 'ession'
	}

	` when calling a function that's a prop of a composite,
		we need to remember that AccessorOp is just a binary op
		and the function call precedes it in priority `
	obj.fn := () => 'xyz'
	obj.fz := f => (f() + f())

	t((obj.fn)(), 'xyz')
	t((obj.fz)(obj.fn), 'xyzxyz')
	t(obj.nonexistent, ())
	t(obj.39, 'clues')
	t(obj.expr, 'ession')

	` nested composites `
	comp := {list: ['hi', 'hello', {what: 'thing'}]}

	` can't just do comp.list.2.what because
		2.what is not a valid identifier.
		these are some other recommended ways `
	t(comp.list.(2).what, 'thing')
	t((comp.list.2).what, 'thing')
	t((comp.list).(2).what, 'thing')
	t(comp.('li' + 'st').0, 'hi')
	t(comp.list.2, {what: 'thing'})
)

m('function, expression, and lexical scope')
(
	thing := 3
	state := {
		thing: 21
	}
	fn := () => thing := 4
	fn2 := thing => thing := 24
	fn3 := () => (
		state.thing := 100
		thing := ~3
	)

	fn()
	fn2()
	fn3()
	(
		thing := 200
	)

	t(fn(), 4)
	t(fn2(), 24)
	t(fn3(), ~3)
	t(thing, 3)
	t(state.thing, 100)
)

m('match expressions')
(
	x := 'what ' + string(1 + 2 + 3 + 4) :: {
		'what 10' -> 'what 10'
		_ -> '??'
	}
	t(x, 'what 10')

	x := [1, 2, [3, 4, ['thing']], {a: ['b']}]
	t(x, [1, 2, [3, 4, ['thing']], {a: ['b']}])
)

m('accessing properties strangely, accessing nonexistent properties')
(
	t({1: ~1}.1, ~1)
	t(['y', 'z'].1, ('z'))
	t({1: 4.2}.'1', 4.2)
	t({1: 4.2}.(1), 4.2)
	t({1: 'hi'}.(1.0000), 'hi')
	` also: composite parts can be empty `
	t([_, _, 'hix'].('2'), 'hix')
	t(string({test: 4200.00}.('te' + 'st')), '4200')
	` types should be accounted for in equality `
	t(string({test: 4200.00}.('te' + 'st')) = 4200, false)

	dashed := {'test-key': 14}
	t(string(dashed.('test-key')), '14')
)

m('calling functions with mismatched argument length')
(
	tooLong := (a, b, c, d, e) => a + b
	tooShort := (a, b) => a + b

	t(tooLong(1, 2), 3)
	t(tooShort(9, 8, 7, 6, 5, 4, 3), 17)
)

m('argument order of evaluation')
(
	acc := []
	fn := (x, y) => (acc.len(acc) := x, y)

	t(fn(fn(fn(fn('i', '?'), 'h'), 'g'), 'k'), 'k')
	t(acc, ['i', '?', 'h', 'g'])
)

m('empty identifier "_" in arguments and functions')
(
	emptySingle := _ => 'snowman'
	emptyMultiple := (_, a, _, b) => a + b

	t(emptySingle(), 'snowman')
	t(emptyMultiple('bright', 'rain', 'sky', 'bow'), 'rainbow')
)

m('comment syntaxes')
(
	`` t(wrong, wrong)
	` t(wrong, more wrong) `
	t(`hidden` '...thing', '...thing')
	t(len('include `cmt` thing'), 19)
)

m('more complex pattern matching')
(
	t([_, [2, _], 6], [10, [2, 7], 6])
	t({
		hi: 'hello'
		bye: {
			good: 'goodbye'
		}
	}, {
		hi: _
		bye: {
			good: _
		}
	})
	t([_, [2, _], 6, _], [10, [2, 7], 6, 0])
	t({6: 9, 7: _}, {6: _, 7: _})
)

m('order of operations')
(
	t(1 + 2 - 3 + 5 - 3, 2)
	t(1 + 2 * 3 + 5 - 3, 9)
	t(10 - 2 * 16/4 + 3, 5)
	t(3 + (10 - 2) * 4, 35)
	t(1 + 2 + (4 - 2) * 3 - (~1), 10)
	t(1 - ~(10 - 3 * 3), 2)
	t(10 - 2 * 24 % 20 / 8 - 1 + 5 + 10/10, 14)
	t(1 & 5 | 4 ^ 1, (1 & 5) | (4 ^ 1))
	t(1 + 1 & 5 % 3 * 10, (1 + 1) & ((5 % 3) * 10))
)

m('logic composition correctness')
(
	` and `
	t(1 & 4, 0, 'num & num')
	t(2 & 3, 2, 'num & num')
	t(true & true, true, 't & t')
	t(true & false, false, 't & f')
	t(false & true, false, 'f & t')
	t(false & false, false, 'f & f')

	` or `
	t(1 | 4, 5, 'num | num')
	t(2 | 3, 3, 'num | num')
	t(true | true, true, 't | t')
	t(true | false, true, 't | f')
	t(false | true, true, 'f | t')
	t(false | false, false, 'f | f')

	` xor `
	t(2 ^ 7, 5, 'num ^ num')
	t(2 ^ 3, 1, 'num ^ num')
	t(true ^ true, false, 't ^ t')
	t(true ^ false, true, 't ^ f')
	t(false ^ true, true, 'f ^ t')
	t(false ^ false, false, 'f ^ f')
)

m('object keys / list, std.clone')
(
	clone := load('std').clone
	obj := {
		first: 1
		second: 2
		third: 3
	}
	list := ['red', 'green', 'blue']

	ks := {
		first: false
		second: false
		third: false
	}
	ky := keys(obj)
	` keys are allowed to be out of insertion order
		-- composites are unordered maps`
	ks.(ky.0) := true
	ks.(ky.1) := true
	ks.(ky.2) := true
	t([ks.first, ks.second, ks.third], [true, true, true])
	t(len(keys(obj)), 3)

	cobj := clone(obj)
	obj.fourth := 4
	clist := clone(list)
	list.(len(list)) := 'alpha'

	t(len(keys(obj)), 4)
	t(len(keys(cobj)), 3)
	t(len(list), 4)
	t(len(clist), 3)
)

m('composite pass by refernece / mutation check')
(
	clone := load('std').clone

	obj := [1, 2, 3]
	twin := obj ` by reference `
	clone := clone(obj) ` cloned (by value) `

	obj.len(obj) := 4
	obj.len(obj) := 5
	obj.len(obj) := 6

	t(len(obj), 6)
	t(len(twin), 6)
	t(len(clone), 3)
)

m('number & composite/list -> string conversions')
(
	stringList := load('std').stringList

	t(string(3.14), '3.14000000')
	t(string(42), '42')
	t(string(true), 'true')
	t(string(false), 'false')
	t(string('hello'), 'hello')
	t(string([0]), '{0: 0}')

	t(number('3.14'), 3.14)
	t(number('-42'), ~42)
	t(number(true), 1)
	t(number(false), 0)

	t(string({a: 3.14}), '{a: 3.14000000}')
	t(stringList([5, 4, 3, 2, 1]), '[5, 4, 3, 2, 1]')
)

m('function/composite equality checks')
(
	` function equality `
	fn1 := () => (3 + 4, 'hello')
	fn2 := () => (3 + 4, 'hello')

	t(fn1 = fn1, true)
	t(fn1 = fn2, false)

	` composite equality `
	comp1 := {1: 2, hi: '4'}
	comp2 := {1: 2, hi: '4'}
	list1 := [1, 2, 3, 4, 5]
	list2 := [1, 2, 3, 4, 5]

	t(comp1 = comp2, true)
	t(list1 = list2, true)
)

m('type() builtin function')
(
	t(type('hi'), 'string')
	t(type(3.14), 'number')
	t(type([0, 1, 2]), 'composite')
	t(type({hi: 'what'}), 'composite')
	t(type(() => 'hi'), 'function')
	t(type(type), 'function')
	t(type(out), 'function')
	t(type(()), '()')
)

m('stdlib range/slice/append/join functions and stringList')
(
	std := load('std')
	stringList := std.stringList
	sliceList := std.sliceList
	range := std.range
	reverse := std.reverse
	slice := std.slice
	join := std.join

	sl := (l, s, e) => stringList(sliceList(l, s, e))
	list := range(10, ~1, ~1)
	str := 'abracadabra'

	t(sl(list, 0, 5), '[10, 9, 8, 7, 6]')
	t(sl(list, ~5, 2), '[10, 9]')
	t(sl(list, 7, 20), '[3, 2, 1, 0]')
	t(sl(list, 20, 1), '[]')

	` redefine list using range and reverse, to t those `
	list := reverse(range(0, 11, 1))

	t(stringList(join(
		sliceList(list, 0, 5), sliceList(list, 5, 200)
	)), '[10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0]')
	t(stringList(join(
		[1, 2, 3]
		join([4, 5, 6], ['a', 'b', 'c'])
	)), '[1, 2, 3, 4, 5, 6, a, b, c]')

	t(slice(str, 0, 5), 'abrac')
	t(slice(str, ~5, 2), 'ab')
	t(slice(str, 7, 20), 'abra')
	t(slice(str, 20, 1), '')
)

m('ascii <-> char point conversions and string encode/decode')
(
	std := load('std')
	encode := std.encode
	decode := std.decode

	s1 := 'this is a long piece of string
	with weird line
	breaks
	'
	s2 := ''
	s3 := 'AaBbCcDdZzYyXx123456789!@#$%^&*()_+-='

	` note: at this point, we only care about ascii, not full Unicode `
	t(point('a'), 97)
	t(char(65), 'A')
	t(decode(encode(decode(encode(s1)))), s1)
	t(decode(encode(decode(encode(s2)))), s2)
	t(decode(encode(decode(encode(s3)))), s3)
)

m('functional list reducers: map, filter, reduce, each, reverse, join/append')
(
	std := load('std')

	map := std.map
	filter := std.filter
	reduce := std.reduce
	each := std.each
	reverse := std.reverse
	append := std.append
	join := std.join

	list := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

	t(map(list, n => n * n), [1, 4, 9, 16, 25, 36, 49, 64, 81, 100])
	t(filter(list, n => n % 2 = 0), [2, 4, 6, 8, 10])
	t(reduce(list, (acc, n) => acc * n, 1), 3628800)
	t(reverse(list), [10, 9, 8, 7, 6, 5, 4, 3, 2, 1])
	t(join(list, list), [1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
	
	` each doesn't return anything meaningful `
	acc := {
		str: ''
	}
	twice := f => x => (f(x), f(x))
	each(list, twice(n => acc.str := acc.str + string(n)))
	t(acc.str, '1122334455667788991010')

	` append mutates `
	append(list, list)
	t(list, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10])
)

m('std.format -- the standard library formatter / templater')
(
	std := load('std')
	f := std.format
	stringList := std.stringList

	values := {
		first: 'ABC'
		'la' + 'st': 'XYZ'
		thingOne: 1
		thingTwo: stringList([5, 4, 3, 2, 1])
		'magic+eye': 'add_sign'
	}

	t(f('', {}), '')
	t(f('one two {{ first }} four', values), 'one two ABC four')
	t(f('new
	{{ sup }} line', {sup: 42}), 'new
	42 line')
	t(f(
		' {{thingTwo}}+{{ magic+eye }}  '
		values
	), ' [5, 4, 3, 2, 1]+add_sign  ')
	t(f(
		'{{last }} {{ first}} {{ thing One }} {{ thing Two }}'
		values
	), 'XYZ ABC 1 [5, 4, 3, 2, 1]')
	t(
		f('{ {  this is not } {{ thingOne } wut } {{ nonexistent }}', values)
		'{ {  this is not } 1 ()'
	)
)

` end test suite, print result `
(s.end)()
