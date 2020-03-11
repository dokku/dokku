## OLD README
First give you a full example, I will explain every command below.

	session := sh.NewSession()
	session.Env["PATH"] = "/usr/bin:/bin"
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Alias("ll", "ls", "-l")
	session.ShowCMD = true // enable for debug
	var err error
	err = session.Call("ll", "/")
	if err != nil {
		log.Fatal(err)
	}
	ret, err := session.Capture("pwd", sh.Dir("/home")) # wraper of session.Call
	if err != nil {
		log.Fatal(err)
	}
	# ret is "/home\n"
	fmt.Println(ret)

create a new Session

	session := sh.NewSession()

use alias like this

	session.Alias("ll", "ls", "-l") # like alias ll='ls -l'

set current env like this

	session.Env["BUILD_ID"] = "123" # like export BUILD_ID=123

set current directory

	session.Set(sh.Dir("/")) # like cd /

pipe is also supported

	session.Command("echo", "hello\tworld").Command("cut", "-f2")
	// output should be "world"
	session.Run()

test, the build in command support

	session.Test("d", "dir") // test dir
	session.Test("f", "file) // test regular file

with `Alias Env Set Call Capture Command` a shell scripts can be easily converted into golang program. below is a shell script.

	#!/bin/bash -
	#
	export PATH=/usr/bin:/bin
	alias ll='ls -l'
	cd /usr
	if test -d "local"
	then
		ll local | awk '{print $1, $NF}'
	fi

convert to golang, will be

	s := sh.NewSession()
	s.Env["PATH"] = "/usr/bin:/bin"
	s.Set(sh.Dir("/usr"))
	s.Alias("ll", "ls", "-l")
	if s.Test("d", "local") {
		s.Command("ll", "local").Command("awk", "{print $1, $NF}").Run()
	}
