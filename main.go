package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"time"

	"github.com/Merovius/bor/sandbox"
	_ "github.com/Merovius/bor/sandbox/plain"
	"github.com/Merovius/go-tap"
)

type stats struct {
	SystemTime time.Duration `json:"system_time"`
	UserTime   time.Duration `json:"user_time"`
}

type suiteWrap struct {
	Name  string    `json:"name"`
	Suite Testsuite `json:"suite"`
	Stats stats     `json:"stats"`
}

var (
	elog *log.Logger
)

func init() {
	elog = log.New(os.Stderr, "", log.LstdFlags)
}

func HandleConnection(conn *net.TCPConn) {
	conn.SetLinger(conf.Linger)
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
		elog.Println("Could not create buildpath:", err)
		return
	}
	defer os.RemoveAll(builddir)

	buildsuite := suiteWrap{Name: "Building", Suite: Testsuite{Ok: false, Tests: make([]*tap.Testline, 1)}}
	test := &tap.Testline{Num: 1, Description: "Building"}
	buildsuite.Suite.Tests[0] = test

	var suites []suiteWrap

	defer func() {
		enc := json.NewEncoder(conn)
		if err = enc.Encode(suites); err != nil {
			elog.Println("Could not encode: ", err)
		}
	}()

	cmd := sandbox.Command(conf.SandboxDriver, "make", "all")
	cmd.SetDir(builddir)
	out, err := cmd.CombinedOutput()
	buildsuite.Stats.SystemTime = cmd.ProcessState().SystemTime()
	buildsuite.Stats.UserTime = cmd.ProcessState().UserTime()

	if !cmd.ProcessState().Success() {
		test.Ok = false
		test.Diagnostic += string(out)
		if err != nil {
			if test.Diagnostic[len(test.Diagnostic)-1] != '\n' {
				test.Diagnostic += "\n"
			}
			test.Diagnostic += err.Error()
		}
		suites = append(suites, buildsuite)
		return
	}
	test.Ok = true
	buildsuite.Suite.Ok = true

	suites = append(suites, buildsuite)

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
		utime := cmd.ProcessState().UserTime()
		stime := cmd.ProcessState().SystemTime()
		log.Printf("%d %d\n", utime, stime)

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
		wrap := suiteWrap{Name: fi.Name(), Suite: Testsuite(*suite)}
		wrap.Stats.SystemTime = stime
		wrap.Stats.UserTime = utime

		suites = append(suites, wrap)
	}
}

func main() {
	var err error
	flag.Parse()

	if err = ReadConfig(); err != nil {
		elog.Fatal(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", conf.TCPListen)
	if err != nil {
		elog.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		elog.Fatal(err)
	}
	defer l.Close()
	log.Println("Listening on", conf.TCPListen)

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			elog.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
