# Build Tracking

> [!IMPORTANT]
> New as of 0.34.0

Every deploy that flows through Dokku - whether triggered by `git push`, `ps:rebuild`, `ps:restart`, `config:set`, or `git:from-archive` / `git:from-image` / `git:sync` / `git:load-image` - is recorded as a structured build record on disk. The `builds` plugin lets operators inspect what is currently deploying, look up the result of an old deploy, and stream the captured build log without depending on `journalctl`.

```
builds:cancel <app>                               # Cancel a running build for an app
builds:info <app> <build-id> [--format json]      # Show details for a single build
builds:list [<app>] [--format json] [--kind ...] [--status ...]
                                                  # List builds (running across all apps, or running + history for one)
builds:output <app> [<build-id>|current]          # Show build output (tail for live, cat for finished)
builds:prune <app> [--all-apps]                   # Reap abandoned records and apply retention
builds:report [<app>] [<flag>]                    # Display a build report
builds:set [--global|<app>] <key> [<value>]       # Set or clear a builds property
```

## Build records

Every build is identified by a sortable base36 ULID-style id (`DOKKU_BUILD_ID`) generated at deploy start. For each build, Dokku writes:

- `$DOKKU_LIB_ROOT/data/builds/<app>/<build-id>.json` - the structured record
- `$DOKKU_LIB_ROOT/data/builds/<app>/<build-id>.log` - the captured stdout/stderr of the deploy

Output is also tagged into syslog as `dokku-<build-id>` so `journalctl -t dokku-<build-id>` continues to work. The on-disk log file is the durable source of truth and is read for `builds:output` even when journald has rotated old entries away.

### Record schema

```json
{
  "id": "01j8c4xv7bk5w3",
  "app": "myapp",
  "kind": "build",
  "pid": 12345,
  "started_at": "2026-04-30T13:50:00Z",
  "finished_at": "2026-04-30T13:51:14Z",
  "status": "succeeded",
  "source": "git-hook",
  "exit_code": 0
}
```

- **kind** - `build` for paths that produce a new image (`git push`, `git:*`, `ps:rebuild`); `deploy` for paths that re-deploy an existing image (`ps:restart`, `ps:start`, `dokku deploy`, `config:set`).
- **status** - on-disk values are `running | succeeded | failed | canceled`. `abandoned` is a fifth display-only value computed at read time for `running` records whose PID is no longer alive; it is never persisted to the record.
- **source** - the user-typed command that originated the deploy (e.g. `git-hook`, `ps:restart`, `config-redeploy`, `git:sync`).

## Listing builds

Without an app argument, `builds:list` shows currently-running builds across every app on the host:

```shell
dokku builds:list
```

```
=====> Currently running builds
App        Build ID        Kind   PID    Source     Started
myapp      01j8c4xv7bk5w3  build  12345  git-hook   2026-04-30T13:50:00Z
```

With an app, `builds:list` returns running builds plus the most recent finalized records up to the configured retention:

```shell
dokku builds:list myapp
```

Filter by kind or status:

```shell
dokku builds:list myapp --kind build
dokku builds:list myapp --status running
dokku builds:list --format json
```

## Inspecting a single build

```shell
dokku builds:info myapp 01j8c4xv7bk5w3
```

```
=====> Build 01j8c4xv7bk5w3
       Build ID:   01j8c4xv7bk5w3
       App:        myapp
       Kind:       build
       Status:     succeeded
       PID:        12345
       Source:     git-hook
       Started:    2026-04-30T13:50:00Z
       Finished:   2026-04-30T13:51:14Z
       Duration:   1m14s
       Exit Code:  0
       Log:        /var/lib/dokku/data/builds/myapp/01j8c4xv7bk5w3.log
```

JSON output includes the same fields plus a computed `log_path`:

```shell
dokku builds:info myapp 01j8c4xv7bk5w3 --format json | jq .
```

## Streaming build output

```shell
# tail -f the live build, or cat the log for a finished one
dokku builds:output myapp 01j8c4xv7bk5w3

# resolve "current" from the in-flight deploy
dokku builds:output myapp current
```

If the on-disk log file is missing (for example, a build from before this plugin was installed), `builds:output` falls back to `journalctl -t dokku-<build-id>`.

## Cancelling a build

`builds:cancel` reads the active `.deploy.lock` for an app, looks up the matching build record, and sends `SIGQUIT` to the deploy's process group. The record is finalized as `canceled` (or `failed` if the process had already exited without finalizing - in that case Dokku marks the record so it doesn't sit in `running` forever).

```shell
dokku builds:cancel myapp
```

If the record is already finalized when cancel runs, no signal is sent and the record is left untouched.

## Retention

Build records are pruned by count, not by age. The default retention is **20** records per app. Live in-flight deploys are never pruned regardless of count.

Set a per-app retention:

```shell
dokku builds:set myapp retention 50
```

Set the global default:

```shell
dokku builds:set --global retention 10
```

Clear an override (falls back to the global value, then the default):

```shell
dokku builds:set myapp retention
```

## Manual pruning

`builds:prune` invokes the same logic that runs at the end of every deploy - it reaps abandoned records (status=running with a dead PID, finalized as `failed`) and trims the directory to the configured retention. Useful after lowering retention via `builds:set` or to clean up after a host reboot:

```shell
dokku builds:prune myapp
dokku builds:prune --all-apps
```

## Reports

```shell
dokku builds:report
dokku builds:report myapp
dokku builds:report myapp --build-status
```

Available flags:

- `--build-id`, `--build-kind`, `--build-status`, `--build-pid`, `--build-source`, `--build-started-at`, `--build-finished-at`, `--build-exit-code`: details of the most recent build for the app
- `--builds-retention`: per-app retention override (empty if none)
- `--builds-global-retention`: global retention override (empty if none)
- `--builds-computed-retention`: the resolved retention applied to this app

`--build-status` returns the **display** status, so an abandoned in-flight build shows `abandoned` rather than `running`. The raw on-disk status is only visible by reading the JSON record directly.
