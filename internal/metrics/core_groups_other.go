//go:build !darwin

package metrics

func hostCoreGroups() ([]CoreGroup, map[int]string) { return nil, nil }
