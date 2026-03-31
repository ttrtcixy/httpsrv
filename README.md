# HTTP Server Package

[Russian](./README.ru.md)

Small Go package for running an HTTPS server with graceful shutdown.

## How it works

The package wraps `net/http.Server` and configures:

- address from `Host` and `Port`
- TLS (minimum TLS 1.3) with certificate and private key
- HTTP timeouts and max header size from config
- graceful shutdown with timeout (`ShutdownTimeout`)

`Start()` runs `ListenAndServeTLS`, and `Close()` shuts the server down with context timeout handling.

## Example usage

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpsrv "github.com/wk/pkg/http_server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg := &httpsrv.Config{
		Host:              "127.0.0.1",
		Port:              "8443",
		ShutdownTimeout:   5 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
		WriteTimeout:      10 * time.Second,
		TLSCertificate:    "cert.pem",
		PrivateKey:        "key.pem",
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := httpsrv.New(logger, cfg, handler)

	go func() {
		if err := server.Start(context.Background()); err != nil {
			logger.Error("server stopped", slog.Any("error", err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	_ = server.Close(context.Background())
}
```
