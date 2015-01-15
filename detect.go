// Host detection.

package embd

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

// The Host type represents all the supported host types.
type Host string

const (
	// HostNull reprents a null host.
	HostNull Host = ""

	// HostRPi represents the RaspberryPi.
	HostRPi = "Raspberry Pi"

	// HostBBB represents the BeagleBone Black.
	HostBBB = "BeagleBone Black"

	// HostGalileo represents the Intel Galileo board.
	HostGalileo = "Intel Galileo"

	// HostCubieTruck represents the Cubie Truck.
	HostCubieTruck = "CubieTruck"

	// HostRadxa represents the Radxa board.
	HostRadxa = "Radxa"
)

func execOutput(name string, arg ...string) (output string, err error) {
	var out []byte
	if out, err = exec.Command(name, arg...).Output(); err != nil {
		return
	}
	output = strings.TrimSpace(string(out))
	return
}

func nodeName() (string, error) {
	return execOutput("uname", "-n")
}

func parseVersion(str string) (major, minor, patch int, err error) {
	versionNumber := strings.Split(str, "-")
	parts := strings.Split(versionNumber[0], ".")
	len := len(parts)

	if major, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, 0, err
	}
	if minor, err = strconv.Atoi(parts[1]); err != nil {
		return 0, 0, 0, err
	}
	if len > 2 {
		part := parts[2]
		part = strings.TrimSuffix(part, "+")
		if patch, err = strconv.Atoi(part); err != nil {
			return 0, 0, 0, err
		}
	}

	return major, minor, patch, err
}

func kernelVersion() (major, minor, patch int, err error) {
	output, err := execOutput("uname", "-r")
	if err != nil {
		return 0, 0, 0, err
	}

	return parseVersion(output)
}

func getPiRevision() (int, error) {
	//default return code of a rev2 board
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return 4, err
	}
	for _, line := range strings.Split(string(cpuinfo), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "Revision" {
			rev, err := strconv.ParseInt(fields[2], 16, 8)
			return int(rev), err
		}
	}
	return 4, nil
}

func cpuInfo() (string, string, int, error) {
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", "", 0, err
	}
	revision := 0
	model := ""
	hardware := ""

	for _, line := range strings.Split(string(cpuinfo), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) <= 0 {
			continue
		}
		if strings.HasPrefix(fields[0], "Revision") {
			rev, err := strconv.ParseInt(fields[1], 16, 8)
			if err != nil {
				continue
			}
			revision = int(rev)
		} else if strings.HasPrefix(fields[0], "Hardware") {
			hardware = fields[1]
		} else if strings.HasPrefix(fields[0], "model name") {
			model = fields[1]
		}
	}
	return model, hardware, revision, nil
}

// DetectHost returns the detected host and its revision number.
func DetectHost() (Host, int, error) {

	major, minor, patch, err := kernelVersion()

	if err != nil {
		return HostNull, 0, err
	}

	if major < 3 || (major == 3 && minor < 8) {
		return HostNull, 0, fmt.Errorf("embd: linux kernel versions lower than 3.8 are not supported. you have %v.%v.%v", major, minor, patch)
	}
	
	model, hardware, revision, err := cpuInfo()
	if err != nil {
		return HostNull, 0, err
	}

	var host Host = HostNull
	var rev int = 0
	
	if strings.Contains(model, "ARMv7") && (strings.Contains(hardware,"AM33XX") || strings.Contains(hardware,"AM335X")) {
		host = HostBBB
		rev = revision
	} else if strings.Contains(hardware,"BCM2708") {
		host = HostRPi
		rev = revision
	} else {
		return HostNull, 0, fmt.Errorf("embd: your host %q : %q is not supported at this moment. please request support at https://github.com/kidoman/embd/issues", host, model)
	}

	return host, rev, nil
}
