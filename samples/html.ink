` html rendering library `

std := load('std')

log := std.log
cat := std.cat
map := std.map
f := std.format

spreadAttrs := attrs => cat(map(keys(attrs), k => (
	v := attrs.(k)
	f('{{ key }}="{{ value }}"', {key: k, value: v})
)), ' ')

el := tag =>
	classes =>
	attrs =>
	children =>
	f('<{{ tag }} class="{{ classes }}" {{ spread }}>{{ children }}</{{ tag }}>', {
		tag: tag
		classes: cat(classes, ' ')
		spread: spreadAttrs(attrs)
		children: cat(children, '')
	})

title := el('title')([])({})
meta := el('meta')
link := el('link')

h1 := el('h1')
h2 := el('h2')
h3 := el('h3')

p := el('p')
em := el('em')
strong := el('strong')

div := el('div')
span := el('span')

` simple div helper `
d := div([])({})

` html wrapper helper `
html := (head, body) => '<!doctype html>' + el('head')([])({})(head) + el('body')([])({})(body)

` example usage `
log(
	html(
		title('Test page')
		d([
			h1(['title'])({itemprop: 'title'})('Hello, World!')
			p(['body'])({})('this is a body paragraph')
		])
	)
)
