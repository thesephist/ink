` a primitive HTTP server `

std := load('std')

log := std.log
encode := std.encode

close := listen('0.0.0.0:8080', evt => (
	log(evt)

	evt.type :: {
		'error' -> log('Error: ' + evt.message)
		'req' -> (evt.end)({
			status: 200
			headers: {'Content-Type': 'text/plain'}
			body: encode('Hello, World!')
		})
	}
))

wait(5, close)
