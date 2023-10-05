package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nsf/jsondiff"
)

func JsonCompare(t *testing.T, result interface{}, expectedJsonStr string) {
	outJsonStr, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		t.Fatal("error marshaling the result: ", err)
	}
	diffOpts := jsondiff.DefaultConsoleOptions()
	res, diff := jsondiff.Compare([]byte(expectedJsonStr), []byte(outJsonStr), &diffOpts)

	if res != jsondiff.FullMatch {
		fmt.Println("The real output with ident --->")
		fmt.Println(string(outJsonStr))
		t.Errorf("The expected result is not equal to what we have: \n %s", diff)
	}
}
