package main

import (
	"bufio"
	"flag"
	"fmt"
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
		scanner := bufio.NewScanner(os.Stdin)
		var alertmanagerConfigBuilder strings.Builder

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				break
			}

			alertmanagerConfigBuilder.WriteString(line)
			alertmanagerConfigBuilder.WriteString("\n")
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		alertmanagerConfig = alertmanagerConfigBuilder.String()
	}

	var c config.Config
	if err := yaml.Unmarshal([]byte(alertmanagerConfig), &c); err != nil {
		log.Fatal(err)
	}

	for name, l := range labelSets {
		log.Printf("Testing: %s", name)

		if err := l.Validate(); err != nil {
			log.Fatal(err)
		}

		routeTree := dispatch.NewRoute(c.Route, nil)
		routes := routeTree.Match(l)

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
			log.Printf("Send to receiver: %s", receiver)
		}
	}
}
