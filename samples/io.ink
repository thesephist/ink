` filesystem i/o demo `

SOURCE := 'eval.go'
TARGET := 'sub.go'

` std imports `
std := load('std')

log := std.log
slice := std.slice
decode := std.decode
rf := std.readFile
wf := std.writeFile

` we're going to copy main.go to sub.go,
	and we're going to buffer it `
BUFSIZE := 1024 ` bytes `

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
			log('copying --> ' + slice(decode(evt.data), 0, 8) + '...')

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

` copy main.go to sub.go`
copy(SOURCE, TARGET)
log('copy scheduled.')

` delete the file, since we don't need it `
log('Delete scheduled at ' + string(time()))
wait(4, () => (
	log('Delete fired at ' + string(time()))
	delete('sub.go', evt => evt.type :: {
		'error' -> log('Encountered an error deleting: ' + evt.message)
		'end' -> log('Safely deleted the generated file')
	}))
)

` as test, schedule a copy-back task in between copy and delete `
log('Copy-back scheduled at ' + string(time()))
wait(2, () => (
	log('Copy-back fired at ' + string(time()))
	rf(TARGET, data => data :: {
		() -> log('Error copying-back WRITEME.md')
		_ -> wf(SOURCE, data, () => log('Copy-back done!'))
	})
))
