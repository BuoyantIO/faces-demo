# Emissary and Linkerd Resilience Patterns

This is the documentation - and executable code! - for a demo of resilience
patterns using Emissary-ingress and Linkerd. The easiest way to use this file
is to execute it with [demosh].

Things in Markdown comments are safe to ignore when reading this later. When
executing this with [demosh], things after the horizontal rule below (which
is just before a commented `@SHOW` directive) will get displayed.

[demosh]: https://github.com/BuoyantIO/demosh

When you use `demosh` to run this file, your cluster will be checked for you.

<!-- set -e >
<!-- @import ../demosh/demo-tools.sh -->
<!-- @import ../demosh/check-requirements.sh -->

<!-- @start_livecast -->

---
<!-- @SHOW -->

# Linkerd Resilience Patterns

We're going to show a couple of simple Linkerd resilience techniques using the
Faces demo (from https://github.com/BuoyantIO/faces-demo):

- _Retries_ automatically repeat requests that fail; and
- _Timeouts_ cut off requests that take too long.

Though both are very simple, they can be extremely effective.

Let's start with a quick look at Faces in the web browser. You'll be able to
see that it's in pretty sorry shape immediately.

<!-- @wait -->
<!-- @show_4 -->
<!-- @wait -->
<!-- @clear -->
<!-- @show_composite -->

By eye, we can see that this is not a well-behaving application, but we don't
have a good way to see where the failures are, or what exactly is failing. If
we look at the Linkerd Viz dashboard at this point, the best it can do is to
tell us that we have workloads running, but that it can't tell us anything
about them.

<!-- @browser_then_composite -->

## Enter the Mesh

Let's bring Linkerd into play so that we can start answering these questions.

Our browser talks to Emissary, which in turn talks to the Faces app itself, so
let's get both Emissary and Faces into the mesh. We'll start with Emissary.

<!-- @wait -->

First we'll annotate the emissary namespace to tell Linkerd to include it...

```bash
kubectl annotate ns emissary linkerd.io/inject=enabled
```

...then, we need to restart Emissary's deployments.

```bash
kubectl rollout restart -n emissary deployment
kubectl rollout status -n emissary deployment
```

At this point, a look at the Linkerd Viz dashboard will show that Emissary's
pods are meshed.

<!-- @browser_then_composite -->

Next, we'll repeat that process for the Faces namespace.

```bash
kubectl annotate ns faces linkerd.io/inject=enabled
kubectl rollout restart -n faces deployment
kubectl rollout status -n faces deployment
```

If we look at Linkerd Viz now, it can immediately show us that the Faces
application is now meshed. Given a few seconds, it'll be able to give us some
hard data about what's going wrong, and about how bad things are.

<!-- @browser_then_composite -->

## Retries

Let's start by going after the red frowning faces: those are the ones where
the face service itself is failing. We'll tell Linkerd that it's OK to retry
connections to the face service, by adding a ServiceProfile resource:

```bash
bat k8s/02-retries/face-profile.yaml
```

Linkerd uses a _retry budget_: it computes how much of the total traffic it's
sending to a workload is retries, and as long as that fraction doesn't exceed
the retry budget (20% by default), Linkerd will just keep retrying. Let's
apply that and see what happens.

```bash
kubectl apply -f k8s/02-retries/face-profile.yaml
```

<!-- @wait_clear -->

So that helped quite a bit! Let's continue by adding a retry for the smiley
service, too, to try to get rid of the cursing faces:

```bash
bat k8s/02-retries/smiley-profile.yaml
kubectl apply -f k8s/02-retries/smiley-profile.yaml
```

<!-- @wait_clear -->

We can do the same for the color service, to get rid of the grey backgrounds.

```bash
bat k8s/02-retries/color-profile.yaml
kubectl apply -f k8s/02-retries/color-profile.yaml
```

<!-- @wait_clear -->

Note that if we head back to the Viz dashboard, we'll still see the real
success rate, _without_ factoring in the effect of the retries.

<!-- @browser_then_composite -->

If we really want to see the effective success rate, we can use the `linkerd
viz routes` command. Here, we'll check the difference between the real and
effective success rates for traffic from the `face` Deployment to the `smiley`
Deployment:

```bash
linkerd viz routes -o wide -n faces deploy/face --to deploy/smiley
```

<!-- @show_terminal -->

Note that we see both the real and effective success rates, _and_ the real and
effective request rate. You can clearly see that retries actually _increase_
the load on the services, since they cause more requests: they're not about
protecting the service, they're about **improving the experience of the
client**.

<!-- @wait_clear -->
<!-- @show_composite -->

## Timeouts

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

Let's apply that and see how things go.

```bash
kubectl apply -f k8s/03-timeouts/color-profile.yaml
```

<!-- @wait_clear -->

Let's continue by adding a timout to the smiley service. The face service
will show a smiley-service timeout as a sleeping face.

```bash
diff -u99 --color k8s/{02-retries,03-timeouts}/smiley-profile.yaml
kubectl apply -f k8s/03-timeouts/smiley-profile.yaml
```

<!-- @wait_clear -->

Finally, we'll add a timeout that lets the GUI decide what to do if the face
service itself takes too long.

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
diff -u99 --color k8s/{02-retries,03-timeouts}/face-profile.yaml
kubectl apply -f k8s/03-timeouts/face-profile.yaml
```

<!-- @wait_clear -->

# SUMMARY

We've used Linkerd to take a very, very broken application and turn it into
something the user might actually have an OK experience with. Fixing the
application is, of course, still necessary!! but making the user experience
better is a good thing.

<!-- @wait -->
<!-- @show_slides -->
