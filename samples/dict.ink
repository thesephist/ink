` composite (dict) `

log := load('std').log

obj := {
	5: 'fifth item',
	hi: 'hi text',
	hello: 'hello text',
	what3: 3.14,
}

log(obj.hi)
log(obj.hello)
log(obj.what3)
log(obj.5)


` composite (list) `

arr := [3, 2, 1, 'four']

log(arr.2)
log(arr.3)

` property access and assignment `

main := () => (
	obj.hi := 8
	out('should be 8: ')
	log(obj.hi)
)
main()
