package mapreduce

import (
	"hash/fnv"
	"os"
	"encoding/json"
	"io/ioutil"
)

func closeFiles(outFiles []*os.File) {
	for _, file  := range outFiles {
		file.Close()
	}
}

func doMap(
	jobName string, // the name of the MapReduce job
	mapTaskNumber int, // which map task this is
	inFile string,
	nReduce int, // the number of reduce task that will be run ("R" in the paper)
	mapF func(file string, contents string) []KeyValue,
) {

	outFiles := make([]*os.File, nReduce)
	encoders := make([]*json.Encoder, nReduce)

	for reduceTask  := range outFiles {
		outName := reduceName(jobName, mapTaskNumber, reduceTask)
		outFile, _ := os.Create(outName)

		outFiles[reduceTask] = outFile
		encoder := json.NewEncoder(outFile)
		encoders[reduceTask] = encoder

	}

	defer closeFiles(outFiles) 

	dat, _ := ioutil.ReadFile(inFile)

	kvs := mapF(inFile, string(dat))

	for _, kv := range kvs {
		target := ihash(kv.Key) % uint32(nReduce)
		encoders[target].Encode(&kv)
	}
}

func ihash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
