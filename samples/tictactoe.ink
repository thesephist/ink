` interactive terminal tic tac toe in Ink `

std := load('std')

log := std.log
scan := std.scan
f := std.format
sliceList := std.sliceList
map := std.map
reduce := std.reduce
filter := std.filter

` async version of a while(... condition, ... predicate)
	that takes a callback `
asyncWhile := (cond, do) => (sub := () => cond() :: {
	true -> do(sub)
	false -> ()
})()

` shorthand tools for getting players and player labels `
Player := {x: 1, o: 2}
Label := [' ', 'x', 'o']
` make letters appear bolder / fainter on the board `
bold := c => '[0;1m' + c + '[0;0m'
grey := c => '[33;2m' + c + '[0;0m'

` create a new game board + state `
newBoard := () => [
	1 ` current player turn `
	0, 0, 0
	0, 0, 0
	0, 0, 0
]

` format string to print board state `
BoardFormat :=
'{{ 1 }} │ {{ 2 }} │ {{ 3 }}
──┼───┼──
{{ 4 }} │ {{ 5 }} │ {{ 6 }}
──┼───┼──
{{ 7 }} │ {{ 8 }} │ {{ 9 }}
'
` format-print board state `
stringBoard := bd => f(
	BoardFormat
	map(bd, (player, idx) => Label.(player) :: {
		' ' -> grey(string(idx))
		_ -> bold(Label.(player))
	})
)

` winning placement combinations for a single player `
Combinations := [
	` horizontal `
	[1, 2, 3]
	[4, 5, 6]
	[7, 8, 9]

	` vertical `
	[1, 4, 7]
	[2, 5, 8]
	[3, 6, 9]

	` diagonal `
	[1, 5, 9]
	[3, 5, 7]
]
` returns -1 if no win, 0 if tie, or winner player ID `
Result := {
	None: ~1
	Tie: 0
	X: Player.x
	O: Player.o
}
checkBoard := bd => (
	checkIfPlayerWon := player => (
		isPlayer := row => row = [player, player, player]
		combinationToValues := combo => map(combo, idx => bd.(idx))
		possibleRows := map(Combinations, combinationToValues)
		didWin := len(filter(possibleRows, isPlayer)) > 0

		didWin
	)

	checkIfPlayerWon(Player.x) :: {
		true -> Result.X
		_ -> checkIfPlayerWon(Player.o) :: {
			true -> Result.O
			_ -> (
				` check if game ended in a tie `
				takenCells := filter(sliceList(bd, 1, 10), val => ~(val = 0))
				len(takenCells) :: {
					9 -> Result.Tie
					_ -> Result.None
				}
			)
		}
	}
)

` take one player turn, mutates game state `
stepBoard! := (bd, cb) => scan(s => idx := number(s) :: {
	` not a number, try again `
	() -> stepBoard!(bd, cb)
	_ -> idx > 0 & idx < 10 :: {
		` number in range, make a move `
		true -> bd.(idx) :: {
			` the given cell is empty, make a move `
			0 -> (
				bd.(number(s)) := getPlayer(bd)
				setPlayer(bd, nextPlayer(bd))
				cb()
			)
			` the cell is already occupied, try again `
			_ -> (
				log(f('{{ idx }} is already taken!', {idx: idx}))
				out(f('Move for player {{ player }}: ', {
					player: Label.(getPlayer(bd))
				}))
				stepBoard!(bd, cb)
			)
		}
		` number not in range, try again `
		false -> (
			log('Enter a number 0 < n < 10.')
			out(f('Move for player {{ player }}: ', {
				player: Label.(getPlayer(bd))
			}))
			stepBoard!(bd, cb)
		)
	}
})

` get/set/modify player turn state from the game board `
getPlayer := bd => bd.0
setPlayer := (bd, pl) => bd.0 := pl
nextPlayer := bd => Label.(getPlayer(bd)) :: {
	'x' -> Player.o
	_ -> Player.x
}

` divider used to delineate each turn in the UI `
Divider := '
>---------------<
'

` run a single game `
log('Welcome to Ink tic-tac-toe!')
bd := newBoard()
asyncWhile(
	() => checkBoard(bd) :: {
		(Result.None) -> true
		_ -> (
			log(Divider)
			checkBoard(bd) :: {
				(Result.Tie) -> log('x and o tied!')
				(Result.X) -> log('x won!')
				(Result.O) -> log('o won!')
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
			player: Label.(getPlayer(bd))
		}))
		stepBoard!(bd, cb)
	)
)
