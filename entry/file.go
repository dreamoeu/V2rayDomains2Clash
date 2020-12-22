package entry

import (
	"io/ioutil"
	"log"
	"path"
	"strings"
)

func ParseFile(file string) (*Entry, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	entry := &Entry{}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.SplitN(line, "@", 2)[0]
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		l := &Line{}

		switch {
		case !strings.Contains(line, ":"):
			l.Type = Suffix
			l.Payload = line
		case strings.HasPrefix(line, "include:"):
			l.Type = Include
			l.Payload = line[8:]
		case strings.HasPrefix(line, "full:"):
			l.Type = Full
			l.Payload = line[5:]
		default:
			log.Println("Unsupported line", line)
			continue
		}

		entry.Lines = append(entry.Lines, l)
	}

	return entry, nil
}

func BuildCache(baseDir string) (map[string]*Entry, error) {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	r := map[string]*Entry{}

	for _, file := range files {
		entry, err := ParseFile(path.Join(baseDir, file.Name()))
		if err != nil {
			return nil, err
		}

		r[file.Name()] = entry
	}

	return r, nil
}
