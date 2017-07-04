cd src/raft
GOPATH=~/6.824
export GOPATH

#report race conditions
go test -race

###part 2A
go test -run 2A

> out


###part 2B
go test -run 2B

go test -run 2B > out
