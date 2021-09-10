## Simple VPS

1) Copy binary from releases to a VPS and make it executable.

2) Create a token:
```bash
head -c 16 /dev/urandom | shasum | cut -d" " -f1
```

3) Create a service:
`/etc/systemd/system/inlets.service`
```bash
[Unit]
Description=Inlets Server Service
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=1
StartLimitInterval=0
EnvironmentFile=/etc/default/inlets
ExecStart=/usr/local/bin/inlets server --port=8080 --token="<mytoken>"

[Install]
WantedBy=multi-user.target
```

4) Start the service:
```bash
systemctl daemon-reload
systemctl start inlets
systemctl enable inlets

systemctl status inlets
```

5) Use [Caddy](https://caddyserver.com/docs/install#debian-ubuntu-raspbian) to provide a revere proxy with Certificate for the exit node.
Always prefix your client websocket connection with `--url wss://` instead of `--url ws://` for a secure tunnel!!
`/etc/caddy/Caddyfile`
```bash
dev.example.com
reverse_proxy 127.0.0.1:8080
```

6) Restart Caddy:
```bash
systemctl restart caddy
systemctl status caddy
```

7) Sample for a Kubernetes/Docker-Client:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: inlets
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: inlets
  template:
    metadata:
      labels:
        app: inlets
    spec:
      containers:
      - name: inlets
        image: ghcr.io/cubed-it/inlets:latest
        imagePullPolicy: IfNotPresent
        command: ["inlets"]
        args:
        - "client"
        - "--upstream=dev.example.com=http://nginx-ingress-ingress-nginx-controller.kube-system:80"
        - "--remote=wss://dev.example.com"
        - "--token=<mytoken>"
```