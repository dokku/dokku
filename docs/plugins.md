# Plugins

Dokku itself is built out of plugins and uses [pluginhook](https://github.com/progrium/pluginhook) for its plugin system. In essence a plugin is a collection of scripts that will be run based on naming convention.

Let's take a quick look at the current dokku nginx plugin that's shipped with dokku by default.

    nginx-vhosts/
    ├── commands     # contains additional commands
    ├── install      # runs on dokku installation
    └── post-deploy  # runs after an app is deployed

## Installing a plugin

```bash
cd /var/lib/dokku/plugins
git clone <git url>
dokku plugins-install
```

> todo: add a command to dokku to install a plugin, given a git repository `dokku plugin:install <git url>`?

## Creating your own plugin

If you create your own plugin:

1. take a look at the plugins shipped with dokku and hack away!
2. upload your plugin to github with a repository name in form of `dokku-<name>` (e.g. `dokku-mariadb`)
3. edit this page and add a link to it below!
4. subscribe to the [dokku development blog](http://progrium.com) to be notified about API changes and releases

## Community plugins

Note: The following plugins have been supplied by our community and may not have been tested by dokku maintainers.

[Kloadut]: https://github.com/Kloadut
[jeffutter]: https://github.com/jeffutter
[jezdez]: https://github.com/jezdez
[blag]: https://github.com/blag
[luxifer]: https://github.com/luxifer
[stuartpb]: https://github.com/stuartpb
[hughfletcher]: https://github.com/hughfletcher
[Aomitayo]: https://github.com/Aomitayo
[jlachowski]: https://github.com/jlachowski
[apmorton]: https://github.com/apmorton
[statianzo]: https://github.com/statianzo
[sehrope]: https://github.com/sehrope
[wmluke]: https://github.com/wmluke
[krisrang]: https://github.com/krisrang
[scottatron]: https://github.com/scottatron
[pauldub]: https://github.com/pauldub
[dyson]: https://github.com/dyson
[musicglue]: https://github.com/musicglue
[motin]: https://github.com/motin
[neam]: https://github.com/neam
[cedricziel]: https://github.com/cedricziel
[pnegahdar]: https://github.com/pnegahdar
[fgrehm]: https://github.com/fgrehm
[alexanderbeletsky]: https://github.com/alexanderbeletsky
[thrashr888]: https://github.com/thrashr888
[robv]: https://github.com/robv
[rlaneve]: https://github.com/rlaneve
[cef]: https://github.com/cef
[alex-sherwin]: https://github.com/alex-sherwin
[F4-Group]: https://github.com/F4-Group
[fermuch]: https://github.com/fermuch
[Zenedith]: https://github.com/Zenedith
[heichblatt]: https://github.com/heichblatt
[michaelshobbs]: https://github.com/michaelshobbs
[iskandar]: https://github.com/iskandar
[mikexstudios]: https://github.com/mikexstudios
[cameron-martin]: https://github.com/cameron-martin
[matto1990]: https://github.com/matto1990
[ohardy]: https://github.com/ohardy
[agco-adm]: https://github.com/agco-adm

### Datastores

| Plugin | Author | Compatibility |
| --- | --- | --- |
| [CouchDB](https://github.com/racehub/dokku-couchdb-plugin) | [RaceHub](https://github.com/racehub) | Compatible with 0.2.0 |
| [MariaDB](https://github.com/Kloadut/dokku-md-plugin) | [Kloadut][] | Compatible with 0.2.0 |
| [MariaDB (single container)](https://github.com/ohardy/dokku-mariadb) | [ohardy][] | Compatible with 0.2.0 |
| [MongoDB (single container)](https://github.com/jeffutter/dokku-mongodb-plugin) | [jeffutter][] | |
| [PostgreSQL](https://github.com/Kloadut/dokku-pg-plugin) | [Kloadut][] | Compatible with 0.2.0 |
| [PostgreSQL (single container)](https://github.com/jeffutter/dokku-postgresql-plugin) | [jeffutter][] | This plugin creates a single postgresql container that all your apps can use. Thus only one instance of postgresql running (good for servers without a ton of memory). |
| [PostgreSQL](https://github.com/jezdez/dokku-postgres-plugin) | [jezdez][] | Compatible with 0.2.0 |
| [PostgreSQL (single container)](https://github.com/ohardy/dokku-psql) | [ohardy][] | Compatible with 0.2.0 |
| [PostGIS](https://github.com/fermuch/dokku-pg-plugin) | [fermuch][] | |
| [Redis](https://github.com/jezdez/dokku-redis-plugin) | [jezdez][] | Requires https://github.com/rlaneve/dokku-link; compatible with 0.2.0 |
| [Redis](https://github.com/luxifer/dokku-redis-plugin) | [luxifer][] | |
| [Redis (single container)](https://github.com/ohardy/dokku-redis) | [ohardy][] | Compatible with 0.2.0 |
| [RethinkDB](https://github.com/stuartpb/dokku-rethinkdb-plugin) | [stuartpb][] | 2014-02-22: targeting dokku @ [latest][217d00a]; will fail with Dokku earlier than [28de3ec][]. |
| [RiakCS (single container)](https://github.com/jeffutter/dokku-riakcs-plugin) | [jeffutter][] | Incompatible with 0.2.0 (checked at [dccee02][]) |
| [MySQL](https://github.com/hughfletcher/dokku-mysql-plugin) | [hughfletcher][] | |
| [Memcached](https://github.com/jezdez/dokku-memcached-plugin) | [jezdez][] | Compatible with 0.2.0 |
| [Neo4j](https://github.com/Aomitayo/dokku-neo4j-plugin) | [Aomitayo][] | |
| [RabbitMQ](https://github.com/jlachowski/dokku-rabbitmq-plugin) | [jlachowski][] | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [RabbitMQ (single container)](https://github.com/jlachowski/dokku-rabbitmq-single-plugin) | [jlachowski][] | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [Postgresql](https://github.com/jlachowski/dokku-pg-plugin) | [jlachowski][] | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [Memcached](https://github.com/jlachowski/dokku-memcached-plugin) | [jlachowski][] | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [Varnish](https://github.com/Zenedith/dokku-varnish-plugin) | [Zenedith][] | Varnish cache between nginx and application with base configuration|

[dccee02]: https://github.com/jeffutter/dokku-riakcs-plugin/commit/dccee02702e7001851917b7814e78a99148fb709

### Process Managers

| Plugin | Author | Compatibility |
| --- | --- | --- |
| [Circus](https://github.com/apmorton/dokku-circus) | [apmorton][] | |
| [Shoreman ](https://github.com/statianzo/dokku-shoreman) | [statianzo][] | Compatible with 0.2.0 |
| [Supervisord](https://github.com/statianzo/dokku-supervisord) | [statianzo][] | Compatible with 0.2.0 |
| [Logging Supervisord](https://github.com/sehrope/dokku-logging-supervisord) | [sehrope][] | Works with dokku @ [c77cbf1][] - no 0.2.0 compatibility |
| [Forego](https://github.com/iskandar/dokku-forego) | [iskandar][] | Compatible with 0.2.x |

[c77cbf1]: https://github.com/progrium/dokku/commit/c77cbf1d3ae07f0eafb85082ed7edcae9e836147
[28de3ec]: https://github.com/progrium/dokku/commit/28de3ecaa3231a223f83fd8d03f373308673bc40

### Dokku Features

| Plugin | Author | Compatibility |
| --- | --- | --- |
| [Custom Domains](https://github.com/neam/dokku-custom-domains) | [motin][] | Compatible with 0.2.* and master |
| [Multiple Domains](https://github.com/wmluke/dokku-domains-plugin)<sup>4</sup> | [wmluke][] | Compatible with 0.2.0 . |
| [git rev-parse HEAD in env](https://github.com/nornagon/dokku-git-rev) | [nornagon](https://github.com/nornagon) | Compatible with 0.2.0 |
| [Rebuild application](https://github.com/scottatron/dokku-rebuild) | [scottatron][] | Compatible with 0.2.x |
| [Multi-Buildpack](https://github.com/pauldub/dokku-multi-buildpack) | [pauldub][] | |
| [Docker Options](https://github.com/dyson/dokku-docker-options) | [dyson][] | Requires dokku >= [c77cbf1][] |
| [Persistent Storage](https://github.com/dyson/dokku-persistent-storage) | [dyson][] | Requires dokku >= [c77cbf1][] |
| [Volume (persistent storage)](https://github.com/ohardy/dokku-volume) | [ohardy][] | Compatible with 0.2.0 |
| [user-env-compile](https://github.com/musicglue/dokku-user-env-compile)<sup>1</sup>  | [musicglue][] | Compatible with dokku master branch |
| [user-env-compile](https://github.com/motin/dokku-user-env-compile)<sup>1</sup>  | [motin][] | Compatible with 0.2.1 |
| [registry](https://github.com/agco-adm/dokku-registry)<sup>5</sup>  | [agco-adm](https://github.com/agco-adm) |  |
| [SSH Deployment Keys](https://github.com/cedricziel/dokku-deployment-keys)<sup>2</sup>  | [cedricziel][] | 2014-01-17: compatible with upstream/master |
| [SSH Hostkeys](https://github.com/cedricziel/dokku-hostkeys-plugin)<sup>3</sup>  | [cedricziel][] | 2014-01-17: compatible with upstream/master |
| [Link Containers](https://github.com/rlaneve/dokku-link) | [rlaneve][] | Requires dokku >= [c77cbf1][] |
| [Host Port binding](https://github.com/stuartpb/dokku-bind-port) | [stuartpb][] | Requires dokku >= [c77cbf1][]. 2014-02-17: [a043e98][] targeting dokku @ [latest][217d00a] |
| [Docker name](https://github.com/alex-sherwin/dokku-name) | [alex-sherwin][] | Requires dokku >= [c77cbf1][] |
| [No VHOST](https://github.com/alex-sherwin/dokku-novhost) | [alex-sherwin][] | Requires dokku >= [c77cbf1][] |
| [Docker Direct](https://github.com/heichblatt/dokku-docker-direct) | [heichblatt][] | |
| [Ports](https://github.com/heichblatt/dokku-ports) | [heichblatt][] | |
| [Debug](https://github.com/heichblatt/dokku-debug) | [heichblatt][] | |
| [Pre-Deploy Tasks](https://github.com/michaelshobbs/dokku-app-predeploy-tasks) | [michaelshobbs][] | |
| [Hostname](https://github.com/michaelshobbs/dokku-hostname) | [michaelshobbs][] | |
| [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration) | [motin][] | Compatible with 0.3.1+ |
| [Multiple domains and per-app custom nginx.conf](https://github.com/mikexstudios/dokku-nginx-alt) | [mikexstudios][] | Works with v0.2.3 |
| [Supply env vars to buildpacks](https://github.com/cameron-martin/dokku-build-env)| [cameron-martin][] | Works with v0.2.3 |
| [HTTP Auth Secure Apps](https://github.com/matto1990/dokku-secure-apps) | [matto1990][] | Works with v0.2.3 |
| [app-url](https://github.com/mikecsh/dokku-app-url) | [mikecsh](https://github.com/mikecsh) | Works with 0.2.0 |
| [dokku-registry](https://github.com/agco-adm/dokku-registry) | [agco-adm][] | |

[8fca220]: https://github.com/progrium/dokku/commit/8fca2204edb0017796d6915ca9157c05b1238e28
[217d00a]: https://github.com/progrium/dokku/commit/217d00a1bc47a7e24d8847617bb08a1633025fc7
[98332de]: https://github.com/dyson/dokku-persistent-storage/commit/98332de4b5b640610bee535f4d5260263074e18b
[a043e98]: https://github.com/stuartpb/dokku-bind-port/commit/a043e9892f4815b6525c850131e09fd64db5c1fa

<sup>1</sup> Similar to the heroku-labs feature (see https://devcenter.heroku.com/articles/labs-user-env-compile)

<sup>2</sup> Adds the possibility to add SSH deployment keys to receive private hosted packages

<sup>3</sup> Adds the ability to add custom hosts to the containers known_hosts file to be able to ssh them easily (useful with deployment keys)

<sup>4</sup> Conflicts with [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration)

<sup>5</sup> On Heroku similar functionality is offered by the [heroku-labs pipeline feature](https://devcenter.heroku.com/articles/labs-pipelines), which allows you to promote builds across multiple environments (staging -> production)

### Other Add-ons

| Plugin | Author | Compatibility |
| --- | --- | --- |
| [Node](https://github.com/pnegahdar/dokku-node) | [pnegahdar](https://github.com/pnegahdar) | |
| [Node](https://github.com/ademuk/dokku-nodejs) | [ademuk](https://github.com/ademuk) | |
| [Chef cookbooks](https://github.com/fgrehm/chef-dokku) | [fgrehm][] | |
| [Bower install](https://github.com/alexanderbeletsky/dokku-bower-install) | [alexanderbeletsky][] | |
| [Bower/Grunt](https://github.com/thrashr888/dokku-bower-grunt-build-plugin) | [thrashr888][] | |
| [Bower/Gulp](https://github.com/gdi2290/dokku-bower-gulp-build-plugin) | [gdi2290](https://github.com/gdi2290) | |
| [Elasticsearch](https://github.com/robv/dokku-elasticsearch) | [robv][] | |
| [Elasticsearch](https://github.com/jezdez/dokku-elasticsearch-plugin) | [jezdez][] | Compatible with 0.2.0 |
| [Elasticsearch](https://github.com/blag/dokku-elasticsearch-plugin)<sup>1</sup> | [blag][] | Compatible with 0.2.0 |
| [HipChat Notifications](https://github.com/cef/dokku-hipchat) | [cef][] | |
| [Graphite/statsd] (https://github.com/jlachowski/dokku-graphite-plugin) | [jlachowski][] | |
| [APT](https://github.com/F4-Group/dokku-apt) | [F4-Group][] | |
| [User ACL](https://github.com/mlebkowski/dokku-acl) | [Maciej Łebkowski](https://github.com/mlebkowski) | |
| [PrimeCache](https://github.com/darkpixel/dokku-prime-cache) | [darkpixel](https://github.com/darkpixel/) | |

<sup>1</sup> Forked from [jezdez/dokku-elasticsearch-plugin](https://github.com/jezdez/dokku-elasticsearch-plugin): uses Elasticsearch 1.2 (instead of 0.90), doesn't depend on dokku-link, runs as elasticsearch user instead of root, and turns off multicast autodiscovery for use in a VPS environment.
