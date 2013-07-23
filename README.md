bor - Automated testing of problem-assignments
==============================================

Installation
============

```shell
$ # Get and build the software
$ export GOPATH="$HOME/go"
$ go get github.com/Merovius/bor/cppunit
$ # Copy the relevant files to their default locations
$ cd "$GOPATH/src/github.com/Merovius/bor/
$ cp bor.conf /etc
$ mkdir -p /usr/share/bor
$ cp cppunit/Makefile.tpl /usr/share/bor
$ cp cppunit/cppunit_main.cpp /usr/share/bor
```

Updating
========

```shell
$ go get -u github.com/Merovius/bor/cppunit
```

Configuration
=============

The default config-path is `/etc/bor.conf` but you can give alternative config
via the `-config` option.

There is a [default-config](bor.conf) with comments to document the various
options.

Using bor/cppunit
=================

You just start the `cppunit`-binary that was built when you got bor:

```shell
$ cppunit [-config /path/to/alternate/config]
```

`cppunit` will listen on the configured interface/port for incoming
connections. On every connection it expects a JSON-list of maps, each
describing one solution with corresponding testsuites. Every solution has the
properties `name` (an arbitrary (but unique) name slashes), `content` (the
solution .cpp-file base64-encoded) and `suites` (a list of maps, each
consisting of a unique `name` and a base64 encoded `content`). See
[cppunit/example](cppunit/example) for an example of how to write
solutions/testsuites and [cppunit/example.json](cppunit/example.json) for the
JSON-representation.

`cppunit` will then write every solution to a temporary build-dir. It will then
build all solutions and link them with
[cppunit_main.cpp](cppunit/cppunit_main.cpp) to create a testsuite-executable.
This will then be run and the testresults will be collected and send back in
JSON-form. After this it will close the connection, so you have to make a new
one to start another testrun.

Example output:
```JSON
[
  {
    "name": "Building",
    "suite": {
      "ok": true,
      "tests": [
        {
          "description": "Building",
          "diagnostic": "",
          "ok": true
        }
      ]
    },
    "stats": {
        "system_time": 184000000,
        "user_time": 1820000000
    }
  },
  {
    "name": "exercise2_tests",
    "suite": {
      "ok": false,
      "tests": [
        {
          "description": "Exercise2Test::FibPos",
          "diagnostic": "equality assertion failed\nExpected: 2584\nActual  : 4181"
          ,"ok": false
        },
        {
          "description": "Exercise2Test::Fib1",
          "diagnostic": "",
          "ok": true
        }
      ]
    },
    "stats": {
        "system_time": 0,
        "user_time": 0
    }
  },
  {
    "name": "exercise1_tests",
    "suite": {
      "ok": true,
      "tests": [
        {
          "description": "Exercise1Test::PosChoosePos",
          "diagnostic": "",
          "ok": true
        },
        {
          "description": "Exercise1Test::PosChooseO",
          "diagnostic": "",
          "ok": true
        },
        {
          "description": "Exercise1Test::PosChooseSame",
          "diagnostic": "",
          "ok": true
        },
        {
          "description": "Exercise1Test::PosChooseGreater",
          "diagnostic": "",
          "ok": true
        }
      ]
    },
    "stats": {
        "system_time": 0,
        "user_time": 0
    }
  }
]
```
Note the failure in `Exercise2Test::FibPos`: The person writing the test
obviously expected `fib(0) == 0 && fib(1) == 1`, while the person writing the
solution started with `fib(0) == 1 && fib(1) == 1`, thus producing an
off-by-one error.

The `stats`-property of a suite gives usage-statistics (currently the system-
and usertime in nanoseconds).
