# Faces Demo

This is the Faces demo application. It has a single-page web GUI that presents
a grid of cells, each of which _should_ show a smiling face on a green
background. Spoiler alert: installed exactly as committed to this repo, that
isn't what you'll get -- many, many things can go wrong, and will. The point
of the demo is let you try to fix things.

In here you will find:

- `create-cluster.sh`, a shell script to create a `k3d` cluster and prep it by
  running `setup-cluster.sh`.

- `setup-cluster.sh`, a shell script to set up an empty cluster with [Linkerd],
  [Emissary-ingress], and the Faces app.
   - These things are installed in a demo configuration: read and think
     **carefully** before using this demo as background for a production
     installation! In particular:
      - We use `sed` to force everything to just one replica when installing
        Emissary -- **DON'T** do that in production.
      - We only configure HTTP, not HTTPS. Again, **DON'T** do this in
        production.

- `DEMO.md`, a Markdown file for the resilience demo presented live for a
  couple of events. The easiest way to use `DEMO.md` is to run it with
  [demosh].

   - (You can also run `create-cluster.sh` and `setup-cluster.sh` with
     [demosh], but they're fine with `bash` as well. Realize that all the
     `#@` comments are special to [demosh] and ignored by `bash`.)

## To try this yourself:

- Make sure `$KUBECONFIG` is set correctly.

- If you need to, run `bash create-cluster.sh` to create a new `k3d` cluster to
  use.
   - **Note:** `create-cluster.sh` will delete any existing `k3d` cluster named
     "faces".

- If you already have an empty cluster to use, you can run `bash setup-cluster.sh`
  to initialize it.

- Play around!! Assuming that you're using k3d, the Faces app is reachable at
  http://localhost/faces/ and the Linkerd Viz dashboard is available at
  http://localhost/

   - If you're not using k3d, instead of `localhost` use the IP or DNS name of
     the `emissary-ingress` service in the `emissary` namespace.

   - Remember, HTTPS is **not** configured.

- To run the demo as we've given it before, check out [DEMO.md]. The easiest
  way to use that is to run it with [demosh].

[Linkerd]: https://linkerd.io
[Emissary-ingress]: https://www.getambassador.io/docs/emissary/
[DEMO.md]: DEMO.md
[demosh]: https://github.com/BuoyantIO/demosh
