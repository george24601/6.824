ct -run 2B src/raft
GOPATH=~/6.824
export GOPATH

#report race conditions
go test -race

###part 2A
go test -run 2A > out


###part 2B
go test -run 2B > out


####part 2C
simple case
go test -run 'TestPersist12C' > out

#this one often fails!
go test -run 'TestReliableChurn2C' > out

go test -run 2C > out
