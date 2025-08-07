# HTTP Hello World

This is a simple Rust Wasm example that responds with a "Hello World" message for each request.

## Prerequisites

- `cargo` 1.82
- [`wash`](https://wasmcloud.com/docs/installation) 0.36.1
- `wasmtime` >=25.0.0 (if running with wasmtime)

## Building

```bash
wash build
```

## Running with wasmCloud

```shell
wash dev
```

```shell
curl http://127.0.0.1:8000
```

## Adding Capabilities

To learn how to extend this example with additional capabilities, see the [Adding Capabilities](https://wasmcloud.com/docs/tour/adding-capabilities?lang=rust) section of the wasmCloud documentation.

## Now adding to a K8s cluster

```bash
helm upgrade --install \
    wasmcloud-platform \
    --values https://raw.githubusercontent.com/wasmCloud/wasmcloud/main/charts/wasmcloud-platform/values.yaml \
    oci://ghcr.io/wasmcloud/charts/wasmcloud-platform \
    --dependency-update

kubectl wait --for=condition=available --timeout=600s deploy -l app.kubernetes.io/name=wadm
kubectl wait --for=condition=available --timeout=600s deploy -l app.kubernetes.io/name=wasmcloud-operator
```

Next install a host:

```bash
helm upgrade --install \
    wasmcloud-platform \
    --values https://raw.githubusercontent.com/wasmCloud/wasmcloud/main/charts/wasmcloud-platform/values.yaml \
    oci://ghcr.io/wasmcloud/charts/wasmcloud-platform \
    --dependency-update \
    --set "hostConfig.enabled=true"
```

Then apply the rusty app:

```bash
kubectl apply -f wadm.yaml
```
