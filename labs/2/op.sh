ct -run 2B src/raft
GOPATH=~/6.824
export GOPATH

#report race conditions, MAKE SURE YOU RUN IT AT EVERY STEP
go test -race

###part 2A
go test -run 2A > out


###part 2B
go test -run 2B > out


####part 2C
#simple case 
go test -run 'TestPersist12C' > out

#often race case
go test -race -run 'TestPersist22C' > out 2>error

#this one often fails!
go test -run 'TestReliableChurn2C' > out

go test -race -run 'TestFigure8Unreliable2C'

go test -run 2C > out
