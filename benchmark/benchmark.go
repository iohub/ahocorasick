package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	_ "net/http/pprof"

	"github.com/anknown/ahocorasick"
	"github.com/cloudflare/ahocorasick"
	"github.com/iohub/Ahocorasick"
)

const zhDict = "./cn/dictionary.txt"
const zhText = "./cn/text.txt"
const enDict = "./en/dictionary.txt"
const enText = "./en/text.txt"

func calcTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

func readBytes(filename string) ([][]byte, error) {
	dict := [][]byte{}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0660)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)
	for {
		l, err := r.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		l = bytes.TrimSpace(l)
		dict = append(dict, l)
	}

	return dict, nil
}

func readRunes(filename string) ([][]rune, error) {
	dict := [][]rune{}

	f, err := os.OpenFile(filename, os.O_RDONLY, 0660)
	if err != nil {
		return nil, err
	}

	r := bufio.NewReader(f)
	for {
		l, err := r.ReadBytes('\n')
		if err != nil || err == io.EOF {
			break
		}
		l = bytes.TrimSpace(l)
		dict = append(dict, bytes.Runes(l))
	}

	return dict, nil
}

func testA(dictName, textName string) {
	dict, err := readBytes(dictName)
	if err != nil {
		fmt.Println(err)
		return
	}
	content, err := ioutil.ReadFile(textName)
	if err != nil {
		fmt.Println(err)
		return
	}
	mem := new(runtime.MemStats)
	runtime.GC()
	runtime.ReadMemStats(mem)
	before := mem.HeapAlloc
	var m *ahocorasick.Matcher
	func() {
		defer calcTime(time.Now(), "cloudflare/ahocorasick [build]")
		m = ahocorasick.NewMatcher(dict)
	}()

	func() {
		defer calcTime(time.Now(), "cloudflare/ahocorasick [match]")
		m.Match(content)
	}()

	runtime.GC()
	runtime.ReadMemStats(mem)
	after := mem.HeapAlloc
	fmt.Printf("cloudflare/ahocorasick [mem] took %d KBytes\n", (after-before)/1024)
}

func testB(dictName, textName string) {
	dict, err := readRunes(enDict)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := ioutil.ReadFile(enText)
	if err != nil {
		fmt.Println(err)
		return
	}

	contentRune := bytes.Runes([]byte(content))
	mem := new(runtime.MemStats)
	runtime.GC()
	runtime.ReadMemStats(mem)
	before := mem.HeapAlloc
	var m *goahocorasick.Machine
	func() {
		defer calcTime(time.Now(), "anknown/ahocorasick [build]")
		m = new(goahocorasick.Machine)
		if err := m.Build(dict); err != nil {
			fmt.Println(err)
			return
		}
	}()

	func() {
		defer calcTime(time.Now(), "anknown/ahocorasick [match]")
		m.MultiPatternSearch(contentRune, false)
	}()

	runtime.GC()
	runtime.ReadMemStats(mem)
	after := mem.HeapAlloc
	fmt.Printf("anknown/ahocorasick [mem] took %d KBytes\n", (after-before)/1024)
}

func testC(dictName, textName string) {
	dict, err := readBytes(enDict)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := ioutil.ReadFile(enText)
	if err != nil {
		fmt.Println(err)
		return
	}
	mem := new(runtime.MemStats)
	runtime.GC()
	runtime.ReadMemStats(mem)
	before := mem.HeapAlloc
	var m *cedar.Matcher

	func() {
		defer calcTime(time.Now(), "iohub/ahocorasick [build]")
		m = cedar.NewMatcher()
		for i, b := range dict {
			m.Insert(b, i)
		}
		m.Compile()
	}()

	func() {
		defer calcTime(time.Now(), "iohub/ahocorasick [match]")
		m.Match(content)
	}()

	runtime.GC()
	runtime.ReadMemStats(mem)
	after := mem.HeapAlloc
	fmt.Printf("iohub/ahocorasick [mem] took %d KBytes\n", (after-before)/1024)
}

func main() {
	/*
		f, err := os.Create("benchmark.bin")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	*/
	fmt.Println("\nBenchmark in english dict and text")
	testA(enDict, enText)
	testB(enDict, enText)
	testC(enDict, enText)

	fmt.Println("\nBenchmark in chinese dict and text")
	testA(zhDict, zhText)
	testB(zhDict, zhText)
	testC(zhDict, zhText)

	http.ListenAndServe(":8080", http.DefaultServeMux)
}
