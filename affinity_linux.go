package bedis

import (
	"fmt"

	"golang.org/x/sys/unix"
)

var avaliableCPU = make(map[int]bool)

func init() {
	mask := new(unix.CPUSet)
	unix.SchedGetaffinity(0, mask)
	// range 128 cpu
	for i := 0; i < 128; i++ {
		if mask.IsSet(i) {
			avaliableCPU[i] = true
		}
	}
}

func setCPUAffinity(pid int) (int, error) {
	if len(avaliableCPU) < 2 {
		return 0, fmt.Errorf("cannot set cpu affinity, cpu set < 2")
	}

	for cpu := range avaliableCPU {
		mask := new(unix.CPUSet)
		mask.Set(cpu) // bind this core
		if err := unix.SchedSetaffinity(pid, mask); err != nil {
			return 0, fmt.Errorf("set %w", err)
		}
		delete(avaliableCPU, cpu)
		return cpu, nil
	}

	return 0, nil
}

func givebackCPU(cpu int) {
	avaliableCPU[cpu] = true
}
