package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/mikejoh/routest/internal/buildinfo"
	"github.com/mikejoh/routest/internal/ui"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/dispatch"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

type routestOptions struct {
	version    bool
	labels     string
	configFile string
	ui         bool
	port       int
}

type LabelSets map[string]model.LabelSet

func main() {
	var routestOpts routestOptions
	flag.BoolVar(&routestOpts.version, "version", false, "Print the version number.")
	flag.StringVar(&routestOpts.labels, "labels", "", "Comma separated labels in the form of: label=value. Example: -labels=env=dev,severity=critical")
	flag.StringVar(&routestOpts.configFile, "file", "", "Path to the Alertmanager configuration file.")
	flag.BoolVar(&routestOpts.ui, "ui", false, "Launch interactive web UI for testing routes in the browser.")
	flag.IntVar(&routestOpts.port, "port", 0, "Port for the web UI (default: random available port).")
	flag.Parse()

	if routestOpts.version {
		fmt.Println(buildinfo.Get())
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	alertmanagerConfig, err := readConfig(routestOpts.configFile)
	if err != nil {
		slog.Error("failed to read config", "error", err)
		os.Exit(1)
	}
	if alertmanagerConfig == "" {
		slog.Error("alertmanager config must be provided either via file or stdin")
		os.Exit(1)
	}

	var c config.Config
	if err := yaml.Unmarshal([]byte(alertmanagerConfig), &c); err != nil {
		slog.Error("failed to unmarshal alertmanager config", "error", err)
		os.Exit(1)
	}

	if c.Route == nil {
		slog.Error("alertmanager config is missing a route definition")
		os.Exit(1)
	}

	if routestOpts.ui {
		srv, err := ui.NewServer(&c, routestOpts.port)
		if err != nil {
			slog.Error("failed to create UI server", "error", err)
			os.Exit(1)
		}
		slog.Info("Opening UI in browser", "url", srv.URL())
		ui.OpenBrowser(srv.URL())
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("UI server error", "error", err)
			os.Exit(1)
		}
		return
	}

	if routestOpts.labels == "" {
		slog.Error("no labels provided, use -labels flag (or -ui for interactive mode)")
		os.Exit(1)
	}

	labelSets := make(LabelSets)
	labels := strings.Split(routestOpts.labels, ",")
	labelSet := make(model.LabelSet)
	for _, label := range labels {
		labelParts := strings.Split(label, "=")
		if len(labelParts) != 2 {
			slog.Error("invalid label format", "label", label)
			os.Exit(1)
		}
		labelSet[model.LabelName(labelParts[0])] = model.LabelValue(labelParts[1])
	}
	labelSets["custom"] = labelSet

	for _, labelSet := range labelSets {
		slog.Info("Testing with labels", "labels", labelSet)

		if err := labelSet.Validate(); err != nil {
			slog.Error("label set validation failed", "error", err)
			os.Exit(1)
		}

		routeTree := dispatch.NewRoute(c.Route, nil)
		routes := routeTree.Match(labelSet)

		var results []string
		for _, route := range routes {
			results = append(results, route.RouteOpts.Receiver)
			if !route.Continue {
				break
			}
		}

		sort.Strings(results)
		for _, receiver := range results {
			slog.Info("Matches", "receiver", receiver)
		}
	}
}

func readConfig(file string) (string, error) {
	if file == "" {
		slog.Info("No config file provided, reading from stdin...")
		stat, err := os.Stdin.Stat()
		if err != nil {
			return "", fmt.Errorf("stat stdin: %w", err)
		}
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return "", fmt.Errorf("read stdin: %w", err)
			}
			return string(b), nil
		}
		return "", nil
	}

	slog.Info("Reading config file", "path", file)
	if _, err := os.Stat(file); err != nil {
		return "", fmt.Errorf("config file not accessible: %w", err)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("read config file: %w", err)
	}
	return string(b), nil
}
