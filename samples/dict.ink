` composite (dict) `

obj := {
    5: 'fifth item',
    hi: 'hi text',
    hello: 'hello text',
    what3: 3.14,
}

log(obj.hi)
log(obj.hello)
log(string(obj.what3))
log(obj.5)


` composite (list) `

arr := [3, 2, 1, 'four']

log(string(arr.2))
log(arr.3)

` property access and assignment `

obj.hi := 8
out('should be 8: ')
log(string(obj.hi))
