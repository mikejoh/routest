# routest

`routest` - Test your Alertmanager route configuration! 🔔

## Install

### From source

1. `git clone https://github.com/mikejoh/routest.git`
2. `cd routest`
3. `make build` (places the compiled binary in `./build/`
4. `make install` (copies the compiled binary to `~/.local/bin`)

### Download and run

1. Download (using version `0.1.0` as an example):

```bash
curl -LO https://github.com/mikejoh/routest/releases/download/0.1.0/routest_0.1.0_linux_amd64.tar.gz
```

2. Unpack:

```bash
tar xzvf routest_0.1.3_linux_amd64.tar.gz
```

3. Run:

```bash
./routest -version
```

## Usage

From `stdin`:

```bash
kubectl get secrets -n kube-prometheus-stack alertmanager -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | routest -labels="mylabel=myvalue,severity=critical" -

2025/08/14 08:46:53 INFO No config file provided, reading from stdin...
2025/08/14 08:46:53 INFO Testing with labels labels=mylabel=myvalue,severity=critical
2025/08/14 08:46:53 INFO Matches receiver receiver=send_to_receiver
```

From a file:

```bash
routest -file "alertmanager.yaml" -labels="mylabel=myvalue,severity=critical"

2025/08/14 08:46:53 INFO Reading config file path=alertmanager.yaml
2025/08/14 08:46:53 INFO Testing with labels="{mylabel=\"myvalue\",severity=\"critical\"}"
2025/08/14 08:46:53 INFO Matches receiver=send_to_receiver
```
