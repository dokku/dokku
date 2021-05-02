# Plugins

Dokku itself is built out of plugins and uses [plugn](https://github.com/dokku/plugn) for its plugin system. In essence a plugin is a collection of scripts that will be run based on naming convention.

Let's take a quick look at the current Dokku nginx plugin that's shipped with Dokku by default.

    nginx-vhosts/
    ├── plugin.toml  # plugin metadata
    ├── commands     # contains additional commands
    ├── install      # runs on Dokku installation
    └── post-deploy  # runs after an app is deployed

## Installing a plugin

[See the plugin management documentation](/docs/advanced-usage/plugin-management.md).

## Creating your own plugin

[See the full documentation](/docs/development/plugin-creation.md).

## Official Plugins

The following plugins are available and provided by Dokku maintainers.  Please file issues against their respective issue trackers.

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [CouchDB](https://github.com/dokku/dokku-couchdb)                                                 | [dokku][]             | 0.4.0+                |
| [Elasticsearch](https://github.com/dokku/dokku-elasticsearch-plugin)                              | [dokku][]             | 0.4.0+                |
| [Grafana/Graphite/Statsd](https://github.com/dokku/dokku-graphite-grafana)                        | [dokku][]             | 0.4.0+                |
| [MariaDB](https://github.com/dokku/dokku-mariadb-plugin)                                          | [dokku][]             | 0.4.0+                |
| [Memcached](https://github.com/dokku/dokku-memcached-plugin)                                      | [dokku][]             | 0.4.0+                |
| [Mongo](https://github.com/dokku/dokku-mongo-plugin)                                              | [dokku][]             | 0.4.0+                |
| [MySQL](https://github.com/dokku/dokku-mysql-plugin)                                              | [dokku][]             | 0.4.0+                |
| [Nats](https://github.com/dokku/dokku-nats)                                                       | [dokku][]             | 0.4.0+                |
| [Postgres](https://github.com/dokku/dokku-postgres-plugin)                                        | [dokku][]             | 0.4.0+                |
| [RabbitMQ](https://github.com/dokku/dokku-rabbitmq-plugin)                                        | [dokku][]             | 0.4.0+                |
| [Redis](https://github.com/dokku/dokku-redis-plugin)                                              | [dokku][]             | 0.4.0+                |
| [RethinkDB](https://github.com/dokku/dokku-rethinkdb-plugin)                                      | [dokku][]             | 0.4.0+                |
| [Copy Files to Image](https://github.com/dokku/dokku-copyfiles-to-image)                          | [dokku][]             | 0.4.0+                |
| [HTTP Auth](https://github.com/dokku/dokku-http-auth)                                             | [dokku][]             | 0.4.0+                |
| [Let's Encrypt](https://github.com/dokku/dokku-letsencrypt)                                       | [dokku][]             | 0.4.0+                |
| [Maintenance mode](https://github.com/dokku/dokku-maintenance)                                    | [dokku][]             | 0.4.0+                |
| [Redirect](https://github.com/dokku/dokku-redirect)                                               | [dokku][]             | 0.4.0+                |
| [Registry](https://github.com/dokku/dokku-registry)                                               | [dokku][]             | 0.12.0+

## Community plugins

> Warning: The following plugins have been supplied by our community and may not have been tested by Dokku maintainers.

[256dpi]: https://github.com/256dpi
[abossard]: https://github.com/dudagroup
[ademuk]: https://github.com/ademuk
[agco-adm]: https://github.com/agco-adm
[alessio]: https://github.com/alessio
[alex-sherwin]: https://github.com/alex-sherwin
[alexanderbeletsky]: https://github.com/alexanderbeletsky
[alexkruegger]: https://github.com/alexkruegger
[Aluxian]: https://github.com/Aluxian
[Aomitayo]: https://github.com/Aomitayo
[apmorton]: https://github.com/apmorton
[artofrawr]: https://github.com/artofrawr
[badsyntax]: https://github.com/badsyntax
[basgys]: https://github.com/basgys
[Benjamin-Dobell]: https://github.com/Benjamin-Dobell
[blag]: https://github.com/blag
[cameron-martin]: https://github.com/cameron-martin
[candlewaster]: https://notabug.org/candlewaster
[cedricziel]: https://github.com/cedricziel
[cef]: https://github.com/cef
[cjblomqvist]: https://github.com/cjblomqvist
[crisward]: https://github.com/crisward
[cu12]: https://github.com/cu12
[darkpixel]: https://github.com/darkpixel
[dokku]: https://github.com/dokku
[dokku-community]: https://github.com/dokku-community
[dyson]: https://github.com/dyson
[fermuch]: https://github.com/fermuch
[fgrehm]: https://github.com/fgrehm
[Flink]: https://github.com/Flink
[fomojola]: https://github.com/fomojola
[gdi2290]: https://github.com/gdi2290
[hughfletcher]: https://github.com/hughfletcher
[iamale]: https://github.com/iamale
[ignlg]: https://github.com/ignlg
[iloveitaly]: https://github.com/iloveitaly
[investtools]: https://github.com/investtools
[iskandar]: https://github.com/iskandar
[jagandecapri]: https://github.com/jagandecapri
[jeffutter]: https://github.com/jeffutter
[jlachowski]: https://github.com/jlachowski
[josegonzalez]: https://github.com/josegonzalez
[Kloadut]: https://github.com/Kloadut
[krisrang]: https://github.com/krisrang
[luxifer]: https://github.com/luxifer
[m0rth1um]: https://github.com/m0rth1um
[Maciej Łebkowski]: https://github.com/mlebkowski
[matto1990]: https://github.com/matto1990
[mbreit]: https://github.com/mbreit
[mbriskar]: https://github.com/mbriskar
[michaelshobbs]: https://github.com/michaelshobbs
[mikecsh]: https://github.com/mikecsh
[mikexstudios]: https://github.com/mikexstudios
[mimischi]: https://github.com/mimischi
[mixxorz]: https://github.com/mixxorz
[mlebkowski]: https://github.com/mlebkowski
[motin]: https://github.com/motin
[mrname]: https://github.com/mrname
[musicglue]: https://github.com/musicglue
[neam]: https://github.com/neam
[nickcharlton]: https://github.com/nickcharlton
[nickstenning]: https://github.com/nickstenning
[nornagon]: https://github.com/nornagon
[ohardy]: https://github.com/ohardy
[pauldub]: https://github.com/pauldub
[pnegahdar]: https://github.com/pnegahdar
[RaceHub]: https://github.com/racehub
[ribot]: https://github.com/ribot
[rlaneve]: https://github.com/rlaneve
[robv]: https://github.com/robv
[scottatron]: https://github.com/scottatron
[sehrope]: https://github.com/sehrope
[sekjun9878]: https://github.com/sekjun9878
[sgulseth]: https://github.com/sgulseth
[sseemayer]: https://github.com/sseemayer
[statianzo]: https://github.com/statianzo
[stuartpb]: https://github.com/stuartpb
[thrashr888]: https://github.com/thrashr888
[wmluke]: https://github.com/wmluke
[Zenedith]: https://github.com/Zenedith
[fteychene]: https://github.com/fteychene
[sarendsen]: https://github.com/sarendsen
[baikunz]: https://github.com/baikunz
[lazyatom]: https://github.com/lazyatom
[ollej]: https://github.com/ollej

### Datastores

#### Relational

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [MariaDB](https://github.com/Kloadut/dokku-md-plugin)                                             | [Kloadut][]           | 0.3.x                 |
| [MariaDB (single container)](https://github.com/ohardy/dokku-mariadb)                             | [ohardy][]            | 0.3.x                 |
| [MariaDB (single container)](https://github.com/krisrang/dokku-mariadb)                           | [krisrang][]          | 0.3.26+               |
| [PostgreSQL](https://github.com/jlachowski/dokku-pg-plugin)                                       | [jlachowski][]        | 0.3.x                 |
| [PostgreSQL (single container)](https://github.com/ohardy/dokku-psql)                             | [ohardy][]            | 0.3.x                 |
| [PostgreSQL (single container)](https://github.com/Flink/dokku-psql-single-container)             | [Flink][]             | 0.3.26+               |

#### Caching

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [Nginx Cache](https://github.com/Aluxian/dokku-nginx-cache)                                       | [Aluxian][]           | 0.5.0+                |
| [Redis (single container)](https://github.com/ohardy/dokku-redis)                                 | [ohardy][]            | 0.3.x                 |
| [Varnish](https://github.com/Zenedith/dokku-varnish-plugin)                                       | [Zenedith][]          | Varnish cache between nginx and application with base configuration|

#### Queuing

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [RabbitMQ](https://github.com/jlachowski/dokku-rabbitmq-plugin)                                   | [jlachowski][]        | 0.3.x                 |
| [RabbitMQ (single container)](https://github.com/jlachowski/dokku-rabbitmq-single-plugin)         | [jlachowski][]        | 0.3.x                 |
| [ElasticMQ (SQS compatible)](https://github.com/cu12/dokku-elasticmq)                             | [cu12][]              | 0.5.0+                |
| [VerneMQ (MQTT Broker)](https://github.com/SpinifexGroup/dokku-vernemq)                           | [mrname][]            | 0.4.0+                |

#### Other

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [etcd](https://github.com/basgys/dokku-etcd)                                                      | [basgys][]            | 0.4.x                 |
| [FakeSNS](https://github.com/cu12/dokku-fake_sns)                                                 | [cu12][]              | 0.5.0+                |
| [InfluxDB](https://github.com/basgys/dokku-influxdb)                                              | [basgys][]            | 0.4.x                 |
| [RethinkDB](https://github.com/stuartpb/dokku-rethinkdb-plugin)                                   | [stuartpb][]          | 0.3.x                 |
| [Headless Chrome](https://github.com/lazyatom/dokku-chrome)                                       | [lazyatom][]          | 0.8.1+                |

[dccee02]: https://github.com/jeffutter/dokku-riakcs-plugin/commit/dccee02702e7001851917b7814e78a99148fb709
[c77cbf1]: https://github.com/dokku/dokku/commit/c77cbf1d3ae07f0eafb85082ed7edcae9e836147
[28de3ec]: https://github.com/dokku/dokku/commit/28de3ecaa3231a223f83fd8d03f373308673bc40

### Plugins Implementing New Dokku Functionality

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [App name as env](https://github.com/cjblomqvist/dokku-app-name-env)                              | [cjblomqvist][]       | 0.3.x                 |
| [Docker Direct](https://github.com/josegonzalez/dokku-docker-direct)                              | [josegonzalez][]      | 0.4.0+                |
| [Dokku Clone](https://github.com/crisward/dokku-clone)                                            | [crisward][]          | 0.4.0+                |
| [Dokku Copy App Config Files](https://github.com/josegonzalez/dokku-supply-config)                | [josegonzalez][]      | 0.4.0+                |
| [Dockerfile custom path](https://github.com/mimischi/dokku-dockerfile)                            | [mimischi][]          | 0.8.0+                               |
| [Dokku Require](https://github.com/crisward/dokku-require)<sup>1</sup>                            | [crisward][]          | 0.4.0+                |
| [Global Certificates](https://github.com/josegonzalez/dokku-global-cert)                          | [josegonzalez][]      | 0.5.0+                |
| [Graduate (Environment Management)](https://github.com/glassechidna/dokku-graduate)               | [Benjamin-Dobell][]   | 0.4.0+                |
| [Haproxy tcp load balancer](https://github.com/256dpi/dokku-haproxy)                              | [256dpi][]            | 0.4.0+                |
| [Hostname](https://github.com/michaelshobbs/dokku-hostname)                                       | [michaelshobbs][]     | 0.4.0+                |
| [HTTP Auth Secure Apps](https://github.com/matto1990/dokku-secure-apps)                           | [matto1990][]         | 0.4.0+                |
| [Monit (Health Checks)](https://github.com/mbreit/dokku-monit)                                    | [mbreit][]            | 0.8.0+                |
| [Nuke Containers](https://github.com/josegonzalez/dokku-nuke)                                     | [josegonzalez][]      | 0.4.0+                |
| [Open App Ports](https://github.com/josegonzalez/dokku-ports)                                     | [josegonzalez][]      | 0.3.x                 |
| [Proctype Filter](https://github.com/michaelshobbs/dokku-proctype-filter)                         | [michaelshobbs][]     | 0.4.0+                |
| [robots.txt](https://notabug.org/candlewaster/dokku-robots.txt)                                   | [candlewaster][]      | 0.4.x                 |
| [SSH Deployment Keys](https://github.com/cedricziel/dokku-deployment-keys)<sup>2</sup>            | [cedricziel][]        | 0.4.0+                |
| [SSH Hostkeys](https://github.com/cedricziel/dokku-hostkeys-plugin)<sup>3</sup>                   | [cedricziel][]        | 0.3.x                 |
| [Application build hook](https://github.com/fteychene/dokku-build-hook)                           | [fteychene][]         | 0.4.0+                 |
| [Post Deploy Script](https://github.com/baikunz/dokku-post-deploy-script)                         | [baikunz][]           | 0.4.0+                 |
| [Auto Sync](https://github.com/IdeaSynthesis/dokku-autosync)<sup>4</sup>                          | [fomojola][]          | 0.8.1+                |
| [Deploy Webhook](https://github.com/IdeaSynthesis/dokku-deploy-webhook)<sup>5</sup>               | [fomojola][]          | 0.8.1+                |

[217d00a]: https://github.com/dokku/dokku/commit/217d00a1bc47a7e24d8847617bb08a1633025fc7

<sup>1</sup> Extends app.json support to include creating volumes and creating / linking databases on push

<sup>2</sup> Adds the possibility to add SSH deployment keys to receive private hosted packages

<sup>3</sup> Adds the ability to add custom hosts to the containers known_hosts file to be able to ssh them easily (useful with deployment keys)

<sup>4</sup> Adds the ability to sync an application repo with a remote Github repo (useful for automated rebuilds without needing a git push from an external system

<sup>5</sup> Adds the ability to invoke a post-deploy webhook with the IP, port and app name, all with a single config:set).

### Other Plugins

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [Airbrake deploy](https://github.com/Flink/dokku-airbrake-deploy)                                 | [Flink][]             | 0.4.0+                |
| [APT](https://github.com/dokku-community/dokku-apt)                                               | [dokku-community][]   | 0.18.x+               |
| [Bower install](https://github.com/alexanderbeletsky/dokku-bower-install)                         | [alexanderbeletsky][] | 0.3.x                 |
| [Bower/Grunt](https://github.com/thrashr888/dokku-bower-grunt-build-plugin)                       | [thrashr888][]        | 0.3.x                 |
| [Bower/Gulp](https://github.com/gdi2290/dokku-bower-gulp-build-plugin)                            | [gdi2290][]           | 0.3.x                 |
| [Bower/Gulp](https://github.com/jagandecapri/dokku-bower-gulp-build-plugin)                       | [jagandecapri][]      | 0.3.x                 |
| [Builders: bower, compass, gulp, grunt](https://github.com/ignlg/dokku-builders-plugin)           | [ignlg][]             | 0.4.0+                |
| [Chef cookbook](https://github.com/nickcharlton/dokku-cookbook)                                   | [nickcharlton][]      |                       |
| [Docker auto persist volumes](https://github.com/Flink/dokku-docker-auto-volumes)                 | [Flink][]             | 0.4.0+                |
| [Hostname](https://github.com/michaelshobbs/dokku-hostname)                                       | [michaelshobbs][]     | 0.4.0+                |
| [Limit (Resource management)](https://github.com/sarendsen/dokku-limit)                           | [sarendsen][]         | 0.9.0+                |
| [Logspout](https://github.com/michaelshobbs/dokku-logspout)                                       | [michaelshobbs][]     | 0.4.0+                |
| [Syslog](https://github.com/michaelshobbs/dokku-syslog)                                           | [michaelshobbs][]     | 0.10.4+               |
| [Long Timeout](https://github.com/investtools/dokku-long-timeout-plugin)                          | [investtools][]       | 0.4.0+                |
| [Monit](https://github.com/cjblomqvist/dokku-monit)                                               | [cjblomqvist][]       | 0.3.x                 |
| [Monorepo](https://github.com/iamale/dokku-monorepo)                                              | [iamale][]            | 0.4.0+                |
| [Docker Monorepo](https://github.com/tianhuil/dokku-docker-monorepo)                              | [tianhuil][]          | 0.1.0+                |
| [Multi Dockerfile](https://github.com/artofrawr/dokku-multi-dockerfile)                           | [artofrawr][]         | 0.4.0+                |
| [Node](https://github.com/ademuk/dokku-nodejs)                                                    | [ademuk][]            | 0.3.x                 |
| [Node](https://github.com/pnegahdar/dokku-node)                                                   | [pnegahdar][]         | 0.3.x                 |
| [Rollbar](https://github.com/iloveitaly/dokku-rollbar)                                            | [iloveitaly][]        | 0.5.0+                |
| [Slack Notifications](https://github.com/ribot/dokku-slack)                                       | [ribot][]             | 0.4.0+                |
| [Telegram Notifications](https://github.com/m0rth1um/dokku-telegram)                              | [m0rth1um][]          | 0.4.0+                |
| [Tor](https://github.com/michaelshobbs/dokku-tor)                                                 | [michaelshobbs][]     | 0.4.0+                |
| [User ACL](https://github.com/dokku-community/dokku-acl)                                          | [Maciej Łebkowski][]  | 0.4.0+                |
| [Webhooks](https://github.com/nickstenning/dokku-webhooks)                                        | [nickstenning][]      | 0.3.x                 |
| [Wkhtmltopdf](https://github.com/mbriskar/dokku-wkhtmltopdf)                                      | [mbriskar][]          | 0.4.0+                |
| [Dokku Wordpress](https://github.com/dokku-community/dokku-wordpress)                             | [dokku-community][]      | 0.4.0+                |
| [Access](https://github.com/mainto/dokku-access)                                                  | [mainto](https://github.com/mainto)            | 0.4.0+                |
| [Dokku Nginx Trust Proxy](https://github.com/kingsquare/dokku-nginx-vhost-trustproxy)             | [kingsquare](https://github.com/kingsquare)  | 0.4.0+ |
| [Fonts](https://github.com/ollej/dokku-fonts)             | [ollej]  | 0.19.11+ |
| [Discourse](https://github.com/badsyntax/dokku-discourse)                                         | [badsyntax][]         | 0.21.4+               |


### Deprecated Plugins

The following plugins have been removed as their functionality is now in Dokku Core.

| Plugin                                                                                            | Author                | In Dokku Since                            |
| ------------------------------------------------------------------------------------------------- | --------------------- | ----------------------------------------- |
| [App User](https://github.com/michaelshobbs/dokku-app-user)                                       | [michaelshobbs][]     | v0.7.1 (herokuish 0.3.18)                 |
| [Custom Domains](https://github.com/neam/dokku-custom-domains)                                    | [motin][]             | v0.3.10 (domains plugin)                  |
| [Debug](https://github.com/josegonzalez/dokku-debug)                                              | [josegonzalez][]      | v0.3.9 (trace command)                    |
| [Docker Options](https://github.com/dyson/dokku-docker-options)                                   | [dyson][]             | v0.3.17 (docker-options plugin)           |
| [Dokku Name](https://github.com/alex-sherwin/dokku-name)                                          | [alex-sherwin][]      | v0.4.2 (named containers plugin)          |
| [Events Logger](https://github.com/alessio/dokku-events)                                          | [alessio][]           | v0.3.21 (events plugin)                   |
| [git rev-parse HEAD in env](https://github.com/dokku-community/dokku-git-rev)                     | [cjblomqvist][]       | v0.12.0 (enhanced core git plugin)        |
| [Host Port binding](https://github.com/stuartpb/dokku-bind-port)                                  | [stuartpb][]          | v0.3.17 (docker-options plugin)           |
| [Link Containers](https://github.com/rlaneve/dokku-link)                                          | [rlaneve][]           | v0.3.17 (docker-options plugin)           |
| [List Containers](https://github.com/josegonzalez/dokku-list)                                     | [josegonzalez][]      | v0.3.14 (ps plugin)                       |
| [Multi-Buildpack](https://github.com/pauldub/dokku-multi-buildpack)                               | [pauldub][]           | v0.4.0 (herokuish)                        |
| [Multiple Domains](https://github.com/wmluke/dokku-domains-plugin)<sup>1</sup>                    | [wmluke][]            | v0.3.10 (domains plugin)                  |
| [Named-containers](https://github.com/Flink/dokku-named-containers)                               | [Flink][]             | v0.4.2 (named-containers plugin)          |
| [Nginx-Alt](https://github.com/mikexstudios/dokku-nginx-alt)                                      | [mikexstudios][]      | v0.3.10 (domains plugin)                  |
| [Persistent Storage](https://github.com/dyson/dokku-persistent-storage)                           | [dyson][]             | v0.3.17 (docker-options plugin)           |
| [Pre-Deploy Tasks](https://github.com/michaelshobbs/dokku-app-predeploy-tasks)                    | [michaelshobbs][]     | v0.5.0 (deployment tasks)                 |
| [PrimeCache](https://github.com/darkpixel/dokku-prime-cache)                                      | [darkpixel][]         | v0.3.0 (zero downtime deploys)            |
| [Process Manager: Circus](https://github.com/apmorton/dokku-circus)                               | [apmorton][]          | v0.3.14/0.7.0 (ps/restart policy plugin)  |
| [Process Manager: Forego](https://github.com/Flink/dokku-forego)                                  | [Flink][]             | v0.3.14/0.7.0 (ps plugin)                 |
| [Process Manager: Forego](https://github.com/iskandar/dokku-forego)                               | [iskandar][]          | v0.3.14/0.7.0 (ps plugin)                 |
| [Process Manager: Logging Supervisord](https://github.com/sehrope/dokku-logging-supervisord)      | [sehrope][]           | v0.3.14/0.7.0 (ps plugin)                 |
| [Process Manager: Shoreman ](https://github.com/statianzo/dokku-shoreman)                         | [statianzo][]         | v0.3.14/0.7.0 (ps plugin)                 |
| [Process Manager: Supervisord](https://github.com/statianzo/dokku-supervisord)                    | [statianzo][]         | v0.3.14/0.7.0 (ps plugin)                 |
| [Rebuild application](https://github.com/scottatron/dokku-rebuild)                                | [scottatron][]        | v0.3.14 (ps plugin)                       |
| [Reset mtime](https://github.com/mixxorz/dokku-docker-reset-mtime)                                | [mixxorz][]           | Docker 1.8+                               |
| [Supply env vars to buildpacks](https://github.com/cameron-martin/dokku-build-env)<sup>2</sup>    | [cameron-martin][]    | v0.3.9 (build-env plugin)                 |
| [user-env-compile](https://github.com/motin/dokku-user-env-compile)<sup>2</sup>                   | [motin][]             | v0.3.9 (build-env plugin)                 |
| [user-env-compile](https://github.com/musicglue/dokku-user-env-compile)<sup>2</sup>               | [musicglue][]         | v0.3.9 (build-env plugin)                 |
| [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration)    | [motin][]             | v0.3.10 (domains plugin)                  |
| [Volume (persistent storage)](https://github.com/ohardy/dokku-volume)                             | [ohardy][]            | v0.5.0 (storage plugin)                   |

<sup>1</sup> Conflicts with [VHOSTS Custom Configuration](https://github.com/neam/dokku-nginx-vhosts-custom-configuration)
<sup>2</sup> Similar to the heroku-labs feature (see https://devcenter.heroku.com/articles/labs-user-env-compile)

[a043e98]: https://github.com/stuartpb/dokku-bind-port/commit/a043e9892f4815b6525c850131e09fd64db5c1fa


### Unmaintained Plugins

The following plugins are no longer maintained by their developers.

| Plugin                                                                                            | Author                | Compatibility         |
| ------------------------------------------------------------------------------------------------- | --------------------- | --------------------- |
| [app-url](https://github.com/mikecsh/dokku-app-url)                                               | [mikecsh][]           | Works with 0.2.0      |
| [Chef cookbooks](https://github.com/fgrehm/chef-dokku)                                            | [fgrehm][]            |                       |
| [CouchDB (multi containers)](https://github.com/Flink/dokku-couchdb-multi-containers)             | [Flink][]             | 0.4.0+                |
| [CouchDB](https://github.com/racehub/dokku-couchdb-plugin)                                        | [RaceHub][]           | Compatible with 0.2.0 |
| [Dokku Copy App Config Files](https://github.com/alexkruegger/dokku-app-configfiles)              | [alexkruegger][]      | Compatible with 0.3.17+ |
| [Dokku Registry](https://github.com/agco-adm/dokku-registry)                                      | [agco-adm][]          | 0.4.0+
| [Elasticsearch](https://github.com/robv/dokku-elasticsearch)                                      | [robv][]              | Not compatible with >= 0.3.0 (still uses /home/git) |
| [Elasticsearch](https://github.com/blag/dokku-elasticsearch-plugin)<sup>1</sup>                   | [blag][]              | Compatible with 0.2.0 |
| [Graphite/statsd](https://github.com/jlachowski/dokku-graphite-plugin)                            | [jlachowski][]        | < 0.4.0               |
| [HipChat Notifications](https://github.com/cef/dokku-hipchat)                                     | [cef][]               |                       |
| [Memcached](https://github.com/Flink/dokku-memcached-plugin)                                      | [Flink][]             | 0.4.0+                |
| [MongoDB (single container)](https://github.com/jeffutter/dokku-mongodb-plugin)                   | [jeffutter][]         |                       |
| [MySQL](https://github.com/hughfletcher/dokku-mysql-plugin)                                       | [hughfletcher][]      |                       |
| [Neo4j](https://github.com/Aomitayo/dokku-neo4j-plugin)                                           | [Aomitayo][]          |                       |
| [PostGIS](https://github.com/fermuch/dokku-pg-plugin)                                             | [fermuch][]           |                       |
| [PostgreSQL (single container)](https://github.com/jeffutter/dokku-postgresql-plugin)             | [jeffutter][]         | This plugin creates a single postgresql container that all your apps can use. Thus only one instance of postgresql running (good for servers without a ton of memory). |
| [RiakCS (single container)](https://github.com/jeffutter/dokku-riakcs-plugin)                     | [jeffutter][]         | Incompatible with 0.2.0 (checked at [dccee02][]) |
| [Redis](https://github.com/luxifer/dokku-redis-plugin)                                            | [luxifer][]           |                       |
| [Redis](https://github.com/sekjun9878/dokku-redis-plugin)                                         | [sekjun9878][]        | 0.3.26+               |

<sup>1</sup> Forked from [jezdez/dokku-elasticsearch-plugin](https://github.com/jezdez/dokku-elasticsearch-plugin): uses Elasticsearch 1.2 (instead of 0.90), doesn't depend on dokku-link, runs as elasticsearch user instead of root, and turns off multicast autodiscovery for use in a VPS environment.
