` filesystem i/o demo `

` std imports `
std := load('std')

log := std.log
slice := std.slice
decode := std.decode

` we're going to copy README.md to WRITEME.md,
	and we're going to buffer it `
BUFSIZE := 512 ` bytes `

` main routine that reads/writes through buffer
	and recursively copies data. This is also tail-recursive `
copy := (in, out) => incrementalCopy(in, out, 0)
incrementalCopy := (src, dest, offset) => read(src, offset, BUFSIZE, evt => (
	evt.type :: {
		'error' -> log('Encountered an error reading: ' + evt.message)
		'data' -> (
			` compute data size from data response `
			dataLength := len(evt.data)

			` log progress `
			log('copying --> ' + slice(decode(evt.data), 0, 50) + '...')

			` write the read bit, and recurse back to reading `
			write(dest, offset, evt.data, evt => evt.type :: {
				'error' -> log('Encountered an error writing: ' + evt.message)
				'end' -> dataLength = BUFSIZE :: {
					true -> incrementalCopy(src, dest, offset + dataLength)
				}
			})
		)
	}
))

` copy README.md to WRITEME.md `
copy('README.md', 'WRITEME.md')
log('copy scheduled.')

` delete the file, since we don't need it `
wait(2, () => delete('WRITEME.md', evt => evt.type :: {
	'error' -> log('Encountered an error deleting: ' + evt.message)
	'end' -> log('Safely deleted the generated file')
}))
log('delete scheduled.')
