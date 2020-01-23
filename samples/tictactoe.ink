` interactive terminal tic tac toe in Ink `

std := load('std')

log := std.log
scan := std.scan
f := std.format
sliceList := std.sliceList
map := std.map
filter := std.filter

` async version of a while(... condition, ... predicate)
	that takes a callback `
asyncWhile := (cond, do) => (sub := () => cond() :: {
	true -> do(sub)
	false -> ()
})()

` shorthand tools for getting players and player labels `
Player := {
	x: 1
	o: 2
}
Label := ['.', 'x', 'o']
label := id => Label.(id)

` create a new game board `
newBoard := () => [
	1 ` current player turn `
	0, 0, 0
	0, 0, 0
	0, 0, 0
]

` format-print board state `
stringBoard := bd => f(
'{{ 1 }}    {{ 2 }}    {{ 3 }}

{{ 4 }}    {{ 5 }}    {{ 6 }}

{{ 7 }}    {{ 8 }}    {{ 9 }}
'
	map(bd, label)
)

` winning placement combinations for a single player `
Combinations := [
	[1, 2, 3]
	[4, 5, 6]
	[7, 8, 9]

	[1, 4, 7]
	[2, 5, 8]
	[3, 6, 9]

	[1, 5, 9]
	[3, 5, 7]
]
` returns -1 if no win, 0 if tie, or winner player ID `
checkBoard := bd => (
	checkIfPlayerWon := player => (
		isPlayer := row => row = [player, player, player]
		combinationToValues := combo => map(combo, idx => bd.(idx))
		possibleRows := map(Combinations, combinationToValues)
		didWin := len(filter(possibleRows, isPlayer)) > 0

		didWin
	)

	checkIfPlayerWon(Player.x) :: {
		true -> Player.x
		_ -> checkIfPlayerWon(Player.o) :: {
			true -> Player.o
			_ -> (
				` check if tie `
				takenCells := filter(sliceList(bd, 1, 10), val => ~(val = 0))
				len(takenCells) :: {
					9 -> 0
					_ -> ~1
				}
			)
		}
	}
)

` take one player turn `
stepBoard! := (bd, cb) => scan(s => (
	idx := number(s) :: {
		() -> stepBoard!(bd, cb)
		_ -> idx > 0 & idx < 10 :: {
			true -> bd.(idx) :: {
				0 -> (
					bd.(number(s)) := getPlayer(bd)
					setPlayer(bd, nextPlayer(bd))
					cb()
				)
				_ -> (
					log(f('{{ idx }} is already taken!', {idx: idx}))
					out(f('Move for player {{ player }}: ', {
						player: label(getPlayer(bd))
					}))
					stepBoard!(bd, cb)
				)
			}
			false -> (
				log('Enter a number 0 < n < 10.')
				out(f('Move for player {{ player }}: ', {
					player: label(getPlayer(bd))
				}))
				stepBoard!(bd, cb)
			)
		}
	}
))

` get/set/modify player turn state from the game board `
getPlayer := bd => bd.0
setPlayer := (bd, pl) => bd.0 := pl
nextPlayer := bd => label(getPlayer(bd)) :: {
	'x' -> Player.o
	_ -> Player.x
}

` divider used to delineate each turn in the UI `
Divider := '
>---------------<
'

` run a single game `
run := () => (
	log('Welcome to Ink tic-tac-toe!')

	bd := newBoard()
	asyncWhile(
		() => checkBoard(bd) :: {
			~1 -> true
			_ -> (
				log(Divider)
				checkBoard(bd) :: {
					0 -> log('x and o tied!')
					(Player.x) -> log('x won!')
					(Player.o) -> log('o won!')
					_ -> log('Unrecognized result: ' + checkBoard(bd))
				}
				log('')
				log(stringBoard(bd))
				false
			)
		}
		cb => (
			log(Divider)
			log(stringBoard(bd))
			out(f('Move for player {{ player }}: ', {
				player: label(getPlayer(bd))
			}))
			stepBoard!(bd, cb)
		)
	)
)

run()
