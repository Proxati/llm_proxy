package transformers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

	"github.com/proxati/llm_proxy/v2/config"
	"github.com/proxati/llm_proxy/v2/proxy/addons/transformers/runners"
	"github.com/proxati/llm_proxy/v2/schema"
)

// FileProvider is a transformer provider that reads a file and executes it as a command in a
// subprocess. The command should read a schema.LogDumpContainer JSON object from stdin and write
// a schema.LogDumpContainer object in JSON to stdout.
type FileProvider struct {
	logger            *slog.Logger
	TransformerConfig *config.Transformer
	commandRunner     runners.Provider
}

// NewFileProvider creates a new FileProvider object with the given logger and transformer config
func NewFileProvider(logger *slog.Logger, ctx context.Context, Tcfg *config.Transformer) (*FileProvider, error) {
	if Tcfg == nil {
		return nil, errors.New("transformer config object is nil")
	}
	logger = logger.WithGroup("FileProvider").With("ServiceName", Tcfg.Name)

	commandPath := Tcfg.URL.Path
	if commandPath == "" {
		return nil, errors.New("command path is empty")
	}

	cr, err := runners.NewPosixCommand(logger, ctx, commandPath, Tcfg.Concurrency, Tcfg.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to create command runner: %w", err)
	}

	fp := &FileProvider{
		logger:            logger,
		TransformerConfig: Tcfg,
		commandRunner:     cr,
	}

	// health check the command via the command runner
	if err := fp.commandRunner.HealthCheck(ctx); err != nil {
		return nil, fmt.Errorf("initial health check failed: %w", err)
	}

	return fp, nil
}

// String returns a string representation of the FileProvider object
func (ht *FileProvider) String() string {
	return fmt.Sprintf("FileProvider{Name: %s, commandRunner: %s}", ht.TransformerConfig.Name, ht.commandRunner)
}

// HealthCheck wraps the command runner's health check function
func (ht *FileProvider) HealthCheck(ctx context.Context) error {
	return ht.commandRunner.HealthCheck(ctx)
}

// Transform uses the command runner to execute the command with the given request and response
// objects.
func (ht *FileProvider) Transform(
	ctx context.Context,
	logger *slog.Logger,
	oldReq *schema.ProxyRequest,
	oldResp *schema.ProxyResponse,
	newReq *schema.ProxyRequest,
	newResp *schema.ProxyResponse,

) (*schema.ProxyRequest, *schema.ProxyResponse, error) {
	logger = logger.WithGroup("FileProvider.Transform").With("ServiceName", ht.TransformerConfig.Name)
	req := oldReq // default to operating on the original request/response
	resp := oldResp

	// check if another transformer has previously updated the request
	if newReq != nil {
		logger.Debug("Transforming an updated request, and ignoring the original request", "newReq", newReq)
		req = newReq
	}

	if newResp != nil {
		logger.Debug("Transforming an updated response, and ignoring the original response", "newResp", newResp)
		resp = newResp
	}

	newLDC := schema.NewLogDumpContainerEmpty()
	newLDC.Request = req
	newLDC.Response = resp

	if req == nil {
		return nil, nil, errors.New("unable to transform, request object is nil")
	}

	logger.Debug("Starting transformation", "request", req, "response", resp)

	// convert the req to a json string, which will be sent via stdin to the command
	// the command should return a schema.LogDumpContainer object in json format.
	jsonBytes, err := newLDC.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed to marshal LDC to JSON: %v", err)
	}

	// run the command
	out, err := ht.commandRunner.Run(ctx, bytes.NewReader(jsonBytes))

	// check for errors
	if err != nil {
		return nil, nil, fmt.Errorf("unable to run command: %w", err)
	}
	logger.Debug("Command executed successfully", "output", string(out))

	// create a new schema.LogDumpContainer from the output
	ldc := &schema.LogDumpContainer{}
	if err := ldc.UnmarshalJSON(out); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal LogDumpContainer: %w", err)
	}

	// return the new request and response, if they exist and are not nil
	if ldc.Request != nil {
		logger.DebugContext(ctx, "Transformer updated request", "request", ldc.Request)
		newReq = ldc.Request
	}

	if ldc.Response != nil {
		logger.DebugContext(ctx, "Transformer updated response", "response", ldc.Response)
		newResp = ldc.Response
	}

	return newReq, newResp, nil
}

// Close currently does nothing for the FileProvider
func (ht *FileProvider) Close() error {
	return nil
}

// GetTransformerConfig returns the transformer config object
func (ht *FileProvider) GetTransformerConfig() config.Transformer {
	return *ht.TransformerConfig
}
