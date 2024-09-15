package runners

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/bulkhead"
	"github.com/proxati/llm_proxy/v2/internal/fileutils"
)

// PosixCommand is a command runner that executes a command on a POSIX system.
type PosixCommand struct {
	logger          *slog.Logger
	command         string
	sha256sum       string
	concurrency     int
	bulkhead        bulkhead.Bulkhead[any]
	bulkheadTimeout time.Duration
}

func NewPosixCommand(logger *slog.Logger, ctx context.Context, command string, concurrency int, bulkHeadTimeout time.Duration) (*PosixCommand, error) {
	cr := &PosixCommand{
		logger:      logger.With("commandRunner", command),
		command:     command,
		concurrency: concurrency,
	}

	// check the commands sha256sum, store it in the provider so we can check if the command has changed
	sha256sum, err := fileutils.ComputeSHA256FromFileContentsContext(ctx, command)
	if err != nil {
		return nil, fmt.Errorf("unable to compute sha256sum of command: %w", err)
	}
	cr.sha256sum = sha256sum

	if cr.concurrency > 0 {
		// Create a Bulkhead with a limit of N concurrent executions
		bh := bulkhead.Builder[any](uint(concurrency)).
			// WithMaxWaitTime(bhTimeout).
			OnFull(func(e failsafe.ExecutionEvent[any]) {
				cr.logger.Warn("Bulkhead full")
			}).
			Build()
		cr.bulkhead = bh
		cr.bulkheadTimeout = bulkHeadTimeout
	}

	return cr, nil
}

// Run executes the command with the given context and stdin.
// There are two timeout values to consider:
// - The context timeout, which is the maximum time the command is allowed to run.
// - The bulkhead timeout, which is the maximum time to wait in line for a permit to execute the command.
func (cr *PosixCommand) Run(ctx context.Context, stdin *bytes.Reader) ([]byte, error) {
	if cr.concurrency > 0 {
		// Attempt to acquire a permit
		if err := cr.bulkhead.AcquirePermitWithMaxWait(ctx, cr.bulkheadTimeout); err != nil {
			// Failed to acquire a permit; return an error
			return nil, fmt.Errorf("unable to acquire bulkhead permit: %w", err)
		}
		// Ensure the permit is released after execution
		defer cr.bulkhead.ReleasePermit()
	}

	// Execute the command
	cmd := exec.CommandContext(ctx, cr.command)
	cmd.Stdin = stdin

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	return output, nil
}

func (cr *PosixCommand) HealthCheck(ctx context.Context) error {
	// check if the command exists
	if _, err := exec.LookPath(cr.command); err != nil {
		return fmt.Errorf("command not found: %w", err)
	}

	// check if the command is executable
	if err := fileutils.IsExecutable(cr.command); err != nil {
		return fmt.Errorf("command is not executable: %w", err)
	}

	// check the commands sha256sum, and compare it to the stored sha256sum
	sha256sum, err := fileutils.ComputeSHA256FromFileContentsContext(ctx, cr.command)
	if err != nil {
		return fmt.Errorf("unable to compute the SHA256 checksum: %w", err)
	}

	if sha256sum != cr.sha256sum {
		return fmt.Errorf("SHA256 checksum of the command has changed: %w", err)
	}

	return nil
}

func (cr *PosixCommand) String() string {
	return fmt.Sprintf(
		"unixPathCommandRunner{Command: %s, Concurrency: %d, BulkheadTimeout: %s}",
		cr.command, cr.concurrency, cr.bulkheadTimeout)
}
