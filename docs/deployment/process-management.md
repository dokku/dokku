# Process/Container management

Dokku supports rudimentary process (really container) management via the `ps` plugin.

```
ps <app>                                        List processes running in app container(s)
ps:start <app>                                  Start app container(s)
ps:stop <app>                                   Stop app container(s)
ps:restart <app>                                Restart app container(s)
ps:restartall                                   Restart all deployed app containers
```

*NOTE*: As of v0.3.14, `dokku deploy:all` in now deprecated by `ps:restartall` and will be removed in a future version.
