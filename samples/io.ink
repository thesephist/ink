` filesystem i/o demo `

` std imports `
std := load('std')

log := std.log
slice := std.slice
decode := std.decode
map := std.map
each := std.each
f := std.format
stringList := std.stringList
rf := std.readFile
wf := std.writeFile

SOURCE := 'pkg/ink/eval.go'
TARGET := 'test_io.go'

` we're going to copy SOURCE to TARGET
	and we're going to buffer it `
BUFSIZE := 4096 ` bytes `

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
			log('copying --> ' + slice(evt.data, 0, 8) + '...')

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

copy(SOURCE, TARGET)
log('Copy scheduled at ' + string(time()))

` delete the file, since we don't need it `
wait(1, () => (
	log('Delete fired at ' + string(time()))
	delete(TARGET, evt => evt.type :: {
		'error' -> log('Encountered an error deleting: ' + evt.message)
		'end' -> log('Safely deleted the generated file')
	}))
)
log('Delete scheduled at ' + string(time()))

` as concurrency test, schedule a copy-back task in between copy and delete `
wait(0.5, () => (
	log('Copy-back fired at ' + string(time()))
	rf(TARGET, data => data :: {
		() -> log('Error copying-back ' + TARGET)
		_ -> wf(SOURCE, data, () => log('Copy-back done!'))
	})
))
log('Copy-back scheduled at ' + string(time()))

` while scheduled tasks are running, create and delete a directory `
testdir := 'ink_io_test_dir'
make(testdir, evt => evt.type :: {
	'error' -> log('dir() error: ' + evt.message)
	'end' -> (
		log('Created test directory...')
		delete(testdir, evt => evt.type :: {
			'error' -> log('delete() of dir error: ' + evt.message)
			'end' -> log('Deleted test directory.')
		})
	)
})

` test stat: show file data for README.md, samples/, and current dir `
each(['.', 'samples', 'README.md', 'fake.txt'], path => stat(path, evt => evt.type :: {
	'error' -> log('Error stat ' + path + ': ' + evt.message)
	'data' -> evt.data :: {
		() -> log(f('{{ path }} does not exist', {path: path}))
		_ -> log(f('{{ name }}{{ sep }}: {{ len }}B mod:{{ mod }}', {
			name: evt.data.name
			len: evt.data.len
			mod: evt.data.mod
			sep: evt.data.dir :: {
				true -> '/'
				false -> ''
			}
		}))
	}
}))

` test dir(): list all samples and file sizes `
dir('./samples', evt => evt.type :: {
	'error' -> log('Error listing samples: ' + evt.message)
	'data' -> log(stringList(map(evt.data, file => f('{{ name }} ({{ len }}B mod:{{ mod }})', file))))
})
