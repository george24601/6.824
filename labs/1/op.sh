######Part I
cd 6.824
export "GOPATH=$PWD"  # go needs $GOPATH to be set to the project's working directory
cd "$GOPATH/src/mapreduce"
go test -run Sequential mapreduce/... 

#to debug, set debugEnabled = true in common.go, and run
env "GOPATH=$PWD/../../" go test -v -run Sequential mapreduce/... 

######Part 2
cd "$GOPATH/src/main"
time go run wc.go master sequential pg-*.txt

#compare it with given result
sort -n -k2 mrtmp.wcseq | tail -10

#remove intermediate 
rm mrtmp.*


#### part 3
#debug is similar
go test -run TestBasic mapreduce/...


##Skip part 4 since it is not very relevent

#run ALL tests
sh src/main/test-mr.sh

