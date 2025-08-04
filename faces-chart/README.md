# Faces Helm Chart

**Faces** is a deliberately broken demo application designed to help you explore and debug the kinds of reliability issues that can happen in real-world microservices — including latency, failures, misconfigurations, and poor observability. It’s a great way to learn how a service mesh like Linkerd can help!

This chart deploys a full instance of the Faces app, composed of multiple services communicating over HTTP and gRPC, and lets you tune behavior via Helm values to simulate different failure scenarios.

---

## Installation

The Faces chart is published as a Helm OCI package. To install it into a Linkerd-injected namespace:

```sh
kubectl create namespace faces
kubectl annotate namespace faces linkerd.io/inject=enabled

helm install faces -n faces \
  oci://ghcr.io/buoyantio/faces-chart --version 2.0.0

kubectl rollout status -n faces deploy
```

To access the GUI:

```sh
kubectl port-forward -n faces svc/faces-gui 8080:80
open http://localhost:8080
```

---

## Configuration

You can customize the demo using Helm `--set` flags or by creating a `values.yaml` file.

### Common Options

| Parameter                | Description                        | Default                     |
| ------------------------ | ---------------------------------- | --------------------------- |
| `defaultImageTag`        | Global default image tag           | Matches `.Chart.AppVersion` |
| `defaultImagePullPolicy` | Global pull policy                 | `IfNotPresent`              |
| `defaultReplicas`        | Default number of pod replicas     | `1`                         |
| `authHeader`             | HTTP header name for user identity | `X-Faces-User`              |

### GUI

| Parameter                | Description                                            | Default                       |
| ------------------------ | ------------------------------------------------------ | ----------------------------- |
| `gui.image`              | GUI image                                              | ` `                           |
| `gui.imageName`          | GUI image name                                         | `ghcr.io/buoyantio/faces-gui` |
| `gui.imageTag`           | GUI image tag                                          | Uses `defaultImageTag`        |
| `gui.imagePullPolicy`    | GUI pull policy                                        | `IfNotPresent`       |
| `gui.replicas`           | Replica count                                          | `defaultReplicas`             |
| `gui.serviceType`        | Service type (`ClusterIP`, `NodePort`, `LoadBalancer`) | `ClusterIP`                   |

### Face

| Parameter            | Description                        | Default                        |
| -------------------- | ---------------------------------- | ------------------------------ |
| `face.image`         | Face backend image                     | ` `                            |
| `face.imageName`     | Face backend image name                | `ghcr.io/buoyantio/faces-face` |
| `face.imagePullPolicy`     | Face pull policy                | `IfNotPresent` |
| `face.errorFraction` | % of requests that fail (0–100)    | `20`                           |
| `face.delayBuckets`  | Comma-separated delay values in ms | *not set*                      |
| `face.smileyService` | Name of smiley service to call     | `smiley`                       |
| `face.colorService`  | Name of color service to call      | `color`                        |

### Ingress

| Parameter                  | Description                        | Default                        |
| -------------------------- | ---------------------------------- | ------------------------------ |
| `ingress.enabled`          | Enabled Ingress workload           | `False`                        |
| `ingress.image`            | Ingress image                      | ` `                            |
| `ingress.imageName`        | Ingress image name                 | `ghcr.io/buoyantio/faces-ingress` |
| `ingress.imageTag`         | Ingress tag name                   | ` `                            |
| `ingress.imagePullPolicy`  | Ingress pull policy                | `IfNotPresent`                 |
| `ingress.cellService`      | Name of cell service to call       | `cell`                         |

### Backend

| Parameter               | Description                             | Default                        |
| ----------------------- | --------------------------------------- | ------------------------------ |
| `backend.errorFraction` | Default error rate for backend services | `20`                           |
| `backend.delayBuckets`  | Default delays in ms                    | `0,5,10,15,20,50,200,500,1500` |

**`delayBuckets`** lets you simulate random latency by specifying a list of delays (in milliseconds). On each request, the app randomly picks one of the values and pauses for that duration before responding. This helps test how your system handles slow or delayed services.

### Smiley / Color Variants

#### Smiley

| Key                                | Description                                          | Default                 |
| ---------------------------------- | ---------------------------------------------------- | ----------------------- |
| `smiley.enabled`                   | Whether to deploy this workload                      | `true`                  |
| `smiley.smiley`                    | Emoji name to return                                 | `Grinning`              |
| `smiley.imageName`                 | Image (defaults to `ghcr.io/buoyantio/faces-smiley`) |                         |
| `smiley.imageTag`                  | Tag for image                                        | *optional*              |
| `smiley.errorFraction`             | Failure percentage                                   | `backend.errorFraction` |
| `smiley.delayBuckets`              | Delay buckets                                        | `backend.delayBuckets`  |

#### Color

| Key                               | Description                                         | Default                 |
| --------------------------------- | --------------------------------------------------- | ----------------------- |
| `color.enabled`                   | Whether to deploy this workload                     | `true`                  |
| `color.color`                     | Name of the color to return (e.g. `blue`)           | `lightblue`             |
| `color.imageName`                 | Image (defaults to `ghcr.io/buoyantio/faces-color`) |                         |
| `color.imageTag`                  | Tag for image                                       | *optional*              |
| `color.errorFraction`             | Failure percentage                                  | `backend.errorFraction` |
| `color.delayBuckets`              | Delay buckets                                       | `backend.delayBuckets`  |

You can enable up to three smiley and color services:

| Parameter         | Description                        | Example         |
| ----------------- | ---------------------------------- | --------------- |
| `smiley.enabled`  | Deploy the main smiley backend     | `true`          |
| `smiley.smiley`   | Override default emoji             | `"RollingEyes"` |
| `smiley2.enabled` | Enable second smiley backend       | `true`          |
| `smiley3.enabled` | Enable third smiley backend        | `false`         |
| `color.enabled`   | Deploy the main color backend      | `true`          |
| `color.color`     | Color name or hex (e.g. `#00ff00`) | `"lightblue"`   |
| `color2.enabled`  | Enable second color backend        | `false`         |
| `color3.enabled`  | Enable third color backend         | `false`         |

### Available Colors and Smileys

You can customize the smiley face and background color used by each cell via Helm values or environment variables. Below are the predefined options supported by the Faces app.

#### Color Options

Use the `COLOR` environment variable or `color.color` Helm value to override the default. These named colors are designed to be distinguishable for users with various types of color vision.

| Name     |
| -------- |
| grey     |
| black    |
| white    |
| darkblue |
| blue     |
| green    |
| yellow   |
| red      |
| purple   |

You may also use any valid hex color code (e.g. `#073359` for a Buoyant Blue).

#### Smiley Options

Use the `SMILEY` environment variable or `smiley.smiley` Helm value to change the emoji.

| Name        | Emoji          |
| ----------- | -------------- |
| Grinning    | 😃 (`U+1F603`) |
| Sleeping    | 😴 (`U+1F634`) |
| Cursing     | 🤬 (`U+1F92C`) |
| Kaboom      | 🤯 (`U+1F92F`) |
| HeartEyes   | 😍 (`U+1F60D`) |
| Neutral     | 😐 (`U+1F610`) |
| RollingEyes | 🙄 (`U+1F644`) |
| Screaming   | 😱 (`U+1F631`) |
| Vomiting    | 🤮 (`U+1F92E`) |

These values are case-sensitive and must match exactly.

### Ingress

| Parameter             | Description                   | Default |
| --------------------- | ----------------------------- | ------- |
| `ingress.enabled`     | Deploy ingress proxy          | `false` |
| `ingress.cellService` | Target service for /face path | `cell`  |

---

## Example Custom Installation

Install Faces with a red background, heart-eyes smiley, and 50% error rate:

```sh
helm install faces -n faces \
  oci://ghcr.io/buoyantio/faces-chart --version 2.0.0 \
  --set color.color="#073359" \
  --set smiley.smiley="HeartEyes" \
  --set face.errorFraction=50
```

Enable second smiley and color services:

```sh
helm upgrade -i faces -n faces \
  oci://ghcr.io/buoyantio/faces-chart --version 2.0.0 \
  --set color.color="#073359" \
  --set smiley.smiley="HeartEyes" \
  --set face.errorFraction=50
  --set smiley2.enabled=true \
  --set smiley2.smiley="RollingEyes" \
  --set color2.enabled=true \
  --set color2.color="green"
```

Split smiley 50/50 between the two services:

```sh
cat <<EOF | kubectl -n faces apply -f -
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: smiley-route
  namespace: faces
spec:
  parentRefs:
    - kind: Service
      group: ""
      name: smiley
      port: 80
  rules:
    - backendRefs:
        - name: smiley2
          group: ""
          port: 80
          weight: 50
        - name: smiley
          group: ""
          port: 80
          weight: 50
EOF
```

Split the colors between two services:

```sh
cat <<EOF | kubectl -n faces apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: color-route
  namespace: faces
spec:
  parentRefs:
    - group: ""
      kind: Service
      name: color
      namespace: faces
      port: 80
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: color2
          namespace: faces
          port: 80
        - group: ""
          kind: Service
          name: color
          namespace: faces
          port: 80
EOF
```

Now put one color service on the edge, and one in the center:


```sh
cat <<EOF | kubectl -n faces apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: color-route
  namespace: faces
spec:
  parentRefs:
    - group: ""
      kind: Service
      name: color
      namespace: faces
      port: 80
  rules:
    - matches:
        - method:
            service: ColorService
            method: Center
      backendRefs:
        - group: ""
          kind: Service
          name: color
          namespace: faces
          port: 80
    - matches:
        - method:
            service: ColorService
            method: Edge
      backendRefs:
        - group: ""
          kind: Service
          name: color2
          namespace: faces
          port: 80
EOF
```

Now put one smiley service on the edge, and one in the center:

```sh
cat <<EOF | kubectl -n faces apply -f -
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: smiley-route
  namespace: faces
spec:
  parentRefs:
    - kind: Service
      group: ""
      name: smiley
      namespace: faces
      port: 80
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /edge
      backendRefs:
        - kind: Service
          group: ""
          name: smiley
          namespace: faces
          port: 80
    - matches:
        - path:
            type: PathPrefix
            value: /center
      backendRefs:
        - kind: Service
          group: ""
          name: smiley2
          namespace: faces
          port: 80
EOF
```

