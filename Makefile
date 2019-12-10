CMD = ./cmd/ink.go
RUN = go run -race ${CMD}
LDFLAGS = -ldflags="-s -w"

all: run test install

# run standard samples
run:
	# we echo in some input for prompt.ink
	echo 'Linus' | ${RUN} \
		samples/fizzbuzz.ink \
		samples/graph.ink \
		samples/basic.ink \
		samples/kv.ink \
		samples/fib.ink \
		samples/newton.ink \
		samples/pi.ink \
		samples/prime.ink \
		samples/quicksort.ink \
		samples/pingpong.ink \
		samples/undefinedme.ink \
		samples/error.ink \
		samples/exec.ink \
		samples/prompt.ink


# run just the minimal test suite
test-mini:
	${RUN} samples/test.ink


# run standard test suites
test:
	${RUN} \
		samples/mangled.ink \
		samples/test.ink \
		samples/io.ink
	# run I/O test under isolated mode -- all ops should still return valid responses
	${RUN} -no-read samples/io.ink
	${RUN} -no-write samples/io.ink
	${RUN} -isolate samples/io.ink
	${RUN} -isolate samples/pingpong.ink
	# test -eval flag
	${RUN} -eval "log:=load('samples/std').log,f:=x=>()=>log('Eval test: '+x),f('passed!')()"


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
