` simple ping-pong request-response test over HTTP `

std := load('std')

log := std.log
encode := std.encode
decode := std.decode
f := std.format

` helper for logging errors `
logErr := msg => log('error: ' + msg)

` start a server `
closeServer := listen('0.0.0.0:9600', evt => evt.type :: {
	'error' -> logErr(evt.message)
	'req' -> (
		log(f('Request ---> {{ data }}', evt))

		dt := evt.data
		end := evt.end
		[dt.method, dt.url, dt.body] :: {
			['POST', '/test', encode('ping')] -> end({
				status: 302 ` test that it doesn't auto-follow redirects `
				headers: {
					'Content-Type': 'text/plain'
					'Location': 'https://dotink.co'
				}
				body: encode('pong')
			})
			_ -> end({
				status: 400
				headers: {
					'Content-Type': 'text/plain'
				}
				body: encode('invalid request!')
			})
		}
	)
})

` send a request `
closeRequest := req({
	method: 'POST'
	url: 'http://127.0.0.1:9600/test'
	headers: {
		'Accept': 'text/html'
	}
	body: encode('ping')
}, evt => evt.type :: {
	'error' -> logErr(evt.message)
	'resp' -> (
		log(f('Response ---> {{ data }}', evt))

		dt := evt.data
		[dt.status, dt.body] :: {
			[302, encode('pong')] -> (
				log('---> ping-pong, success!')
				closeServer()
			)
			_ -> logErr('communication failed!')
		}
	)
})

` half-second timeout on the request `
wait(0.5, closeRequest)
