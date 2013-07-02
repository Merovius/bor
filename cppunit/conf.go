package main

import (
	goconf "code.google.com/p/goconf/conf"
	"flag"
	"fmt"
)

type Conf struct {
	TmpDir           string
	TmpPrefix        string
	MakefileTemplate string
	CppunitMain      string
	SandboxDriver    string
	TCPListen        string
	NumConns         int
}

var (
	conf = Conf{
		"",
		"bor-",
		"/usr/share/bor/Makefile.tpl",
		"/usr/share/bor/cppunit_main.cpp",
		"plain",
		"localhost:7066",
		10,
	}
	confpath = flag.String("config", "/etc/bor.conf", "Config path")
)

func ReadConfig() error {
	cfg, err := goconf.ReadConfigFile(*confpath)
	if err != nil {
		return err
	}

	if str, err := cfg.GetString("cppunit", "TmpDir"); err == nil {
		conf.TmpDir = str
	}
	if str, err := cfg.GetString("cppunit", "TmpPrefix"); err == nil {
		conf.TmpPrefix = str
	}
	if str, err := cfg.GetString("cppunit", "MakefileTemplate"); err == nil {
		conf.MakefileTemplate = str
	}
	if str, err := cfg.GetString("cppunit", "CppunitMain"); err == nil {
		conf.CppunitMain = str
	}
	if str, err := cfg.GetString("default", "SandboxDriver"); err == nil {
		conf.SandboxDriver = str
	} else {
		return fmt.Errorf("You need to specify SandboxDriver")
	}
	if str, err := cfg.GetString("cppunit", "TCPListen"); err == nil {
		conf.TCPListen = str
	} else {
		return fmt.Errorf("You need to specify TCPListen")
	}
	if num, err := cfg.GetInt("cppunit", "NumConns"); err == nil {
		conf.NumConns = num
	}

	return nil
}
