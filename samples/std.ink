` the ink standard library `

log := val => out(string(val) + '
')

scan := callback => (
	acc := ['']
	cb := evt => evt.type :: {
		'end' -> callback(acc.0)
		'data' -> (
			acc.0 :=
				acc.0 + slice(evt.data, 0, len(evt.data) - 1)
			false
		)
	}
	in(cb)
)

` clamp start and end numbers to ranges, such that
	start < end. Utility used in slice/sliceList`
clamp := (start, end, min, max) => (
	start := (start < min :: {
		true -> min
		false -> start
	})
	end := (end < min :: {
		true -> min
		false -> end
	})
	end := (end > max :: {
		true -> max
		false -> end
	})
	start := (start > end :: {
		true -> end
		false -> start
	})

	{
		start: start,
		end: end,
	}
)

` get a substring of a given string `
slice := (str, start, end) => (
	result := ['']

	` bounds checks `
	x := clamp(start, end, 0, len(str))
	start := x.start
	end := x.end

	(sl := idx => idx :: {
		end -> result.0
		_ -> (
			result.0 := result.0 + str.(idx)
			sl(idx + 1)
		)
	})(start)
)

` get a sub-list of a given list `
sliceList := (list, start, end) => (
	result := []

	` bounds checks `
	x := clamp(start, end, 0, len(list))
	start := x.start
	end := x.end

	(sl := idx => idx :: {
		end -> result
		_ -> (
			result.(len(result)) := list.(idx)
			sl(idx + 1)
		)
	})(start)
)

` join one list to the end of another, return the original first list `
append := (base, child) => (
	baseLength := len(base)
	childLength := len(child)
	(append := idx => idx :: {
		childLength -> base
		_ -> (
			base.(baseLength + idx) := child.(idx)
			append(idx + 1)
		)
	})(0)
)

` join one list to the end of another, return the third list `
join := (base, child) => append(clone(base), child)

` clone a composite value `
clone := comp => (
	reduce(keys(comp), (acc, k) => (
		acc.(k) := comp.(k)
		acc
	), {})
)

` tail recursive numeric list -> string converter `
stringList := list => (
	length := len(list)
	stringListRec := (start, acc) => start :: {
		length -> acc
		_ -> stringListRec(
			start + 1
			(acc :: {
				'' -> ''
				_ -> acc + ', '
			}) + string(list.(start))
		)
	}
	'[' + stringListRec(0, '') + ']'
)

` tail recursive reversing a list `
reverse := list => (
	state := [len(list) - 1]
	reduce(list, (acc, item) => (
		acc.(state.0) := item
		state.0 := state.0 - 1
		acc
	), {})
)

` tail recursive map `
map := (list, f) => (
	reduce(list, (l, item) => (
		l.(len(l)) := f(item)
		l
	), {})
)

` tail recursive filter `
filter := (list, f) => (
	reduce(list, (l, item) => (
		f(item) :: {
			true -> l.(len(l)) := item
		}
		l
	), {})
)

` tail recursive reduce `
reduce := (list, f, acc) => (
	length := len(list)
	(reducesub := (idx, acc) => (
		idx :: {
			length -> acc
			_ -> reducesub(
				idx + 1
				f(acc, list.(idx))
			)
		}
	)
	)(0, acc)
)

` encode ascii string into a number list `
encode := str => reduce(
	str
	(acc, char) => (
		acc.(len(acc)) := point(char)
		acc
	)
	{}
)

` decode number list into an ascii string `
decode := data => reduce(
	data
	(acc, cp) => acc + char(cp)
	''
)
