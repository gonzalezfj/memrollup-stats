package process

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gonzalezfj/memrollup-stats/src/internal/config"
)

// ProcessManager defines the interface for process management.
type ProcessManager interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning() bool
	Wait() error
}

// RealProcessManager implements ProcessManager for actual process management.
type RealProcessManager struct {
	Cmd    *exec.Cmd
	config *config.Config
}

// New creates a new process manager.
func New(cfg *config.Config) ProcessManager {
	return &RealProcessManager{config: cfg}
}

// Start implements ProcessManager interface.
func (pm *RealProcessManager) Start(ctx context.Context) error {
	pm.Cmd = exec.CommandContext(ctx, "bash", "-c", pm.config.Command)
	pm.Cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Pipe stdout and stderr to the parent process
	pm.Cmd.Stdout = os.Stdout
	pm.Cmd.Stderr = os.Stderr

	if err := pm.Cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.config.PID = pm.Cmd.Process.Pid
	pm.config.StartTime = time.Now()

	if pm.config.Verbose {
		log.Printf("Started process with PID: %d", pm.config.PID)
	}

	return nil
}

// Stop implements ProcessManager interface.
func (pm *RealProcessManager) Stop() error {
	if pm.Cmd == nil || pm.Cmd.Process == nil {
		return nil
	}

	if pm.config.Verbose {
		log.Printf("Terminating process group: %d", pm.config.PID)
	}

	syscall.Kill(-pm.config.PID, syscall.SIGTERM)
	time.Sleep(500 * time.Millisecond)
	syscall.Kill(-pm.config.PID, syscall.SIGKILL)

	return pm.Cmd.Wait()
}

// IsRunning implements ProcessManager interface.
func (pm *RealProcessManager) IsRunning() bool {
	return pm.Cmd != nil && pm.Cmd.Process != nil
}

// Wait implements ProcessManager interface.
func (pm *RealProcessManager) Wait() error {
	if pm.Cmd == nil {
		return fmt.Errorf("no process to wait for")
	}
	return pm.Cmd.Wait()
}
