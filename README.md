bor - Automated testing of problem-assignments
==============================================

Installation
============

```shell
$ # Get and build the software
$ export GOPATH="$HOME/go"
$ go get github.com/Merovius/bor
$ # Copy the relevant files to their default locations
$ cd "$GOPATH/src/github.com/Merovius/bor/
$ cp bor.conf /etc
$ mkdir -p /usr/share/bor
$ cp share/* /usr/share/bor
```

Updating
========

```shell
$ go get -u github.com/Merovius/bor
```

Configuration
=============

The default config-path is `/etc/bor.conf` but you can give alternative config
via the `-config` option.

There is a [default-config](bor.conf) with comments to document the various
options.

Using bor/cppunit
=================

You just start the `bor`-binary that was built when you got bor:

```shell
$ bor [-config /path/to/alternate/config]
```

`bor` will listen on the configured interface/port for incoming connections. On
every connection it expects a JSON-dictionary with keys "suites" und "files".

The latter should contain a dictionary of files, with the filename as the key
and the gzipped, base64-encoded content of the file.

The former should contain an array of dictionaries, each one describing one
testsuite, having a "name" key and a "link" key, where the latter contains a
list of files (without the .cpp-extension) to link together.

See [examples/small](examples/small) for an example of how to write
solutions/testsuites and
[examples/small/example.json](examples/small/example.json) for the
JSON-representation.

`bor` will then write all given files to a temporary build-dir and build all
testsuites and link them with [share/TAPListener.cpp](share/TAPListener.cpp) to
create a testsuite-executable.
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
      "system_time": 164000000,
      "user_time": 1612000000
    }
  },
  {
    "name": "solution2_tests",
    "suite": {
      "ok": false,
      "tests": [
        {
          "description": "Exercise2Test::FibPos",
          "diagnostic": "equality assertion failed\nExpected: 2584\nActual  : 4181",
          "ok": false
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
    "name": "solution1_tests",
    "suite": {
      "ok": false,
      "tests": null
    },
    "stats": {
      "system_time": 0,
      "user_time": 0
    },
    "error": "Timeout"
  }
]
```
Note the failure in `Exercise2Test::FibPos`: The person writing the test
obviously expected `fib(0) == 0 && fib(1) == 1`, while the person writing the
solution started with `fib(0) == 1 && fib(1) == 1`, thus producing an
off-by-one error.

Also notice the failure of `solution1_tests`: The tests-property is null and
instead there is an `error`-property (plus an output-property, if there is any
`output`). This kind of in-band-signalling is used for notification of unusual
failures during execution, for example a timeout or getting killed because of
an invalid systemcall or a segmentation fault.

The `stats`-property of a suite gives usage-statistics (currently the system-
and usertime in nanoseconds).
