# Ahocorasick 
[![Go Report Card](https://goreportcard.com/badge/github.com/iohub/Ahocorasick?style=flat-square)](https://goreportcard.com/report/github.com/iohub/Ahocorasick) [![Build Status](https://semaphoreci.com/api/v1/iohub/ahocorasick/branches/master/badge.svg)](https://semaphoreci.com/iohub/ahocorasick) [![GoCover](http://gocover.io/_badge/github.com/iohub/Ahocorasick)](http://gocover.io/github.com/iohub/Ahocorasick) [![GoDoc](https://godoc.org/github.com/iohub/Ahocorasick?status.svg)](https://godoc.org/github.com/iohub/Ahocorasick)

Package `Ahocorasick` implementes fast, compact and low memory used aho-corasick algorithm based on double-array trie. and also supports visualizing inner data structures by [graphviz](http://graphviz.org) 

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
	req := m.Match(seq)
	for _, item := range req {
		key := m.TokenOf(seq, item)
		fmt.Printf("key:%s value:%d\n", key, item.Value.(int))
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
<img src="https://github.com/iohub/Ahocorasick/blob/master/pictures/trie.png" alt="GitHub" width="420" height="460" /> 

* aho-corasick 
<img src="https://github.com/iohub/Ahocorasick/blob/master/pictures/dfa.png" alt="GitHub" width="420" height="460" />


## Chinese words segment demo

Build demo test

```shell
go test .
./Ahocorasick.test
```

demo output
```go
Searching 一丁不识一丁点C++的T桖中华人民共和国人民解放军轰炸南京长江大桥
key:一 value:7
key:一丁 value:17
key:丁 value:3317
key:一丁不识 value:18
key:不识 value:9890
key:识 value:290279
key:一 value:7
key:不识一丁 value:9891
key:一丁 value:17
key:丁 value:3317
key:一丁点 value:19
key:丁点 value:3519
key:点 value:214913
key:C++ value:5
key:的 value:233716
key:中 value:13425
key:中华 value:13663
key:华 value:63497
key:华人 value:63545
key:人 value:25372
key:中华人民 value:13667
key:人民 value:25881
key:民 value:195603
key:中华人民共和国 value:13668
key:人民共和国 value:25891
key:共和国 value:44163
key:国 value:88227
key:国人 value:88295
key:人 value:25372
key:人民 value:25881
key:民 value:195603
key:解 value:287247
key:解放 value:287374
key:放 value:160645
key:人民解放军 value:25927
key:解放军 value:287381
key:军 value:46587
key:轰 value:302585
key:轰炸 value:302617
key:炸 value:214859
key:南 value:64826
key:南京 value:64860
key:京 value:24978
key:长 value:321363
key:长江 value:321685
key:江 value:198725
key:南京长江大桥 value:64899
key:长江大桥 value:321699
key:大桥 value:98737
key:桥 value:187704
```


## Benchmark

ahocorasick golang implementation: [`cloudflare`](https://github.com/cloudflare/ahocorasick) [`anknown`](https://github.com/anknown/ahocorasick) [`iohub`](https://github.com/iohub/Ahocorasick)

  ![image](https://github.com/iohub/Ahocorasick/blob/master/pictures/merge_from_ofoct.jpg)


How to run benchmark

```
git clone https://github.com/iohub/Ahocorasick
cd benchmark
go get github.com/cloudflare/ahocorasick
go get github.com/anknown/ahocorasick
go get github.com/iohub/Ahocorasick
go build .
./benchmark
```
