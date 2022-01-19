```sh
git pull --tags --all
```

```sh
step certificate create "RITMS ROOT" root-ca.crt root-ca.key --kty RSA --profile root-ca --insecure --no-password
step certificate create "RITMS CA" intermediate-ca.crt intermediate-ca.key --kty RSA --profile intermediate-ca --ca root-ca.crt --ca-key root-ca.key --insecure --no-password
step certificate create console.svc.ritms.online leaf.crt leaf.key --san *.console.svc.ritms.online --kty RSA --profile leaf --ca intermediate-ca.crt --ca-key intermediate-ca.key --insecure --no-password
```

```sh
kubectl create secret tls teleport-ca --key intermediate-ca.key --cert intermediate-ca.crt --namespace teleport-cluster
kubectl create secret tls teleport-tls --key leaf.key --cert leaf.crt --namespace teleport-cluster
```

```sh
devspace deploy --namespace teleport-cluster
```
