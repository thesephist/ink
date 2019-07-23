` ink language test suite,
	built on the suite library for testing `

s := (load('suite').suite)(
	'Ink language and standard library'
)

` short helper functions on the suite `
m := s.mark
t := s.test

` load std once for all tests `
std := load('std')

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
	obj.fz := f => f() + f()

	t((obj.fn)(), 'xyz')
	t((obj.('fn'))(), 'xyz')
	t((obj.fz)(obj.fn), 'xyzxyz')
	t(obj.nonexistent, ())
	t(obj.(~10), ())
	t(obj.39, 'clues')
	t(obj.expr, 'ession')

	` string index access `
	t(('hello').0, 'h')
	t(('what').3, 't')
	t(('hi').(~1), ())
	t(('hello, world!').len('hello, world!'), ())

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

m('tail call optimizations and thunk unwrap order')
(
	acc := ['']

	appender := prefix => str => acc.0 := acc.0 + prefix + str
	f1 := appender('f1_')
	f2 := appender('f2_')

	sub := () => (
		f1('hi')
		(
			f2('what')
		)
		f3 := () => (
			f2('hg')
			f1('bb')
		)
		f1('sup')
		f2('sample')
		f3()
		f2('xyz')
	)

	sub()

	t(acc.0, 'f1_hif2_whatf1_supf2_samplef2_hgf1_bbf2_xyz')
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
	t(string({test: 4200.00}.('te' + 'st')).1, '2')
	t(string({test: 4200.00}.('te' + 'st')).10, ())
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

m('object keys / list, mutable strings, std.clone')
(
	clone := std.clone
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
	list.len(list) := 'alpha'

	t(len(keys(obj)), 4)
	t(len(keys(cobj)), 3)
	t(len(list), 4)
	t(len(clist), 3)

	` len() should count the number of keys on a composite,
		not just integer indexes like ECMAScript `
	t(len({
		0: 1
		1: 'order'
		2: 'natural'
	}), 3)
	t(len({
		hi: 'h'
		hello: 'he'
		thing: 'th'
		what: 'w'
	}), 4)
	t(len({
		0: 'hi'
		'1': 100
		3: 'x'
		5: []
		'word': 0
	}), 5)

	str := 'hello'
	twin := str
	ccpy := str + '' ` should yield a new copy `
	tcpy := '' + twin
	copy := clone(str)
	str.2 := 'xx'
	copy.2 := 'yy'

	t(str, 'hexxo')
	t(twin, 'hexxo')
	t(ccpy, 'hello')
	t(tcpy, 'hello')
	t(copy, 'heyyo')
)

m('string/composite pass by reference / mutation check')
(
	clone := std.clone

	obj := [1, 2, 3]
	twin := obj ` by reference `
	clone := clone(obj) ` cloned (by value) `

	obj.len(obj) := 4
	obj.len(obj) := 5
	obj.len(obj) := 6

	t(len(obj), 6)
	t(len(twin), 6)
	t(len(clone), 3)

	t(clone.hi := 'x', {
		0: 1
		1: 2
		2: 3
		hi: 'x'
	})

	str := 'hello, world'
	str2 := '' + str
	str.5 := '!'
	str.8 := 'lx'
	str.2 := ''
	str.3 := ''
	str.len(str) := 'x?'

	t(str, 'hello! wlxldx?')
	t(str2, 'hello, world')

	str := 'hi'
	t(str.1 := 'what', 'hwhat')
	t(str, 'hwhat')

	str := '00000000'
	mut := (i, s) => (
		str.(i) := s
	)
	mut(4, 'AAA')
	mut(8, 'YYY')
	t(str, '0000AAA0YYY')
)

m('number & composite/list -> string conversions')
(
	stringList := std.stringList

	t(string(3.14), '3.14000000')
	t(string(42), '42')
	t(string(true), 'true')
	t(string(false), 'false')
	t(string('hello'), 'hello')
	t(string([0]), '{0: 0}')

	t(number('3.14'), 3.14)
	t(number('03.140000'), 3.14)
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
	complist := {0: 1, 1: 2, 2: 3, 3: 4, 4: 5}

	t(comp1 = comp2, true)
	t(list1 = list2, true)
	t(comp1 = list1, false)
	t(comp1 = {1: '4', 2: 2}, false)
	t(comp1 = {}, false)
	t(list1 = [1, 2, 3], false)
	t(list1 = complist, true)
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

m('stdlib range/slice/append/join/cat functions and stringList')
(
	stringList := std.stringList
	sliceList := std.sliceList
	range := std.range
	reverse := std.reverse
	slice := std.slice
	join := std.join
	cat := std.cat

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

	t(cat([], '--'), '')
	t(cat(['hello'], '--'), 'hello')
	t(cat(['hello', 'world'], '--'), 'hello--world')
	t(cat(['hello', 'world,hi'], ','), 'hello,world,hi')
	t(cat(['good', 'bye', 'friend'], ''), 'goodbyefriend')
	t(cat(['good', 'bye', 'friend'], ', '), 'good, bye, friend')
	t(cat({
		0: 'first'
		1: 'last'
	}, ' and '), 'first and last')
)

m('hexadecimal conversions, hex & xeh')
(
	hex := std.hex
	xeh := std.xeh

	` base cases `
	t(hex(0), '0')
	t(hex(66), '42')
	t(hex(256), '100')
	t(hex(1998), '7ce')
	t(hex(3141592), '2fefd8')
	t(xeh('fff'), 4095)
	t(xeh('a2'), 162)

	` hex should floor non-integer inputs `
	t(hex(16.8), '10')
	t(hex(1998.123), '7ce')

	` recoverability `
	t(xeh(hex(390420)), 390420)
	t(xeh(hex(9230423903)), 9230423903)
	t(hex(xeh('fffab123')), 'fffab123')
	t(hex(xeh('0000ab99ff33')), 'ab99ff33')
	t(hex(xeh(hex(xeh('aabbef')))), 'aabbef')
	t(xeh(hex(xeh(hex(201900123)))), 201900123)
)

m('ascii <-> char point conversions and string encode/decode')
(
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
	t(encode('ab'), [97, 98])
	t(decode([65, 67, 66]), 'ACB')
	t(decode(encode(decode(encode(s1)))), s1)
	t(decode(encode(decode(encode(s2)))), s2)
	t(decode(encode(decode(encode(s3)))), s3)
)

m('functional list reducers: map, filter, reduce, each, reverse, join/append')
(
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

m('json ser/de')
(
	clone := std.clone

	json := load('json')
	ser := json.ser
	de := json.de

	` primitives `
	t(ser(()), 'null')
	t(ser(''), '""')
	t(ser('world'), '"world"')
	t(ser('es"c a"pe
me'), '"es\\"c a\\"pe\\nme"')
	t(ser(true), 'true')
	t(ser(false), 'false')
	t(ser(12), '12')
	t(ser(3.14), string(3.14))
	t(ser(~2.4142), string(~2.4142))
	t(ser(x => x), 'null')
	t(ser({}), '{}')
	t(ser([]), '{}')

	t(de('null'), ())
	t(de('neh'), ())
	t(de('true'), true)
	t(de('trxx'), ())
	t(de('false'), false)
	t(de('fah'), ())
	t(de('true_32'), true)
	t(de('"thing"'), 'thing')
	t(de('"es\\"c a\\"pe\\nme"'), 'es"c a"pe
me')
	t(de('""'), '')
	t(de('"my"what"'), 'my')
	t(de('-59.413'), ~59.413)
	t(de('10-14.2'), 10)
	t(de('1.2.3'), ())
	t(de('[50, -100]'), [50, ~100])

	` strange whitespace, commas, broken input `
	t(de('	" string"	 '), ' string')
	t(de('   6.'), 6)
	t(de(' .90'), 0.9)
	t(de('   ["first", 2, true, ]	'), ['first', 2, true])
	t(de('"start '), ())
	t(de('{"a": b, "12": 3.41}'), ())
	t(de('{"a": b  "12": 3.41 '), ())
	t(de('[1, 2  3.24.253, fals}'), ())
	t(de('[1, 2, 3.24.253, false'), ())

	` serialize light object `
	s := ser({a: 'b', c: ~4.251})
	first := '{"a":"b","c":-4.25100000}'
	second := '{"c":-4.25100000,"a":"b"}'
	t(s = first | s = second, true)

	s := ser([2, false])
	first := '{"0":2,"1":false}'
	second := '{"1":false,"0":2}'
	t(s = first | s = second, true)

	` complex serde `
	obj := {
		ser: 'de'
		apple: 'dessert'
		func: () => ()
		x: ['train', false, 'car', true, {x: ['y', 'z']}]
		32: 'thirty-two'
		nothing: ()
	}
	objr := clone(obj)
	objr.func := ()
	list := ['a', true, {c: 'd', e: 32.14}, ['f', {}, (), ~42]]
	t(de(ser(obj)), objr)
	t(de(ser(list)), list)

	list.1 := obj
	listr := clone(list)
	listr.1 := objr
	t(de(ser(de(ser(list)))), listr)
)

` end test suite, print result `
(s.end)()
