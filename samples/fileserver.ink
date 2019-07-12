#!/usr/bin/env ink

` a primitive HTTP static file server `

DIR := '.'
PORT := 7800

` short non-comprehensive list of MIME types `
TYPES := {
	` text formats `
	html: 'text/html'
	js: 'text/javascript'
	css: 'text/css'
	txt: 'text/plain'
	md: 'text/plain'
	` serve ink source code as plain text`
	ink: 'text/plain'

	` image formats `
	jpg: 'image/jpeg'
	jpeg: 'image/jpeg'
	png: 'image/png'
	gif: 'image/gif'
	svg: 'image/svg+xml'

	` other misc `
	pdf: 'application/pdf'
	zip: 'application/zip'
}

std := load('std')

log := std.log
encode := std.encode
readRawFile := std.readRawFile

close := listen('0.0.0.0:' + string(PORT), evt => (
	evt.type :: {
		'error' -> log('Server error: ' + evt.message)
		'req' -> (
			` normalize path `
			path := DIR + evt.data.url
			path := (path.(len(path) - 1) :: {
				'/' -> path + 'index.html'
				_ -> path
			})

			log(evt.data.method + ': ' + evt.data.url + ', type: ' + getType(path))
		
			` pre-define callback to readRawFile `
			readHandler := fileBody => fileBody :: {
				() -> (
					log('-> ' + path + ' not found')
					(evt.end)({
						status: 404
						headers: {
							'Content-Type': 'text/plain'
							'X-Served-By': 'ink-serve'
						}
						body: encode('not found'),
					})
				)
				_ -> (
					log('-> ' + evt.data.url + ' served')
					(evt.end)({
						status: 200
						headers: {
							'Content-Type': getType(path)
							'X-Served-By': 'ink-serve'
						}
						body: fileBody,
					})
				)
			}

			evt.data.method :: {
				'GET' -> readRawFile(path, readHandler)
				_ -> (
					` if other methods, just drop the request `
					log('-> ' + evt.data.url + ' dropped')
					(evt.end)({
						status: 405
						headers: {
							'Content-Type': 'text/plain'
							'X-Served-By': 'ink-serve'
						}
						body: encode('method not allowed'),
					})
				)
			}
		)
	}
))

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
