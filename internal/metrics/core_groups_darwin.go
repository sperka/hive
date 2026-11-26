package metrics

import (
	"fmt"
	"os/exec"

	"golang.org/x/sys/unix"
)

func hostCoreGroups() ([]CoreGroup, map[int]string) {
	levels, err := unix.SysctlUint32("hw.nperflevels")
	if err != nil || levels < 2 {
		return nil, nil
	}
	var groups []CoreGroup
	for level := range levels {
		key := fmt.Sprintf("hw.perflevel%d.", level)
		name, err := unix.Sysctl(key + "name")
		if err != nil || name == "" {
			return nil, nil
		}
		count, err := unix.SysctlUint32(key + "physicalcpu")
		if err != nil || count == 0 {
			return nil, nil
		}
		groups = append(groups, CoreGroup{Name: name, Count: int(count)})
	}
	// The CPU nodes pair the Mach logical CPU ID with its hardware cluster type.
	registry, err := exec.Command("/usr/sbin/ioreg", "-a", "-l", "-r", "-p", "IODeviceTree", "-n", "cpus", "-d", "2").Output()
	if err != nil {
		return groups, nil
	}
	return groups, registryCoreTypes(registry, groups)
}
