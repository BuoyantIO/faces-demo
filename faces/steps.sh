cdiff() {
	diff -U 9999 "$1" "$2" | ./diffc
}

#@SHOW

#@clear

# RETRIES
#
# Let's tell Emissary to add a retry when we call the
# face service.
#@echo
#@noshow
cdiff k8s/{01-base,02-retries}/face-mapping.yaml
#@wait
#@echo

kubectl apply -f k8s/02-retries/face-mapping.yaml
#@wait
#@clear

# RETRIES
#
# Great -- how about using Emissary to add a retry when
# we call the smiley service, to get rid of the cursing
# faces?
#@echo
#@noshow
cdiff k8s/{01-base,02-retries}/smiley-mapping.yaml
#@wait
#@echo

kubectl apply -f k8s/02-retries/smiley-mapping.yaml
#@wait
#@clear

# RETRIES
#
# That didn't work because Emissary doesn't ever see the
# call from the face service to the smiley service.
# Instead, we'll ask Linkerd to do the retries.
#@echo
#@noshow
cdiff k8s/{01-base,02-retries}/smiley-profile.yaml
#@wait
#@echo

kubectl apply -f k8s/02-retries/smiley-profile.yaml
#@wait
#@clear

# RETRIES
#
# We'll ask Linkerd to do retries for the color service,
# too, to get rid of the grey backgrounds.
#@echo
#@noshow
cdiff k8s/{01-base,02-retries}/color-profile.yaml
#@wait
#@echo

kubectl apply -f k8s/02-retries/color-profile.yaml
#@wait
#@clear

# TIMEOUTS
#
# Things are better, but still too slow. Let's add
# timeouts, starting from the bottom of the call graph
# this time.
#
# You'll see pink backgrounds here: timeouts aren't about
# protecting the service, they're about providing agency
# to the client. Here, the face service chooses to show
# timeouts as a pink background.
#@echo
#@noshow
cdiff k8s/{02-retries,03-timeouts}/color-profile.yaml
#@wait
#@echo

kubectl apply -f k8s/03-timeouts/color-profile.yaml
#@wait
#@clear

# TIMEOUTS
#
# Let's continue with the faces service. The faces
# service shows a timeout as a sleeping face.
#@echo
#@noshow
cdiff k8s/{02-retries,03-timeouts}/smiley-profile.yaml
#@wait
#@echo

kubectl apply -f k8s/03-timeouts/smiley-profile.yaml
#@wait
#@clear

# TIMEOUTS
#
# Finally, we'll add a timeout when Emissary calls the
# faces service.
#
# Here, the client is the web app. The web app is OK
# showing the user older data rather than showing them a
# failure, but for the moment it will also put a counter
# in the corner so that we can see when it's doing so for
# demo purposes.
#@echo
#@noshow
cdiff k8s/{02-retries,03-timeouts}/face-mapping.yaml
#@wait
#@echo

kubectl apply -f k8s/03-timeouts/face-mapping.yaml
#@wait
#@clear

# RATELIMITS
#
# Suppose someone adds some code to the faces service
# that makes it collapse under too much load, rather than
# just getting slower?
#
# As it happens, our faces service has exactly that code,
# which we'll enable now.
kubectl set env deploy -n faces face MAX_RATE=8.5
#@wait
#@clear

# RATELIMITS
#
# We can tell Emissary to enforce a rate limit for
# requests to the faces service. This is both protecting
# the service and providing agency to the client: here,
# our web app is going to handle rate limits just like it
# handles timeouts.
#@echo
#@noshow
cdiff k8s/{03-timeouts,04-ratelimits}/face-mapping.yaml
#@wait
#@echo

kubectl apply -f k8s/04-ratelimits/face-mapping.yaml
#@wait
#@clear

# SUMMARY
#
# We've used both Emissary and Linkerd to take a very,
# very broken application and turn it into something the
# user might actually have an OK experience with. Fixing
# the application is, of course, still necessary!! but
# making the user experience better is a good thing.

#@wait
