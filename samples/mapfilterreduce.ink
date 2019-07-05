` map, filter, reduce demos `

std := load('std')

log := std.log
stringList := std.stringList

map := std.map
filter := std.filter
reduce := std.reduce
reverse := std.reverse
join := std.join

list := [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

log('Mapped 1-10 list, squared
-> ' + stringList(map(list, n => n * n)))

log('Filtered 1-10 list, evens
-> ' + stringList(filter(list, n => n % 2 = 0)))

log('Reduced 1-10 list, multiplication
-> ' + string(reduce(list, (acc, n) => acc * n, 1)))

log('Reversing 1-10 list
-> ' + stringList(reverse(list)))

log('Adding the list to itself
-> ' + stringList(join(list, list)))
