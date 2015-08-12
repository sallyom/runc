## Systemd RegisterMachine

`RegisterMachine` implements a container manager that makes use of systemd-machined.  On this branch `RegisterMachine` and `TerminateMachine` are called using prestart and poststop hooks added to `runc`.  This doc assumes you have docker version 1.3 or later installed and your GOPATH set.  To use the systemd-machined methods do the following (run as root for Docker commands):

## 1. Clone this branch/build runC
build requires seccomp.h so
```
sudo yum install libseccomp-devel
```
```
cd <to your github.com directory>
mkdir opencontainers
cd opencontainers
git clone https://github.com/sallyom/runc.git
cd runc
git checkout -b registermachine origin/registermachine 
```
You have to build from the registermachine branch, not master, so now that you are in the registermachine branch:
```
make
sudo make install
```
## 2. Get up and running with a runC container:

To test using `centos` image follow these steps:
* Download the `centos` image: `docker pull centos`
* Create a container from that image and export its contents to a tar file:
```
docker export $(docker create centos) > centos.tar
```
* Untar the contents to create your filesystem directory.  Create a directory that will eventually contain your rootfs directory and a config.json file:
```
cd <back to GOPATH/src>
mkdir runcRun (or name it whatever you want)
cd runcRun
mkdir rootfs
tar -C rootfs -xf /ABS PATH TO /centos.tar
```

* **Create the config file**

   `ls runcRun` should list `rootfs`

   `ls rootfs` should show the centos filesystem

   * Create a file called `config.json`.  To do so run the following:
   ```
   /usr/local/bin/runc spec > config.json
   ```
   
   This will produce a config file similar to [OCF Container JSON Format](https://github.com/sallyom/runc/tree/registermachine#ocf-container-json-format)

   * Finally, to implement systemd-machined in your runC containers, you must give the path to the executable `regmach` to the prestart and poststop hooks in the config file as follows:
   ```
   cd runcRun
   vim config.json
   ```
   find the `hooks` section and make the following changes to config.json:
```
"hooks": {
    "prestart": [{
        "path": "ABS PATH TO /runc/regmach",
        "args": ["register"]
    }],
    "poststop": [{
        "path": "ABS PATH TO /runc/regmach",
        "args": ["terminate"]
    }]

},
```
## 3. Run a runC container and  register/terminate with machinectl

This example starts a runC container with the container ID of test (you can give it any unique ID)
```
cd runcRun
/usr/local/bin/runc --id test
```

At this point you will be placed in a shell.  Open a new terminal (while container is running) and view the machinectl output!

`machinectl`

You should see your container `test` listed.  
`machinectl --help` to see what else you can do with registered machines.

Now, exit the container by typing `exit` at the shell prompt.
Go back to the other terminal and see that machine was terminated upon container exit.
`machinectl` should no longer list `test`
