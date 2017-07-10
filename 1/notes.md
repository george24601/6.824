Preview Question
----------
How soon after it receives the first file of intermediate data can a reduce worker start calling the application's Reduce function? Explain your answer.


Review Questions
----------
Sample word count MapReduce(MR) pesudo-code
Why MR requires the function purely functional
Why many more input splits than workers?
During the whole MR process, what files are written into GFS?
What if the master accidentally starts *two* Map() workers on same input?
 What if two Reduce() workers for the same partition of intermediate data?
Examples that does not fit MR model

BEFORE YOU START part 1, you need to understand
1.Slice in Go

2.Pointer in Go

3.Map in Go

```
//To read the whole file
dat, err := ioutil.ReadFile("/tmp/dat")
str = string(dat)

//To write to file and serialize in JSON
file,err := os.Create("outPath" )
encoder := json.NewEncoder(file)
//Encode accepts pointer to object, otherwise it is copy by value
encoder.Encode(&object)
file.Close

//To read from file in JSON and deserialize from JSON
file,err := os.Open("outPath" )
decoder := json.NewDecoder(file)

//read it token by token
for {
	var obj Type
	//aceepts a pointer similar to encoder
	err := decoder.Decode(&obj)
	if(err != nil)
	{
		break //hits EOF
	}
}
file.Close
```

Before you start part 3, you need to understand
1. goroutine

2. channel 

3. RPC in go

See master.go , and worker.go for how RPC is used in this library
