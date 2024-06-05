# routest

`routest` - Test your Alertmanager route configuration.

## Install

1. `make build`
2. `make install`

## Usage

Example (from `stdin` and Alertmanager configuration from Kubernetes):
```
kubectl get secrets -n kube-prometheus-stack alertmanager -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d | routest -labels="mylabel=myvalue,severity=critical" -

2024/06/04 23:21:40 Testing with labels: mylabel=myvalue,severity=critical
2024/06/04 23:21:40 Matches receiver: send_to_receiver
```
