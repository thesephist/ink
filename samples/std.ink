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

` like Python's range(), but no optional arguments `
range := (start, end, step) => (
	span := end - start
	sub := (i, v, acc) => (v - start) / span < 1 :: {
		true -> (
			acc.(i) := v
			sub(i + 1, v + step, acc)
		)
		false -> acc
	}

	` preempt potential infinite loops `
	(end - start) / step > 0 :: {
		true -> sub(0, start, [])
		false -> []
	}
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
		start: start
		end: end
	}
)

` get a substring of a given string `
slice := (str, start, end) => (
	result := ['']

	` bounds checks `
	x := clamp(start, end, 0, len(str))
	start := x.start
	end := x.end

	(sl := i => i :: {
		end -> result.0
		_ -> (
			result.0 := result.0 + str.(i)
			sl(i + 1)
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

	(sl := i => i :: {
		end -> result
		_ -> (
			result.(len(result)) := list.(i)
			sl(i + 1)
		)
	})(start)
)

` join one list to the end of another, return the original first list `
append := (base, child) => (
	baseLength := len(base)
	childLength := len(child)
	(sub := i => i :: {
		childLength -> base
		_ -> (
			base.(baseLength + i) := child.(i)
			sub(i + 1)
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
	(sub := (i, acc) => i :: {
			length -> acc
			_ -> sub(
				i + 1
				f(acc, list.(i))
			)
	})(0, acc)
)

` for-each loop over a list `
each := (list, f) => (
	length := len(list)
	(sub := i => i :: {
		length -> ()
		_ -> (
			f(list.(i))
			sub(i + 1)
		)
	})(0)
)

` encode ascii string into a number list
	we don't use reduce here, because this has to be fast,
	and we implement an optimized loop instead with minimal copying `
encode := str => (
	acc := [{}]
	strln := len(str)
	(sub := i => i :: {
		strln -> acc.0
		_ -> (
			(acc.0).(i) := point(str.(i))
			sub(i + 1)
		)
	})(0)
)

` decode number list into an ascii string `
decode := data => reduce(data, (acc, cp) => acc + char(cp), '')

` utility for reading an entire file
	readFile does not lean on readRawFile because a string-based
	implementation can be more efficient here `
readFile := (path, callback) => (
	BUFSIZE := 4096 ` bytes `
	sent := [false]
	(accumulate := (offset, acc) => read(path, offset, BUFSIZE, evt => (
		sent.0 :: {false -> (
			evt.type :: {
				'error' -> (
					sent.0 := true
					callback(())
				)
				'data' -> (
					dataLen := len(evt.data)
					dataLen = BUFSIZE :: {
						true -> accumulate(offset + dataLen, acc + decode(evt.data))
						false -> (
							sent.0 := true
							callback(acc + decode(evt.data))
						)
					}
				)
			}
		)}
	)))(0, '')
)

` utility for reading an entire file without string conversion `
readRawFile := (path, callback) => (
	BUFSIZE := 4096 ` bytes `
	sent := [false]
	(accumulate := (offset, acc) => read(path, offset, BUFSIZE, evt => (
		sent.0 :: {false -> (
			evt.type :: {
				'error' -> (
					sent.0 := true
					callback(())
				)
				'data' -> (
					dataLen := len(evt.data)
					dataLen = BUFSIZE :: {
						true -> accumulate(offset + dataLen, append(acc, evt.data))
						false -> (
							sent.0 := true
							callback(append(acc, evt.data))
						)
					}
				)
			}
		)}
	)))(0, [])
)

` utility for writing an entire file
	is not buffered, because it's simpler, but may cause jank later
	we'll address that if/when it becomes a performance issue `
writeFile := (path, data, callback) => writeRawFile(path, encode(data), callback)

` analogue of readRawFile for writeFile `
writeRawFile := (path, data, callback) => (
	sent := [false]
	write(path, 0, data, evt => (
		sent.0 :: {false -> (
			sent.0 := true
			evt.type :: {
				'error' -> callback(())
				'end' -> callback(true)
			}
		)}
	))
)

` template formatting with {{ key }} constructs `
format := (raw, values) => (
	` parser state `
	state := {
		` current position in raw `
		idx: 0
		` parser internal state:
			0 -> normal
			1 -> seen one {
			2 -> seen two {
			3 -> seen a valid }
		`
		which: 0
		` buffer for currently reading key `
		key: ''
		` result build-up buffer `
		buf: ''
	}

	` helper function for appending to state.buf `
	append := c => state.buf := state.buf + c

	` read next token, update state `
	readNext := () => (
		c := raw.(state.idx)

		state.which :: {
			0 -> c :: {
				'{' -> state.which := 1
				_ -> append(c)
			}
			1 -> c :: {
				'{' -> state.which := 2
				` if it turns out that earlier brace was not
					a part of a format expansion, just backtrack `
				_ -> (
					append('{' + c)
					state.which := 0
				)
			}
			2 -> c :: {
				'}' -> (
					` insert key value `
					state.buf := state.buf + string(values.(state.key))
					state.key := ''
					state.which := 3
				)
				` ignore spaces in keys -- not allowed `
				' ' -> ()
				_ -> state.key := state.key + c
			}
			3 -> c :: {
				'}' -> state.which := 0
				` ignore invalid inputs -- treat them as nonexistent `
				_ -> ()
			}
		}

		state.idx := state.idx + 1
	)

	` main recursive sub-loop `
	max := len(raw)
	(sub := () => state.idx < max :: {
		true -> (
			readNext()
			sub()
		)
		false -> state.buf
	})()
)
