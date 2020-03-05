#!/usr/bin/env ink
clear := '__cleared'

` ink language test suite,
	built on the suite library for testing `

s := (load('suite').suite)(
	'Ink language and standard library'
)

` short helper functions on the suite `
m := s.mark
t := s.test

` load std & str once for all tests `
std := load('std')
str := load('str')

m('eval with #!/usr/bin/env ink')
(
	` check that the line immediately following #!/... still runs okay `
	t('eval with #!/usr/bin/env ink does not miss lines', clear, '__cleared')
)

m('value equality')
(
	` with primitives `
	t('() = ()', () = (), true)
	t('() = bool', () = false, false)
	t('number = number', 1 = 1.000, true)
	t('number = number', 100 = 1000, false)
	t('empty string = empty string', '' = '', true)
	t('string = string', 'val' = 'val', true)
	t('string = string', '' = 'empty', false)
	t('number = string', '23' = 23, false)
	t('string = number', 23 = '23', false)
	t('bool = bool', false = false, true)
	t('bool = bool', true = false, false)
	t('list = list', ['first', '_second'] = ['first', '_second'], true)
	t('list = list', ['first', '_second'] = ['first', '_second', '*third'], false)
	t('composite = composite', {} = {}, true)
	t('composite = list', {} = [], true)
	t('composite = ()', {} = (), false)

	fn := () => 1
	fn2 := () => 1
	t('function = function', fn = fn, true)
	t('function = function', fn = fn2, false)
	t('builtin fn = builtin fn', len = len, true)
	t('builtin fn = builtin fn', len = string, false)

	` to empty identifier `
	t('_ = _', _ = _, true)
	t('bool = _', true = _, true)
	t('_ = bool', _ = false, true)
	t('number = _', 0 = _, true)
	t('_ = number', _ = 3, true)
	t('string = _', '' = _, true)
	t('_ = string', _ = '', true)
	t('() = _', () = _, true)
	t('_ = ()', _ = (), true)
	t('composite = _', {} = _, true)
	t('_ = composite', _ = {}, true)
	t('_ = list', _ = [_], true)
	t('function = _', (() => ()) = _, true)
	t('_ = function', _ = (() => ()), true)
	t('builtin fn = _', len = _, true)
	t('_ = builtin fn', _ = len, true)
)

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

	t('calling composite property', (obj.fn)(), 'xyz')
	t('composite property by string value', (obj.('fn'))(), 'xyz')
	t('composite property by property access', (obj.fz)(obj.fn), 'xyzxyz')
	t('nonexistent composite key is ()', obj.nonexistent, ())
	t('composite property by number value', obj.(~10), ())
	t('composite property by number literal', obj.39, 'clues')
	t('composite property by identifier', obj.expr, 'ession')

	` string index access `
	t('string index access at 0', ('hello').0, 'h')
	t('string index access', ('what').3, 't')
	t('out of bounds string index access (negative)'
		('hi').(~1), ())
	t('out of bounds string index access (too large)'
		('hello, world!').len('hello, world!'), ())

	` nested composites `
	comp := {list: ['hi', 'hello', {what: 'thing'}]}

	` can't just do comp.list.2.what because
		2.what is not a valid identifier.
		these are some other recommended ways `
	t('nested composite value access with number value'
		comp.list.(2).what, 'thing')
	t('nested composite value access with string value'
		comp.list.('2').what, 'thing')
	t('nested composite value access, parenthesized'
		(comp.list.2).what, 'thing')
	t('nested composite value access, double-parenthesized'
		(comp.list).(2).what, 'thing')
	t('string at index in computed string', comp.('li' + 'st').0, 'hi')
	t('nested property access returns composite', comp.list.2, {what: 'thing'})

	` modifying composite in chained accesses `
	comp.list.4 := 'oom'
	comp.list.(2).what := 'arg'

	t('modifying composite at key leaves others unchanged', comp.list.4, 'oom')
	t('modifying composite at key', comp.list.(2).what, 'arg')
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

	t('function body forms a new scope', fn(), 4)
	t('function body forms a new scope, assignment', fn2(), 24)
	t('function body with expression list forms a new scope', fn3(), ~3)
	t('assignment in child frames are isolated', thing, 3)
	t('modifying composites in scope from child frames causes mutation', state.thing, 100)
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

	t('tail optimized thunks are unwrapped in correct order'
		acc.0, 'f1_hif2_whatf1_supf2_samplef2_hgf1_bbf2_xyz')
)

m('match expressions')
(
	x := ('what ' + string(1 + 2 + 3 + 4) :: {
		'what 10' -> 'what 10'
		_ -> '??'
	})
	t('match expression follows matched clause', x, 'what 10')

	x := ('what ' + string(1 + 2 + 3 + 4) :: {
		'what 11' -> 'what 11'
		_ -> '??'
	})
	t('match expression follows through to empty identifier', x, '??')

	x := [1, 2, [3, 4, ['thing']], {a: ['b']}]
	t('composite deep equality after match expression'
		x, [1, 2, [3, 4, ['thing']], {a: ['b']}])
)

m('accessing properties strangely, accessing nonexistent properties')
(
	t('property access with number literal', {1: ~1}.1, ~1)
	t('list access with number literal', ['y', 'z'].1, ('z'))
	t('property access with bare string literal', {1: 4.2}.'1', 4.2)
	t('property access with number value', {1: 4.2}.(1), 4.2)
	t('property access with decimal number value', {1: 'hi'}.(1.0000), 'hi')

	` also: composite parts can be empty `
	t('composite parts can be empty', [_, _, 'hix'].('2'), 'hix')
	t('property access with computed string'
		string({test: 4200.00}.('te' + 'st')), '4200')
	t('nested property access with computed string'
		string({test: 4200.00}.('te' + 'st')).1, '2')
	t('nested property access with computed string, nonexistent key'
		string({test: 4200.00}.('te' + 'st')).10, ())

	dashed := {'test-key': 14}
	t('property access with string literal that is not valid identifier'
		string(dashed.('test-key')), '14')
)

m('calling functions with mismatched argument length')
(
	tooLong := (a, b, c, d, e) => a + b
	tooShort := (a, b) => a + b

	t('function call with too few arguments', tooLong(1, 2), 3)
	t('function call with too many arguments', tooShort(9, 8, 7, 6, 5, 4, 3), 17)
)

m('argument order of evaluation')
(
	acc := []
	fn := (x, y) => (acc.len(acc) := x, y)

	t('function arguments are evaluated in order, I'
		fn(fn(fn(fn('i', '?'), 'h'), 'g'), 'k'), 'k')
	t('function arguments are evaluated in order, II'
		acc, ['i', '?', 'h', 'g'])
)

m('empty identifier "_" in arguments and functions')
(
	emptySingle := _ => 'snowman'
	emptyMultiple := (_, a, _, b) => a + b

	t('_ is a valid argument placeholder', emptySingle(), 'snowman')
	t('_ can be used multiple times in a single function as argument placeholders'
		emptyMultiple('bright', 'rain', 'sky', 'bow'), 'rainbow')
)

m('comment syntaxes')
(
	`` t(wrong, wrong)
	ping := 'pong'
	` t(wrong, more wrong) `
	t('single line (line-lead) comments are recognized', ping, 'pong')
	t('inline comments are recognized', `hidden` '...thing', '...thing')
	t('inline comments terminate correctly', len('include `cmt` thing'), 19)
)

m('more complex pattern matching')
(
	t('nested list pattern matches correctly', [_, [2, _], 6], [10, [2, 7], 6])
	t('nested composite pattern matches correctly', {
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
	t('nested list pattern matches with empty identifiers', [_, [2, _], 6, _], [10, [2, 7], 6, _])
	t('composite pattern matches with empty identifiers', {6: 9, 7: _}, {6: _, 7: _})
)

m('order of operations')
(
	t('addition/subtraction', 1 + 2 - 3 + 5 - 3, 2)
	t('multiplication over addition/subtraction', 1 + 2 * 3 + 5 - 3, 9)
	t('multiplication/division', 10 - 2 * 16/4 + 3, 5)
	t('parentheses', 3 + (10 - 2) * 4, 35)
	t('parentheses and negation', 1 + 2 + (4 - 2) * 3 - (~1), 10)
	t('negating parenthesized expressions', 1 - ~(10 - 3 * 3), 2)
	t('modulus in bare expressions', 10 - 2 * 24 % 20 / 8 - 1 + 5 + 10/10, 14)
	t('logical operators', 1 & 5 | 4 ^ 1, (1 & 5) | (4 ^ 1))
	t('logical operators, arithmetic, and parentheses', 1 + 1 & 5 % 3 * 10, (1 + 1) & ((5 % 3) * 10))
)

m('string lexicographical comparisons')
(
	t('less-than, I','a' < 'b', true)
	t('less-than, II','x' < 'A', false)
	` shorter strings are lesser `
	t('less-than, III','x long str' < 'A', false)

	t('greater-than, I', 'E' > 'A', true)
	t('greater-than, II', '0' > '_', false)
	t('greater-than, III', 'xxxxx' > 'xxx', true)

	t('empty strings', '' < ' ', true)
	t('uppercase < lowercase', 'S' < 's', true)
	t('non-printed byte arrays', char(253) > char(252), true)
)

m('min/max')
(
	min := std.min
	max := std.max

	t('min of list of 1', min([~30]), ~30)
	t('minimum of list', min([39, 254, 5, ~2, 0, 3]), ~2)

	t('max of list of 1', max([101]), 101)
	t('maximum of list', max([39, 254, 5, ~2, 0, 3]), 254)

	t('min of array of same', min([2, 2, 2, 2, 2, 2, 2, 2]), 2)
	t('max of array of same', min([2, 2, 2, 2, 2, 2, 2, 2]), 2)
)

m('logic composition correctness, std.some/std.every')
(
	` and `
	t('number & number, I', 1 & 4, 0)
	t('number & number, II', 2 & 3, 2)
	t('t & t', true & true, true)
	t('t & f', true & false, false)
	t('f & t', false & true, false)
	t('f & f', false & false, false)

	` or `
	t('number | number, I', 1 | 4, 5)
	t('number | number, II', 2 | 3, 3)
	t('t | t', true | true, true)
	t('t | f', true | false, true)
	t('f | t', false | true, true)
	t('f | f', false | false, false)

	` xor `
	t('number ^ number, I', 2 ^ 7, 5)
	t('number ^ number, II', 2 ^ 3, 1)
	t('t ^ t', true ^ true, false)
	t('t ^ f', true ^ false, true)
	t('f ^ t', false ^ true, true)
	t('f ^ f', false ^ false, false)

	` std.some and std.every `
	some := std.some
	every := std.every

	t('std.some() of empty list is false', some([]), false)
	t('std.every() of empty list is true', every([]), true)
	t('std.some() is true if at least one in list is true'
		some([false, false, true, false]), true)
	t('std.some() is false if none in list is true'
		some([false, false, false, false]), false)
	t('std.every() is true if all in list is true'
		every([true, true, true, true, true]), true)
	t('std.every() is false if at least one in list is false'
		every([true, true, true, false, true]), false)
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
	t('keys() builtin for composite returns keys'
		[ks.first, ks.second, ks.third], [true, true, true])
	t('keys() builtin for composite returns all keys'
		len(keys(obj)), 3)

	cobj := clone(obj)
	obj.fourth := 4
	clist := clone(list)
	list.len(list) := 'alpha'

	t('std.clone does not affect original composite', len(keys(obj)), 4)
	t('std.clone creates a new copy of composite', len(keys(cobj)), 3)
	t('std.clone does not affect original list', len(list), 4)
	t('std.clone creates a new copy of list', len(clist), 3)

	` len() should count the number of keys on a composite,
		not just integer indexes like ECMAScript `
	t('len() builtin on manually indexed composite', len({
		0: 1
		1: 'order'
		2: 'natural'
	}), 3)
	t('len() builtin on non-number keyed composite', len({
		hi: 'h'
		hello: 'he'
		thing: 'th'
		what: 'w'
	}), 4)
	t('len() builtin counts non-consecutive integer keys', len({
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

	t('std.clone does not affect original string', str, 'hexxo')
	t('define op does not create a copy of string', twin, 'hexxo')
	t('concatenation via + creates a copy of string, I', ccpy, 'hello')
	t('concatenation via + creates a copy of string, II', tcpy, 'hello')
	t('std.clone creates a copy of string', copy, 'heyyo')
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

	t('std.clone does not affect original composite', len(obj), 6)
	t('define op does not create a copy of composite', len(twin), 6)
	t('std.clone creates a copy of composite', len(clone), 3)

	t('assignment to composite key returns composite itself, updated', clone.hi := 'x', {
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

	t('assigning to string indexes mutates original string', str, 'hello! wlxldx?')
	t('concatenating to string with + creates a copy of string', str2, 'hello, world')

	str := 'hi'
	t('assigning to string index returns the string itself, updated', str.1 := 'what', 'hwhat')
	t('assigning to string index modifies itself', str, 'hwhat')

	str := '00000000'
	mut := (i, s) => (
		str.(i) := s
	)
	mut(4, 'AAA')
	t('assigning to string index with more than one char modifies multiple indexes'
		str, '0000AAA0')
	mut(8, 'YYY')
	t('assigning to string index with more than one char appends as necessary'
		str, '0000AAA0YYY')
)

m('number & composite/list -> string conversions')
(
	stringList := std.stringList

	t('string(number) truncates to 8th decimal digit', string(3.14), '3.14000000')
	t('string(number) truncates decimal point for integers', string(42), '42')
	t('string(true)', string(true), 'true')
	t('string(false)', string(false), 'false')
	t('string(string) returns itself', string('hello'), 'hello')
	t('string(list) returns string(composite)', string([0]), '{0: 0}')

	t('number(string) correctly parses decimal number', number('3.14'), 3.14)
	t('number(string) deals with leading and trailing zeroes', number('03.140000'), 3.14)
	t('number(string) deals with negative numbers', number('-42'), ~42)
	t('number(true) = 1', number(true), 1)
	t('number(false) = 0', number(false), 0)
	t('number(composite) = 0', number([]), 0)
	t('number(function) = 0', number(() => 100), 0)
	t('number(builtin fn) = 0', number(len), 0)

	t('string(composite)', string({a: 3.14}), '{a: 3.14000000}')

	result := string([3, 'two'])
	p1 := '{0: 3, 1: \'two\'}'
	p2 := '{1: \'two\', 0: 3}'
	t('string(composite) containing string and multiple keys', result = p1 | result = p2, true)

	t('stringList(list) for nested list', stringList(['fine', ['not']]), '[fine, {0: \'not\'}]')
)

m('function/composite equality checks')
(
	` function equality `
	fn1 := () => (3 + 4, 'hello')
	fnc := fn1
	fn2 := () => (3 + 4, 'hello')

	t('functions are equal if they are the same function'
		fn1 = fnc, true)
	t('functions are different if they are defined separately, even if same effect'
		fn1 = fn2, false)

	` composite equality `
	comp1 := {1: 2, hi: '4'}
	comp2 := {1: 2, hi: '4'}
	list1 := [1, 2, 3, 4, 5]
	list2 := [1, 2, 3, 4, 5]
	complist := {0: 1, 1: 2, 2: 3, 3: 4, 4: 5}

	t('deep composite equality', comp1 = comp2, true)
	t('deep list equality', list1 = list2, true)
	t('deep composite inequality, I', comp1 = list1, false)
	t('deep composite inequality, II', comp1 = {1: '4', 2: 2}, false)
	t('composite = {}', comp1 = {}, false)
	t('deep list inequality, I', list1 = [1, 2, 3], false)
	t('deep list inequality, II', list1 = complist, true)
)

m('type() builtin function')
(
	t('type(string)', type('hi'), 'string')
	t('type(number)', type(3.14), 'number')
	t('type(list) (composite)', type([0, 1, 2]), 'composite')
	t('type(composite)', type({hi: 'what'}), 'composite')
	t('type(function)', type(() => 'hi'), 'function')
	t('type(builtin fn) (function), I', type(type), 'function')
	t('type(builtin fn) (function), II', type(out), 'function')
	t('type(()) = ()', type(()), '()')
)

m('std.range/slice/append/join/cat and stringList')
(
	stringList := std.stringList
	range := std.range
	reverse := std.reverse
	slice := std.slice
	join := std.join
	cat := std.cat

	` slice returns copies `
	(
		st := '12345'
		li := [1, 2, 3, 4, 5]

		stc := slice(st, 0, len(st))
		lic := slice(li, 0, len(li))
		stc.2 := 'x'
		lic.2 := 'x'

		t('slice(string) should make a copy', st, '12345')
		t('slice(string) should return a copy', stc, '12x45')
		t('slice(list) should make a copy', li, [1, 2, 3, 4, 5])
		t('slice(list) should return a copy', lic, [1, 2, 'x', 4, 5])
	)

	sl := (l, s, e) => stringList(slice(l, s, e))
	list := range(10, ~1, ~1)
	str := 'abracadabra'

	t('slice(list)', sl(list, 0, 5), '[10, 9, 8, 7, 6]')
	t('slice with OOB lower bound', sl(list, ~5, 2), '[10, 9]')
	t('slice with OOB upper bound', sl(list, 7, 20), '[3, 2, 1, 0]')
	t('slice with OOB both bounds', sl(list, 20, 1), '[]')

	` redefine list using range and reverse, to t those `
	list := reverse(range(0, 11, 1))

	t('join() homogeneous lists', stringList(join(
		slice(list, 0, 5), slice(list, 5, 200)
	)), '[10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0]')
	t('join() heterogeneous lists', stringList(join(
		[1, 2, 3]
		join([4, 5, 6], ['a', 'b', 'c'])
	)), '[1, 2, 3, 4, 5, 6, a, b, c]')

	t('slice() from 0', slice(str, 0, 5), 'abrac')
	t('slice() from nonzero', slice(str, 2, 4), 'ra')
	t('slice() with OOB lower bound', slice(str, ~5, 2), 'ab')
	t('slice() with OOB upper bound', slice(str, 7, 20), 'abra')
	t('slice() with OOB both bounds', slice(str, 20, 1), '')

	t('cat() empty list', cat([], '--'), '')
	t('cat() single-element list', cat(['hello'], '--'), 'hello')
	t('cat() double-element list', cat(['hello', 'world'], '--'), 'hello--world')
	t('cat() list containing delimiter', cat(['hello', 'world,hi'], ','), 'hello,world,hi')
	t('cat() with empty string delimiter', cat(['good', 'bye', 'friend'], ''), 'goodbyefriend')
	t('cat() with comma separator', cat(['good', 'bye', 'friend'], ', '), 'good, bye, friend')
	t('cat() with manually indexed composite', cat({
		0: 'first'
		1: 'last'
	}, ' and '), 'first and last')
)

m('hexadecimal conversions, hex & xeh')
(
	hex := std.hex
	xeh := std.xeh

	` base cases `
	t('hex(0)', hex(0), '0')
	t('hex(42)', hex(66), '42')
	t('hex(256)', hex(256), '100')
	t('hex(1998)', hex(1998), '7ce')
	t('hex(3141592)', hex(3141592), '2fefd8')
	t('xeh(fff)', xeh('fff'), 4095)
	t('xeh(a2)', xeh('a2'), 162)

	` hex should floor non-integer inputs `
	t('hex() of fractional number, I', hex(16.8), '10')
	t('hex() of fractional number, II', hex(1998.123), '7ce')

	` recoverability `
	t('xeh(hex()), I', xeh(hex(390420)), 390420)
	t('xeh(hex()), II', xeh(hex(9230423903)), 9230423903)
	t('hex(xeh()), I', hex(xeh('fffab123')), 'fffab123')
	t('hex(xeh()), II', hex(xeh('0000ab99ff33')), 'ab99ff33')
	t('hex(xeh(hex(xeh()))), I', hex(xeh(hex(xeh('aabbef')))), 'aabbef')
	t('hex(xeh(hex(xeh()))), II', xeh(hex(xeh(hex(201900123)))), 201900123)
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
	t('point(a)', point('a'), 97)
	t('char(65)', char(65), 'A')
	t('encode(ab)', encode('ab'), [97, 98])
	t('decode() => ACB', decode([65, 67, 66]), 'ACB')
	t('repeated decode/encode, I', decode(encode(decode(encode(s1)))), s1)
	t('repeated decode/encode, II', decode(encode(decode(encode(s2)))), s2)
	t('repeated decode/encode, III', decode(encode(decode(encode(s3)))), s3)
)

m('std list: map/filter/reduce[Back]/each/reverse/flatten, join/append')
(
	map := std.map
	filter := std.filter
	reduce := std.reduce
	reduceBack := std.reduceBack
	each := std.each
	reverse := std.reverse
	flatten := std.flatten
	append := std.append
	join := std.join

	list := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

	t('std.map', map(list, n => n * n), [1, 4, 9, 16, 25, 36, 49, 64, 81, 100])
	t('std.filter', filter(list, n => n % 2 = 0), [2, 4, 6, 8, 10])
	t('std.reduce', reduce(list, (acc, n) => acc + string(n), '')
		'12345678910')
	t('std.reduceBack', reduceBack(list, (acc, n) => acc + string(n), '')
		'10987654321')
	t('std.flatten', flatten([[1, 2, 3], [4], [], [[5], 6, 7, [8, 9, 10]]])
		[1, 2, 3, 4, [5], 6, 7, [8, 9, 10]])
	t('std.reverse', reverse(list), [10, 9, 8, 7, 6, 5, 4, 3, 2, 1])
	t('std.join', join(list, list), [1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10])

	` passing index in callback `
	t('std.map passes index to callback', map(list, (_, i) => i)
		[0, 1, 2, 3, 4, 5, 6, 7, 8, 9])
	t('std.filter passes index to callback', filter(list, (_, i) => i % 2 = 1)
		[2, 4, 6, 8, 10])
	t('std.reduce passes index to callback'
		reduce(list, (acc, _, i) => acc + string(i), ''), '0123456789')
	t('std.reduceBack passes index to callback'
		reduceBack(list, (acc, _, i) => acc + string(i), ''), '9876543210')
	(
		eachAcc := []
		each(list, (_, i) => eachAcc.len(eachAcc) := i)
		t('std.each passes index to callback', eachAcc
			[0, 1, 2, 3, 4, 5, 6, 7, 8, 9])
	)
	
	` each doesn't return anything meaningful `
	acc := {
		str: ''
	}
	twice := f => x => (f(x), f(x))
	each(list, twice(n => acc.str := acc.str + string(n)))
	t('std.each', acc.str, '1122334455667788991010')

	` append mutates `
	append(list, list)
	t('std.append', list, [1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
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

	t('std.format empty string', f('', {}), '')
	t('std.format single value', f('one two {{ first }} four', values), 'one two ABC four')
	t('std.format with newlines in string', f('new
	{{ sup }} line', {sup: 42}), 'new
	42 line')
	t('std.format with non-terminated slot ignores rest'
		f('now {{ then now', {then: 'then'}), 'now ')
	t('std.format with unusual (tighter) spacing', f(
		' {{thingTwo}}+{{ magic+eye }}  '
		values
	), ' [5, 4, 3, 2, 1]+add_sign  ')
	t('std.format with unusual (tighter) spacing and more replacements', f(
		'{{last }} {{ first}} {{ thing One }} {{ thing Two }}'
		values
	), 'XYZ ABC 1 [5, 4, 3, 2, 1]')
	t(
		'std.format with non-format braces'
		f('{ {  this is not } {{ thingOne } wut } {{ nonexistent }}', values)
		'{ {  this is not } 1 ()'
	)
)

m('uuid -- uuid v4 generator')
(
	uuid := load('uuid').uuid

	xeh := std.xeh
	range := std.range
	map := std.map
	every := std.every
	reduce := std.reduce
	split := str.split

	uuids := map(range(0, 200, 1), uuid)

	` every character should be a hex character or "-" `
	isValidChar := s => s :: {
		'-' -> true
		_ -> ~(xeh(s) = ())
	}
	everyCharIsHex := every(map(
		uuids
		u => every(map(split(u, ''), c => isValidChar(c)))
	))
	t('uuid() validity, hexadecimal range set', everyCharIsHex, true)

	` test for uniqueness (kinda) `
	collisions? := reduce(
		map(range(0, 200, 1), () => [uuid(), uuid()])
		(acc, us) => acc | us.0 = us.1
		false
	)
	t('uuid() validity, rare collisions', collisions?, false)

	` correct length, formatting `
	format? := u => map(u, x => x) = [
		_, _, _, _, _, _, _, _, '-'
		_, _, _, _, '-'
		_, _, _, _, '-'
		_, _, _, _, '-'
		_, _, _, _, _, _, _, _, _, _, _, _
	]
	everyIsFormatted := every(map(
		uuids
		u => format?(u)
	))
	t('uuid() validity, correct string formatting', everyIsFormatted, true)
)

m('json ser/de')
(
	clone := std.clone

	json := load('json')
	ser := json.ser
	de := json.de

	` primitives `
	t('ser null', ser(()), 'null')
	t('ser ""', ser(''), '""')
	t('ser string', ser('world'), '"world"')
	t('ser escaped string', ser('es"c a"pe
me'), '"es\\"c a\\"pe\\nme"')
	t('ser true', ser(true), 'true')
	t('ser false', ser(false), 'false')
	t('ser number', ser(12), '12')
	t('ser fractional number', ser(3.14), string(3.14))
	t('ser negative number', ser(~2.4142), string(~2.4142))
	t('ser function => null', ser(x => x), 'null')
	t('ser empty composite', ser({}), '{}')
	t('ser empty list => composite', ser([]), '{}')

	t('de null', de('null'), ())
	t('de invalid JSON, null-ish', de('neh'), ())
	t('de true', de('true'), true)
	t('de invalid JSON, true-ish, I', de('trxx'), ())
	t('de false', de('false'), false)
	t('de invalid JSON, false-ish', de('fah'), ())
	t('de invalid JSON, true-ish, II', de('true_32'), true)
	t('de string', de('"thing"'), 'thing')
	t('de escaped string', de('"es\\"c a\\"pe\\nme"'), 'es"c a"pe
me')
	t('de empty string', de('""'), '')
	t('de interrupted string', de('"my"what"'), 'my')
	t('de negative number', de('-59.413'), ~59.413)
	t('de interrupted number', de('10-14.2'), 10)
	t('de invalid number', de('1.2.3'), ())
	t('de list of numbers', de('[50, -100]'), [50, ~100])

	` strange whitespace, commas, broken input `
	t('de string with surrounding whitespace', de('   " string"	 '), ' string')
	t('de number with surrounding whitespace and decimal point', de('   6.'), 6)
	t('de fractional number without leading zero', de(' .90'), 0.9)
	t('de list with surrounding whitespace', de('   ["first", 2, true, ]	'), ['first', 2, true])
	t('de non-terminated string', de('"start '), ())
	t('de invalid object values', de('{"a": b, "12": 3.41}'), ())
	t('de non-terminated object literal', de('{"a": b  "12": 3.41 '), ())
	t('de non-terminated literal symbols', de('[1, 2  3.24.253, fals}'), ())
	t('de non-terminated list literal', de('[1, 2, 3.24.253, false'), ())

	` serialize light object `
	s := ser({a: 'b', c: ~4.251})
	first := '{"a":"b","c":-4.25100000}'
	second := '{"c":-4.25100000,"a":"b"}'
	t('ser composite', s = first | s = second, true)

	s := ser([2, false])
	first := '{"0":2,"1":false}'
	second := '{"1":false,"0":2}'
	t('ser list', s = first | s = second, true)

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
	t('ser complex composite', de(ser(obj)), objr)
	t('de ser complex composite', de(ser(list)), list)

	list.1 := obj
	listr := clone(list)
	listr.1 := objr
	t('de ser de ser complex list', de(ser(de(ser(list)))), listr)
)

m('str.upper/lower/digit/letter/ws? -- checked char ranges')
(
	upper? := str.upper?
	lower? := str.lower?
	digit? := str.digit?
	letter? := str.letter?
	ws? := str.ws? 

	every := std.every
	some := std.some
	map := std.map

	t('upper? verifies uppercase letters'
		every(map('ABCDEFGHIJKLMNOPQRSTUVWXYZ', upper?)), true)
	t('upper? rejects non-uppercase-letters'
		some(map('onawfepd913043?-~\'!/.,;()$@)%', upper?)), false)
	t('lower? verifies lowercase letters'
		every(map('abcdefghijklmnopqrstuvwxyz', lower?)), true)
	t('lower? rejects non-lowercase-letters'
		some(map('ONAWFEPD913043?-~\'!/.,;()$@)%', lower?)), false)
	t('digit? verifies digits'
		every(map('0123456789', digit?)), true)
	t('digit? rejects non-digits, including punctuations'
		some(map('~@!#@$%^()&?!.;,-', digit?)), false)
	t('letter? verifies all alphabet letters'
		every(map('ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz', letter?))
		true)
	t('letter? rejects non-letters'
		some(map('913043?-~\'!/.,;()$@)%', upper?)), false)
	t('ws? verifies whitespace characters'
		every(map('	   
', ws?)), true)
	t('ws? rejects all non-whitespace'
		some(map('jafsioSINDFOEJ#@%@()_#9u40529' + char(250), ws?)), false)

	hasPrefix? := str.hasPrefix?
	hasSuffix? := str.hasSuffix?

	t('hasPrefix? detects prefix'
		hasPrefix?('programming', 'prog'), true)
	t('hasPrefix? returns true for empty prefix'
		hasPrefix?('programming', ''), true)
	t('hasPrefix? returns true if s = prefix'
		hasPrefix?('programming', 'programming'), true)
	t('hasPrefix? returns false if not prefix'
		hasPrefix?('programming', 'progx'), false)

	t('hasSuffix? detects suffix'
		hasSuffix?('programming', 'mming'), true)
	t('hasSuffix? returns true for empty suffix'
		hasSuffix?('programming', ''), true)
	t('hasSuffix? returns true if s = suffix'
		hasSuffix?('programming', 'programming'), true)
	t('hasSuffix? returns false if not suffix'
		hasSuffix?('programming', 'science'), false)

	matchesAt? := str.matchesAt?

	t('matchesAt? returns true for empty substring'
		matchesAt?('some substring', ''), true)
	t('matchesAt? returns true if string matches at idx'
		matchesAt?('some substring', 'substr', 5), true)
	t('matchesAt? returns false if string matches not at idx'
		matchesAt?('some substring', 'substr', 2), false)
	t('matchesAt? returns false if no match'
		matchesAt?('some substring', 'other', 5), false)

	index := str.index

	t('index = 0 for empty string', index('quick brown fox', ''), 0)
	t('index returns index of substring'
		index('quick brown fox', 'ick'), 2)
	t('index returns 0 if matches whole string'
		index('quick brown fox', 'quick brown fox'), 0)
	t('index returns ~1 if no match'
		index('quick brown fox', 'lazy dog'), ~1)
	t('index returned is first occurrence'
		index('quick brown fox', 'o'), 8)
	t('index works if substring longer than string'
		index('quick brown fox', 'jumps over the lazy dog'), ~1)

	contains? := str.contains?

	t('contains? = true for empty string'
		contains?('quick brown fox', ''), true)
	t('contains? = true if string fits substring'
		contains?('quick brown fox', 'fox'), true)
	t('contains? = true if substring fits multiple times'
		contains?('quick brown fox', 'o'), true)
	t('contains? = false if not contained'
		contains?('quick brown fox', 'lazy dog'), false)

	lower := str.lower
	upper := str.upper
	title := str.title
	given := 'MIXED case StrinG with ?!~:punct'

	t('lower transforms string to lowercase'
		lower(given), 'mixed case string with ?!~:punct')
	t('upper transforms string to uppercase'
		upper(given), 'MIXED CASE STRING WITH ?!~:PUNCT')
	t('title returns uppercase first + lowercase rest'
		title(given), 'Mixed case string with ?!~:punct')

	replace := str.replace

	t('replace is no-op if empty string'
		replace('he stared in amazement', '', '__')
		'he stared in amazement')
	t('replace replaces all instances of given substring'
		replace('he stared in amazement', 'e', 'j')
		'hj starjd in amazjmjnt')
	t('replace works for multi-character substring'
		replace('he is staring in amazement', 'in', 'xx')
		'he is starxxg xx amazement')
	t('replace accounts for different old/new substring lengths'
		replace('he is staring in amazement', 'in', 'wonder')
		'he is starwonderg wonder amazement')
	t('replace deals gracefully with overlapping matches'
		replace('wow what a sight, wow', 'ow', 'wow')
		'wwow what a sight, wwow')
	t('replace works if new substring is empty'
		replace('wow what a sight, wow', 'wow', '')
		' what a sight, ')
	t('replace works even if new str contains recursive match'
		replace('a {} b {} c {}', '{}', '{}-{}')
		'a {}-{} b {}-{} c {}-{}')

	split := str.split

	t('split splits string into letters if empty'
		split('alphabet', '')
		['a', 'l', 'p', 'h', 'a', 'b', 'e', 't'])
	t('splits with given delimiter'
		split('a,b,cde,fg', ',')
		['a', 'b', 'cde', 'fg'])
	t('splits with empty strings if delimiter in start or end'
		split(', original taste, ', ', ')
		['', 'original taste', ''])
	t('returns one chunk if no match of delimiter found'
		split('no taste whatsoever!', 'grand')
		['no taste whatsoever!'])

	trimPrefix := str.trimPrefix
	trimSuffix := str.trimSuffix
	trim := str.trim

	t('trimPrefix is a no-op with empty string', 
		trimPrefix('???????what???', ''), '???????what???')
	t('trimPrefix trims given prefix until it does not prefix'
		trimPrefix('???????what???', '?'), 'what???')
	t('trimPrefix works with multi-char prefix'
		trimPrefix('abababacdef', 'ab'), 'acdef')
	t('trimPrefix only trims whole multiples of prefix'
		trimPrefix('aaaaaaaadef', 'aaa'), 'aadef')

	t('trimSuffix is a no-op with empty string', 
		trimSuffix('???????what???', ''), '???????what???')
	t('trimSuffix trims given suffix until it does not suffix'
		trimSuffix('???????what???', '?'), '???????what')
	t('trimSuffix works with multi-char suffix'
		trimSuffix('abacdefabcabab', 'ab'), 'abacdefabc')
	t('trimSuffix only trims whole multiples of suffix'
		trimSuffix('xxxyyyyyyyy', 'yyy'), 'xxxyy')


	t('trim trims given string from both sides'
		trim('????what?????', '?'), 'what')
	t('trim is a no-op with empty string'
		trim('????what?????', ''), '????what?????')
	t('trim trims whole multiples of substring from both sides'
		trim('????what?????', '???'), '?what??')
)

m('load() import semantics')
(
	A := load('load_dedup')
	B := load('load_dedup/load_dedup_child')
	getObjA := A.getObj
	getObjB := B.getObj

	t('load() from different contexts should be deduplicated'
		getObjA() = getObjB(), true)
	t('load() of different source files should still return different contexts'
		A = B, false)
)

m('args() list')
(
	hasSuffix? := str.hasSuffix?

	as := args()
	t('args() returns a list'
		type(as), 'composite')
	t('first item in args() is the Ink executable'
		hasSuffix?(as.0, 'ink'), true)
	t('second item in args() is the test script'
		hasSuffix?(as.1, 'test.ink'), true)
)

` end test suite, print result `
(s.end)()
