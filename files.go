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

type File struct {
	b []byte
	r io.Reader
}

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

type Suite struct {
	Name string   `json:"name"`
	Link []string `json:"link"`
}

type Message struct {
	Suites []Suite         `json:"suites"`
	Files  map[string]File `json:"files"`
}

func CreateBuildDir(msg Message) (build string, err error) {
	numgo := 0
	godone := make(chan bool, 100)

	build, err = ioutil.TempDir(conf.TmpDir, conf.TmpPrefix)
	if err != nil {
		return
	}

	mk, err := os.Create(path.Join(build, "Makefile"))
	if err != nil {
		return
	}
	defer mk.Close()

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
		// We pass the reader as a parameter, to prevent a race with the closure
		go func(r io.Reader) {
			io.Copy(dst, r)
			dst.Close()
			godone <- true
		}(content.r)
	}

	// Write Makefile
	var testprogs []string
	for _, suite := range msg.Suites {
		if len(suite.Link) == 0 {
			return build, fmt.Errorf("No files to link given in suite", suite.Name)
		}
		link := strings.Join(suite.Link, ".o ") + ".o"
		fmt.Fprintf(mk, "%s: cppunit_main.o %s\n", suite.Name, link)
		fmt.Fprintf(mk, "\t$(CXX) $(CXXFLAGS) $(LDFLAGS) -o %s %s cppunit_main.o\n\n", suite.Name, link)

		testprogs = append(testprogs, suite.Name)
	}

	src, err := os.Open(conf.CppunitMain)
	if err != nil {
		return
	}
	dst, err := os.Create(path.Join(build, "cppunit_main.cpp"))
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

	fmt.Fprintf(mk, "all: %s\n", strings.Join(testprogs, " "))
	mk.Close()

	for ; numgo > 0; numgo-- {
		<-godone
	}
	return build, nil
}
