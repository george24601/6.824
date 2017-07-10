cd src/kvraft
GOPATH=~/6.824
export GOPATH

go test> out 2>err

go test -run "TestBasic"


go test -run "TestConcurrent"



go test -run "TestPersistPartitionUnreliable" >out 2>err



go test -run "TestBasic" > out 2>err
go test -run "TestPersistConcurrent" >out 2>err
go test -run "TestConcurrent" >out 2>err
