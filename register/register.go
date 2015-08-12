// +build linux

package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/godbus/dbus"
)

var conn *dbus.Conn

func main() {
	args := os.Args
	// get/decode info from stdin
	incoming, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("Read os.Stdin error:%v", err)
	}
	inbytes := bytes.NewBufferString(string(incoming))
	var jmap map[string]interface{}
	json.NewDecoder(inbytes).Decode(&jmap)
	pid := jmap["init_process_pid"].(float64)
	id := jmap["id"].(string)
	// name nested
	// rootfs nested
	config := jmap["config"].(map[string]interface{})
	root_directory := config["rootfs"].(string)
	cgroups := config["cgroups"].(map[string]interface{})
	name := cgroups["name"].(string)

	// ensure id is a hex string at least 32 chars
	passId, err := Validate(id, name, pid, root_directory)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if args[1] == "register" {
		if err = RegisterMachine(name, passId, int(pid), root_directory); err != nil {
			log.Fatalf("Register machine failed: %v", err)
		}
	}

	if args[1] == "terminate" {
		if err := TerminateMachine(name); err != nil {
			log.Fatalf("TerminateMachine failed: %v", err)
		}
	}
}

func Validate(id string, name string, pid float64, rootfs string) (string, error) {
	varStr := fmt.Sprintf("ID:%s, Name:%s, PID:%v, Root_directory:%s", id, name, pid, rootfs)
	if len(id) == 0 || len(name) == 0 || len(rootfs) == 0 {
		return "", fmt.Errorf("Error: Invalid  %s", varStr)
	}
	for len(id) < 32 {
		id += "0"
	}

	return hex.EncodeToString([]byte(id)), nil
}

// RegisterMachine with systemd on the host system
func RegisterMachine(name string, id string, pid int, root_directory string) error {
	var (
		av  []byte
		err error
	)
	if conn == nil {
		conn, err = dbus.SystemBus()
		if err != nil {
			return err
		}
	}

	av, err = hex.DecodeString(id[0:32])
	if err != nil {
		return err
	}
	obj := conn.Object("org.freedesktop.machine1", "/org/freedesktop/machine1")
	return obj.Call("org.freedesktop.machine1.Manager.RegisterMachine", 0, name, av, "runc.service", "container", uint32(pid), root_directory).Err
	return nil
}

// TerminateMachine registered with systemd on the host system
func TerminateMachine(name string) error {
	var err error
	if conn == nil {
		conn, err = dbus.SystemBus()
		if err != nil {
			return err
		}
	}
	obj := conn.Object("org.freedesktop.machine1", "/org/freedesktop/machine1")
	return obj.Call("org.freedesktop.machine1.Manager.TerminateMachine", 0, name).Err
	return nil
}
