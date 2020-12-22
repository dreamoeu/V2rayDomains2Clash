package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"

	"github.com/kr328/domains2providers/entry"
	"github.com/kr328/domains2providers/trie"
)

func main() {
	data := path.Join(".", "domain-list-community", "data")
	generated := path.Join(".", "generated")

	_ = os.MkdirAll(generated, 0755)

	files, err := ioutil.ReadDir(data)
	if err != nil {
		log.Println("Open domain list:", err.Error())

		return
	}

	cache, err := entry.BuildCache(data)

	for _, file := range files {
		t := trie.New()

		if err := putAllDomains(cache, t, file.Name()); err != nil {
			log.Println("Put domain list", err.Error())

			return
		}

		domains := t.Dump()

		sort.Strings(domains)

		output, err := os.OpenFile(path.Join(generated, file.Name()+".yaml"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Println("Open output file", file.Name(), ":", err.Error())

			return
		}

		_, _ = output.WriteString("# Generated from https://github.com/v2fly/domain-list-community/tree/master/data/" + file.Name() + "\n\n")
		_, _ = output.WriteString("payload:\n")

		for _, domain := range domains {
			if _, err := output.WriteString(fmt.Sprintf("  - \"%s\"\n", domain)); err != nil {
				panic(err.Error())
			}
		}

		_ = output.Close()
	}
}

func putAllDomains(cache map[string]*entry.Entry, trie *trie.Trie, name string) error {
	root := cache[name]
	if root == nil {
		return fmt.Errorf("entry %s not found", name)
	}

	for _, l := range root.Lines {
		switch l.Type {
		case entry.Include:
			if err := putAllDomains(cache, trie, l.Payload); err != nil {
				return err
			}
		case entry.Full:
			_ = trie.Insert(l.Payload, true)
		case entry.Suffix:
			_ = trie.Insert(l.Payload, false)
		}
	}

	return nil
}
