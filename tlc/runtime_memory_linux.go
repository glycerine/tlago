package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func tlcPhysicalMemoryBytes() (int64, error) {
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err != nil {
		return 0, err
	}
	total := uint64(info.Totalram) * uint64(info.Unit)
	if total == 0 || total > uint64(^uint64(0)>>1) {
		return 0, fmt.Errorf("invalid physical memory size %d", total)
	}
	memory := int64(total)
	memberships, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return memory, nil
	}
	mounts, err := os.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return memory, nil
	}
	// A process may be in a nested cgroup and a container mount may expose only
	// a subtree. Every visible ancestor's hard limit constrains the heap budget.
	unescape := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)
	for _, line := range strings.Split(string(memberships), "\n") {
		entry := strings.SplitN(line, ":", 3)
		if len(entry) != 3 {
			continue
		}
		for _, mount := range strings.Split(string(mounts), "\n") {
			fields := strings.Fields(mount)
			sep := -1
			for i, f := range fields {
				if f == "-" {
					sep = i
					break
				}
			}
			if sep < 6 || sep+3 >= len(fields) {
				continue
			}
			file := ""
			switch fields[sep+1] {
			case "cgroup2":
				if entry[0] == "0" && entry[1] == "" {
					file = "memory.max"
				}
			case "cgroup":
				if strings.Contains(","+entry[1]+",", ",memory,") && strings.Contains(","+fields[sep+3]+",", ",memory,") {
					file = "memory.limit_in_bytes"
				}
			}
			if file == "" {
				continue
			}
			root, base := filepath.Clean(unescape.Replace(fields[3])), filepath.Clean(unescape.Replace(fields[4]))
			member := filepath.Clean(entry[2])
			rel, err := filepath.Rel(root, member)
			if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
				continue
			}
			dir := filepath.Join(base, rel)
			for {
				if data, err := os.ReadFile(filepath.Join(dir, file)); err == nil {
					if limit, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil && limit > 0 && limit < memory {
						memory = limit
					}
				}
				if dir == base {
					break
				}
				dir = filepath.Dir(dir)
			}
		}
	}
	return memory, nil
}
