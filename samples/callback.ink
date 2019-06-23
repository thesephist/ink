parent := num => (
    log('running parent with ' + string(num))
    num2 => (
        log('running child with ' + string(num2))
        log('but i know parent was ' + string(num))
    )
)

parent(5)(12)
