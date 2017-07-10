package mapreduce

import (
	"os"
	"encoding/json"
//	"fmt"
)

func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	kToVs := make(map[string][]string)

	//iterate through each mapper and find potential reducer file assigned to this reducer by each mapper
	for mapTaskNumber := 0; mapTaskNumber < nMap; mapTaskNumber++ {
		inFileName := reduceName(jobName, mapTaskNumber, reduceTaskNumber)
		file, _ := os.Open(inFileName)

		decoder := json.NewDecoder(file)

		for {
			var kv KeyValue 
			err := decoder.Decode(&kv)

			if err != nil{
				break //EOF will return this
			}


			k := kv.Key
			_ , ok := kToVs[k] 

			//k does not exist in the map yet
			if !ok {
				kToVs[k] = make([]string, 0)
			}

			kToVs[k] =append(kToVs[k], kv.Value) 
		}

		file.Close()
	}

	outFileName := mergeName(jobName, reduceTaskNumber)
	file, _ := os.Create(outFileName)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for k, v := range kToVs {
		res := reduceF(k, v)
		encoder.Encode(KeyValue{k, res})
	}
}
