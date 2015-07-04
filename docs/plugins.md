# Plugins

Dokku itself is built out of plugins and uses [pluginhook](https://github.com/progrium/pluginhook) for its plugin system. In essence a plugin is a collection of scripts that will be run based on naming convention.

Let's take a quick look at the current dokku nginx plugin that's shipped with dokku by default.

    nginx-vhosts/
    ├── commands     # contains additional commands
    ├── install      # runs on dokku installation
    └── post-deploy  # runs after an app is deployed

## Installing a plugin

```shell
cd /var/lib/dokku/plugins
git clone <git url>
dokku plugins-install
```

> todo: add a command to dokku to install a plugin, given a git repository `dokku plugin:install <git url>`?

## Creating your own plugin

[See the full documentation](http://progrium.viewdocs.io/dokku/development/plugin-creation).

## Community plugins

Note: The following plugins have been supplied by our community and may not have been tested by dokku maintainers.

[agco-adm]: https://github.com/agco-adm
[ademuk]: https://github.com/ademuk
[alessio]: https://github.com/alessio
[alex-sherwin]: https://github.com/alex-sherwin
[alexanderbeletsky]: https://github.com/alexanderbeletsky
[Aomitayo]: https://github.com/Aomitayo
[apmorton]: https://github.com/apmorton
[blag]: https://github.com/blag
[cameron-martin]: https://github.com/cameron-martin
[cedricziel]: https://github.com/cedricziel
[cef]: https://github.com/cef
[cjblomqvist]: https://github.com/cjblomqvist
[darkpixel]: https://github.com/darkpixel
[dyson]: https://github.com/dyson
[F4-Group]: https://github.com/F4-Group
[fermuch]: https://github.com/fermuch
[fgrehm]: https://github.com/fgrehm
[gdi2290]: https://github.com/gdi2290
[heichblatt]: https://github.com/heichblatt
[hughfletcher]: https://github.com/hughfletcher
[iskandar]: https://github.com/iskandar
[jeffutter]: https://github.com/jeffutter
[jezdez]: https://github.com/jezdez
[jlachowski]: https://github.com/jlachowski
[krisrang]: https://github.com/krisrang
[Kloadut]: https://github.com/Kloadut
[luxifer]: https://github.com/luxifer
[mlebkowski]: https://github.com/mlebkowski
[matto1990]: https://github.com/matto1990
[michaelshobbs]: https://github.com/michaelshobbs
[mikecsh]: https://github.com/mikecsh
[mikexstudios]: https://github.com/mikexstudios
[motin]: https://github.com/motin
[musicglue]: https://github.com/musicglue
[neam]: https://github.com/neam
[nickstenning]: https://github.com/nickstenning
[nornagon]: https://github.com/nornagon
[ohardy]: https://github.com/ohardy
[pauldub]: https://github.com/pauldub
[pnegahdar]: https://github.com/pnegahdar
[RaceHub]: https://github.com/racehub
[rlaneve]: https://github.com/rlaneve
[robv]: https://github.com/robv
[scottatron]: https://github.com/scottatron
[sehrope]: https://github.com/sehrope
[statianzo]: https://github.com/statianzo
[stuartpb]: https://github.com/stuartpb
[thrashr888]: https://github.com/thrashr888
[wmluke]: https://github.com/wmluke
[Zenedith]: https://github.com/Zenedith
[sekjun9878]: https://github.com/sekjun9878
[Flink]: https://github.com/Flink
[ribot]: https://github.com/ribot
[Benjamin-Dobell]: https://github.com/Benjamin-Dobell
[jagandecapri]: https://github.com/jagandecapri

### Datastores

#### Relational

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [MariaDB](https://github.com/Kloadut/dokku-md-plugin)                                             | [Kloadut][]           | Compatible with 0.2.0 |
| [MariaDB (single container)](https://github.com/ohardy/dokku-mariadb)                             | [ohardy][]            | Compatible with 0.2.0 |
| [MySQL](https://github.com/hughfletcher/dokku-mysql-plugin)                                       | [hughfletcher][]      |                       |
| [PostgreSQL](https://github.com/Kloadut/dokku-pg-plugin)                                          | [Kloadut][]           | Compatible with 0.2.0 |
| [PostgreSQL](https://github.com/jezdez/dokku-postgres-plugin)                                     | [jezdez][]            | Compatible with 0.2.0 |
| [PostgreSQL](https://github.com/jlachowski/dokku-pg-plugin)                                       | [jlachowski][]        | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [PostgreSQL (single container)](https://github.com/jeffutter/dokku-postgresql-plugin)             | [jeffutter][]         | This plugin creates a single postgresql container that all your apps can use. Thus only one instance of postgresql running (good for servers without a ton of memory). |
| [PostgreSQL (single container)](https://github.com/ohardy/dokku-psql)                             | [ohardy][]            | Compatible with 0.2.0 |
| [PostgreSQL (single container)](https://github.com/Flink/dokku-psql-single-container)             | [Flink][]             | Single Postgresql container with official Postgresql docker image. Compatible with 0.3.16 |
| [PostGIS](https://github.com/fermuch/dokku-pg-plugin)                                             | [fermuch][]           |                       |

#### Caching

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [Memcached](https://github.com/jezdez/dokku-memcached-plugin)                                     | [jezdez][]            | Compatible with 0.2.0 |
| [Memcached](https://github.com/jlachowski/dokku-memcached-plugin)                                 | [jlachowski][]        | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [Redis](https://github.com/jezdez/dokku-redis-plugin)                                             | [jezdez][]            | Requires https://github.com/rlaneve/dokku-link; compatible with 0.2.0 |
| [Redis](https://github.com/luxifer/dokku-redis-plugin)                                            | [luxifer][]           |                       |
| [Redis](https://github.com/sekjun9878/dokku-redis-plugin)                                         | [sekjun9878][]        | A better Redis plugin with automatic instance creation and Dokku Link support
| [Redis (single container)](https://github.com/ohardy/dokku-redis)                                 | [ohardy][]            | Compatible with 0.2.0 |
| [Varnish](https://github.com/Zenedith/dokku-varnish-plugin)                                       | [Zenedith][]          | Varnish cache between nginx and application with base configuration|

#### Queuing

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [RabbitMQ](https://github.com/jlachowski/dokku-rabbitmq-plugin)                                   | [jlachowski][]        | IP & PORT available directly in linked app container env variables (requires link plugin)|
| [RabbitMQ (single container)](https://github.com/jlachowski/dokku-rabbitmq-single-plugin)         | [jlachowski][]        | IP & PORT available directly in linked app container env variables (requires link plugin)|

#### Other

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [CouchDB](https://github.com/racehub/dokku-couchdb-plugin)                                        | [RaceHub][]           | Compatible with 0.2.0 |
| [Elasticsearch](https://github.com/robv/dokku-elasticsearch)                                      | [robv][]              | Not compatible with >= 0.3.0 (still uses /home/git) |
| [Elasticsearch](https://github.com/jezdez/dokku-elasticsearch-plugin)                             | [jezdez][]            | Compatible with 0.2.0 to 0.3.13 |
| [Elasticsearch](https://github.com/blag/dokku-elasticsearch-plugin)<sup>1</sup>                   | [blag][]              | Compatible with 0.2.0 |
| [MongoDB (single container)](https://github.com/jeffutter/dokku-mongodb-plugin)                   | [jeffutter][]         |                       |
| [Neo4j](https://github.com/Aomitayo/dokku-neo4j-plugin)                                           | [Aomitayo][]          |                       |
| [RethinkDB](https://github.com/stuartpb/dokku-rethinkdb-plugin)                                   | [stuartpb][]          | 2014-02-22: targeting dokku @ [latest][217d00a]; will fail with Dokku earlier than [28de3ec][]. |
| [RiakCS (single container)](https://github.com/jeffutter/dokku-riakcs-plugin)                     | [jeffutter][]         | Incompatible with 0.2.0 (checked at [dccee02][]) |

[dccee02]: https://github.com/jeffutter/dokku-riakcs-plugin/commit/dccee02702e7001851917b7814e78a99148fb709

### Process Managers

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [Circus](https://github.com/apmorton/dokku-circus)                                                | [apmorton][]          |                       |
| [Forego](https://github.com/iskandar/dokku-forego)                                                | [iskandar][]          | Compatible with 0.2.x |
| [Logging Supervisord](https://github.com/sehrope/dokku-logging-supervisord)                       | [sehrope][]           | Works with dokku @ [c77cbf1][] - no 0.2.0 compatibility |
| [Monit](https://github.com/cjblomqvist/dokku-monit)                                               | [cjblomqvist][]       |                       |
| [Shoreman ](https://github.com/statianzo/dokku-shoreman)                                          | [statianzo][]         | Compatible with 0.2.0 |
| [Supervisord](https://github.com/statianzo/dokku-supervisord)                                     | [statianzo][]         | Compatible with 0.2.0 |

[c77cbf1]: https://github.com/progrium/dokku/commit/c77cbf1d3ae07f0eafb85082ed7edcae9e836147
[28de3ec]: https://github.com/progrium/dokku/commit/28de3ecaa3231a223f83fd8d03f373308673bc40

### Dokku Features

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [app-url](https://github.com/mikecsh/dokku-app-url)                                               | [mikecsh][]           | Works with 0.2.0      |
| [Docker Direct](https://github.com/heichblatt/dokku-docker-direct)                                | [heichblatt][]        |                       |
| [Dokku Name](https://github.com/alex-sherwin/dokku-name)                                          | [alex-sherwin][]      | dokku >= [c77cbf1][]  |
| [Dokku Registry](https://github.com/agco-adm/dokku-registry)<sup>1</sup>                          | [agco-adm][]          |                       |
| [git rev-parse HEAD in env](https://github.com/cjblomqvist/dokku-git-rev)                         | [cjblomqvist][]       | Compatible with 0.3.0 |
| [Graduate (Environment Management)](https://github.com/glassechidna/dokku-graduate)               | [Benjamin-Dobell][]   | dokku >= v0.3.14      |
| [HTTP Auth](https://github.com/Flink/dokku-http-auth)                                             | [Flink][]             |                       |
| [HTTP Auth Secure Apps](https://github.com/matto1990/dokku-secure-apps)                           | [matto1990][]         | Works with v0.2.3     |
| [Hostname](https://github.com/michaelshobbs/dokku-hostname)                                       | [michaelshobbs][]     |                       |
| [Link Containers](https://github.com/rlaneve/dokku-link)                                          | [rlaneve][]           | dokku >= [c77cbf1][]  |
| [Maintenance mode](https://github.com/Flink/dokku-maintenance)                                    | [Flink][]             |                       |
| [Multi-Buildpack](https://github.com/pauldub/dokku-multi-buildpack)                               | [pauldub][]           |                       |
| [Ports](https://github.com/heichblatt/dokku-ports)                                                | [heichblatt][]        |                       |
| [Pre-Deploy Tasks](https://github.com/michaelshobbs/dokku-app-predeploy-tasks)                    | [michaelshobbs][]     |                       |
| [SSH Deployment Keys](https://github.com/cedricziel/dokku-deployment-keys)<sup>2</sup>            | [cedricziel][]        | 2014-01-17: compatible with upstream/master |
| [SSH Hostkeys](https://github.com/cedricziel/dokku-hostkeys-plugin)<sup>3</sup>                   | [cedricziel][]        | 2014-01-17: compatible with upstream/master |
| [Volume (persistent storage)](https://github.com/ohardy/dokku-volume)                             | [ohardy][]            | Compatible with 0.2.0 |

[217d00a]: https://github.com/progrium/dokku/commit/217d00a1bc47a7e24d8847617bb08a1633025fc7

<sup>1</sup> On Heroku similar functionality is offered by the [heroku-labs pipeline feature](https://devcenter.heroku.com/articles/labs-pipelines), which allows you to promote builds across multiple environments (staging -> production)

<sup>2</sup> Adds the possibility to add SSH deployment keys to receive private hosted packages

<sup>3</sup> Adds the ability to add custom hosts to the containers known_hosts file to be able to ssh them easily (useful with deployment keys)

### Other Plugins

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [Airbrake deploy](https://github.com/Flink/dokku-airbrake-deploy)                                 | [Flink][]             |                       |
| [APT](https://github.com/F4-Group/dokku-apt)                                                      | [F4-Group][]          |                       |
| [Chef cookbooks](https://github.com/fgrehm/chef-dokku)                                            | [fgrehm][]            |                       |
| [Bower install](https://github.com/alexanderbeletsky/dokku-bower-install)                         | [alexanderbeletsky][] |                       |
| [Bower/Grunt](https://github.com/thrashr888/dokku-bower-grunt-build-plugin)                       | [thrashr888][]        |                       |
| [Bower/Gulp](https://github.com/gdi2290/dokku-bower-gulp-build-plugin)                            | [gdi2290][]           |                       |
| [Bower/Gulp](https://github.com/jagandecapri/dokku-bower-gulp-build-plugin)                       | [jagandecapri][]      |                       |
| [HipChat Notifications](https://github.com/cef/dokku-hipchat)                                     | [cef][]               |                       |
| [Graphite/statsd](https://github.com/jlachowski/dokku-graphite-plugin)                            | [jlachowski][]        |                       |
| [Logspout](https://github.com/michaelshobbs/dokku-logspout)                                       | [michaelshobbs][]     |                       |
| [Node](https://github.com/pnegahdar/dokku-node)                                                   | [pnegahdar][]         |                       |
| [Node](https://github.com/ademuk/dokku-nodejs)                                                    | [ademuk][]            |                       |
| [Rails logs](https://github.com/Flink/dokku-rails-logs)                                           | [Flink][]             |                       |
| [Slack Notifications](https://github.com/ribot/dokku-slack)                                       | [ribot][]             |                       |
| [User ACL](https://github.com/mlebkowski/dokku-acl)                                               | [Maciej Łebkowski][]  |                       |
| [Webhooks](https://github.com/nickstenning/dokku-webhooks)                                        | [nickstenning][]      |                       |
| [Wordpress](https://github.com/dudagroup/dokku-wordpress-template)                                | [abossard][]          | Dokku dev, mariadb, volume, domains |

<sup>1</sup> Forked from [jezdez/dokku-elasticsearch-plugin](https://github.com/jezdez/dokku-elasticsearch-plugin): uses Elasticsearch 1.2 (instead of 0.90), doesn't depend on dokku-link, runs as elasticsearch user instead of root, and turns off multicast autodiscovery for use in a VPS environment.

### Deprecated Plugins

The following plugins have been removed as their functionality is now in Dokku Core.

| Plugin                                                                                            | Author                | In Dokku Since                  |
| ------------------------------------------------------------------------------------------------- | --------------------- | ------------------------------- |
| [Custom Domains](https://github.com/neam/dokku-custom-domains)                                    | [motin][]             | v0.3.10 (domains plugin)        |
| [Debug](https://github.com/heichblatt/dokku-debug)                                                | [heichblatt][]        | v0.3.9 (trace command)          |
| [Docker Options](https://github.com/dyson/dokku-docker-options)                                   | [dyson][]             | v0.3.17 (docker-options plugin) |
| [Events Logger](https://github.com/alessio/dokku-events)                                          | [alessio][]           | v0.3.21 (events plugin)         |
| [Host Port binding](https://github.com/stuartpb/dokku-bind-port)                                  | [stuartpb][]          | v0.3.17 (docker-options plugin) |
| [Multiple Domains](https://github.com/wmluke/dokku-domains-plugin)<sup>1</sup>                    | [wmluke][]            | v0.3.10 (domains plugin)        |
| [Nginx-Alt](https://github.com/mikexstudios/dokku-nginx-alt)                                      | [mikexstudios][]      | v0.3.10 (domains plugin)        |
| [Persistent Storage](https://github.com/dyson/dokku-persistent-storage)                           | [dyson][]             | v0.3.17 (docker-options plugin) |
| [PrimeCache](https://github.com/darkpixel/dokku-prime-cache)                                      | [darkpixel][]         | v0.3.0 (zero downtime deploys)  |
| [Rebuild application](https://github.com/scottatron/dokku-rebuild)                                | [scottatron][]        | v0.3.14 (ps plugin)             |
| [Supply env vars to buildpacks](https://github.com/cameron-martin/dokku-build-env)<sup>2</sup>    | [cameron-martin][]    | v0.3.9 (build-env plugin)       |
| [user-env-compile](https://github.com/musicglue/dokku-user-env-compile)<sup>2</sup>               | [musicglue][]         | v0.3.9 (build-env plugin)       |
| [user-env-compile](https://github.com/motin/dokku-user-env-compile)<sup>2</sup>                   | [motin][]             | v0.3.9 (build-env plugin)       |
| [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration)    | [motin][]             | v0.3.10 (domains plugin)        |


<sup>1</sup> Conflicts with [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration)
<sup>2</sup> Similar to the heroku-labs feature (see https://devcenter.heroku.com/articles/labs-user-env-compile)

[a043e98]: https://github.com/stuartpb/dokku-bind-port/commit/a043e9892f4815b6525c850131e09fd64db5c1fa
