package rule

import (
	"io/ioutil"
	"path"
	"strings"
)

func ParseFile(file string) (*Ruleset, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	set := &Ruleset{}

	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		rule := &Rule{}

		switch {
		case !strings.Contains(fields[0], ":"):
			rule.Type = Suffix
			rule.Payload = fields[0]
		case strings.HasPrefix(fields[0], "include:"):
			rule.Type = Include
			rule.Payload = fields[0][len("include:"):]
		case strings.HasPrefix(fields[0], "full:"):
			rule.Type = Full
			rule.Payload = fields[0][len("full:"):]
		default:
			println("Unsupported rule: " + line)
			continue
		}

		var tags []string

		for i := 1; i < len(fields); i++ {
			if strings.HasPrefix(fields[i], "@") {
				tags = append(tags, fields[i][len("@"):])
			}
		}

		rule.Tags = tags

		set.Rules = append(set.Rules, rule)
	}

	return set, nil
}

func ParseDirectory(directory string) (map[string]*Ruleset, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	r := map[string]*Ruleset{}

	for _, file := range files {
		entry, err := ParseFile(path.Join(directory, file.Name()))
		if err != nil {
			return nil, err
		}

		r[file.Name()] = entry
	}

	return r, nil
}
