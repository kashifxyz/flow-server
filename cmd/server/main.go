package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/cache"
	"github.com/kashifxyz/flow-server/internal/config"
	"github.com/kashifxyz/flow-server/internal/database"
	"github.com/kashifxyz/flow-server/internal/jobs"
	"github.com/kashifxyz/flow-server/internal/modules/health"
	"github.com/kashifxyz/flow-server/internal/realtime"
	"github.com/kashifxyz/flow-server/internal/routes"
	"github.com/kashifxyz/flow-server/pkg/logger"
	"github.com/kashifxyz/flow-server/pkg/mail"
	"github.com/kashifxyz/flow-server/pkg/storage"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		return err
	}
	if err := config.ApplyFlags(&cfg, args, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		os.Stderr.WriteString(err.Error() + "\n")
		return err
	}

	proxies, badProxies := parseTrustedProxies(cfg.TrustedProxies)
	auth.SetTrustedProxies(proxies)

	log := logger.New(cfg.Env, cfg.LogLevel)
	for _, bad := range badProxies {
		log.Warn().Str("value", bad).Msg("ignoring invalid FLOW_TRUSTED_PROXIES entry")
	}
	log.Info().
		Str("mode", cfg.Mode).
		Str("url", cfg.URL).
		Str("addr", cfg.HTTPAddr).
		Msg("server starting")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error().Err(err).Msg("database connect failed")
		return err
	}
	defer db.Close()
	log.Info().Msg("database connected")

	rdb, err := connectRedis(ctx, cfg, log)
	if err != nil {
		log.Error().Err(err).Msg("redis connect failed")
		return err
	}
	if rdb != nil {
		defer func() { _ = rdb.Close() }()
	}

	var sender mail.Sender = mail.LogSender{Log: log}
	if cfg.Mail.Enabled() {
		sender = mail.SMTPSender{Config: mail.SMTPConfig{
			Host:       cfg.Mail.SMTPHost,
			Port:       cfg.Mail.SMTPPort,
			User:       cfg.Mail.SMTPUser,
			Password:   cfg.Mail.SMTPPassword,
			From:       cfg.Mail.From,
			FromName:   cfg.Mail.FromName,
			Encryption: cfg.Mail.Encryption,
		}}
		log.Info().Str("from", cfg.Mail.From).Str("host", cfg.Mail.SMTPHost).Msg("mail smtp enabled")
	} else {
		log.Info().Msg("mail smtp disabled (set FLOW_MAIL_SMTP_PASSWORD)")
	}

	s3Client, err := storage.New(ctx, storage.Config{
		Endpoint:     cfg.S3.Endpoint,
		Region:       cfg.S3.Region,
		Bucket:       cfg.S3.Bucket,
		AccessKey:    cfg.S3.AccessKey,
		SecretKey:    cfg.S3.SecretKey,
		UsePathStyle: cfg.S3.UsePathStyle,
	})
	if err != nil {
		log.Error().Err(err).Msg("s3 client failed")
		return err
	}
	if s3Client != nil {
		if err := storage.Ping(ctx, s3Client, cfg.S3.Bucket); err != nil {
			log.Error().Err(err).Msg("s3 bucket check failed")
			return err
		}
		log.Info().Str("bucket", cfg.S3.Bucket).Str("endpoint", cfg.S3.Endpoint).Msg("s3 connected")
	} else {
		log.Info().Msg("s3 disabled (set FLOW_S3_SECRET_KEY)")
	}

	authSvc, err := auth.NewService(db, sender, cfg.URL, log)
	if err != nil {
		log.Error().Err(err).Msg("auth service failed")
		return err
	}
	sessions := &auth.Sessions{Service: authSvc, Log: log}
	hub := realtime.NewHub()

	handler := routes.New(routes.Deps{
		Config:        cfg,
		Log:           log,
		DB:            db,
		S3:            s3Client,
		Redis:         rdb,
		Hub:           hub,
		Sessions:      sessions,
		AuthService:   authSvc,
		Mail:          sender,
		PublicURL:     cfg.URL,
		SecureCookies: cfg.IsProd() || strings.HasPrefix(cfg.URL, "https://"),
		Health: health.Module{
			DB:    db,
			Redis: rdb,
			Log:   log,
		},
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	go jobs.Scheduler{
		Log:    log,
		Worker: jobs.Worker{Log: log},
	}.Run(ctx)

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("http listening")
		errCh <- srv.ListenAndServe()
	}()

	var httpErr error
	select {
	case <-ctx.Done():
		log.Info().Msg("shutdown signal received")
	case httpErr = <-errCh:
	}

	stop()
	hub.Shutdown() // WebSocket connections are hijacked; http.Server.Shutdown never sees them
	shutCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownWait)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("http shutdown failed")
		_ = srv.Close()
		return err
	}
	if httpErr == nil {
		select {
		case httpErr = <-errCh:
		case <-shutCtx.Done():
		}
	}
	if httpErr != nil && !errors.Is(httpErr, http.ErrServerClosed) {
		log.Error().Err(httpErr).Msg("http server failed")
		return httpErr
	}
	log.Info().Msg("server stopped")
	return nil
}

// parseTrustedProxies turns FLOW_TRUSTED_PROXIES entries (CIDRs, or bare IPs
// treated as /32 or /128) into prefixes auth.clientIP will trust
// X-Forwarded-For from. Entries that parse as neither are reported back so
// the caller can log them, rather than silently trusting nothing or crashing
// on a config typo.
func parseTrustedProxies(raw []string) (prefixes []netip.Prefix, invalid []string) {
	for _, r := range raw {
		if p, err := netip.ParsePrefix(r); err == nil {
			prefixes = append(prefixes, p)
			continue
		}
		if addr, err := netip.ParseAddr(r); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		invalid = append(invalid, r)
	}
	return prefixes, invalid
}

func connectRedis(ctx context.Context, cfg config.Config, log zerolog.Logger) (*redis.Client, error) {
	if !cfg.RedisEnabled() {
		log.Info().Msg("redis disabled")
		return nil, nil
	}
	client, err := cache.New(ctx, cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("redis connected")
	return client, nil
}
