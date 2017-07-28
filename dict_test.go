package cedar

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type item struct {
	key   []byte
	value int
}

var trie = NewCedar()

func loadDict() []item {
	var dict []item
	f, err := os.Open("testdata/dict.txt")
	if err != nil {
		panic("failed to open testdata/dict.txt")
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
