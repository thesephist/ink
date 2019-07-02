` scan() / in() based prompt demo `

ask := (question, cb) => (
	log(question)
	scan(cb)
)

ask('What\'s your name?', name => (
	log('Great to meet you, ' + name + '!')
))
