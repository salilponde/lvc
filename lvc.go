package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func Branch() {
	println("branch")
}

func Commit() {
	println("commit")
}

func Config() {
	println("config")
}

func History() {
	println("history")
}

func Init() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(dir)
	if len(files) > 0 {
		println("Directory not empty! lvc repository can only be initialized in an empty directory.")
	}

	os.Mkdir(".lvc", os.ModeDir)
	os.MkdirAll(".lvc/objects", os.ModeDir)
}

func Status() {
	println("status")
}

func ReadBlob(args []string) {
	if len(args) < 1 || len(args) > 2 {
		println("Invalid arguments!\nUsage: lvc readblob <sha> [filename]")
		return
	}

	sha := args[0]

	filename := "./.lvc/objects/" + sha[0:2] + "/" + sha[2:]

	if !FileExists(filename) {
		fmt.Printf("File %s does not exist!\n", filename)
		return
	}

	compressed, err := ioutil.ReadFile(filename)

	reader := bytes.NewReader(compressed)
	r, err := zlib.NewReader(reader)
	check(err)

	var content bytes.Buffer
	io.Copy(&content, r)

	header, err := content.ReadString('\u0000')
	check(err)

	datalengthstring := strings.TrimSuffix(strings.Split(header, " ")[1], "\u0000")

	datalength, err := strconv.Atoi(datalengthstring)
	check(err)

	data := make([]byte, datalength)
	bytesread, err := content.Read(data)
	check(err)

	if bytesread != datalength {
		println("Corrupt object")
		return
	}

	if len(args) == 2 {
		outfilename := args[1]
		ioutil.WriteFile(outfilename, data, os.ModePerm)
	} else {
		print(string(data))
	}
}

func WriteBlob(args []string) {
	if len(args) != 1 {
		println("Invalid arguments!\nUsage: lvc writeblob <filename>")
		return
	}

	filename := args[0]

	if !FileExists(filename) {
		fmt.Printf("File %s does not exist!\n", filename)
		return
	}

	fi, err := os.Stat(filename)
	check(err)

	header := fmt.Sprintf("blob %d\u0000", fi.Size())

	data, err := ioutil.ReadFile(filename)
	check(err)

	content := append([]byte(header)[:], data[:]...)
	hasher := sha1.New()
	hasher.Write(content)
	shabytes := hasher.Sum(nil)
	sha := fmt.Sprintf("%x", shabytes)

	println(sha)

	objectdir := ".lvc/objects/" + sha[0:2]
	objectfile := objectdir + "/" + sha[2:]

	var compressed bytes.Buffer

	w := zlib.NewWriter(&compressed)
	w.Write(content)
	w.Close()

	os.MkdirAll(objectdir, os.ModeDir)
	ioutil.WriteFile(objectfile, []byte(compressed.Bytes()), os.ModePerm)
}

func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		println("Little Version Control v0.1")
		return
	}

	cmd := args[0]

	switch cmd {
	case "b", "branch":
		Branch()
	case "com", "commit", "checkin":
		Commit()
	case "con", "config":
		Config()
	case "h", "history", "log":
		History()
	case "i", "init":
		Init()
	case "s", "status":
		Status()

	case "readblob":
		ReadBlob(args[1:])
	case "writeblob":
		WriteBlob(args[1:])
	}
}
