` minimal quicksort implementation
	using hoare partition `

std := load('std')

map := std.map
clone := std.clone

sortBy := (v, pred) => (
	vPred := map(v, pred)
	partition := (v, lo, hi) => (
		pivot := vPred.(lo)
		lsub := i => (vPred.(i) < pivot) :: {
			true -> lsub(i + 1)
			false -> i
		}
		rsub := j => (vPred.(j) > pivot) :: {
			true -> rsub(j - 1)
			false -> j
		}
		(sub := (i, j) => (
			i := lsub(i)
			j := rsub(j)
			(i < j) :: {
				false -> j
				true -> (
					` inlined swap! `
					tmp := v.(i)
					tmpPred := vPred.(i)
					v.(i) := v.(j)
					v.(j) := tmp
					vPred.(i) := vPred.(j)
					vPred.(j) := tmpPred

					sub(i + 1, j - 1)
				)
			}
		))(lo, hi)
	)
	(quicksort := (v, lo, hi) => len(v) :: {
		0 -> v
		_ -> (lo < hi) :: {
			false -> v
			true -> (
				p := partition(v, lo, hi)
				quicksort(v, lo, p)
				quicksort(v, p + 1, hi)
			)
		}
	})(v, 0, len(v) - 1)
)

sort! := v => sortBy(v, x => x)

sort := v => sort!(clone(v))

` TEST `
range := std.range
log := std.log
list := std.stringList

rint := () => floor(rand() * 500)
L := map(range(0, 250, 1), rint)
Before := clone(L)
log('before quicksort: ' + list(L))
log('after quicksort: ' + list(sort(L)))
log('before intact?: ' + (L :: {Before -> 'yes', _ -> 'no'}))
sort!(L)
log('after mutable sort: ' + list(L))
