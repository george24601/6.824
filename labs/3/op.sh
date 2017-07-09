cd src/kvraft
GOPATH=~/6.824
export GOPATH
go test

go test -run "TestBasic"

go test -run "TestBasic" > out 2>err

go test -run "Concurrent"

go test -run "Concurrent" >out 2>err

go test -run "TestPersistConcurrent" >out 2>err
