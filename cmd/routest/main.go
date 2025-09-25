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
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/dispatch"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

type routestOptions struct {
	version    bool
	labels     string
	configFile string
}

type LabelSets map[string]model.LabelSet

func main() {
	var routestOpts routestOptions
	flag.BoolVar(&routestOpts.version, "version", false, "Print the version number.")
	flag.StringVar(&routestOpts.labels, "labels", "", "Comma separated labels in the form of: label=value. Example: -labels=env=dev,severity=critical")
	flag.StringVar(&routestOpts.configFile, "file", "", "Path to the Alertmanager configuration file.")
	flag.Parse()

	if routestOpts.version {
		fmt.Println(buildinfo.Get())
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}

	labelSets := make(LabelSets)
	if routestOpts.labels != "" {
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
	}

	var alertmanagerConfig string

	if routestOpts.configFile == "" {
		slog.Info("No config file provided, reading from stdin...")

		stat, err := os.Stdin.Stat()
		if err != nil {
			slog.Error("failed to stat stdin", "error", err)
			os.Exit(1)
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				slog.Error("failed to read from stdin", "error", err)
				os.Exit(1)
			}

			alertmanagerConfig = string(bytes)
		}
	} else {
		slog.Info("Reading config file", "path", routestOpts.configFile)

		if _, err := os.Stat(routestOpts.configFile); os.IsNotExist(err) || os.IsPermission(err) {
			slog.Error("config file does not exist or is not accessible", "path", routestOpts.configFile, "error", err)
			os.Exit(1)
		}

		bytes, err := os.ReadFile(routestOpts.configFile)
		if err != nil {
			slog.Error("failed to read config file", "path", routestOpts.configFile, "error", err)
			os.Exit(1)
		}

		alertmanagerConfig = string(bytes)
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
			if route.Continue {
				continue
			} else {
				break
			}
		}

		sort.Strings(results)
		for _, receiver := range results {
			slog.Info("Matches", "receiver", receiver)
		}
	}
}
