#!/bin/bash

#create soft link
ln -s tests/cvl_test.go cvl_test.go
ln -s tests/jsondata_test.go jsondata_test.go

profiling=""
testcase=""
coverpkgs="-coverpkg=cvl,cvl/internal/util,cvl/internal/yparser"

if [ "${BUILD}:" != ":" ] ; then
	go test -v -c -gcflags="all=-N -l" 
fi

if [ "${TESTCASE}:" != ":" ] ; then
	testcase="-run ${TESTCASE}"
fi

if [ "${PROFILE}:" != ":" ] ; then
	profiling="-bench=. -benchmem -cpuprofile profile.out"
fi

#Run test and display report
if [ "${NOREPORT}:" != ":" ] ; then
	go test  -v -cover ${coverpkgs} ${testcase}
elif [ "${COVERAGE}:" != ":" ] ; then
	go test  -v -cover -coverprofile coverage.out ${coverpkgs} ${testcase}
	go tool cover -html=coverage.out
else
	go test  -v -cover -json ${profiling} ${testcase} | tparse -smallscreen -all
fi

#With profiling 
#go test  -v -cover -json -bench=. -benchmem -cpuprofile profile.out | tparse -smallscreen -all

#delete soft link
rm -rf cvl_test.go jsondata_test.go

