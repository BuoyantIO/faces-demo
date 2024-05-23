# Edge Stack and Linkerd Resilience Patterns

This is the documentation - and executable code! - for a demo of resilience
patterns using Edge Stack and Linkerd. The easiest way to use this file
is to execute it with [demosh].

Things in Markdown comments are safe to ignore when reading this later. When
executing this with [demosh], things after the horizontal rule below (which
is just before a commented `@SHOW` directive) will get displayed.

[demosh]: https://github.com/BuoyantIO/demosh

When you use `demosh` to run this file, your cluster will be checked for you.

<!-- set -e >
<!-- @import demosh/demo-tools.sh -->
<!-- @import demosh/check-requirements.sh -->

<!-- @start_livecast -->

---
<!-- @SHOW -->

# Edge Stack and Linkerd Resilience Patterns

## Rate Limits, Retries, and Timeouts

We're going to show various resilience techniques using the Faces demo (from
<https://github.com/BuoyantIO/faces-demo>):

- _Retries_ automatically repeat requests that fail;
- _Timeouts_ cut off requests that take too long; and
- _Rate limits_ protect services by restricting the amount of traffic that can
  flow through to a service.

All are important techniques for resilience, and all can be applied - at
various points in the call stack - by infrastructure components like the
ingress controller and/or service mesh.

<!-- @wait_clear -->

## Installing Linkerd

We're going to install Linkerd first -- that lets us install Edge Stack and
Faces directly into the mesh, rather than installing and then meshing as a
separate step.

### A digression on Linkerd releases

There are two kinds of Linkerd releases: _edge_ and _stable_. The Linkerd
project itself only produces edge releases, which show up every week or so and
always have the latest and greatest features and fixes directly from the
`main` branch. Stable releases are produced by the vendor community around
Linkerd, and are the way to go for full support.

For this demo, we'll use the stable Buoyant Enterprise for Linkerd (BEL)
distribution. You'll need a free Buoyant account for this -- if you don't
already have one, hit up <https://enterprise.buoyant.io> to get set up.

<!-- @wait_clear -->

### Installing the CLI

Once you have your Buoyant account set up, and you've set the environment
variables it'll tell you about, it's time to get the Linkerd CLI installed!
The CLI makes it easy to work with Linkerd, and it's installed with this
one-liner that will download the latest BEL CLI and get it set up to run:

```bash
curl --proto '=https' --tlsv1.2 -sSfL https://enterprise.buoyant.io/install | sh
linkerd version --proxy --client --short
```

Once that's done, you'll need to add the CLI directory to your PATH:

```bash
export PATH=$PATH:$HOME/.linkerd2/bin
```

and then we can make sure that this cluster really can run Linkerd:

```bash
linkerd check --pre
```

<!-- @wait_clear -->

## Installing the Linkerd CRDs

Linkerd uses Custom Resource Definitions (CRDs) to extend Kubernetes. After
verifying that the cluster is ready to run Linkerd, we next need to install
the CRDs. We do this by running `linkerd install --crds`, which will output
the CRDs that need to be installed so that we can apply them to the cluster.
(The Linkerd CLI will never directly modify the cluster.)

```bash
linkerd install --crds | kubectl apply -f -
```

As you can see in the output above, Linkerd doesn't actually install many
CRDs, and in fact it can add security and observability to an application
without using _any_ of these CRDs. However, they're necessary for more
advanced usage.

<!-- @wait_clear -->

## Installing Linkerd and Linkerd Viz

Now that the CRDs are installed, we can install Linkerd itself.

```bash
linkerd install | kubectl apply -f -
```

We're also going to install Linkerd Viz: this is an optional component that
provides a web-based dashboard for Linkerd. It's a great way to see what's
happening in your cluster, so we'll install it as well.

(For the moment, we need to tell the CLI to explicitly install Viz 2.14.10.)

```bash
linkerd viz install --set linkerdVersion=stable-2.14.10 | kubectl apply -f -
```

Just like Linkerd itself, this will start the installation and return
immediately, so - again - we'll use `linkerd check` to make sure all is well.

```bash
linkerd check
```

So far so good -- let's take a look at the Viz dashboard just to make sure.

```bash
linkerd viz dashboard
```

<!-- @clear -->

## Installing Edge Stack

At this point, Linkerd is up and running, so we'll continue by installing
Edge Stack, which works pretty much the same way as Linkerd: we install
Edge Stack's CRDs first, then we install Edge Stack itself.

We want Edge Stack to be in the Linkerd mesh from the start, so we'll begin by
creating Edge Stack's namespace and annotating it such that any new Pods in that
namespace will automatically be injected with the Linkerd proxy.

```bash
kubectl create namespace ambassador
kubectl annotate namespace ambassador linkerd.io/inject=enabled
```

After that, we can install Edge Stack's CRDs.

```bash
kubectl apply -f https://app.getambassador.io/yaml/edge-stack/3.10.2/aes-crds.yaml && \
kubectl wait --timeout=90s --for=condition=available deployment emissary-apiext -n emissary-system
```

Once that's done, we can install Edge Stack itself. We'll deliberately run just
a single replica (this makes things simpler if you're running a local
cluster!), and we'll wait for Edge Stack to be running before continuing.

```bash
helm repo add datawire https://app.getambassador.io && \
helm repo update && \
helm install edge-stack --namespace ambassador datawire/edge-stack \
  --set emissary-ingress.replicaCount=1 \
  --set emissary-ingress.agent.cloudConnectToken=$CLOUD_CONNECT_TOKEN && \
kubectl -n ambassador wait --for condition=available --timeout=90s deploy -l product=aes
```

With this, Edge Stack is running -- but it needs some configuration to be
useful.

<!-- @wait_clear -->

## Configuring Edge Stack

First things first: let's tell Edge Stack which ports and protocols we want to
use. Specifically, we'll tell it to listen for HTTP on port 8080 and 8443, and
to accept any hostname. This is not great for production, but it's fine for
us.

```bash
bat aes-yaml/listeners.yaml
kubectl apply -f aes-yaml/listeners.yaml
```

Next up, we need to set up rate limiting. For every request, Edge Stack checks if
rate limiting should be applied based on our ratelimit configuration.

```bash
bat aes-yaml/ratelimit.yaml
kubectl apply -f aes-yaml/ratelimit.yaml
```

Finally, we want Edge Stack to give us access to the Linkerd Viz dashboard.

```bash
bat aes-yaml/linkerd-viz-mapping.yaml
kubectl apply -f aes-yaml/linkerd-viz-mapping.yaml
```

With that, Edge Stack should be good to go! We can test it by going to check out
the Linkerd Viz dashboard again _without_ using the `linkerd viz dashboard`
command -- just going to the IP address of the `Edge Stack` service from a
browser should load up the dashboard.

<!-- @browser_then_terminal -->

## Installing Faces

Finally, let's install Faces! This is pretty simple: we'll create and annotate
the namespace as before, then use Helm to install Faces:

```bash
kubectl create namespace faces
kubectl annotate namespace faces linkerd.io/inject=enabled

helm install faces -n faces \
     oci://ghcr.io/buoyantio/faces-chart --version 1.3.0

kubectl rollout status -n faces deploy
```

We'll also install basic Mappings and ServiceProfiles for the Faces workloads:

```bash
bat k8s/01-base/*-mapping.yaml
bat k8s/01-base/*-profile.yaml
kubectl apply -f k8s/01-base
```

And with that, let's take a quick look at Faces in the web browser. You'll be
able to see that it's in pretty sorry shape, and you'll be able to look at the
Linkerd dashboard to see how much traffic it generates.

<!-- @browser_then_terminal -->

## RETRIES

Let's start by going after the red frowning faces: those are the ones where
the face service itself is failing. We can tell Edge Stack to retry those when
they fail, by adding a `retry_policy` to the Mapping for `/face/`:

```bash
diff -u99 --color k8s/{01-base,02-retries}/face-mapping.yaml
```

We'll apply those...

```bash
kubectl apply -f k8s/02-retries/face-mapping.yaml
```

...then go take a look at the results in the browser.

<!-- @browser_then_terminal -->

## RETRIES continued

So that helped quite a bit: it's not perfect, because Edge Stack will only retry
once, but it definitely cuts down on problems! Let's continue by adding a
retry for the smiley service, too, to try to get rid of the cursing faces:

```bash
diff -u99 --color k8s/{01-base,02-retries}/smiley-mapping.yaml
```

Let's apply those and go take a look in the browser.

```bash
kubectl apply -f k8s/02-retries/smiley-mapping.yaml
```

<!-- @browser_then_terminal -->

## RETRIES continued

That... had no effect. If we take a look back at the overall application
diagram, the reason is clear...

<!-- @wait -->
<!-- @show_slides -->
<!-- @wait -->
<!-- @clear -->
<!-- @show_terminal -->

...Edge Stack never talks to the smiley service! so telling Edge Stack to retry
the failed call will never work.

Instead, we need to tell Linkerd to do the retries, by adding `isRetryable` to
the `ServiceProfile` for the smiley service:

```bash
diff -u99 --color k8s/{01-base,02-retries}/smiley-profile.yaml
```

This is different from the Edge Stack config because Linkerd uses a _retry
budget_ instead of a counter: as long as the total number of retries doesn't
exceed the budget, Linkerd will just keep retrying. Let's apply that and take
a look.

```bash
kubectl apply -f k8s/02-retries/smiley-profile.yaml
```

<!-- @browser_then_terminal -->

## RETRIES continued

That works great. Let's do the same for the color service.

```bash
diff -u99 --color k8s/{01-base,02-retries}/color-profile.yaml
kubectl apply -f k8s/02-retries/color-profile.yaml
```

And, again, back to the browser to check it out.

<!-- @browser_then_terminal -->

## RETRIES continued

Finally, let's go back to the browser to take a look at the load on the
services now. Retries actually _increase_ the load on the services, since they
cause more requests: they're not about protecting the service, they're about
**improving the experience of the client**.

<!-- @browser_then_terminal -->

## TIMEOUTS

Things are a lot better already! but... still too slow, which we can see as
those cells that are fading away. Let's add some timeouts, starting from the
bottom of the call graph this time.

Again, timeouts are not about protecting the service: they are about
**providing agency to the client** by giving the client a chance to decide
what to do when things take too long. In fact, like retries, they _increase_
the load on the service.

We'll start by adding a timeout to the color service. This timeout will give
agency to the face service, as the client of the color service: when a call to
the color service takes too long, the face service will show a pink background
for that cell.

```bash
diff -u99 --color k8s/{02-retries,03-timeouts}/color-profile.yaml
```

Let's apply that and then switch back to the browser to see what's up.

```bash
kubectl apply -f k8s/03-timeouts/color-profile.yaml
```

<!-- @browser_then_terminal -->

## TIMEOUTS continued

Let's continue by adding a timout to the smiley service. The face service
will show a smiley-service timeout as a sleeping face.

```bash
diff -u99 --color k8s/{02-retries,03-timeouts}/smiley-profile.yaml
kubectl apply -f k8s/03-timeouts/smiley-profile.yaml
```

<!-- @browser_then_terminal -->

## TIMEOUTS continued

Finally, we'll add a timeout that lets the GUI decide what to do if the face
service itself takes too long. We'll use Edge Stack for this (although we
could've used Linkerd, since Edge Stack is itself in the mesh).

When the GUI sees a timeout talking to the face service, it will just keep
showing the user the old data for awhile. There are a lot of applications
where this makes an enormous amount of sense: if you can't get updated data,
the most recent data may still be valuable for some time! Eventually, though,
the app should really show the user that something is wrong: in our GUI,
repeated timeouts eventually lead to a faded sleeping-face cell with a pink
background.

For the moment, too, the GUI will show a counter of timed-out attempts, to
make it a little more clear what's going on.

```bash
diff -u99 --color k8s/{02-retries,03-timeouts}/face-mapping.yaml
kubectl apply -f k8s/03-timeouts/face-mapping.yaml
```

<!-- @browser_then_terminal -->

## RATELIMITS

Given retries and timeouts, things look better -- still far from perfect, but
better. Suppose, though, that someone now adds some code to the face service
that makes it just completely collapse under heavy load? Sadly, this is often
all-too-easy to mistakenly do.

Let's simulate this. The face service has internal functionality to limit its
abilities under load when we set the `MAX_RATE` environment variable, so we'll
do that now:

```bash
kubectl set env deploy -n faces face MAX_RATE=9.0
```

Once that's done, we can take a look in the browser to see what happens.

<!-- @browser_then_terminal -->

## RATELIMITS continued

Since the face service is right on the edge, we can have Edge Stack enforce a
rate limit on requests to the face service. This is both protecting the
service (by reducing the traffic) **and** providing agency to the client (by
providing a specific status code when the limit is hit). Here, our web app is
going to handle rate limits just like it handles timeouts.

The most important thing about Edge Stack's rate limit configuration
here is to realize that we're actually providing a **label** on the requests,
and that the external rate limit service is counting traffic with that label
to decide what response to hand back.

```bash
diff -u99 --color k8s/{03-timeouts,04-ratelimits}/face-mapping.yaml
```

For this demo, our rate limit service is preconfigured to allow 8 requests per
second. Let's apply this and see how things look:

```bash
kubectl apply -f k8s/04-ratelimits/face-mapping.yaml
```

<!-- @browser_then_terminal -->

# SUMMARY

We've used both Edge Stack and Linkerd to take a very, very broken application
and turn it into something the user might actually have an OK experience with.
Fixing the application is, of course, still necessary!! but making the user
experience better is a good thing.

<!-- @wait -->
<!-- @show_slides -->
