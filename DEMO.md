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
<!-- @import demosh/demo-tools.sh -->
<!-- @import demosh/check-requirements.sh -->

<!-- @start_livecast -->

The "diff -u99 --color" function does a simple colorized diff.

```bash
diff -u99 --color() {
	diff -U 9999 "$1" "$2" | ./diffc
}
```

---
<!-- @SHOW -->

# Emissary and Linkerd Resilience Patterns
## Rate Limits, Retries, and Timeouts

We're going to show various resilience techniques using the Faces demo (from
https://github.com/BuoyantIO/faces-demo):

- _Rate limits_ protect services by restricting the amount of traffic that can
  flow through to a service;
- _Retries_ automatically repeat requests that fail; and
- _Timeouts_ cut off requests that take too long.

All are important techniques for resilience, and all can be applied - at
various points in the call stack - by infrastructure components like the
ingress controller and/or service mesh.

Let's start with a quick look at Faces in the web browser. You'll be able to
see that it's in pretty sorry shape, and you'll be able to look at the Linkerd
dashboard to see how much traffic it generates.

<!-- @browser_then_terminal -->

## RETRIES

Let's start by going after the red frowning faces: those are the ones where
the face service itself is failing. We can tell Emissary to retry those when
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

So that helped quite a bit: it's not perfect, because Emissary will only retry
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

...Emissary never talks to the smiley service! so telling Emissary to retry
the failed call will never work.

Instead, we need to tell Linkerd to do the retries, by adding `isRetryable` to
the `ServiceProfile` for the smiley service:

```bash
diff -u99 --color k8s/{01-base,02-retries}/smiley-profile.yaml
```

This is different from the Emissary version because Linkerd uses a _retry
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

Let's continue by adding a timout to the smiley service. The faces service
will show a smiley-service timeout as a sleeping face.

```bash
diff -u99 --color k8s/{02-retries,03-timeouts}/smiley-profile.yaml
kubectl apply -f k8s/03-timeouts/smiley-profile.yaml
```

<!-- @browser_then_terminal -->

## TIMEOUTS continued

Finally, we'll add a timeout that lets the GUI decide what to do if the faces
service itself takes too long. We'll use Emissary for this (although we
could've used Linkerd, since Emissary is itself in the mesh).

When the GUI sees a timeout talking to the faces service, it will just keep
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
better. Suppose, though, that someone now adds some code to the faces service
that makes it just completely collapse under heavy load? Sadly, this is often
all-too-easy to mistakenly do.

Let's simulate this. The faces service has internal functionality to limit its
abilities under load when we set the `MAX_RATE` environment variable, so we'll
do that now:

```bash
kubectl set env deploy -n faces face MAX_RATE=8.5
```

Once that's done, we can take a look in the browser to see what happens.

<!-- @browser_then_terminal -->

## RATELIMITS continued

Since the faces service is right on the edge, we can have Emissary enforce a
rate limit on requests to the faces service. This is both protecting the
service (by reducing the traffic) **and** providing agency to the client (by
providing a specific status code when the limit is hit). Here, our web app is
going to handle rate limits just like it handles timeouts.

Actually setting the rate limit is one of the messier bits of Emissary: the
most important thing here is to realize that we're actually providing a
**label** on the requests, and that the external rate limit service is
counting traffic with that label to decide what response to hand back.

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

We've used both Emissary and Linkerd to take a very, very broken application
and turn it into something the user might actually have an OK experience with.
Fixing the application is, of course, still necessary!! but making the user
experience better is a good thing.

<!-- @wait -->
<!-- @show_slides -->
