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

// FileProvider is a transformer provider that reads a file and executes it as a command.
type FileProvider struct {
	logger            *slog.Logger
	TransformerConfig *config.Transformer
	ctx               context.Context
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

	cr, err := runners.NewUnixCommand(logger, commandPath, Tcfg.Concurrency, Tcfg.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("unable to create command runner: %w", err)
	}

	fp := &FileProvider{
		logger:            logger,
		TransformerConfig: Tcfg,
		ctx:               ctx,
		commandRunner:     cr,
	}

	// health check the command via the command runner
	if err := fp.commandRunner.HealthCheck(); err != nil {
		return nil, fmt.Errorf("initial health check failed: %w", err)
	}

	return fp, nil
}

func (ht *FileProvider) String() string {
	return fmt.Sprintf("FileProvider{Name: %s, commandRunner: %s}", ht.TransformerConfig.Name, ht.commandRunner)
}

func (ht *FileProvider) HealthCheck() error {
	return ht.commandRunner.HealthCheck()
}

func (ht *FileProvider) Transform(
	ctx context.Context,
	req *schema.ProxyRequest,
	newReq *schema.ProxyRequest,
	newResp *schema.ProxyResponse,
) (*schema.ProxyRequest, *schema.ProxyResponse, error) {
	if req == nil {
		return nil, nil, errors.New("unable to transform, request object is nil")
	}

	logger := ht.logger.With("req", req, "newReq", newReq, "newResp", newResp)
	logger.Debug("Transforming")

	// convert the req to a json string, which will be sent via stdin to the command
	// the command should return a schema.LogDumpContainer object in json format.
	jsonBytes, err := req.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed to marshal ProxyRequest to JSON: %v", err)
	}

	// run the command
	out, err := ht.commandRunner.Run(ht.ctx, bytes.NewReader(jsonBytes))

	// check for errors
	if err != nil {
		return nil, nil, fmt.Errorf("unable to run command: %w", err)
	}

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

func (ht *FileProvider) Close() error {
	return nil
}

func (ht *FileProvider) GetTransformerConfig() config.Transformer {
	return *ht.TransformerConfig
}
