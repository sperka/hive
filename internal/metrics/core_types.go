package metrics

import (
	"bytes"

	"howett.net/plist"
)

func registryCoreTypes(data []byte, groups []CoreGroup) map[int]string {
	var roots []struct {
		Children []struct {
			ID      *int   `plist:"logical-cpu-id"`
			Cluster []byte `plist:"cluster-type"`
		} `plist:"IORegistryEntryChildren"`
	}
	if _, err := plist.Unmarshal(data, &roots); err != nil || len(roots) != 1 {
		return nil
	}
	types := make(map[int]string)
	counts := make(map[string]int)
	for _, node := range roots[0].Children {
		if node.ID == nil {
			continue
		}
		var name string
		switch string(bytes.TrimRight(node.Cluster, "\x00")) {
		case "P":
			name = "Performance"
		case "E":
			name = "Efficiency"
		default:
			return nil
		}
		if _, duplicate := types[*node.ID]; duplicate {
			return nil
		}
		types[*node.ID] = name
		counts[name]++
	}
	// Mach returns load counters in logical ID order. Verify every index and
	// cross-check the registry against the independent perflevel counts.
	for id := range len(types) {
		if _, ok := types[id]; !ok {
			return nil
		}
	}
	if len(counts) != len(groups) {
		return nil
	}
	for _, group := range groups {
		if counts[group.Name] != group.Count {
			return nil
		}
	}
	return types
}
