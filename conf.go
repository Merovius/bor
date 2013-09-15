package main

import (
	goconf "code.google.com/p/goconf/conf"
	"flag"
	"fmt"
	"github.com/Merovius/bor/sandbox"
	"time"
)

type Conf struct {
	TmpDir           string
	TmpPrefix        string
	MakefileTemplate string
	CppunitMain      string
	MakeSandbox      string
	TestSandbox      string
	TCPListen        string
	NumConns         int
	Linger           int
	MakeTimeout      time.Duration
	TestTimeout      time.Duration
}

var (
	conf = Conf{
		"",
		"bor-",
		"/usr/share/bor/Makefile.tpl",
		"/usr/share/bor/cppunit_main.cpp",
		"plain",
		"easysandbox",
		"localhost:7066",
		10,
		5,
		5 * time.Second,
		time.Second,
	}
	confpath = flag.String("config", "/etc/bor.conf", "Config path")
)

func ReadConfig() error {
	cfg, err := goconf.ReadConfigFile(*confpath)
	if err != nil {
		return err
	}

	if str, err := cfg.GetString("default", "TmpDir"); err == nil {
		conf.TmpDir = str
	}
	if str, err := cfg.GetString("default", "TmpPrefix"); err == nil {
		conf.TmpPrefix = str
	}
	if str, err := cfg.GetString("default", "MakefileTemplate"); err == nil {
		conf.MakefileTemplate = str
	}
	if str, err := cfg.GetString("default", "CppunitMain"); err == nil {
		conf.CppunitMain = str
	}
	if str, err := cfg.GetString("default", "MakeSandbox"); err == nil {
		conf.MakeSandbox = str
	}
	if str, err := cfg.GetString("default", "TestSandbox"); err == nil {
		conf.TestSandbox = str
	}
	if str, err := cfg.GetString("default", "TCPListen"); err == nil {
		conf.TCPListen = str
	} else {
		return fmt.Errorf("You need to specify TCPListen")
	}
	if num, err := cfg.GetInt("default", "NumConns"); err == nil {
		conf.NumConns = num
	}
	if num, err := cfg.GetInt("default", "Linger"); err == nil {
		conf.Linger = num
	}
	if str, err := cfg.GetString("default", "MakeTimeout"); err == nil {
		to, err := time.ParseDuration(str)
		if err != nil {
			return fmt.Errorf("Could not parse duration: %s", str)
		}
		conf.MakeTimeout = to
	}
	if str, err := cfg.GetString("default", "TestTimeout"); err == nil {
		to, err := time.ParseDuration(str)
		if err != nil {
			return fmt.Errorf("Could not parse duration: %s", str)
		}
		conf.TestTimeout = to
	}

	if err = sandbox.Config(cfg); err != nil {
		return err
	}

	return nil
}
