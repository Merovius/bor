package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type File []byte

func (f *File) UnmarshalJSON(b64 []byte) error {
	b64 = b64[1 : len(b64)-1]
	*f = make([]byte, (len(b64)*3)/4)
	base64.StdEncoding.Decode(*f, b64)
	return nil
}

type Suite struct {
	Name    string `json:"name"`
	Content File   `json:"content"`
}

type Solution struct {
	Name    string  `json:"name"`
	Content File    `json:"content"`
	Suites  []Suite `json:"suites"`
}

func CreateBuildDir(sols []Solution) (build string, err error) {
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

	var testprogs []string
	for _, sol := range sols {
		dst, err := os.Create(path.Join(build, sol.Name) + ".cpp")
		if err != nil {
			return build, err
		}

		src := bytes.NewBuffer(sol.Content)

		numgo++
		go func() {
			io.Copy(dst, src)
			dst.Close()
			godone <- true
		}()

		for _, ts := range sol.Suites {
			dst, err := os.Create(path.Join(build, ts.Name) + ".cpp")
			if err != nil {
				return build, err
			}

			src := bytes.NewBuffer(ts.Content)

			numgo++
			go func() {
				io.Copy(dst, src)
				dst.Close()
				godone <- true
			}()

			fmt.Fprintf(mk, "%s: %s.cpp %s.o cppunit_main.o\n", ts.Name, ts.Name, sol.Name)
			fmt.Fprintf(mk, "\t$(CXX) $(CXXFLAGS) $(LDFLAGS) -o %s %s.cpp %s.o cppunit_main.o\n\n", ts.Name, ts.Name, sol.Name)

			testprogs = append(testprogs, ts.Name)
		}
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

	fmt.Fprintf(mk, "all:")
	for _, t := range testprogs {
		fmt.Fprintf(mk, " %s", t)
	}
	fmt.Fprintf(mk, "\n")
	mk.Close()

	for ; numgo > 0; <-godone {
		numgo--
	}
	return build, nil
}
