package gocommons

import (
	"fmt"
	"testing"
)

func InitResult(name string) string {
	return fmt.Sprintf("%s\t:", name)
}

func HandleResult(t *testing.T, success bool, result string) {
	if success != true {
		result = fmt.Sprintf("%sFAIL", result)
	} else {
		result = fmt.Sprintf("%sPASS", result)
	}
	fmt.Println(result)
	if success != true {
		t.Fail()
	}
}
