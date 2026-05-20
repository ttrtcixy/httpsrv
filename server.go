package httpsrv

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Host              string
	Port              string
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	WriteTimeout      time.Duration
	TLSCertificate    string
	PrivateKey        string
	Addr 			  string // Pre-formatted address
}

type Server struct {
	log *slog.Logger
	cfg *Config
	srv *http.Server
}

func New(slogLog *slog.Logger, cfg *Config, handler http.Handler) *Server {
	cfg.Addr = net.JoinHostPort(cfg.Host, cfg.Port)

	// не будет проблем из-за того что логер может быть буферезированным
	log := slog.NewLogLogger(slogLog.Handler(), slog.LevelError)

	s := &Server{
		cfg: cfg,
		log: slogLog.WithGroup("http_server"),
	}

	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS13, ServerName: "localhost"}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
		//BaseContext: func(listener net.Listener) context.Context {
		//	return ctx
		//},
		TLSConfig:         tlsCfg,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ErrorLog:          log,
	}

	s.srv = srv

	return s
}

// Start http server
func (s *Server) Start(_ context.Context) error {
	s.log.LogAttrs(nil, slog.LevelInfo, "Starting http server on", slog.String("addr", s.cfg.Addr))

	return s.srv.ListenAndServeTLS(s.cfg.TLSCertificate, s.cfg.PrivateKey)
}

// Close http server
func (s *Server) Close(ctx context.Context) error {
	const op = "http.server.close()"

	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.log.LogAttrs(nil, slog.LevelWarn, "server shutdown timed out, some connections were forced to close", slog.Duration("timeout", s.cfg.ShutdownTimeout))
			return nil
		}

		return fmt.Errorf("%s -> %w", op, err)
	}

	return nil
}
