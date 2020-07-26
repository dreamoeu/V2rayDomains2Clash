package main

import (
	"bufio"
	"fmt"
	"github.com/kr328/domains2providers/trie"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func main() {
	data := path.Join(".", "domain-list-community", "data")
	generated := path.Join(".", "generated")

	_ = os.MkdirAll(generated, 0755)

	files, err := ioutil.ReadDir(data)
	if err != nil {
		panic(err.Error())
	}

	for _, file := range files {
		t := trie.New()

		if err := appendFileToTrie(data, file.Name(), t); err != nil {
			panic(err.Error())
		}

		file, err := os.OpenFile(path.Join(generated, file.Name()+".yaml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err.Error())
		}

		header := fmt.Sprintf("https://github.com/v2ray/domain-list-community/blob/master/data/%s\n\npayload:\n", file.Name())

		if _, err := file.WriteString(header); err != nil {
			panic(err.Error())
		}

		domains := t.Dump()

		for _, domain := range domains {
			if _, err := file.WriteString(fmt.Sprintf("  - \"+.%s\"\n", domain)); err != nil {
				panic(err.Error())
			}
		}

		_ = file.Close()
	}
}

func appendFileToTrie(base, fileName string, t *trie.DomainTrie) error {
	file, err := os.Open(path.Join(base, fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		lineBytes, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return nil
			}

			return err
		}

		line := string(lineBytes)

		line = strings.SplitN(line, "#", 2)[0]
		line = strings.SplitN(line, "@", 2)[0]
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "include:") {
			if err := appendFileToTrie(base, line[len("include:"):], t); err != nil {
				return err
			}
		} else {
			_ = t.Insert(line, true)
		}
	}
}
