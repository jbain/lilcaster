# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./...          # build all packages
go test ./...           # run all tests
go test ./internal/ffmpeg/... -run TestBuild   # run a single test
go mod tidy             # tidy dependencies
GOPROXY=direct go get github.com/jbain/ezec@main  # update ezec (proxy doesn't list versions)
```

## Architecture

lilcaster is a thin CLI wrapper around ffmpeg for livestream automation. The entry point is `main.go` → `cmd.Execute()`.

**Config layer (`internal/config`)** — loads `lilcaster.yml` via viper (4-location precedence: `--config` flag → `$LILCASTER_CONFIG` → `./lilcaster.yml` → `~/.config/lilcaster/lilcaster.yml`). Uses `gopkg.in/yaml.v3` directly (not viper's mapstructure) so the custom `UnmarshalYAML` on `FilterEntry` runs. `Filter` is a sum type: `ScaleFilter`, `TimestampFilter`, `CustomFilter` — all dispatched by the `type:` YAML field.

**ffmpeg layer (`internal/ffmpeg`)** — four focused files:
- `resolve.go` — resolves `Endpoint.Path`: executes `script://` paths, stubs `avfoundation://`, passes everything else through. Returns `ResolvedEndpoint{Original, Path}` so SIGHUP can re-resolve and SIGUSR1 can reuse.
- `filter.go` — `Build([]FilterEntry) string` produces the `-vf` value from the filter list.
- `progress.go` — `Parser` reads ffmpeg's `pipe:3` `key=value` progress stream. Blocks are terminated by `progress=continue` or `progress=end`; only then does it swap pending → live state and fire the callback. `Render()` formats the status line.
- `args.go` + `ffmpeg.go` — `BuildArgs` assembles the full ffmpeg argv (`-re` before each `-i` to cap playback at native framerate; `-stream_loop` before `-i` when looping). `Process` wraps ezec: a `CallbackConsumer` feeds `Parser.Feed` on fd 3; a `LastNConsumer` (capacity 10) retains the last 10 stderr lines for error reporting; verbose mode adds a second stderr consumer that prints live.

**Runner (`internal/runner`)** — `Run()` orchestrates the lifecycle: resolve → build → start → signal loop. Signals: `SIGINT`/`SIGTERM` kill and exit, `SIGHUP` re-resolves all `script://` endpoints and restarts, `SIGUSR1` reuses cached paths and restarts. A 250 ms ticker renders the progress line with `\r\033[K` (carriage return + erase-to-EOL). On non-zero ffmpeg exit, prints the error and the last 10 stderr lines.

**CLI (`cmd`)** — `root.go` holds the persistent `--config` flag and `SilenceUsage: true` (prevents cobra from printing usage on runtime errors). `scenario.go` implements `lilcaster scenario <name> [--source] [--sink] [--verbose]`.

## Key design decisions

- `SliceArgs` (not ezec's `StringArgs`) is used to avoid whitespace-splitting of paths with spaces.
- `LastNConsumer` is always wired to stderr even in non-verbose mode — this is intentional so error output is always available, not just when `--verbose` is set.
- `viper.ConfigFileUsed()` + `yaml.Unmarshal` rather than `viper.Unmarshal` — viper's mapstructure won't trigger custom YAML unmarshalers.
