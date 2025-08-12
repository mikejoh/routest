# routest

`routest` - Test your Alertmanager route configuration! 🔔

_This tool will only be able to read a config file from `stdin` at the moment._

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

Example (from `stdin` and Alertmanager configuration from Kubernetes):

```
kubectl get secrets -n kube-prometheus-stack alertmanager -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | routest -labels="mylabel=myvalue,severity=critical" -

2024/06/04 23:21:40 Testing with labels: mylabel=myvalue,severity=critical
2024/06/04 23:21:40 Matches receiver: send_to_receiver
```

## Todo

* Add support for reading from a file.
* Add support for reading from a URL.
