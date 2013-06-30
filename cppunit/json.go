package main

import (
	"encoding/json"
	"github.com/Merovius/go-tap"
)

type Testsuite tap.Testsuite

func (t *Testsuite) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	m["ok"] = t.Ok

	var tests []map[string]interface{}
	for _, tl := range t.Tests {
		tm := make(map[string]interface{})
		tm["ok"] = tl.Ok
		tm["description"] = tl.Description
		tm["diagnostic"] = tl.Diagnostic
		tests = append(tests, tm)
	}
	m["tests"] = tests
	return json.Marshal(m)
}
