package main

import (
	"flag"
	"fmt"
	"io"
	"log"
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
	version bool
	labels  string
}

type LabelSets map[string]model.LabelSet

func main() {
	var routestOpts routestOptions
	flag.BoolVar(&routestOpts.version, "version", false, "Print the version number.")
	flag.StringVar(&routestOpts.labels, "labels", "", "Comma separated labels in the form of: label=value. Example: -label-set=env=dev,severity=critical")
	flag.Parse()

	if routestOpts.version {
		fmt.Println(buildinfo.Get())
		os.Exit(0)
	}

	labelSets := make(LabelSets)
	if routestOpts.labels != "" {
		labels := strings.Split(routestOpts.labels, ",")
		labelSet := make(model.LabelSet)
		for _, label := range labels {
			labelParts := strings.Split(label, "=")
			if len(labelParts) != 2 {
				log.Fatalf("invalid label format: %s", label)
			}

			labelSet[model.LabelName(labelParts[0])] = model.LabelValue(labelParts[1])
		}

		labelSets["custom"] = labelSet
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}

	var alertmanagerConfig string
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatalf("failed to read from stdin: %s", err)
		}

		alertmanagerConfig = string(bytes)
	}

	var c config.Config
	if err := yaml.Unmarshal([]byte(alertmanagerConfig), &c); err != nil {
		log.Fatal(err)
	}

	for _, labelSet := range labelSets {
		log.Printf("Testing with labels: %s", routestOpts.labels)

		if err := labelSet.Validate(); err != nil {
			log.Fatal(err)
		}

		routeTree := dispatch.NewRoute(c.Route, nil)
		routes := routeTree.Match(labelSet)

		results := []string{}
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
			log.Printf("Matches receiver: %s", receiver)
		}
	}
}
