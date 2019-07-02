` scan() / in() based prompt demo `

ask := (question, cb) => (
	log(question)
	scan(cb)
)

ask('What\'s your first name?', first => (
	ask('What about your last name?', last => (
		log('Great to meet you, ' + first + ' ' + last + '!')
	))
))
