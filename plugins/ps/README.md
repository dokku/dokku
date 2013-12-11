dokku-ps
========

A process management plugin for [dokku](https://github.com/progrium/dokku).

Modeled after the process/dyno management of heroku.

Written to solve these problems:

1. Restarting crashed containers.

  Because sometimes the process exits because of something like a memory leak and, until we have it fixed, we just want it to keep going.

  And the idea is to use upstart to handle the containers, as described in [dockers host integration docs](http://docs.docker.io/en/latest/use/host_integration/).

2. Being able to start up non `web` processes.

Bonus problems to solve:

1. Monitor & notify.

  I'm guessing since upstart is monitoring the respawning of containers it should be able to notify me by email, irc or something even cooler.

2. Process scaling.

  If we are able to add each process as a backend to the nginx config we should easily be able to load balance the requests over multiple processes. [This PR](https://github.com/progrium/dokku/pull/267) seems like it would help a lot as we would pretty much just generate `$APP_ROOT/nginx.conf` with the processes defined in `$APP_ROOT/PS`. Something like:

    CONF="$APP_ROOT/nginx.conf"
    echo "upstream $APP {" > $CONF
    grep "^web=" $APP_ROOT/PS | while read line; do
      NUM=${line#*=}
      for ((i=1; i<=$NUM; i++)); do
        PORT=$(docker port $APP.web.$i 5000 | sed 's/0.0.0.0://')
        echo "server 127.0.0.1:$PORT;" >> $CONF
      done
    done
    echo "}" >> $CONF
