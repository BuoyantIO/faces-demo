# Faces MCP Server

A Model Context Protocol (MCP) server for the Faces demo that provides
tools to query and update the smiley, color, and face services.

## Tools

- **query_smiley**: Query the smiley service to get the current smiley
  emoji
- **query_color**: Query the color service via gRPC to get the current
  color
- **query_face**: Query the face service to get a combined face response
  (smiley + color)
- **update_smiley_emojis**: Update success and failure emojis for the
  smiley service via HTTP PUT
- **update_color_colors**: Update success and failure colors for the
  color service via gRPC

## Running

The MCP server supports both streaming HTTP (the default) or stdio. The
two offer the same functionality; to choose which one to use, set the
`TRANSPORT` environment variable:

| `$TRANSPORT` | Transport used |
|--------------|----------------|
| unset        | Streaming HTTP |
| `sse`        | Streaming HTTP |
| `stdio`      | Stdio          |

The MCP server must be able to talk to face, smiley, and color directly.
Set `SMILEY_URL`, `COLOR_URL`, and `FACE_URL` in the environment to tell
it where to find them -- the defaults are

* `SMILEY_URL`: `http://smiley`
* `COLOR_URL`: `color:80`
* `FACE_URL`: `http://face`

You can set `PORT` to control which port is used for SSE, with the
default being 3000.

## Building

`gmake images` will do the right thing.

## Installing

Set `mcp.enabled=true` with Helm. Check out `faces-chart/values.yaml` for
more details.

