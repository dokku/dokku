# Process/Container management

Dokku supports rudimentary process (really container) management via the `ps` plugin.

```
ps <app>                                        List processes running in app container(s)
ps:start <app>                                  Start app container(s)
ps:stop <app>                                   Stop app container(s)
ps:restart <app>                                Restart app container(s)
ps:restartall                                   Restart all deployed app containers
```
