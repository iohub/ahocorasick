package cedar

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"testing"
)

type item struct {
	key   []byte
	value int
}

var trie = NewCedar()

func loadDict(t *testing.T) []item {
	t.Helper()

	var dict []item
	f, err := os.Open("testdata/dict.txt")
	if err != nil {
		t.Fatalf("failed to open testdata/dict.txt: %v", err)
	}

	defer f.Close()
	in := bufio.NewReader(f)

	added := make(map[string]struct{})
	var key string
	var freq int
	var pos string
	for {
		_, err := fmt.Fscanln(in, &key, &freq, &pos)
		if err == io.EOF {
			break
		}
		if _, ok := added[string(key)]; !ok {
			dict = append(dict, item{[]byte(key), freq})
			added[string(key)] = struct{}{}
		}
	}
	return dict
}
