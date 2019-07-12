package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"net/http"
	_ "net/http/pprof"

	goahocorasick "github.com/anknown/ahocorasick"
	"github.com/cloudflare/ahocorasick"
	cedar "github.com/iohub/ahocorasick"
)

const zhDict = "./cn/dictionary.txt"
const zhText = "./cn/text.txt"
const enDict = "./en/dictionary.txt"
const enText = "./en/text.txt"

func calcTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s\t %s\n", name, elapsed)
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

func testCloudflare(fdict, ftext string) {
	dict, err := readBytes(fdict)
	if err != nil {
		fmt.Println(err)
		return
	}
	content, err := ioutil.ReadFile(ftext)
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
	fmt.Printf("cloudflare/ahocorasick [mem]\t %d KBytes\n", (after-before)/1024)
}

func testAnknown(fdict, ftext string) {
	dict, err := readRunes(fdict)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := ioutil.ReadFile(ftext)
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
	fmt.Printf("anknown/ahocorasick [mem]\t %d KBytes\n", (after-before)/1024)
}

func testIohub(fdict, ftext string) {
	dict, err := readBytes(fdict)
	if err != nil {
		fmt.Println(err)
		return
	}

	content, err := ioutil.ReadFile(ftext)
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
		clen := len(content)
		tlen := 0
		for s := 0; clen > 0; s += tlen {
			tlen := cedar.DefaultTokenBufferSize / 2
			if clen < tlen {
				tlen = clen
			}
			text := content[s:tlen]
			m.Match(text)
			clen -= tlen
		}
	}()

	runtime.GC()
	runtime.ReadMemStats(mem)
	after := mem.HeapAlloc
	fmt.Printf("iohub/ahocorasick [mem]\t\t %d KBytes\n", (after-before)/1024)
}

func main() {

	f, err := os.Create("benchmark.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	go func() {
		http.ListenAndServe(":8787", http.DefaultServeMux)
	}()
	fmt.Println("\nBenchmark in english dict and text")
	// testCloudflare(enDict, enText)
	testAnknown(enDict, enText)
	testIohub(enDict, enText)

	fmt.Println("\nBenchmark in chinese dict and text")
	// testCloudflare(zhDict, zhText)
	testAnknown(zhDict, zhText)
	testIohub(zhDict, zhText)

	fmt.Println("\nCTL+C exit http pprof")
	// time.Sleep(1 * time.Minute)
}
