## go-sh
[![wercker status](https://app.wercker.com/status/009acbd4f00ccc6de7e2554e12a50d84/s "wercker status")](https://app.wercker.com/project/bykey/009acbd4f00ccc6de7e2554e12a50d84)
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/codeskyblue/go-sh)

*If you depend on the old api, see tag: v.0.1*

install: `go get github.com/codeskyblue/go-sh`

Pipe Example:

	package main

	import "github.com/codeskyblue/go-sh"

	func main() {
		sh.Command("echo", "hello\tworld").Command("cut", "-f2").Run()
	}

Because I like os/exec, `go-sh` is very much modelled after it. However, `go-sh` provides a better experience.

These are some of its features:

* keep the variable environment (e.g. export)
* alias support (e.g. alias in shell)
* remember current dir
* pipe command
* shell build-in commands echo & test
* timeout support

Examples are important:

	sh: echo hello
	go: sh.Command("echo", "hello").Run()

	sh: export BUILD_ID=123
	go: s = sh.NewSession().SetEnv("BUILD_ID", "123")

	sh: alias ll='ls -l'
	go: s = sh.NewSession().Alias('ll', 'ls', '-l')

	sh: (cd /; pwd)
	go: sh.Command("pwd", sh.Dir("/")).Run()

	sh: test -d data || mkdir data
	go: if ! sh.Test("dir", "data") { sh.Command("mkdir", "data").Run() }

	sh: cat first second | awk '{print $1}'
	go: sh.Command("cat", "first", "second").Command("awk", "{print $1}").Run()

	sh: count=$(echo "one two three" | wc -w)
	go: count, err := sh.Echo("one two three").Command("wc", "-w").Output()

	sh(in ubuntu): timeout 1s sleep 3
	go: c := sh.Command("sleep", "3"); c.Start(); c.WaitTimeout(time.Second) # default SIGKILL
	go: out, err := sh.Command("sleep", "3").SetTimeout(time.Second).Output() # set session timeout and get output)

	sh: echo hello | cat
	go: out, err := sh.Command("cat").SetInput("hello").Output()

	sh: cat # read from stdin
	go: out, err := sh.Command("cat").SetStdin(os.Stdin).Output()

	sh: ls -l > /tmp/listing.txt # write stdout to file
	go: err := sh.Command("ls", "-l").WriteStdout("/tmp/listing.txt")

If you need to keep env and dir, it is better to create a session

	session := sh.NewSession()
	session.SetEnv("BUILD_ID", "123")
	session.SetDir("/")
	# then call cmd
	session.Command("echo", "hello").Run()
	# set ShowCMD to true for easily debug
	session.ShowCMD = true

By default, pipeline returns error only if the last command exit with a non-zero status. However, you can also enable `pipefail` option like `bash`. In that case, pipeline returns error if any of the commands fail and for multiple failed commands, it returns the error of rightmost failed command.

	session := sh.NewSession()
	session.PipeFail = true
	session.Command("cat", "unknown-file").Command("echo").Run()

By default, pipelines's std-error is set to last command's std-error. However, you can also combine std-errors of all commands into pipeline's std-error using `session.PipeStdErrors = true`.

for more information, it better to see docs.
[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/codeskyblue/go-sh)

### contribute
If you love this project, starring it will encourage the coder. Pull requests are welcome.

support the author: [alipay](https://me.alipay.com/goskyblue)

### thanks
this project is based on <http://github.com/codegangsta/inject>. thanks for the author.

# the reason to use Go shell
Sometimes we need to write shell scripts, but shell scripts are not good at working cross platform,  Go, on the other hand, is good at that. Is there a good way to use Go to write shell like scripts? Using go-sh we can do this now.
