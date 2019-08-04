` standard string library `

std := load('std')

map := std.map
slice := std.slice
reduce := std.reduce
reduceBack := std.reduceBack

` checking if a given character is of a type `
checkRange := (lo, hi) => c => (
	p := point(c)
	lo < p & p < hi
)
upper? := checkRange(point('A') - 1, point('Z') + 1)
lower? := checkRange(point('a') - 1, point('z') + 1)
digit? := checkRange(point('0') - 1, point('9') + 1)
letter? := c => upper?(c) | lower?(c)

` is the char a whitespace? `
ws? := c => point(c) :: {
	` space `
	32 -> true
	` newline `
	10 -> true
	` hard tab `
	9 -> true
	` carriage return `
	13 -> true
	_ -> false
}

hasPrefix? := (s, prefix) => reduce(prefix, (sofar, c, i) => s.(i) = c, true)

hasSuffix? := (s, suffix) => (
	diff := len(s) - len(suffix)
	reduceBack(suffix, (sofar, c, i) => s.(i + diff) = c, true)
)

matchesAt? := (s, substring, idx) => (
	max := len(substring)
	(sub := i => i :: {
		max -> true
		_ -> s.(idx + i) :: {
			(substring.(i))-> sub(i + 1)
			_ -> false
		}
	})(0)
)

index := (s, substring) => (
	max := len(s) - 1
	(sub := i => matchesAt?(s, substring, i) :: {
		true -> i
		false -> i < max :: {
			true -> sub(i + 1)
			false -> ~1
		}
	})(0)
)

contains? := (s, substring) => index(s, substring) > ~1

lower := s => reduce(s, (acc, c, i) => upper?(c) :: {
	true -> acc.(i) := char(point(c) + 32)
	false -> acc.(i) := c
}, '')

upper := s => reduce(s, (acc, c, i) => lower?(c) :: {
	true -> acc.(i) := char(point(c) - 32)
	false -> acc.(i) := c
}, '')

title := s => (
	lowered := lower(s)
	lowered.0 := upper(lowered.0)
)

replaceNonEmpty := (s, old, new) => (
	lold := len(old)
	lnew := len(new)
	(sub := (acc, i) => matchesAt?(acc, old, i) :: {
		true -> sub(
			slice(acc, 0, i) + new + slice(acc, i + lold, len(acc))
			i + lnew
		)
		false -> i < len(acc) :: {
			true -> sub(acc, i + 1)
			false -> acc
		}
	})(s, 0)
)

replace := (s, old, new) => old :: {
	'' -> s
	_ -> replaceNonEmpty(s, old, new)
}

splitNonEmpty := (s, delim) => (
	coll := []
	ldelim := len(delim)
	(sub := (acc, i, last) => matchesAt?(acc, delim, i) :: {
		true -> (
			coll.len(coll) := slice(acc, last, i)
			sub(acc, i + ldelim, i + ldelim)
		)
		false -> i < len(acc) :: {
			true -> sub(acc, i + 1, last)
			false -> coll.len(coll) := slice(acc, last, len(acc))
		}
	})(s, 0, 0)
)

split := (s, delim) => delim :: {
	'' -> map(s, c => c)
	_ -> splitNonEmpty(s, delim)
}

` trim string from start until it does not begin with prefix.
	trimPrefix is more efficient than repeated application of
	hasPrefix? because it minimizes copying. `
trimPrefixNonEmpty := (s, prefix) => (
	max := len(s)
	lpref := len(prefix)
	idx := (sub := i => i < max :: {
		true -> matchesAt?(s, prefix, i) :: {
			true -> sub(i + lpref)
			false -> i
		}
		false -> i
	})(0)
	slice(s, idx, len(s))
)

trimPrefix := (s, prefix) => prefix :: {
	'' -> s
	_ -> trimPrefixNonEmpty(s, prefix)
}

trimSuffixNonEmpty := (s, suffix) => (
	lsuf := len(suffix)
	idx := (sub := i => i > ~1 :: {
		true -> matchesAt?(s, suffix, i - lsuf) :: {
			true -> sub(i - lsuf)
			false -> i
		}
		false -> i
	})(len(s))
	slice(s, 0, idx)
)

trimSuffix := (s, suffix) => suffix :: {
	'' -> s
	_ -> trimSuffixNonEmpty(s, suffix)
}

trim := (s, ss) => trimPrefix(trimSuffix(s, ss), ss)
