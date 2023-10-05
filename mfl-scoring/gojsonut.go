package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nsf/jsondiff"
)

func JSONCompare(t *testing.T, result interface{}, expectedJSONStr string) {
	outJSONStr, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		t.Fatal("error marshaling the result: ", err)
	}
	diffOpts := jsondiff.DefaultConsoleOptions()
	res, diff := jsondiff.Compare([]byte(expectedJSONStr), outJSONStr, &diffOpts)

	if res != jsondiff.FullMatch {
		fmt.Println("The real output with ident --->")
		fmt.Println(string(outJSONStr))
		t.Errorf("The expected result is not equal to what we have: \n %s", diff)
	}
}
