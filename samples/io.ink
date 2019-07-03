` filesystem i/o demo `

` std imports `
std := load('samples/std')

log := std.log
slice := std.slice
decode := std.decode

` we're going to copy README.md to WRITEME.md,
	and we're going to buffer it `

BUFSIZE := 512 `` bytes

` have we read to the end of the file? `
state := {
	offset: 0
	ended: false
}

` main routine that reads/writes through buffer
	and recursively copies data. This is also tail-recursive `
incrementalCopy := (src, dest) => read(src, state.offset, BUFSIZE, evt => (
	evt.type :: {
		'error' -> (
			log('Encountered an error reading: ' + evt.message)
			state.ended := true
		)

		'data'  -> (
			` compute offsets and state `
			dataLength   := len(evt.data)
			ofs          := state.offset
			state.offset := state.offset + dataLength

			` if we read less data than we expected, read ended `
			dataLength :: {
				BUFSIZE -> ()
				_       -> state.ended := true
			}

			` log progress `
			log('copying --> ' + slice(
				decode(evt.data)
				0, 50
			) + '...')

			` write the read bit, and recurse back to reading `
			write(dest, ofs, evt.data, evt => evt.type :: {
				'error' -> (
					log('Encountered an error writing: ' + evt.message)
					state.ended := true
				)
				'end'   -> state.ended :: {
					false -> incrementalCopy(src, dest)
				}
			})
		)
	}
))

` copy README.md to WRITEME.md `
incrementalCopy('README.md', 'WRITEME.md')
