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

type UnixCommand struct {
	logger          *slog.Logger
	command         string
	sha256sum       string
	concurrency     int
	timeout         time.Duration
	bulkhead        bulkhead.Bulkhead[any]
	bulkheadTimeout time.Duration
}

func NewUnixCommand(logger *slog.Logger, command string, concurrency int, timeout time.Duration) (*UnixCommand, error) {
	cr := &UnixCommand{
		logger:      logger.With("commandRunner", command),
		command:     command,
		concurrency: concurrency,
		timeout:     timeout,
	}

	// check the commands sha256sum, store it in the provider so we can check if the command has changed
	sha256sum, err := fileutils.ComputeSHA256FromFileContents(command)
	if err != nil {
		return nil, fmt.Errorf("unable to compute sha256sum of command: %w", err)
	}
	cr.sha256sum = sha256sum

	// Create a Bulkhead with a limit of N concurrent executions
	bh := bulkhead.Builder[any](uint(concurrency)).
		// WithMaxWaitTime(bhTimeout).
		OnFull(func(e failsafe.ExecutionEvent[any]) {
			cr.logger.Warn("Bulkhead full")
		}).
		Build()
	cr.bulkhead = bh
	cr.bulkheadTimeout = timeout + 1*time.Minute

	return cr, nil
}

// Run executes the command with the given context and stdin.
// There are two timeout values to consider:
// - The context timeout, which is the maximum time the command is allowed to run.
// - The bulkhead timeout, which is the maximum time to wait in line for a permit to execute the command.
func (cr *UnixCommand) Run(ctx context.Context, stdin *bytes.Reader) (outPut []byte, err error) {
	runCtx, cancel := context.WithTimeout(ctx, cr.timeout) // command running timeout
	defer cancel()

	// wait for a permit to execute the command
	if err = cr.bulkhead.AcquirePermitWithMaxWait(ctx, cr.bulkheadTimeout); err != nil {
		defer cr.bulkhead.ReleasePermit()
		cmd := exec.CommandContext(runCtx, cr.command)
		cmd.Stdin = stdin
		return cmd.Output()
	}

	if err != nil {
		return nil, fmt.Errorf("error while attempting to acquire bulkhead permit: %w", err)
	}

	return nil, fmt.Errorf("unable to acquire bulkhead permit")
}

func (cr *UnixCommand) HealthCheck() error {
	// check if the command exists
	if _, err := exec.LookPath(cr.command); err != nil {
		return fmt.Errorf("command not found: %w", err)
	}

	// check if the command is executable
	if err := fileutils.IsExecutable(cr.command); err != nil {
		return fmt.Errorf("command is not executable: %w", err)
	}

	// check the commands sha256sum, and compare it to the stored sha256sum
	sha256sum, err := fileutils.ComputeSHA256FromFileContents(cr.command)
	if err != nil {
		return fmt.Errorf("unable to get sha256sum: %w", err)
	}

	if sha256sum != cr.sha256sum {
		return fmt.Errorf("sha256 checksum of the command has changed: %w", err)
	}

	return nil
}

func (cr *UnixCommand) String() string {
	return fmt.Sprintf(
		"unixPathCommandRunner{Command: %s, Concurrency: %d, Timeout: %s}",
		cr.command, cr.concurrency, cr.timeout)
}
