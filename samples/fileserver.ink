#!/usr/bin/env ink

` an http static file server
	with support for directory indexes `

std := load('std')

log := std.log
f := std.format
slice := std.slice
cat := std.cat
map := std.map
each := std.each
readFile := std.readFile

DIR := '.'
PORT := 7800
ALLOWINDEX := true

` short non-comprehensive list of MIME types `
TYPES := {
	` text formats `
	html: 'text/html; charset=utf-8'
	js: 'text/javascript; charset=utf-8'
	css: 'text/css; charset=utf-8'
	txt: 'text/plain; charset=utf-8'
	md: 'text/plain; charset=utf-8'
	` serve go & ink source code as plain text`
	ink: 'text/plain; charset=utf-8'
	go: 'text/plain; charset=utf-8'

	` image formats `
	jpg: 'image/jpeg'
	jpeg: 'image/jpeg'
	png: 'image/png'
	gif: 'image/gif'
	svg: 'image/svg+xml'

	` other misc `
	pdf: 'application/pdf'
	zip: 'application/zip'
	json: 'application/json'
}

` prepare standard header `
hdr := attrs => (
	base := {
		'X-Served-By': 'ink-serve'
		'Content-Type': 'text/plain'
	}
	each(keys(attrs), k => base.(k) := attrs.(k))
	base
)

` is this path a path to a directory? `
dirPath? := path => path.(len(path) - 1) :: {
	'/' -> true
	_ -> false
}

` main server handler `
close := listen('0.0.0.0:' + string(PORT), evt => evt.type :: {
	'error' -> log('server error: ' + evt.message)
	'req' -> (
		log(f('{{ method }}: {{ url }}', evt.data))

		` set up timer `
		start := time()
		` trim the elapsed-time millisecond count at 2-3 decimal digits `
		getElapsed := () => slice(string(floor((time() - start) * 1000000) / 1000), 0, 5)

		` normalize path `
		url := trimQP(evt.data.url)

		` respond to file request `
		evt.data.method :: {
			'GET' -> handlePath(url, DIR + url, evt.end, getElapsed)
			_ -> (
				` if other methods, just drop the request `
				log('  -> ' + evt.data.url + ' dropped')
				(evt.end)({
					status: 405
					headers: hdr({})
					body: 'method not allowed'
				})
			)
		}
	)
})

` handles requests to path with given parameters `
handlePath := (url, path, end, getElapsed) => stat(path, evt => evt.type :: {
	'error' -> (
		log(f('  -> {{ url }} led to error in {{ ms }}ms: {{ error }}', {
			url: url
			ms: getElapsed()
			error: evt.message
		}))
		end({
			status: 500
			headers: hdr({})
			body: 'server error'
		})
	)
	'data' -> handleStat(url, path, evt.data, end, getElapsed)
})

` handles requests to directories '/' `
handleStat := (url, path, data, end, getElapsed) => data :: {
	` means file didn't exist `
	() -> (
		log(f('  -> {{ url }} not found in {{ ms }}ms', {
			url: url
			ms: getElapsed()
		}))
		end({
			status: 404
			headers: hdr({})
			body: 'not found'
		})
	)
	{dir: true, name: _, len: _} -> dirPath?(path) :: {
		true -> handleDir(url, path, data, end, getElapsed)
		false -> (
			log(f('  -> {{ url }} returned redirect to {{ url }}/ in {{ ms }}ms', {
				url: url
				ms: getElapsed()
			}))
			end({
				status: 301
				headers: hdr({
					'Location': url + '/'
				})
				body: ''
			})
		)
	}
	{dir: false, name: _, len: _} -> readFile(path, data => handleFileRead(url, path, data, end, getElapsed))
}

` handles requests to readFile() `
handleFileRead := (url, path, data, end, getElapsed) => data :: {
	() -> (
		log(f('  -> {{ url }} failed read in {{ ms }}ms', {
			url: url
			ms: getElapsed()
		}))
		end({
			status: 500
			headers: hdr({})
			body: 'server error'
		})
	)
	_ -> (
		fileType := getType(path)
		log(f('  -> {{ url }} ({{ type }}) served in {{ ms }}ms', {
			url: url
			type: fileType
			ms: getElapsed()
		}))
		end({
			status: 200
			headers: hdr({
				'Content-Type': getType(path)
			})
			body: data
		})
	)
}

` handles requests to directories '/' `
handleDir := (url, path, data, end, getElapsed) => (
	ipath := path + 'index.html'
	stat(ipath, evt => evt.type :: {
		'error' -> (
			log(f('  -> {{ url }} (index) led to error in {{ ms }}ms: {{ error }}', {
				url: url
				ms: getElapsed()
				error: evt.message
			}))
			end({
				status: 500
				headers: hdr({})
				body: 'server error'
			})
		)
		'data' -> evt.data :: {
			() -> handleExistingDir(url, path, end, getElapsed)
			` in the off chance that /index.html is a dir, just render index `
			{dir: true, name: _, len: _} -> handleExistingDir(url, path, end, getElapsed)
			{dir: false, name: _, len: _} -> handlePath(url, ipath, end, getElapsed)
		}
	})
)

` handle a directory we stat() confirmed to exist `
handleExistingDir := (url, path, end, getElapsed) => ALLOWINDEX :: {
	true -> handleNoIndexDir(url, path, end, getElapsed)
	false -> (
		log(f('  -> {{ url }} not allowed in {{ ms }}ms', {
			url: url
			ms: getElapsed()
		}))
		end({
			status: 403
			headers: hdr({})
			body: 'permission denied'
		})
	)
}

` helpers for rendering the directory index page `
makeIndex := (path, items) => '<title>' + path +
	'</title><style>body{font-family: system-ui,sans-serif}</style><h1>index of <code>' +
	path + '</code></h1><ul>' + items + '</ul>'
makeIndexLi := (fileStat, separator) => '<li><a href="' + fileStat.name + '" title="' + fileStat.name + '">' +
	fileStat.name + separator + ' (' + string(fileStat.len) + ' B)</a></li>'

` handles requests to dir() without /index.html `
handleNoIndexDir := (url, path, end, getElapsed) => dir(path, evt => evt.type :: {
	'error' -> (
		log(f('  -> {{ url }} dir() led to error in {{ ms }}ms: {{ error }}', {
			url: url
			ms: getElapsed()
			error: evt.message
		}))
		end({
			status: 500
			headers: hdr({})
			body: 'server error'
		})
	)
	'data' -> (
		log(f('  -> {{ url }} (index) served in {{ ms }}ms', {
			url: url
			ms: getElapsed()
		}))
		end({
			status: 200
			headers: hdr({
				'Content-Type': 'text/html'
			})
			body: makeIndex(
				slice(path, 2, len(path))
				cat(map(evt.data, fileStat => makeIndexLi(
					fileStat
					fileStat.dir :: {
						true -> '/'
						false -> ''
					}
				)), '')
			)
		})
	)
})

` trim query parameters `
trimQP := path => (
	max := len(path)
	(sub := (idx, acc) => idx :: {
		max -> path
		_ -> path.(idx) :: {
			'?' -> acc
			_ -> sub(idx + 1, acc + path.(idx))
		}
	})(0, '')
)

` given a path, get the MIME type `
getType := path => (
	guess := TYPES.(getPathEnding(path))
	guess :: {
		() -> 'application/octet-stream'
		_ -> guess
	}
)

` given a path, get the file extension `
getPathEnding := path => (
	(sub := (idx, acc) => idx :: {
		0 -> path
		_ -> path.(idx) :: {
			'.' -> acc
			_ -> sub(idx - 1, path.(idx) + acc)
		}
	})(len(path) - 1, '')
)
