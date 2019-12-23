CMD = ./cmd/ink.go
RUN = go run -race ${CMD}
LDFLAGS = -ldflags="-s -w"

all: run test

# run standard samples
run:
	go build -race ${CMD}
	./ink samples/fizzbuzz.ink
	./ink samples/graph.ink
	./ink samples/basic.ink
	./ink samples/kv.ink
	./ink samples/fib.ink
	./ink samples/newton.ink
	./ink samples/pi.ink
	./ink samples/prime.ink
	./ink samples/quicksort.ink
	./ink samples/pingpong.ink
	./ink samples/undefinedme.ink
	./ink samples/error.ink
	./ink samples/exec.ink
	# we echo in some input for prompt.ink testing stdin
	echo 'Linus' | ./ink samples/prompt.ink
	rm ./ink


repl:
	${RUN} -repl


# run just the minimal test suite
test-mini:
	${RUN} samples/test.ink


# run standard test suites
test:
	go build -race ${CMD}
	./ink samples/mangled.ink
	./ink samples/test.ink
	./ink samples/io.ink
	# run I/O test under isolated mode -- all ops should still return valid responses
	# We copy the file in question -- eval.go -- to a temporary location, since
	# no-read and no-write I/O operations will delete the file.
	cp pkg/ink/eval.go tmp.go
	./ink -no-read samples/io.ink
	cp tmp.go pkg/ink/eval.go
	./ink -no-write samples/io.ink
	cp tmp.go pkg/ink/eval.go
	./ink -isolate samples/io.ink
	rm tmp.go
	./ink -isolate samples/pingpong.ink
	./ink -no-exec samples/exec.ink
	# test -eval flag
	./ink -eval "log:=load('samples/std').log,f:=x=>()=>log('Eval test: '+x),f('passed!')()"
	rm ./ink


# build for specific OS target
build-%:
	GOOS=$* GOARCH=amd64 go build ${LDFLAGS} -o ink-$* ${CMD}


# build for all OS targets, useful for releases
build: build-linux build-darwin build-windows build-openbsd


# install on host system
install:
	cp utils/ink.vim ~/.vim/syntax/ink.vim
	go install ${LDFLAGS} ${CMD}
	ls -l `which ink`


# pre-commit hook
precommit:
	go vet ./cmd ./pkg/ink
	go fmt ./cmd ./pkg/ink


# clean any generated files
clean:
	rm -rvf *.bmp ink ink-*
