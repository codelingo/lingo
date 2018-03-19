package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func UserCancelContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		cancel()
		os.Exit(1)
	}()
	return ctx, cancel
}
