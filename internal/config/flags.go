package config

import (
	"flag"
	"fmt"
	"io"
)

func ApplyFlags(cfg *Config, args []string, errOut io.Writer) error {
	fs := flag.NewFlagSet("flow-server", flag.ContinueOnError)
	fs.SetOutput(errOut)

	addr := fs.String("addr", "", "HTTP listen address (overrides FLOW_HTTP_ADDR)")
	mode := fs.String("mode", "", "dev or prod (overrides FLOW_MODE)")
	url := fs.String("url", "", "public app origin, e.g. http://localhost:5173 (overrides FLOW_URL)")
	env := fs.String("env", "", "legacy alias for -mode / FLOW_ENV")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *addr != "" {
		cfg.HTTPAddr = *addr
	}
	if *mode != "" {
		if err := applyMode(cfg, *mode); err != nil {
			return fmt.Errorf("config: %w", err)
		}
	} else if *env != "" {
		if err := applyMode(cfg, *env); err != nil {
			return fmt.Errorf("config: %w", err)
		}
	}
	if *url != "" {
		if err := applyURL(cfg, *url); err != nil {
			return fmt.Errorf("config: %w", err)
		}
	}
	cfg.CORSOrigins = mergeOrigins(cfg.CORSOrigins, cfg.URL)
	if err := cfg.validate(); err != nil {
		return fmt.Errorf("config: %w", err)
	}
	return nil
}
