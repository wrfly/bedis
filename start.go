package bedis

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func (r *builtinRedis) start(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, binPath, r.confPath)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	log.Printf("redis starting")

	// set cpu affinity
	if r.opt.CPUAffinity {
		if cpu, err := setCPUAffinity(cmd.Process.Pid); err != nil {
			log.Printf("set cpu affinity err: %s", err)
		} else {
			defer givebackCPU(cpu)
		}
	}

	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Printf("built-in-redis: %s", in.Text())
		}
	}()

	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Printf("built-in-redis: %s", in.Text())
		}
	}()

	return cmd.Wait()
}

func (r *builtinRedis) checkStatus() error {
	// check if redis started
	for i := 0; i <= 3; i++ {
		if !fileExists(r.socketPath) {
			log.Printf("redis socket %s not exist", r.socketPath)
			time.Sleep(time.Millisecond * 100 * time.Duration(i+1))
			continue
		}
		return nil
	}
	return fmt.Errorf("cannot start redis in 3 seconds")
}
