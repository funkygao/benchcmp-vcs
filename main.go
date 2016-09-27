package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

const (
	GO = "go"
)

var (
	pkg = ""
)

func main() {
	if len(os.Args) > 1 {
		pkg = os.Args[1]
	}

	before, beforeClose, err := Tempfile()
	if err != nil {
		panic(err)
	}
	defer beforeClose()

	after, afterClose, err := Tempfile()
	if err != nil {
		panic(err)
	}
	defer afterClose()

	git := NewGit()

	err = ExecBench(after)
	if err != nil {
		panic(err)
	}

	err = git.BackToThePast()
	if err != nil {
		panic(err)
	}
	defer git.BackToTheFuture()

	err = ExecBench(before)
	if err != nil {
		panic(err)
	}

	fmt.Printf("old revision: %s\n", git.OldRevision())
	if rev := git.NewRevision(); rev != "" {
		fmt.Printf("new revision: %s\n", rev)
	}
	fmt.Println()

	err = ExecCmp(os.Stdout, before.Name(), after.Name())
	if err != nil {
		panic(err)
	}
}

func Tempfile() (*os.File, func(), error) {
	file, err := ioutil.TempFile("", "benchcmp-vcs")
	if err != nil {
		return nil, nil, err
	}
	fn := func() {
		file.Close()
		os.Remove(file.Name())
	}

	return file, fn, nil
}

func ExecBench(w io.Writer, opts ...string) error {
	o := append([]string{"test", "-run=NONE", "-bench=.", "-cpu=1,4,8,16", "-benchmem"}, opts...)
	if pkg != "" {
		o = append(o, pkg)
	}
	out, err := exec.Command(GO, o...).Output()
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func ExecCmp(w io.Writer, before, after string, opts ...string) error {
	CMD := "benchcmp"
	o := append([]string{before, after}, opts...)
	out, err := exec.Command(CMD, o...).Output()
	if err != nil {
		return err
	}

	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}
