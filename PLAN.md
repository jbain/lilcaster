# lilcaster — Design & Implementation Plan

## Overview

`lilcaster` is a CLI tool written in Go that wraps ffmpeg to simplify running and managing
livestreams, especially for testing scenarios where streams must be differentiated, restarted,
and scripted.

---

## Config File

Location priority (highest to lowest):

1. `--config <path>` flag
2. `LILCASTER_CONFIG` env var
3. `./lilcaster.yml` (CWD)
4. `~/.config/lilcaster/lilcaster.yml`

**Schema:**

```yaml
scenarios:
  - name: simple_stream
    sources:
      - path: /path/of/source.mp4
    sinks:
      - path: rtmp://ingest.example.com/abc123/streamkey
    filters:
      - type: scale
        height: "480"
        width: "-2"
      - type: timestamp
      - type: custom
        string: "eq=brightness=0.1"
    loop: -1   # -1 = stream_loop forever, 0 = no loop, N = loop N times
```

---

## Package Structure

Implementation packages live under `internal/` so they aren't part of the public
import surface. Keeping the tree shallow: three internal packages, with related
concerns grouped into multiple files in the same package.

```
lilcaster/
├── main.go
├── cmd/
│   ├── root.go          # cobra root, viper config wiring
│   └── scenario.go      # `scenario` subcommand
└── internal/
    ├── config/
    │   ├── config.go    # Config/Scenario/Endpoint structs + viper-based loader
    │   └── filter.go    # Filter type + custom UnmarshalYAML dispatch
    ├── ffmpeg/
    │   ├── args.go      # SliceArgs + arg-construction helpers
    │   ├── ffmpeg.go    # Process wrapper, ezec consumer wiring, binary discovery
    │   ├── filter.go    # builds -vf filtergraph string from []config.Filter
    │   ├── progress.go  # parses pipe:3 key=value blocks → State, status-line render
    │   └── resolve.go   # script:// resolution; avfoundation stub
    └── runner/
        └── runner.go    # lifecycle: start, signals, restart, display ticker
```

Rationale for the consolidation: `resolve`, `filter`, and `progress` are all
small, single-purpose pieces tightly coupled to the ffmpeg invocation — they
share a package rather than each getting their own directory. `runner` stays
separate because it's the orchestration layer and depends on the others.

---

## Structs (`config`)

```go
type Config struct {
    Scenarios []Scenario `yaml:"scenarios"`
}

type Scenario struct {
    Name    string     `yaml:"name"`
    Sources []Endpoint `yaml:"sources"`
    Sinks   []Endpoint `yaml:"sinks"`
    Filters []Filter   `yaml:"filters"`
    Loop    int        `yaml:"loop"` // -1=forever, 0=no loop, N=N times
}

// Endpoint is the shared shape for sources and sinks. Keeping a single type
// (rather than parallel Source/Sink structs) avoids duplicate fields when
// the schema is identical and lets resolve/runtime code treat both uniformly.
type Endpoint struct {
    Path string `yaml:"path"`
}
```

### Filter — discriminated by `type`

`Filter` carries different fields depending on `type`. Two acceptable
implementations; pick whichever feels cleaner once the field set settles:

**Option A — custom UnmarshalYAML, sum-type style:**

```go
type Filter interface{ kind() string }

type ScaleFilter     struct{ Width, Height string }
type TimestampFilter struct{}
type CustomFilter    struct{ String string }

// UnmarshalYAML on a wrapper peeks at `type:` then decodes into the right struct.
```

**Option B — decode into a `map[string]any`, then dispatch:**

```go
type Filter struct {
    Type string         `yaml:"type"`
    Args map[string]any `yaml:",inline"`
}
```

Either way, the flat-struct-with-conditional-fields shape is avoided so an
invalid combination (e.g. `scale` with no dimensions, `timestamp` with a
stray `string`) can be rejected at decode time. Validation runs in
`config.Load()` before the scenario reaches the runner.

---

## Path Resolution (`internal/ffmpeg/resolve.go`)

Operates on `config.Endpoint` so the same logic covers sources and sinks
(both can use `script://`).

| Input | Action |
|---|---|
| `script:///abs/path.sh` | exec `/abs/path.sh`, stdout → trimmed path |
| `script://rel/path.sh` | exec `./rel/path.sh`, stdout → trimmed path |
| `avfoundation://...` | return error: "avfoundation not yet supported" |
| anything else | pass through verbatim |

```go
type ResolvedEndpoint struct {
    Original config.Endpoint // kept for SIGHUP re-resolution
    Path     string          // resolved path used on SIGUSR1 reuse
}
```

The runner holds two slices (`[]ResolvedEndpoint` for sources and sinks);
SIGHUP re-runs `Resolve` over `Original` for every entry, SIGUSR1 reuses
the cached `Path` values.

---

## Filter Graph (`filter`)

Multiple filters joined with `,` into a single `-vf` value. No `-vf` flag emitted when the
filter list is empty.

| type | ffmpeg output |
|---|---|
| `scale` | `scale=<width>:<height>` |
| `timestamp` | `drawtext=text='%{localtime\:%T}':x=10:y=10:fontsize=24:fontcolor=white:box=1:boxcolor=black@0.5,` `drawtext=text='%{pts\:hms}':x=10:y=44:fontsize=24:fontcolor=white:box=1:boxcolor=black@0.5` |
| `custom` | raw `string` value verbatim |

---

## ezec Integration (`ffmpeg`)

```go
// Custom Args implementation — avoids StringArgs whitespace-splitting issue with paths
type SliceArgs []string
func (s SliceArgs) Args() []string { return s }

// Consumers must be go-started before cmd.Start()
progressConsumer := consumers.NewCallbackConsumer("progress", 64, parser.Feed)
go progressConsumer.Start()
cmd.AddFd([]ezec.LineConsumer{progressConsumer})  // → fd 3

// Stderr: always a callback consumer that prints. Verbose just prints every
// line; non-verbose still uses a callback so we never accumulate an unbounded
// buffer (BufferedConsumer would leak memory on a long-running stream).
stderrConsumer := consumers.NewCallbackConsumer("stderr", 64, func(line string) {
    fmt.Fprintln(os.Stderr, line)
})
go stderrConsumer.Start()
cmd.Stderr = []ezec.LineConsumer{stderrConsumer}
```

Key notes:
- `AddFd` assigns fd 3, 4, 5... in call order — one call is all we need for `-progress pipe:3`
- Consumers drop lines silently when their channel is full (no backpressure)
- `Wait()` blocks until all consumers drain, then reaps the process

---

## Progress Display (`internal/ffmpeg/progress.go`)

ffmpeg writes `key=value\n` lines to `pipe:3`, ending each block with
`progress=continue` or `progress=end`. The parser accumulates lines into a
pending `State`; the terminating `progress=...` line is the signal that the
current iteration is complete and the `State` is whole — at that point the
parser swaps the staged values into the live `State` and fires the callback.
`progress=end` carries the same "block complete" meaning as `continue`; it
additionally indicates ffmpeg is wrapping up, which the runner can observe
to suppress further status-line redraws.

```go
type State struct {
    FPS     float64
    Bitrate string
    OutTime time.Duration
    Speed   string
}
```

Terminal output (non-verbose, `\r` overwrites in place):
```
lilcaster: [simple_stream] 00:01:23 | fps=30.0 | bitrate=2500.0kbits/s | speed=1.0x
```

In verbose mode, the progress line is suppressed to avoid conflict with live ffmpeg output.

---

## Signal Handling (`runner`)

| Signal | Behavior |
|---|---|
| `SIGINT` / `SIGTERM` | Stop ffmpeg, clean up, exit 0 |
| `SIGHUP` | Kill ffmpeg, re-resolve all `script://` paths, restart |
| `SIGUSR1` | Kill ffmpeg, reuse cached resolved paths, restart |

---

## CLI

```
lilcaster scenario <name> [flags]

  --source <path>    Replace ALL sources with a single source
  --sink   <path>    Replace ALL sinks with a single sink
  --verbose          Stream ffmpeg stderr output live
  --config <path>    Config file path
                     (also: $LILCASTER_CONFIG, ./lilcaster.yml,
                      ~/.config/lilcaster/lilcaster.yml)
```

### Cobra scaffold cleanup

The current `cmd/root.go` is the `cobra-cli init` template and needs trimming
before real wiring:

- Drop the `--toggle` flag (`rootCmd.Flags().BoolP("toggle", ...)`)
- Replace the placeholder `Short`/`Long` strings with lilcaster-specific text
- Remove the `Copyright © 2026 NAME HERE` header in `cmd/root.go` and `main.go`
- Wire `cobra.OnInitialize(initConfig)` and add the persistent `--config` flag

### Viper precedence wiring

Viper does not natively chain "explicit file → env var → CWD → XDG-style
default" — it has to be assembled. Implementation sketch for `initConfig()`:

```go
if cfgFile != "" {                              // 1. --config flag
    viper.SetConfigFile(cfgFile)
} else if env := os.Getenv("LILCASTER_CONFIG"); env != "" {
    viper.SetConfigFile(env)                    // 2. env var
} else {
    viper.SetConfigName("lilcaster")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")                    // 3. CWD
    viper.AddConfigPath(                        // 4. ~/.config/lilcaster
        filepath.Join(xdgConfigHome(), "lilcaster"))
}
if err := viper.ReadInConfig(); err != nil { ... }
```

The first two cases use `SetConfigFile` (exact path); the latter two rely on
`AddConfigPath` + search. Mixing them is the part that bites — keep the
branches strictly separate.

---

## Known Limitations / Future Work

- **avfoundation**: `avfoundation://` source scheme is parsed and returns a clear "not yet
  supported" error. When implemented, it will map to `-f avfoundation -i "video:audio"` ffmpeg
  flags.
- **Scriptable filters**: filters do not currently support the `script://` scheme. A future
  iteration could allow a filter's parameters to be driven by script output.
- **Multiple source combining**: multiple sources are passed to ffmpeg as separate `-i` inputs.
  Combining them (overlay, concat, etc.) is left to custom filters.

---

## Testing

Light unit coverage on the pure pieces — no integration test suite planned:

- `internal/ffmpeg/filter.go`: table-driven test covering `scale` /
  `timestamp` / `custom` combinations and the empty-filter case.
- `internal/ffmpeg/progress.go`: feed canned `key=value` blocks, assert the
  resulting `State` and that the callback fires exactly once per block.
- `internal/ffmpeg/resolve.go`: passthrough, `script://` (using a tiny test
  script in `testdata/`), and the avfoundation error case.
- `internal/config`: load a fixture YAML, assert struct values + filter
  validation rejects bad combinations.

The runner, signal loop, and ezec wiring are skipped — they're better
exercised by manual smoke tests against a real ffmpeg.

---

## Implementation Steps

1. **Dependencies** — add cobra, viper, and `github.com/jbain/ezec` to `go.mod`
2. **Scaffold cleanup** — strip `--toggle` flag and template strings from
   `cmd/root.go`; remove `Copyright © 2026 NAME HERE` headers
3. **`internal/config`** — structs (incl. `Endpoint`), `Filter` UnmarshalYAML,
   `Load()` with viper (4-location precedence as sketched above), validation
4. **`internal/ffmpeg/resolve.go`** — `script://` execution, avfoundation stub,
   passthrough; operates on `Endpoint`
5. **`internal/ffmpeg/filter.go`** — `Build([]config.Filter) string` — scale,
   timestamp (two drawtexts), custom
6. **`internal/ffmpeg/progress.go`** — `Parser` (CallbackConsumer-compatible
   `Feed` func), `State`, `Render`
7. **`internal/ffmpeg/{args,ffmpeg}.go`** — `SliceArgs`, `BuildArgs(scenario,
   resolvedSources, resolvedSinks, filterStr) []string`, `Process` type
   wrapping ezec with consumers wired up; binary discovery via `FFMPEG_PATH`
   then `$PATH`
8. **`internal/runner`** — `Run(scenario, overrides, verbose)`: resolve →
   build → start → signal loop → progress ticker → restart logic
9. **`cmd`** — `root.go` (viper wiring per sketch), `scenario.go` (flag
   parsing, scenario lookup, call runner)
10. **`main.go`** — thin entry point: `cmd.Execute()`
