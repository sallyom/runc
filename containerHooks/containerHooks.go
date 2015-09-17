// +build linux

package main

import (
	"io/ioutil"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path"
)

const (
	hookDirPath = "/usr/libexec/docker/hooks.d"
)

func getHooks() ([]os.FileInfo, error) {
	// find any hooks executables
	return ioutil.ReadDir(hookDirPath)
}

func prestart(stdinbytes []byte) error {
	hooks, err := getHooks()
	if err != nil {
		return err
	}
	for _, item := range hooks {
		if item.Mode().IsRegular() {
			if err = run(path.Join(hookDirPath, item.Name()), stdinbytes); err != nil {
				return err
			}
		}
	}
	return nil
}

func poststop(stdinbytes []byte) error {
	hooks, err := getHooks()
	if err != nil {
		return err
	}
	for i := len(hooks) - 1; i >= 0; i-- {
		fn := hooks[i].Name()
		for _, item := range hooks {
			if item.Mode().IsRegular() && fn == item.Name() {
				if err := run(path.Join(hookDirPath, item.Name()), stdinbytes); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func run(hookfile string, incoming []byte) error {
	log.Print("FILE RUN: ", hookfile)
	cmd := exec.Command(hookfile)
	cmd.Args = os.Args
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	stdinPipe.Write(incoming)
	return nil
}

func check(err error) {
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func main() {
	logwriter, e := syslog.New(syslog.LOG_NOTICE, "docker-hooks")
	if e == nil {
		log.SetOutput(logwriter)
	}
	log.Print("HOOKS: ", os.Args[0])

	incoming, err := ioutil.ReadAll(os.Stdin)
	check(err)

	if os.Args[0] == "prestart" {
		err := prestart(incoming)
		check(err)
	}

	if os.Args[0] == "poststop" {
		err := poststop(incoming)
		check(err)
	}
}
