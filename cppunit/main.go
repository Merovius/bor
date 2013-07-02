package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/Merovius/bor/sandbox"
	_ "github.com/Merovius/bor/sandbox/plain"
	"github.com/Merovius/go-tap"
)

type suiteWrap struct {
	Name  string    `json:"name"`
	Suite Testsuite `json:"suite"`
}

var (
	elog *log.Logger
)

func init() {
	elog = log.New(os.Stderr, "", log.LstdFlags)
}

func CreateBuildDir(solsrc, tstsrc string) (build string) {
	numgo := 0
	godone := make(chan bool, 100)

	build, err := ioutil.TempDir(conf.TmpDir, conf.TmpPrefix)
	if err != nil {
		elog.Fatal("Could not create tempdir: ", err)
	}

	mk, err := os.Create(path.Join(build, "Makefile"))
	defer mk.Close()

	sols, err := filepath.Glob(solsrc + "/*.cpp")
	if err != nil {
		elog.Fatal("Could not get list of sources: ", err)
	}

	mktpl, err := os.Open(conf.MakefileTemplate)
	if err != nil {
		elog.Fatal("Could not open Makefile template: ", err)
	}
	defer mktpl.Close()

	_, err = io.Copy(mk, mktpl)
	if err != nil {
		elog.Fatal("Could not copy Makefile template: ", err)
	}

	// Take basenames, copy solutions to builddir
	for i, f := range sols {
		sols[i] = path.Base(f)
		dst, err := os.Create(path.Join(build, sols[i]))
		if err != nil {
			elog.Fatal("Could not open file for writing: ", err)
		}
		src, err := os.Open(f)
		if err != nil {
			elog.Fatal("Could not open file for reading: ", err)
		}

		sols[i] = sols[i][:len(sols[i])-4]
		numgo++
		go func() {
			io.Copy(dst, src)
			dst.Close()
			src.Close()
			godone <- true
		}()

	}

	// For every solution, copy all relevant testsuites and add them to the
	// Makefile
	var testprogs []string
	for _, f := range sols {
		tests, err := filepath.Glob(tstsrc + "/" + f + "_*.cppunit")
		if err != nil {
			elog.Fatal("Could not get list of tests: ", err)
		}

		for _, t := range tests {
			tbase := path.Base(t)
			tbase = tbase[:len(tbase)-8]

			src, err := os.Open(t)
			if err != nil {
				elog.Fatal("Could not open for reading: ", err)
			}

			dst, err := os.Create(path.Join(build, tbase+".cpp"))
			if err != nil {
				elog.Fatal("Could not open for writing: ", err)
			}

			numgo++
			go func() {
				io.Copy(dst, src)
				dst.Close()
				src.Close()
				godone <- true
			}()

			fmt.Fprintf(mk, "%s: %s.cpp %s.o cppunit_main.o\n", tbase, tbase, f)
			fmt.Fprintf(mk, "\t$(CXX) $(CXXFLAGS) $(LDFLAGS) -o %s %s.cpp %s.o cppunit_main.o\n\n", tbase, tbase, f)
			testprogs = append(testprogs, tbase)
		}
	}

	// Copy cppunit_main
	src, err := os.Open(conf.CppunitMain)
	if err != nil {
		elog.Fatal("Could not open for reading: ", err)
	}
	dst, err := os.Create(path.Join(build, "cppunit_main.cpp"))
	if err != nil {
		elog.Fatal("Could not open for writing: ", err)
	}
	numgo++
	go func() {
		io.Copy(dst, src)
		dst.Close()
		src.Close()
		godone <- true
	}()

	// Add all testsuites to the Makefile
	fmt.Fprintf(mk, "all:")
	for _, t := range testprogs {
		fmt.Fprintf(mk, " %s", t)
	}
	fmt.Fprintf(mk, "\n")
	mk.Close()

	for ; numgo > 0; <-godone {
		numgo--
	}

	return
}

func main() {
	var err error
	flag.Parse()

	if flag.NArg() < 2 {
		elog.Fatal("Not enough arguments")
	}

	if err = ReadConfig(); err != nil {
		elog.Fatal(err)
	}

	solsrc := path.Clean(flag.Arg(0))
	tstsrc := path.Clean(flag.Arg(1))

	builddir := CreateBuildDir(solsrc, tstsrc)
	defer os.RemoveAll(builddir)

	buildsuite := Testsuite{Ok: false, Tests: make([]*tap.Testline, 1)}
	test := &tap.Testline{Num: 1, Description: "Building"}
	buildsuite.Tests[0] = test

	var suites []suiteWrap

	defer func() {
		enc := json.NewEncoder(os.Stdout)
		_ = enc.Encode(suites)
	}()

	cmd := sandbox.Command(conf.SandboxDriver, "make", "all")
	cmd.SetDir(builddir)
	out, err := cmd.CombinedOutput()

	if !cmd.ProcessState().Success() {
		test.Ok = false
		test.Diagnostic += string(out)
		suites = append(suites, suiteWrap{"Building", buildsuite})
		return
	}
	test.Ok = true
	buildsuite.Ok = true

	suites = append(suites, suiteWrap{"Building", buildsuite})

	d, err := os.Open(builddir)
	if err != nil {
		elog.Println("Could not open build-dir: ", err)
		return
	}

	fi, err := d.Readdir(-1)
	if err != nil {
		elog.Println("Could not read build-dir: ", err)
		return
	}

	for _, fi := range fi {
		mode := fi.Mode()
		// Skip non-regular and non-executable files
		if mode&os.ModeType != 0 || mode&1 != 1 {
			continue
		}

		cmd := sandbox.Command(conf.SandboxDriver, path.Join(builddir, fi.Name()))
		cmd.SetDir(builddir)
		out, err := cmd.CombinedOutput()
		if err != nil {
			elog.Println("Could not run testsuite: ", err)
			continue
		}

		r := bytes.NewReader(out)
		parser, err := tap.NewParser(r)
		if err != nil {
			elog.Println("Could not parse: ", err)
			continue
		}

		suite, err := parser.Suite()
		if err != nil {
			elog.Println("Could not parse: ", err)
			continue
		}
		suites = append(suites, suiteWrap{fi.Name(), Testsuite(*suite)})
	}
}
