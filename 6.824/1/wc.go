package main

import (
	"fmt"
	"mapreduce"
	"os"
	"strings"
	"strconv"
	"unicode"
)

func mapF(document string, value string) (res []mapreduce.KeyValue) {

	tokens := strings.FieldsFunc(value, func(c rune) bool { 
		return !unicode.IsLetter(c) 
	})

	toReturn := make([]mapreduce.KeyValue,0)	

	for _, token := range tokens {
		toReturn = append(toReturn, mapreduce.KeyValue{token, "1"})
	}

	return toReturn;
}

func reduceF(key string, values []string) string {
	return strconv.Itoa(len(values))
}
