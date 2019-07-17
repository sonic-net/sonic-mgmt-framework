#!/bin/bash

#create soft link
ln -s tests/cvl_test.go cvl_test.go
ln -s tests/jsondata_test.go jsondata_test.go

#Run test and displat report
go test  -v -cover -json | tparse -smallscreen -all
#go test  -v -c   -gcflags="all=-N -l"

#With profiling 
#go test  -v -cover -json -bench=. -benchmem -cpuprofile profile.out | tparse -smallscreen -all

#delete soft link
rm -rf cvl_test.go jsondata_test.go

