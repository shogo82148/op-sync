package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/shogo82148/op-sync/internal/opsync"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	os.Exit(opsync.Run(ctx, os.Args[1:]))
}
