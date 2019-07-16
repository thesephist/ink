` basic key-value storage library
	built on composite values`

log := load('std').log

makeGet := store => key => store.(key)
makeSet := store => (key, val) => store.(key) := val
makeDelete := store => key => store.(key) := ()

create := () => (
	store := {}

	{
		type: 'kv-store'
		store: store
		get: makeGet(store)
		set: makeSet(store)
		delete: makeDelete(store)
	}
)
