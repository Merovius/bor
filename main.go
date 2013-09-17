package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/Merovius/bor/sandbox"
	_ "github.com/Merovius/bor/sandbox/easysandbox"
	_ "github.com/Merovius/bor/sandbox/plain"
	"github.com/Merovius/go-tap"
)

// cmdResult is used to pass some data about the execution of a command through
// a channel
type cmdResult struct {
	n      int
	output []byte
	stats  stats
	err    error
	suite  *Testsuite
}

// Using elog instead of just log makes it easy to employ syslog-capabilities
// etc later, by just replacing this
var (
	elog *log.Logger
)

// Per default we just log to stderr
func init() {
	elog = log.New(os.Stderr, "", log.LstdFlags)
}

// HandleConnection reads a request from a connection, builds the testsuites
// and executes them, aggregating the results and passing them back to the
// connection
func HandleConnection(conn *net.TCPConn) {
	// Usually the operating system handles connection-lingering, but there was
	// a problem once, we attributed to lack of this, so we put it in and saw
	// no need to take it out again
	conn.SetLinger(conf.Linger)

	// Close the connection once we're finished
	defer conn.Close()

	// Read a request from the connection
	var msg Message
	dec := json.NewDecoder(conn)
	err := dec.Decode(&msg)
	if err != nil {
		elog.Println("Could not parse JSON:", err)
		return
	}

	// Create the build-dir and write everything to it
	builddir, err := CreateBuildDir(msg)
	if err != nil {
		elog.Println("Could not create buildpath:", err)
		return
	}
	defer os.RemoveAll(builddir)

	// The buildsuite is always there and tells, wether the build succeded or not
	buildsuite := suiteWrap{Name: "Building", Suite: Testsuite{Ok: false, Tests: make([]*tap.Testline, 1)}}
	test := &tap.Testline{Num: 1, Description: "Building"}
	buildsuite.Suite.Tests[0] = test

	// We collect all suites and send them off after returning, no matter when
	// we return. This makes it easier to handle error cases
	var suites []suiteWrap

	defer func() {
		enc := json.NewEncoder(conn)
		if err = enc.Encode(suites); err != nil {
			elog.Println("Could not encode: ", err)
		}
	}()

	// Run make in the make sandbox. Use -j to parallelize the build
	cmd := sandbox.Command(conf.MakeSandbox, "make", "-j", fmt.Sprintf("%d", runtime.NumCPU()), "all")
	cmd.SetDir(builddir)
	out, err := sandbox.TimeoutCombinedOutput(cmd, 5*time.Second)
	buildsuite.Stats.SystemTime = cmd.ProcessState().SystemTime()
	buildsuite.Stats.UserTime = cmd.ProcessState().UserTime()

	if !cmd.ProcessState().Success() {
		// Build did not succed, give some context and return, writing the
		// build-suite to the connection
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

	// We know what the Testsuites are simply by listing all executable files
	// in the builddir
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

	// The numbers of started goroutines
	numgo := 0

	// The index of the testsuites, because we will not get the results in the
	// right order, we have to keep track in each goroutine, what testsuite was
	// executed by it
	n := len(suites)

	for _, fi := range fi {
		mode := fi.Mode()
		// Skip non-regular and non-executable files
		if mode&os.ModeType != 0 || mode&1 != 1 {
			continue
		}

		// Create a basic suite, already add it to the list of run buildsuites,
		// to preserve ordering
		wrap := suiteWrap{Name: fi.Name()}
		suites = append(suites, wrap)

		// Run the testsuite in the background. We have to pass name and the
		// index as parameters, to prevent races with fi
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

			// Parse the TAP
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

	// Collect the results
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
	// Get the -config flag, if existent
	flag.Parse()

	// Try to read the configfile
	if err = ReadConfig(); err != nil {
		elog.Fatal(err)
	}

	// Listen on the specified interface/port
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

	// Handle connections in the background
	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			elog.Println(err)
			continue
		}
		go HandleConnection(conn)
	}
}
