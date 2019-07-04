` scan() / in() based prompt demo `

log := load('std').log

ask := (question, cb) => (
	log(question)
	scan(cb)
)

ask('What\'s your name?', name => (
	log('Great to meet you, ' + name + '!')
))
