# lilcaster
A lightweight wrapper around ffmpeg to ease livestream testing and automation.

# Goals
- Create a light weight wrapper around ffmpeg, it should be easy to reason about
- It should be flexible and reusable.
- A primary use-case is to test live streaming services where a user will be running many live streams, one after another
  - The streams need to be easy to differentiate, so features like burning in a timestamp and other differentiating features are first class
  - The ability to stop/restart a live stream easily is required
    - when using a scriptable source or sink, we need the option to restart both with and without re-running the scripts
  - filters should also be scriptable
  - All scriptable elements need to be execeedingly un-complicated
- 'magic' is bad. All behaviour should be clear and easy to reason about.

## Scenarios
In the config file, lilcaster.yml there is a list of `scenarios`

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
        string: ""
```

### Sources/Sinks
There can be multiple sources or destinations for a scenario. This allows streaming to multiple places,
or combining multiple sources into a single stream (ie: bug insertion)

**path**
The path of a source or sink can be a simple file path or URL, anything ffmpeg compatible.

Additionally the `script://` scheme is support to execute a script to generate the output:

examples:
```yaml
sources:
  - path: "script:///User/Alice/GetTwitchUrl.sh"
  - path: "script://scripts/GetTwitchURL.sh"
```
Note the triple '/' on the first one to note the absolute path, where as the second is a relative path

#### Source specific:
**avfoundation**
The ability to use an avfoundation device like a webcam and microphone
```yaml
path: "avfoundation://?audio_dev=0&video_dev=6"
```

### Filters
A list of audio/video filters to run the streams through for burned in timestamps, scaling, etc

## Command Usage:

```shell

lilcaster scenario simple_stream #run the `simple_stream` scenario
lilcaster scenario simple_stream --source newsource.mp4 # run `simplestream` but override the source. This will replace ALL sources
lilcaster scenario simple_stream --sink rtmps://example.com/newstreamkey #run simplestream but override the sinks
```

