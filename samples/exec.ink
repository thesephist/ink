` demonstration of Ink calling into other
	programs with exec(), assumes POSIX system `

std := load('std')

log := std.log

handleExec := evt => evt.type :: {
	'error' -> log(evt.message)
	_ -> out(evt.data)
}

` runs without problems `
log('See: Hello, World!')
exec('echo', ['Hello, World!'], '', handleExec)

` swallows stdout correctly `
log('See: nothing')
exec('echo', ['Hello, World!'], '', evt => evt.type :: {
	'error' -> log(evt.message)
	_ -> ()
})

` sets args correctly `
log('See: Goodbye, World!')
exec('echo', ['Goodbye,', 'World!'], '', handleExec)

` runs commands at full paths `
log('See: Hello, Echo!')
exec('/bin/echo', ['Hello, Echo!'], '', handleExec)

` interprets stdin correctly `
log('See: lovin-pasta')
exec('cat', [], 'lovin-pasta', handleExec)

` closes immediately after exec `
(
	log('Should close immediately after exec safely (may not run):')
	close := exec('sleep', ['10'], '', () => log('Closed immediately after exec safely!'))
	close()

	` multiple closes do not fail `
	close()
	close()
)

` closes during execution `
(
	log('Should close during execution safely:')
	close := exec('sleep', ['5'], '', () => log('Closed during execution safely!'))
	wait(1, close)

	` multiple closes do not fail `
	wait(2, close)
)

` closes after execution `
(
	log('Should exit safely, then close:')
	close := exec('sleep', ['1'], '', () => log('Exited safely!'))
	wait(2, () => (
		close()
		log('Closed!')

		` multiple closes do not fail `
		close()
		close()
		close()
	))
)
