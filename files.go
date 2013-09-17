package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// File stores a decoded and uncompressed file
type File struct {
	b []byte
	r io.Reader
}

// UnmarshalJSON reads a gzipped, base64 encoded file from a JSON-string. The
// Uncompressed form is stored as a slice as well as a reader, for convenience
func (f *File) UnmarshalJSON(b64 []byte) error {
	raw := bytes.NewReader(b64[1 : len(b64)-1])
	dec := base64.NewDecoder(base64.StdEncoding, raw)
	unc, err := gzip.NewReader(dec)
	if err != nil {
		return err
	}
	defer unc.Close()

	content, err := ioutil.ReadAll(unc)
	if err != nil {
		return err
	}

	f.b = content
	f.r = bytes.NewReader(content)

	return nil
}

// Message is the type of a Request to bor
type Message struct {
	Suites []Suite         `json:"suites"`
	Files  map[string]File `json:"files"`
}

// Suite contains all information about what files to use in a Testsuite
type Suite struct {
	Name string   `json:"name"`
	Link []string `json:"link"`
}

// CreateBuildDir writes all files in msg as well as the Makefile needed to
// build everything into a temporary directory and returns its name.
func CreateBuildDir(msg Message) (build string, err error) {
	// We use this to parallelize the IO-operations
	numgo := 0
	godone := make(chan bool, 100)

	build, err = ioutil.TempDir(conf.TmpDir, conf.TmpPrefix)
	if err != nil {
		return
	}

	// We create mk at the beginning, so we can directly write all informations
	// to the Makefile, as we create them
	mk, err := os.Create(path.Join(build, "Makefile"))
	if err != nil {
		return
	}
	defer mk.Close()

	// We start of with a Makefile template, containing some variables and the
	// basic rule to build object-files
	mktpl, err := os.Open(conf.MakefileTemplate)
	if err != nil {
		return
	}
	defer mktpl.Close()

	_, err = io.Copy(mk, mktpl)
	if err != nil {
		return
	}

	// Write Files to temporary directory
	for name, content := range msg.Files {
		dst, err := os.Create(path.Join(build, name))
		if err != nil {
			return build, err
		}

		numgo++
		// We write the file in a seperate goroutine to parallelize IO as much as possible.
		// We pass the reader as a parameter, to prevent a race with the closure
		go func(r io.Reader) {
			io.Copy(dst, r)
			dst.Close()
			godone <- true
		}(content.r)
	}

	// This is where most of the Makefile-magic happens. For every Testsuite we
	// create a rule, containing the dependencies. The default-rule for
	// building object-files takes depenency tracking and everything off our
	// hands
	var testprogs []string
	for _, suite := range msg.Suites {
		if len(suite.Link) == 0 {
			return build, fmt.Errorf("No files to link given in suite", suite.Name)
		}
		link := strings.Join(suite.Link, ".o ") + ".o"
		fmt.Fprintf(mk, "%s: TAPListener.o %s\n", suite.Name, link)
		fmt.Fprintf(mk, "\t$(CXX) $(CXXFLAGS) $(LDFLAGS) -o %s %s TAPListener.o\n\n", suite.Name, link)

		// We keep track of all the Testsuites we want to build to put them in
		// the dependency list of the all-target
		testprogs = append(testprogs, suite.Name)
	}

	// We copy the TAPListener.cpp into the builddirectory
	src, err := os.Open(conf.TAPListener)
	if err != nil {
		return
	}
	dst, err := os.Create(path.Join(build, "TAPListener.cpp"))
	if err != nil {
		return
	}

	numgo++
	go func() {
		io.Copy(dst, src)
		dst.Close()
		src.Close()
		godone <- true
	}()

	// Write the all-target. Now just executing make will build all Testsuites
	// including dependencies
	fmt.Fprintf(mk, "all: %s\n", strings.Join(testprogs, " "))
	mk.Close()

	// Wait for all writes to finish
	for ; numgo > 0; numgo-- {
		<-godone
	}

	return build, nil
}
