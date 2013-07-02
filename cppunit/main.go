package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"path"

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

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	var sols []Solution
	dec := json.NewDecoder(conn)
	err := dec.Decode(&sols)
	if err != nil {
		elog.Println("Could not parse JSON:", err)
		return
	}

	builddir, err := CreateBuildDir(sols)
	if err != nil {
		elog.Println("Could not create buildpath")
		return
	}
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

	enc := json.NewEncoder(conn)
	err = enc.Encode(suites)
	if err != nil {
		elog.Println("Could not encode: ", err)
		return
	}
}

func main() {
	var err error
	flag.Parse()

	if err = ReadConfig(); err != nil {
		elog.Fatal(err)
	}

	l, err := net.Listen("tcp", conf.TCPListen)
	if err != nil {
		elog.Fatal(err)
	}
	defer l.Close()
	log.Println("Listening on", conf.TCPListen)

	for {
		conn, err := l.Accept()
		if err != nil {
			elog.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
