` scan() / in() based prompt demo `

std := load('std')

log := std.log
scan := std.scan

ask := (question, cb) => (
	log(question)
	scan(cb)
)

ask('What\'s your name?', name => (
	log('Great to meet you, ' + name + '!')
))
