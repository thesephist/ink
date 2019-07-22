RUN = go run -race .
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
		samples/callback.ink \
		samples/pi.ink \
		samples/prime.ink \
		samples/quicksort.ink \
		samples/pingpong.ink \
		samples/undefinedme.ink \
		samples/error.ink \
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
	# test -eval flag
	${RUN} -eval "log:=load('samples/std').log,f:=x=>()=>log('Eval test: '+x),f('passed!')()"


# build for specific OS target
build-%:
	GOOS=$* GOARCH=amd64 go build ${LDFLAGS} -o ink-$*


# build for all OS targets, useful for releases
build: build-linux build-darwin build-windows build-openbsd


# install on host system
install:
	cp utils/ink.vim ~/.vim/syntax/ink.vim
	go install ${LDFLAGS}
	ls -l `which ink`


# pre-commit hook
precommit:
	go vet .
	go fmt .


# clean any generated files
clean:
	rm -rvf *.bmp ink-*
