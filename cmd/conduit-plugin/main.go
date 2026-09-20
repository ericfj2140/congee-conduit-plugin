package main

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/michmich112/conduit-plugin/embed"
	"github.com/michmich112/conduit-plugin/handler"
	sdk "github.com/michmich112/congee/sdk/plugin"
)

func main() {
	dataDir := os.Getenv(sdk.EnvPluginDataDir)
	if dataDir == "" {
		dataDir = "."
	}
	var e embed.Embedder = embed.Fake{}
	if os.Getenv("CONDUIT_EMBEDDER") != "fake" {
		model := filepath.Join(dataDir, "models", "minilm.onnx")
		if x, err := embed.NewONNX(model); err == nil {
			e = x
		}
	}
	h := handler.New(dataDir, e)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	if err := sdk.Serve(ctx, h); err != nil && ctx.Err() == nil {
		os.Exit(1)
	}
}
