```sh
git pull --tags --all
```

```sh
step certificate create "RITMS ROOT" certs/root-ca.crt certs/root-ca.key --kty RSA --profile root-ca --insecure --no-password
step certificate create "RITMS CA" certs/intermediate-ca.crt certs/intermediate-ca.key --kty RSA --profile intermediate-ca --ca certs/root-ca.crt --ca-key certs/root-ca.key --insecure --no-password
step certificate create console.svc.ritms.online certs/leaf.crt certs/leaf.key --san console.svc.ritms.online --san *.console.svc.ritms.online --kty RSA --profile leaf --ca certs/intermediate-ca.crt --ca-key certs/intermediate-ca.key --insecure --no-password --not-after 43830h
```

```sh
kubectl create namespace teleport-cluster
kubectl create configmap teleport-certs --from-file=certs/root-ca.crt --from-file=certs/intermediate-ca.crt --namespace teleport-cluster
kubectl create secret tls teleport-tls --key certs/leaf.key --cert certs/leaf.crt --namespace teleport-cluster
```

```sh
devspace deploy --namespace teleport-cluster
```

```sh
devspace enter --namespace teleport-cluster
cat >root.role<<EOF
kind: role
metadata:
  name: root
spec:
  allow:
    app_labels:
      '*': '*'
    kubernetes_groups:
      - '*'
    kubernetes_labels:
      '*': '*'
    logins:
      - root
    node_labels:
      '*': '*'
    rules:
      - resources:
          - '*'
        verbs:
          - '*'
    windows_desktop_logins:
      - Administrator
  options:
    client_idle_timeout: never
    disconnect_expired_cert: 'no'
    forward_agent: true
    max_connections: 100
    max_session_ttl: 24h
    max_sessions: 100
    permit_x11_forwarding: true
    port_forwarding: true
version: v4
EOF
tctl create root.role
tctl users add root --roles root
```

```sh
devspace purge --namespace teleport-cluster
kubectl delete namespace teleport-cluster
```
