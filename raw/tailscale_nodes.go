package raw

import (
	"sort"
	"strings"
	"encoding/json"
	"fmt"
	"net/http"
)

var tsNodes = Raw{
	Name:      "tailscale-node-ips",
	Behavior:  "ipcidr",
	SourceUrl: "https://login.tailscale.com/derpmap/default",
}

func LoadTailscaleNodes() ([]*RuleSet, error) {
	resp, err := http.Get(tsNodes.SourceUrl)
	if err != nil {
		return nil, fmt.Errorf("load %s: %s", tsNodes.Name, err.Error())
	}

	if resp.StatusCode/100 != 2 {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("load %s: response %s", tsNodes.Name, resp.Status)
	}

	defer resp.Body.Close()

	type TailscaleNodes struct {
		Regions map[string]struct {
			Nodes []struct {
				Name     string `json:"Name"`
				RegionID int    `json:"RegionID"`
				HostName string `json:"HostName"`
				IPv4     string `json:"IPv4"`
				IPv6     string `json:"IPv6"`
			} `json:"Nodes"`
		} `json:"Regions"`
	}

	var data TailscaleNodes
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode %s: %s", tsNodes.Name, err.Error())
	}

	var result []*RuleSet
	var rules []string

	for _, region := range data.Regions {
		for _, node := range region.Nodes {
			rules = append(rules, node.IPv4+"/32", node.IPv6+"/128")
		}
	}

	sort.Slice(rules, func(i, j int) bool {
		if strings.HasSuffix(rules[i], "/32") && !strings.HasSuffix(rules[j], "/32") {
            return true
        }
        return false
	})

	result = append(result, &RuleSet{
		Raw:   &tsNodes,
		Rules: rules,
	})

	return result, nil
}
