# lilcaster

A lightweight ffmpeg wrapper for livestream automation and testing. Define reusable scenarios in a config file, then start, stop, and restart streams with simple commands and signals.

## Installation

**Homebrew (macOS/Linux):**

```bash
brew tap jbain/lilcaster
brew install lilcaster
```

**Go:**

```bash
go install github.com/jbain/lilcaster@latest
```

Requires `ffmpeg` to be available in your `PATH`.

## Usage

```bash
lilcaster scenario <name>                          # run a scenario
lilcaster scenario <name> --source newsource.mp4  # override all sources
lilcaster scenario <name> --sink rtmp://...        # override all sinks
lilcaster scenario <name> --verbose                # stream ffmpeg stderr live
```

### Signals

| Signal | Effect |
|--------|--------|
| `SIGINT` / `SIGTERM` | Stop the stream and exit |
| `SIGHUP` | Re-run any `script://` sources/sinks and restart |
| `SIGUSR1` | Restart using the same resolved paths (no re-run) |

`SIGHUP` is useful when a stream key has rotated and your script fetches a fresh one. `SIGUSR1` restarts without that overhead.

## Config file

lilcaster looks for `lilcaster.yml` in these locations, in order:

1. Path passed via `--config`
2. `$LILCASTER_CONFIG` environment variable
3. `./lilcaster.yml` (current directory)
4. `~/.config/lilcaster/lilcaster.yml`

### Sample config

```yaml
scenarios:
  - name: test_stream
    sources:
      - path: /path/to/source.mp4
    sinks:
      - path: rtmp://ingest.example.com/live/streamkey
    filters:
      - type: scale
        width: "1280"
        height: "720"
      - type: timestamp
    loop: -1

  - name: multi_sink
    sources:
      - path: script://scripts/get-source.sh
    sinks:
      - path: rtmp://ingest-a.example.com/live/key1
      - path: rtmp://ingest-b.example.com/live/key2
    filters:
      - type: phrase
    loop: 3
```

### Scenario fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Unique name used on the command line |
| `sources` | list | One or more input endpoints |
| `sinks` | list | One or more output endpoints |
| `filters` | list | Video filters applied in order (see below) |
| `loop` | int | Pass `-stream_loop N` to ffmpeg. `-1` loops forever, `0` disables, `3` loops 3 times |

### Endpoints (sources and sinks)

Each endpoint has a `path` and an optional `args` list:

```yaml
sources:
  - path: /videos/clip.mp4
  - path: script://scripts/get-url.sh
    args:
      - "--region"
      - "us-east-1"
```

**Path types:**

- **File path or URL** — anything ffmpeg accepts: `/path/to/file.mp4`, `rtmp://...`, `rtmps://...`, etc.
- **`script://`** — executes a script and uses its stdout as the path. Relative paths are resolved from the working directory; use a leading `/` for absolute paths.

  ```yaml
  # absolute path (note the leading /)
  path: "script:///home/alice/get-stream-key.sh"

  # relative path (resolved from working directory)
  path: "script://scripts/get-stream-key.sh"
  ```

  The script receives any `args` as command-line arguments. Its stdout (trimmed) becomes the endpoint path. On `SIGHUP`, scripts are re-executed to pick up fresh values.

Multiple sources are each passed as separate `-i` inputs to ffmpeg. Multiple sinks are each passed as separate output paths.

## Filters

Filters are applied as ffmpeg `-vf` video filters, in the order listed.

### `scale`

Resizes the video.

```yaml
- type: scale
  width: "1280"
  height: "720"
```

Use `-2` for either dimension to preserve the aspect ratio:

```yaml
- type: scale
  width: "-2"
  height: "480"
```

Both `width` and `height` are required.

### `timestamp`

Burns the current wall-clock time and the stream position into the top-left corner of the video. Takes no additional fields.

```yaml
- type: timestamp
```

The overlay shows two lines: local time (`HH:MM:SS`) and stream position (`H:MM:SS`).

### `phrase`

Burns a randomly generated three-word phrase (e.g. `swift-cedar-orbit`) in a random color into the bottom-left corner. Useful for visually distinguishing streams at a glance. Takes no additional fields.

```yaml
- type: phrase
```

The phrase is chosen once per stream start. On restart (`SIGHUP` or `SIGUSR1`) a new phrase is generated.

### `custom`

Passes an arbitrary ffmpeg video filter string through directly.

```yaml
- type: custom
  string: "eq=brightness=0.1:contrast=1.2"
```

`string` is required and must be a valid ffmpeg filter expression. Multiple `custom` filters are joined with `,` along with any other filters in the list.