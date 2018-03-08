package sh

import (
	"encoding/xml"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestUnmarshalJSON(t *testing.T) {
	var a int
	s := NewSession()
	s.ShowCMD = true
	err := s.Command("echo", []string{"1"}).UnmarshalJSON(&a)
	if err != nil {
		t.Error(err)
	}
	if a != 1 {
		t.Errorf("expect a tobe 1, but got %d", a)
	}
}

func TestUnmarshalXML(t *testing.T) {
	s := NewSession()
	xmlSample := `<?xml version="1.0" encoding="utf-8"?>
<server version="1" />`
	type server struct {
		XMLName xml.Name `xml:"server"`
		Version string   `xml:"version,attr"`
	}
	data := &server{}
	s.Command("echo", xmlSample).UnmarshalXML(data)
	if data.Version != "1" {
		t.Error(data)
	}
}

func TestPipe(t *testing.T) {
	s := NewSession()
	s.ShowCMD = true
	s.Call("echo", "hello")
	err := s.Command("echo", "hi").Command("cat", "-n").Start()
	if err != nil {
		t.Error(err)
	}
	err = s.Wait()
	if err != nil {
		t.Error(err)
	}
	out, err := s.Command("echo", []string{"hello"}).Output()
	if err != nil {
		t.Error(err)
	}
	if string(out) != "hello\n" {
		t.Error("capture wrong output:", out)
	}
	s.Command("echo", []string{"hello\tworld"}).Command("cut", []string{"-f2"}).Run()
}

func TestPipeCommand(t *testing.T) {
	c1 := exec.Command("echo", "good")
	rd, wr := io.Pipe()
	c1.Stdout = wr
	c2 := exec.Command("cat", "-n")
	c2.Stdout = os.Stdout
	c2.Stdin = rd
	c1.Start()
	c2.Start()

	c1.Wait()
	wc, ok := c1.Stdout.(io.WriteCloser)
	if ok {
		wc.Close()
	}
	c2.Wait()
}

func TestPipeInput(t *testing.T) {
	s := NewSession()
	s.ShowCMD = true
	s.SetInput("first line\nsecond line\n")
	out, err := s.Command("grep", "second").Output()
	if err != nil {
		t.Error(err)
	}
	if string(out) != "second line\n" {
		t.Error("capture wrong output:", out)
	}
}

func TestTimeout(t *testing.T) {
	s := NewSession()
	err := s.Command("sleep", "2").Start()
	if err != nil {
		t.Fatal(err)
	}
	err = s.WaitTimeout(time.Second)
	if err != ErrExecTimeout {
		t.Fatal(err)
	}
}

func TestSetTimeout(t *testing.T) {
	s := NewSession()
	s.SetTimeout(time.Second)
	defer s.SetTimeout(0)
	err := s.Command("sleep", "2").Run()
	if err != ErrExecTimeout {
		t.Fatal(err)
	}
}

func TestCombinedOutput(t *testing.T) {
	s := NewSession()
	bytes, err := s.Command("sh", "-c", "echo stderr >&2 ; echo stdout").CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	stringOutput := string(bytes)
	if !(strings.Contains(stringOutput, "stdout") && strings.Contains(stringOutput, "stderr")) {
		t.Errorf("expect output from both output streams, got '%s'", strings.TrimSpace(stringOutput))
	}
}

func TestPipeFail(t *testing.T) {
	sh := NewSession()
	sh.PipeFail = true
	sh.PipeStdErrors = true
	sh.Command("cat", "unknown-file")
	sh.Command("echo")
	if _, err := sh.Output(); err == nil {
		t.Error("expected error")
	}
}
