# Events

Docker provides an _events_ command to show system's real time events. Likewise, Dokku can record events as syslog entries and also provides a plugin to display the last ones.

```
events [-t]                                     Show the last events (-t follows)
events:list                                     List logged events
events:on                                       Enable events logger
events:off                                      Disable events logger
```

## System logs

The _events_ plugin writes log entries to ``/var/log/syslog`` as well as a Dokku-specific logs sub-directory ``/var/log/dokku/events.log``.
