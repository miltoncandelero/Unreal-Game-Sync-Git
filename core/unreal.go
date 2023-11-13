package core

import (
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

const UNREAL_PROCESS = "UnrealEditor"

func IsUnrealRunning() bool {
	allProcs, _ := process.Processes()
	for _, proc := range allProcs {
		procName, _ := proc.Name()
		if strings.Contains(procName, UNREAL_PROCESS) {
			return true
		}
	}
	return false
}
