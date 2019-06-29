` naive in-place quicksort implementation `
` adopted from https://en.wikipedia.org/wiki/Quicksort
    ... there's probably far more elegant and idiomatic solutions `

` main recursive quicksort routine `
quicksort := (list, lo, hi) => (
    lo < hi :: {true -> (
        p := partition(list, lo, hi)
        quicksort(list, lo, p - 1)
        quicksort(list, p + 1, hi)
    )}
)

` Lomuto partition scheme `
partition := (list, lo, hi) => (
    ` arbitrarily pick last value as pivot `
    pivot := list.(hi)
    acc := {
        i: lo
    }

    jLoop := j => j :: {
        hi -> ()
        _ -> (
            list.(j) < pivot :: {true -> (
                swap(list, acc.i, j)
                acc.i := acc.i + 1
            )}
            jLoop(j + 1)
        )
    }
    
    jLoop(lo)

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
buildList := length => (
    length :: {
        0 -> {}
        _ -> (
            smaller := buildList(length - 1)
            smaller.(len(smaller)) := floor(rand() * 100) + 1
            smaller
        )
    }
)

` tail recursive list -> string converter `
stringList := list => '[' + stringListRec(list, 0, '') + ']'
stringListRec := (list, start, acc) => (
    start :: {
        len(list) -> acc
        _ -> stringListRec(
            list
            start + 1
            (acc :: {
                '' -> ''
                _ -> acc + ', '
            }) + string(list.(start))
        )
    }
)

`` main
list := buildList(50)
out('Quicksorting random list: ' + stringList(list) + '
sorted -> ')
sort(list)
log(stringList(list))
