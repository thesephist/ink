` naive in-place quicksort implementation `
` adopted from https://en.wikipedia.org/wiki/Quicksort
	... there's probably far more elegant and idiomatic solutions `

std := load('std')

log := std.log
stringList := std.stringList

` main recursive quicksort routine `
quicksort := (list, lo, hi) => lo < hi :: {
	true -> (
		p := partition(list, lo, hi)
		quicksort(list, lo, p - 1)
		quicksort(list, p + 1, hi)
	)
}

` Lomuto partition scheme `
partition := (list, lo, hi) => (
	` arbitrarily pick last value as pivot `
	pivot := list.(hi)
	acc := {
		i: lo
	}

	loop := j => j :: {
		hi -> ()
		_ -> (
			list.(j) < pivot :: {
				true -> (
					swap(list, acc.i, j)
					acc.i := acc.i + 1
				)
			}
			loop(j + 1)
		)
	}
	
	loop(lo)

	swap(list, acc.i, hi)
	acc.i
)

` swap two places in a given list `
swap := (list, i, j) => (
	last := {
		i: list.(i)
		j: list.(j)
	}
	list.(i) := last.j
	list.(j) := last.i
)

` top-level sorting function for QuickSort `
sort := list => quicksort(list, 0, len(list) - 1)

` random list builder `
buildList := (length, opts) => (
	max := (opts.max :: {
		() -> 1000
		_ -> opts.max
	})

	length :: {
		0 -> {}
		_ -> (
			smaller := buildList(length - 1, {max: max})
			smaller.len(smaller) := floor(rand() * max) + 1
			smaller
		)
	}
)

`` main
list := buildList(100, {})
out('Quicksorting random list: ' + stringList(list) + '
sorted -> ')
sort(list)
log(stringList(list))
