package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Merovius/bor/sandbox"
	_ "github.com/Merovius/bor/sandbox/easysandbox"
	_ "github.com/Merovius/bor/sandbox/plain"
	"github.com/Merovius/go-tap"
)

type stats struct {
	SystemTime time.Duration `json:"system_time"`
	UserTime   time.Duration `json:"user_time"`
}

type suiteWrap struct {
	Name   string    `json:"name"`
	Suite  Testsuite `json:"suite"`
	Stats  stats     `json:"stats"`
	Error  string    `json:"error,omitempty"`
	Output string    `json:"output,omitempty"`
}

type cmdResult struct {
	n      int
	output []byte
	stats  stats
	err    error
	suite  *Testsuite
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

	var msg Message
	dec := json.NewDecoder(conn)
	err := dec.Decode(&msg)
	if err != nil {
		elog.Println("Could not parse JSON:", err)
		return
	}

	builddir, err := CreateBuildDir(msg)
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

	cmd := sandbox.Command(conf.MakeSandbox, "make", "all")
	cmd.SetDir(builddir)
	out, err := sandbox.TimeoutCombinedOutput(cmd, 5*time.Second)
	buildsuite.Stats.SystemTime = cmd.ProcessState().SystemTime()
	buildsuite.Stats.UserTime = cmd.ProcessState().UserTime()

	if !cmd.ProcessState().Success() {
		test.Ok = false
		test.Diagnostic += string(out)
		if err != nil {
			if len(test.Diagnostic) > 0 && !strings.HasSuffix(test.Diagnostic, "\n") {
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

	ch := make(chan cmdResult)

	numgo := 0
	n := len(suites)
	for _, fi := range fi {
		mode := fi.Mode()
		// Skip non-regular and non-executable files
		if mode&os.ModeType != 0 || mode&1 != 1 {
			continue
		}

		wrap := suiteWrap{Name: fi.Name()}
		suites = append(suites, wrap)

		go func(name string, i int) {
			res := cmdResult{n: i}
			cmd := sandbox.Command(conf.TestSandbox, path.Join(builddir, name))
			cmd.SetDir(builddir)
			out, err := sandbox.TimeoutCombinedOutput(cmd, time.Second)
			if err != nil {
				elog.Println("Could not run testsuite: ", err)
				res.err = err
				ch <- res
				return
			}
			res.stats.UserTime = cmd.ProcessState().UserTime()
			res.stats.SystemTime = cmd.ProcessState().SystemTime()

			r := bytes.NewReader(out)
			parser, err := tap.NewParser(r)
			if err != nil {
				res.err = err
				ch <- res
				return
			}

			suite, err := parser.Suite()
			if err != nil {
				res.err = err
				ch <- res
				return
			}

			res.suite = (*Testsuite)(suite)

			ch <- res
			return
		}(fi.Name(), n)
		n++
		numgo++
	}

	for ; numgo > 0; numgo-- {
		res := <-ch
		suite := &suites[res.n]

		if res.err != nil {
			suite.Error = res.err.Error()
			suite.Stats = res.stats
			suite.Output = string(res.output)
			continue
		}

		suite.Stats = res.stats
		suite.Suite = *res.suite
	}
	return
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
