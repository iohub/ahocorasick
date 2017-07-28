# Ahocorasick [![GoDoc](https://godoc.org/github.com/iohub/Ahocorasick?status.svg)](https://godoc.org/github.com/iohub/Ahocorasick)

Package `Ahocorasick` implementes fast, compact and low memory used aho-corasick algorithm based on efficiently-updatable double-array trie (cedar). and also supports visualizing inner data structures by [graphviz](http://graphviz.org) 

`cedar-go` is a [Golang](https://golang.org/) port of [cedar](http://www.tkl.iis.u-tokyo.ac.jp/~ynaga/cedar) which is written in C++ by Naoki Yoshinaga. [`cedar-go`](https://github.com/adamzy/cedar-go) currently implements the `reduced` verion of cedar. 
This package is not thread safe if there is one goroutine doing insertions or deletions. 

## Install
```
go get github.com/iohub/Ahocorasick
```

## Usage

* aho-corasick

```go 
package main

import (
	"fmt"

	"github.com/iohub/Ahocorasick"
)

func main() {
	fmt.Printf("\ntesting in simple case...\n")
	m := cedar.NewMatcher()
	words := []string{
		"she", "he", "her", "hers",
	}
	for i, word := range words {
		m.Insert([]byte(word), i)
	}
	// visualize trie 
	m.Cedar().DumpGraph("trie.gv")
	m.Compile()
	// visualize aho-corasick
	m.DumpGraph("dfa.gv")
	seq := []byte("hershertongher")
	fmt.Printf("searching %s\n", string(seq))
	req := m.Search(seq)
	for _, item := range req {
		fmt.Printf("key:%s value:%d\n", item.Key, item.Value.(int))
	}
}

```

​	output
```
testing in simple case...
searching hershertongher
key:he value:1
key:her value:2
key:hers value:3
key:he value:1
key:she value:0
key:her value:2
key:he value:1
key:her value:2
```

* trie

```go
package main

import (
	"fmt"

	"github.com/iohub/Ahocorasick"
)

func main() {
	fmt.Printf("\ntesting int float32 value...\n")
	cd := NewCedar()
	words := []string{
		"she", "hers", "her", "he",
	}
	for i, word := range words {
		v := float32(i) + 1.880001
		cd.Insert([]byte(word), v)
	}
	ids := cd.PrefixMatch([]byte("hers"), 0)
	for _, id := range ids {
		k, _ := cd.Key(id)
		v, _ := cd.Get(k)
		fmt.Printf("key:%s val:%f\n", string(k), v.(float32))
	} 
}

```
​	output
```
testing int float32 value...
key:he val:4.880001
key:her val:3.880001
key:hers val:2.880001
```

## Visualize

Install graphviz

```shell
# for mac os
brew install graphviz
# for ubuntu
sudo apt-get install graphviz
```

Dump structure in golang
```go
m := cedar.NewMatcher()
m.Insert...
m.Compile...
// dump trie graph in double array trie
m.Cedar().DumpGraph("trie.gv")
// dump aho-corasick dfa graph in double array trie
m.DumpGraph("dfa.gv")
```
Generate png 
```shell
dot -Tpng -o out.png trie.gv
```
* trie

![image](https://github.com/iohub/Ahocorasick/blob/master/pictures/trie.png)

* aho-corasick

![image](https://github.com/iohub/Ahocorasick/blob/master/pictures/dfa.png)
