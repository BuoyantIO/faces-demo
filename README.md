# Faces Demo

This is the Faces demo application. It has a single-page web GUI that presents
a grid of cells, each of which _should_ show a grinning face on a light blue
background. Spoiler alert: installed exactly as committed to this repo, that
isn't what you'll get -- many, many things can go wrong, and will. The point
of the demo is let you try to fix things.

In here you will find:

- `create-cluster.sh`, a shell script to create a `k3d` cluster and prep it by
  running `setup-cluster.sh`.

- `setup-cluster.sh`, a shell script to set up an empty cluster with [Linkerd],
  [Edge Stack], and the Faces app.
  - These things are installed in a demo configuration: read and think
     **carefully** before using this demo as background for a production
     installation! In particular:

    - We deploy Edge Stack with only one replica of everything
    - We use the default self-signed certificate for HTTPS

     These are likely both bad ideas for a production installation.

- `DEMO.md`, a Markdown file for the resilience demo presented live for a
  couple of events. The easiest way to use `DEMO.md` is to run it with
  [demosh].

  - (You can also run `create-cluster.sh` and `setup-cluster.sh` with
     [demosh], but they're fine with `bash` as well. Realize that all the
     `#@` comments are special to [demosh] and ignored by `bash`.)

## To try this yourself

- Make sure `$KUBECONFIG` is set correctly.

- If you need to, run `bash create-cluster.sh` to create a new `k3d` cluster to
  use.
  - **Note:** `create-cluster.sh` will delete any existing `k3d` cluster named
     "faces".

- If you already have an empty cluster to use, you can run `bash setup-cluster.sh`
  to initialize it.

- Play around!! Assuming that you're using k3d, the Faces app is reachable at
  <http://localhost/faces/> and the Linkerd Viz dashboard is available at
  <http://localhost/>

  - If you're not using k3d, instead of `localhost` use the IP or DNS name of
     the `edge-stack` service in the `ambassador` namespace.

  - Remember, HTTPS is **not** configured.

- To run the demo as we've given it before, check out [DEMO.md]. The easiest
  way to use that is to run it with [demosh].

## Architecture

The Faces architecture is fairly simple:

- The `faces-gui` workload, reached on the `/faces/` path, just returns the
  HTML and Javascript for the GUI. The GUI is a single-page webapp that
  displays a grid of cells: for each cell, the GUI calls the `face` workload.

- The `face` workload, reached on the `/face/` path, calls the `smiley`
  workload to get a smiley face and the `color` workload to get a color. It
  then composes the responses together and returns the smiley/color
  combination to the GUI for display.

- The `smiley` workload returns a smiley face. By default, this is a grinning
  smiley, U+1F603, but you can set the `SMILEY` environment variable to any
  key in the `Smileys` map from `constants.go` to get a different smiley.

- The `color` workload returns a color. By default, this is a light blue, but
  you can set the `COLOR` environment variable to any key in the `Colors` map
  from `constants.go` to get a different color, or to any arbitrary hex color
  code (e.g. `#ff0000` for bright red).

  The named colors in the `Colors` map are meant to work for normal color
  vision and for various kinds of colorblindness, and are taken from the
  "Bright" color scheme shown in the "Qualitative // Color Schemes" section of
  <https://personal.sron.nl/~pault/>. For (much) more information, read the
  comments in `pkg/faces/constants.go`. Feedback here is welcome, since the
  Faces authors have normal color vision...

[Edge Stack]: https://www.getambassador.io/docs/edge-stack
[Linkerd]: https://linkerd.io
[DEMO.md]: DEMO.md
[demosh]: https://github.com/BuoyantIO/demosh
