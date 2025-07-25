#!/bin/sh

PROJECT=gpp

VERBOSITY=0
VERBOSITYFLAGS=""
while test "$1" = "-v"; do
	VERBOSITY=$((VERBOSITY+1))
	VERBOSITYFLAGS="$VERBOSITYFLAGS -v"
	shift
done

run()
{
	if test $VERBOSITY -gt 1; then echo "$@"; fi
	"$@" || exit 1
}

printv()
{
	if test $VERBOSITY -gt 0; then echo "$@"; fi
}

# Switch to Go 1.4.
# . go14-env

# NOTE(anton2920): don't like Google spying on me.
GOPROXY=direct; export GOPROXY
GOSUMDB=off; export GOSUMDB

# NOTE(anton2920): disable Go 1.11+ package management.
GO111MODULE=off; export GO111MODULE
GOPATH=`go env GOPATH`:`pwd`; export GOPATH

CGO_ENABLED=0; export CGO_ENABLED

STARTTIME=`date +%s`

case $1 in
	'' | debug)
		CGO_ENABLED=1; export CGO_ENABLED
		run go build $VERBOSITYFLAGS -o $PROJECT -race -gcflags="-N -l" -tags gofadebug
		;;
	clean)
		run rm -f $PROJECT $PROJECT.s $PROJECT.esc $PROJECT.test c.out cpu.pprof mem.pprof
		run go clean
		run rm -rf `go env GOCACHE`
		run rm -rf /tmp/cover*
		;;
	check)
		run $0 $VERBOSITYFLAGS test-race-cover
		run ./$PROJECT.test
		;;
	check-bench)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=8 -test.benchtime=10000x
		;;
	check-bench-cpu)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.cpuprofile=$PROJECT-cpu.pprof -test.count=8 -test.benchtime=10000x
		;;
	check-bench-mem)
		run $0 $VERBOSITYFLAGS test
		run ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.memprofile=$PROJECT-mem.pprof
		;;
	check-bench-tracing)
		run $0 $VERBOSITYFLAGS test-tracing
		run ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=8 -test.benchtime=10000x
		;;
	check-cover)
		run $0 $VERBOSITYFLAGS test-race-cover
		run ./$PROJECT.test -test.coverprofile=c.out
		run go tool cover -html=c.out
		run rm -f c.out
		;;
	check-msan)
		run $0 $VERBOSITYFLAGS test-msan
		run ./$PROJECT.test
		;;
	disas | disasm | disassembly)
		printv go build $VERBOSITYFLAGS -gcflags="-S"
		go build $VERBOSITYFLAGS -o $PROJECT -gcflags="-S" >$PROJECT.s 2>&1
		;;
	esc | escape | escape-analysis)
		printv go build $VERBOSITYFLAGS -gcflags="-m -m"
		go build $VERBOSITYFLAGS -o $PROJECT -gcflags="-m -m" >$PROJECT.m 2>&1
		;;
	fmt)
		if which goimports >/dev/null; then
			run goimports -l -w *.go
		else
			run gofmt -l -s -w *.go
		fi
		;;
	objdump)
		go build $VERBOSITYFLAGS -o $PROJECT
		printv go tool objdump -S $PROJECT
		go tool objdump -S $PROJECT >$PROJECT.s
		;;
	pgo)
		run $0 $VERBOSITYFLAGS test
		check_db_variable

		printv ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=10 -test.cpuprofile=$PROJECT-cpu.pprof
		./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=10 -test.cpuprofile=$PROJECT-cpu.pprof | tee $PROJECT-before.txt

		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -pgo=$PROJECT-cpu.pprof -vet=off

		printv ./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=10 -test.cpuprofile=$PROJECT-cpu.pprof.tmp
		./$PROJECT.test -test.run=^Benchmark -test.benchmem -test.bench=. -test.count=10 -test.cpuprofile=$PROJECT-cpu.pprof.tmp | tee $PROJECT-after.txt

		printv benchstat $PROJECT-before.txt $PROJECT-after.txt
		benchstat $PROJECT-before.txt $PROJECT-after.txt >$PROJECT-diff.txt

		run rm $PROJECT-before.txt $PROJECT-after.txt $PROJECT-cpu.pprof.tmp
		;;
	png)
		run go tool pprof -png $PROJECT-cpu.pprof >$PROJECT-cpu.png
		;;
	profiling)
		run go build $VERBOSITYFLAGS -o $PROJECT -ldflags="-X main.BuildMode=Profiling"
		;;
	release)
		run go build $VERBOSITYFLAGS -o $PROJECT -ldflags="-s -w"
		;;
	test)
		# run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off
		;;
	test-msan)
		CGO_ENABLED=1; export CGO_ENABLED
		# run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off -msan -gcflags="all=-N -l"
		;;
	test-race-cover)
		CGO_ENABLED=1; export CGO_ENABLED
		# run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off -race -cover -gcflags="all=-N -l"
		;;
	test-tracing)
		# run $0 $VERBOSITYFLAGS vet
		run go test $VERBOSITYFLAGS -c -o $PROJECT.test -vet=off -tags gofatrace
		;;
	tracing)
		run go build $VERBOSITYFLAGS -o $PROJECT -tags gofatrace
		;;
	vet)
		run go vet $VERBOSITYFLAGS
		;;
	*)
		echo "Target $1 is not supported!"
		;;
esac

ENDTIME=`date +%s`

echo Done $1 in $((ENDTIME-STARTTIME))s
