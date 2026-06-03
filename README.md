# routest

`routest` - Test your Alertmanager route configuration! 🔔

## Install

### Homebrew (macOS / Linux)

```bash
brew install mikejoh/routest/routest
```

Binaries are published automatically to the [mikejoh/homebrew-routest](https://github.com/mikejoh/homebrew-routest) tap on every tagged release via GoReleaser.

### go install

```bash
go install github.com/mikejoh/routest/cmd/routest@latest
```

### Prebuilt binaries

Archives for Linux, macOS, and Windows are attached to every [GitHub release](https://github.com/mikejoh/routest/releases/latest) automatically by GoReleaser. Download the archive for your platform, unpack, and place the binary on your `PATH`.

### From source

```bash
git clone https://github.com/mikejoh/routest.git
cd routest
make build    # → ./build/routest
make install  # → ~/.local/bin/routest
```

## Usage

### Web UI

Launch an interactive canvas-based flow diagram in your browser. Add alert labels and click **Test Route** to see which receivers match, with animated arrows tracing the full path from alert through receivers to integration types (Slack, PagerDuty, Webhook, etc.).

From a file:

```bash
routest -file alertmanager.yaml -ui
```

From `stdin`:

```bash
kubectl get secrets -n kube-prometheus-stack alertmanager -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | routest -ui
```

Use `-port` to bind to a specific port instead of a random one:

```bash
routest -file alertmanager.yaml -ui -port 8080
```

### CLI

Test a fixed set of labels and print the matching receiver(s) to stdout.

From a file:

```bash
routest -file "alertmanager.yaml" -labels="mylabel=myvalue,severity=critical"

2025/08/14 08:46:53 INFO Reading config file path=alertmanager.yaml
2025/08/14 08:46:53 INFO Testing with labels="{mylabel=\"myvalue\",severity=\"critical\"}"
2025/08/14 08:46:53 INFO Matches receiver=send_to_receiver
```

From `stdin`:

```bash
kubectl get secrets -n kube-prometheus-stack alertmanager -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | routest -labels="mylabel=myvalue,severity=critical"

2025/08/14 08:46:53 INFO No config file provided, reading from stdin...
2025/08/14 08:46:53 INFO Testing with labels labels=mylabel=myvalue,severity=critical
2025/08/14 08:46:53 INFO Matches receiver receiver=send_to_receiver
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-file` | — | Path to the Alertmanager configuration file (reads stdin if omitted) |
| `-labels` | — | Comma-separated label pairs to test, e.g. `severity=critical,env=prod` |
| `-ui` | `false` | Launch interactive web UI in the browser |
| `-port` | random | Port for the web UI server |
| `-version` | — | Print version and exit |
