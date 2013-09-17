package main

import (
	"encoding/json"
	"github.com/Merovius/go-tap"
)

// Testsuite is a type only used for custom JSON-marshalling
type Testsuite tap.Testsuite

// stats contains all available information about the process-execution
type stats struct {
	SystemTime time.Duration `json:"system_time"`
	UserTime   time.Duration `json:"user_time"`
}

// suitWrap wraps the suits to give all the output, bor gives
type suiteWrap struct {
	Name   string    `json:"name"`
	Suite  Testsuite `json:"suite"`
	Stats  stats     `json:"stats"`
	Error  string    `json:"error,omitempty"`
	Output string    `json:"output,omitempty"`
}

// MarshalJSON marshalls a Testsuite into the format used by bor
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
