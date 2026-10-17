# Hana

A local osu! companion that watches your running game client and streams a live,
customizeable leaderboard to your browser.

Hana attaches to a running **osu!stable** or **osu!lazer** process, reads the
currently selected beatmap out of game memory (like
[tosu](https://github.com/tosuapp/tosu)), fetches scores from the osu! API, and
pushes them to a web UI over a WebSocket so you can watch friends' (or global)
scores update in real time while you play.

![screenshot](https://raw.githubusercontent.com/yukiisen/hana/refs/heads/main/screenshot.png)

## Features

- Beatmap changes are picked up every 400 ms and
  broadcast to all connected browsers.
- Built-in `Global` board pulls the
  beatmap's top scores from the osu! API; custom boards track a list of player
  usernames you define.
- Scores are cached to disk with a TTL, so switching back
  and forth between boards doesn't hammer the API.
- Star rating and FC PP for 95–100% accuracy are
  computed client-side with [rosu-pp](https://github.com/maxohn/rosu-pp) compiled
  to WebAssembly.
- The whole tool is a single binary.
- Cross-platform on Linux and Windows.

# NOTES (IMPORTANT)

- This was not tested on windows, some features are just unimplemented.
- A good portion of this is vibecoded using different agents.
- The project is still in very early stages and it might break.
- osu!stable support is not guaranteed on linux because wine exposes 64bits regardless of the game

# Requirements

- Go 1.27+ (to build)
- A running osu! client (stable or lazer) on the same machine
- An osu! OAuth **client credentials** token
  ([create one here](https://osu.ppy.sh/home/account/edit), scope doesn't
  matter, Hana uses `client_credentials` / `public`)
- Linux or Windows

On Linux, reading another process's memory requires ptrace permissions e.g.
run Hana as the same user as osu! and, if needed, allow it:

```sh
sudo setcap cap_sys_ptrace=eip ./hana
```

## Getting started

### 1. Configure

Copy the example config into place:

```sh
# Linux
mkdir -p ~/.config/hana
cp example.config.json ~/.config/hana/config.json

# Windows
# %AppData%\hana\config.json
```

Fill in your API credentials:

```json
{
  "api": {
    "client_id": 12345,
    "client_secret": "your_osu_oauth_client_secret"
  },
  "settings": {
    "leaderboard_wait_time": "2s",
    "score_cache_ttl": "5m"
  },
  "leaderboards": {}
}
```

| Field | Meaning |
| --- | --- |
| `api.client_id` | osu! OAuth application client ID (required) |
| `api.client_secret` | osu! OAuth client secret (required) |
| `settings.leaderboard_wait_time` | Grace period before hitting the API for uncached scores (default `2s`) |
| `settings.score_cache_ttl` | How long a cached score stays fresh (default `5m`) |
| `leaderboards` | Named boards, each a list of osu! usernames (default board: `Global`) |

Leaderboard definitions can also be edited from the web UI at
`/boards.html`; changes are saved back to `config.json`. Keep custom boards
small, each user costs one API request per uncached beatmap.

## Development

```sh
go build ./...
go test ./...
```

Project layout:

```
main.go          server bootstrap, embedded UI, routing
osumem/          osu! process memory reader (stable & lazer)
worker/          game watcher, leaderboard loader, osu! API client, score cache
handler/         WebSocket connection management
config/          config load/save with atomic writes
cache/           on-disk cache for user IDs and scores
paths/           OS-specific config/cache paths
logger/          scoped, leveled logger
public/          web UI (embedded); rosu/ is rosu-pp compiled to WASM
```

PRs are very welcome if you'd like to help reduce the amount of AI code in here.

## Credits

- [tosu](https://github.com/tosuapp/tosu) - memory-reading approach and
  signatures that `osumem` is ported from
- [rosu-pp](https://github.com/maxrh/rosu-pp) /
  [rosu-pp-js](https://github.com/maxrh/rosu-pp-js) - difficulty & performance
  calculation in the browser (MIT, see `public/rosu/LICENSE`)
- [coder/websocket](https://github.com/coder/websocket) — WebSocket server

## License

MIT License
