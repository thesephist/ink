` demonstration of Ink calling into other
	programs with exec(), assumes POSIX system `

std := load('std')

log := std.log

` runs without problems `
log('See: Hello, World!')
exec('echo', ['Hello, World!'], '', s => out(s))

` swallows stdout correctly `
log('See: nothing')
exec('echo', ['Hello, World!'], '', s => ())

` sets args correctly `
log('See: Goodbye, World!')
exec('echo', ['Goodbye,', 'World!'], '', s => out(s))

` runs commands at full paths `
log('See: Hello, Echo!')
exec('/bin/echo', ['Hello, Echo!'], '', s => out(s))

` interprets stdin correctly `
log('See: lovin-pasta')
exec('cat', [], 'lovin-pasta', s => out(s))

` closes before execution `
(
	close := exec('sleep', ['10'], '', s => log('YOU SHOULD NEVER SEE THIS'))
	close()

	` multiple closes do not fail `
	close()
	close()
)

` closes during execution `
(
	log('Should close during execution safely:')
	close := exec('sleep', ['5'], '', s => log('Closed during execution safely!'))
	wait(1, () => close())

	` multiple closes do not fail `
	wait(2, () => close())
)

` closes after execution `
(
	log('Should exit safely, then close:')
	close := exec('sleep', ['1'], '', s => log('Exited safely!'))
	wait(2, () => (
		close()
		log('Closed!')

		` multiple closes do not fail `
		close()
		close()
		close()
	))
)
