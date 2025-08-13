package linux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	possibleCPUsFile = "/sys/devices/system/cpu/possible"
)

// FindPossibleCPUs returns the number of possible cpus on this system.
// if the file at `possibleCPUsFile` does not exist, it uses the `nproc` command instead.
func FindPossibleCPUs() (int, error) {
	if _, err := os.Stat(possibleCPUsFile); err == nil {
		return parseCPUsFromFile(possibleCPUsFile)
	}

	return getCPUsNproc()
}

// getCPUsNproc retrieves the possible cpus using the `nproc` command
func getCPUsNproc() (int, error) {
	output, err := exec.Command("nproc", "--all").Output()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(strings.Trim(string(output), "\n"))
}

// parseCPUsFromFile parses a cpu bitmap obtained from a file at `path`
func parseCPUsFromFile(path string) (int, error) {
	spec, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	n, err := parseCPUs(string(spec))
	if err != nil {
		return 0, fmt.Errorf("can't parse %s: %v", path, err)
	}

	return n, nil
}

// parseCPUs parses the number of cpus from a string produced
// by bitmap_list_string() in the Linux kernel.
// Multiple ranges are rejected, since they can't be unified
// into a single number.
// This is the format of /sys/devices/system/cpu/possible, it
// is not suitable for /sys/devices/system/cpu/online, etc.
func parseCPUs(spec string) (int, error) {
	if strings.Trim(spec, "\n") == "0" {
		return 1, nil
	}

	var low, high int
	n, err := fmt.Sscanf(spec, "%d-%d\n", &low, &high)
	if n != 2 || err != nil {
		return 0, fmt.Errorf("invalid format: %s", spec)
	}
	if low != 0 {
		return 0, fmt.Errorf("CPU spec doesn't start at zero: %s", spec)
	}

	// cpus is 0 indexed
	return high + 1, nil
}
