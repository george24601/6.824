cd src/kvraft
GOPATH=~/6.824
export GOPATH

go test -race > out 2>err


go test > out 2>err

go test -run "TestConcurrent"



go test -run "TestPersistPartitionUnreliable" >out 2>err



go test -run "TestBasic" > out 2>err
go test -run "TestPersistConcurrent" >out 2>err
go test -run "TestConcurrent" >out 2>err


go test -race -run "TestPersistOneClient" >out 2>err

go test -run "TestSnapshotRPC" >out 2>err
go test -run "TestSnapshotSize" >out 2>err

go test -run "TestSnapshotRecover" >out 2>err

go test -run "TestSnapshotRecoverManyClients" >out 2>err

go test -run "TestSnapshotUnreliable" >out 2>err

go test -run "TestSnapshotUnreliableRecover" >out 2>err

go test -run "TestSnapshotUnreliableRecoverConcurrentPartition" >out 2>err

