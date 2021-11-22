package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr: ":8000",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("Graceful handler exit")
					w.WriteHeader(http.StatusOK)
					return
				case <-time.After(1 * time.Second):
					fmt.Println("Hello in a loop")
				}
			}
		}),
		BaseContext: func(_ net.Listener) context.Context {
			return mainCtx
		},
	}
	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}
}
