package cedar

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestDumpMatcher(t *testing.T) {
	fmt.Printf("\ntesting in simple case...\n")
	m := NewMatcher()
	words := []string{
		"she", "he", "her", "hers",
	}
	for i, word := range words {
		m.Insert([]byte(word), i)
	}
	m.Cedar().DumpGraph("trie.gv")
	m.Compile()
	m.DumpGraph("dfa.gv")
	seq := []byte("hershertongher")
	fmt.Printf("searching %s\n", string(seq))
	req := m.Match(seq)
	for _, item := range req {
		key := m.TokenOf(seq, item)
		fmt.Printf("key:%s value:%d\n", key, item.Value.(int))
	}
}

func timeTrack(start time.Time, name string) {
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

func calcTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %s\n", name, elapsed)
}

func testIohub(dictName, textName string) {
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
	m := NewMatcher()

	func() {
		defer calcTime(time.Now(), "iohub/ahocorasick [build]")
		for i, b := range dict {
			m.Insert(b, i)
		}
		m.Compile()
	}()

	func() {
		defer calcTime(time.Now(), "iohub/ahocorasick [match]")
		m.Match(content)
	}()
}

func TestWithDict(t *testing.T) {
	zhDict := "./benchmark/cn/dictionary.txt"
	zhText := "./benchmark/cn/text.txt"
	testIohub(zhDict, zhText)
}

func TestMatcher(t *testing.T) {
	fmt.Printf("\ntesting Insert & Compile & Search in big dictionary...\n")
	m := NewMatcher()
	fmt.Println("Loading...")
	dict := loadDict()
	size := len(dict)
	fmt.Printf("%d key-value pairs in dictionary\n", size)
	fmt.Println("Inserting...")

	func() {
		defer timeTrack(time.Now(), "Insert")
		for i := 0; i < size; i++ {
			m.Insert(dict[i].key, i)
		}
	}()

	// m.da.DumpGraph("bigtrie.py")
	fmt.Println("Compile...")
	func() {
		defer timeTrack(time.Now(), "Compile")
		m.Compile()
	}()
	// m.DumpGraph("bigdfa.py")
	seq := []byte("一丁不识一丁点C++的T桖中华人民共和国人民解放军轰炸南京长江大桥")
	fmt.Printf("Searching %s\n", string(seq))
	req := m.Match(seq)
	for _, item := range req {
		key := m.TokenOf(seq, item)
		fmt.Printf("key:%s value:%d\n", key, item.Value.(int))
	}
}
