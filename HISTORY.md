# History

## 0.4.4

This release adds a few interesting changes:

- The `dokku logs` command now roughly maps to the `heroku logs` command, and supports most available options.
- Native Microsoft Azure support is now available!
- Quite a few shellcheck issues were fixed thanks to @callahad!
- Experimental debian installation support. Going forward, we will try and make dokku compatible with all systemd installations, as well as investigate dockerfile-based deployment to continue simplifying installation.

Thanks to all our contributors for making this release great!

### Bug Fixes

- #1606: @josegonzalez Install plugn 0.2.0 in Makefile installs
- #1643: @Flink Fix generated nginx config when NO_VHOST=1
- #1644: @mmerickel Watch dokku events through a logrotate
- #1647: @callahad Resolve SC2115: 'Use "${var:?}" to ensure this never expands to /'
- #1648: @callahad Resolve SC2154: 'variable is referenced but not assigned'
- #1649: @callahad Resolve SC2164: 'Use cd ... || exit in case cd fails.'
- #1650: @callahad Resolve SC2148: 'target shell is unknown'
- #1651: @callahad Resolve SC2029: 'this expands on client side'
- #1652: @callahad Resolve SC2143: 'Instead of [ -n $(foo | grep bar) ], use foo | grep -q bar'
- #1653: @callahad Resolve SC2145: 'Argument mixes string and array.'
- #1655: @callahad Resolve SC2162: 'read without -r mangles backslashes'
- #1656: @callahad Resolve SC2001: 'See if you can use ${var//search/replace} instead'
- #1660: @callahad Fixup debian/control for Debian
- #1662: @xadh00m Only return users of group 'adm'

### New Features

- #1607: @josegonzalez Dokku support for Debian Jessie installation
- #1610: @kdomanski Add post-stop plugn trigger
- #1612: @Flink Add multiple options to the logs plugin
- #1613: @kdomanski Add ps:restore to start applications which weren't manually stopped
- #1628: @michaelshobbs Move RESTORE variable to DOKKU_APP namespace
- #1634: @callahad Allow installation of bats via homebrew
- #1645: @rvalyi Add ability to access git repo via ssh
- #1664: @michaelshobbs Add $REV to pre-receive-app call

### Documentation

- #1605: @josegonzalez Make commented output a bit more readable
- #1621: @josegonzalez Document `--force` option for `:destroy` commands
- #1623: @sedouard Add Azure Documentation
- #1624: @u2mejc Add trace to help output
- #1626: @josegonzalez Add official dokku-copy-files-to-image plugin
- #1630: @adamwolf Enabling tracing is actually 'dokku trace on'
- #1633: @elia Add a warning regarding the use of `trace on`
- #1635: @callahad Remove deprecated Linode stackscript
- #1657: @kimausloos Various small doc updates

## 0.4.3

This release was mainly a documentation release/bugfix.

One major removal was is that as of 0.4.3, we no longer restart containers automatically on failure via docker. This feature was introduced in 0.4.0, and caused issues with duplicate containers being started on server reboot. Until the docker api for container restarts stabilizes, we will not be able to provide this functionality within Dokku.

If desired, you may replicate this functionality using the official `docker-options` plugin.

### Bug Fixes

- #1560: @darklow Fixes issue where SSL and non-SSL templates cannot be used at the same time
- #1566: @michaelshobbs Fix logic error in enabling nginx
- #1568: @josegonzalez Remove 'connection closed' messages from dokku ssh client
- #1574: @josegonzalez Ensure the user has a valid hostname set during installation
- #1585: @michaelshobbs Ensure dokku can read nginx logs and don't remove other perms
- #1589: @michaelshobbs Patch broken nginx 1.8.0 logrotate script
- #1591: @michaelshobbs Remove docker restart policy until the docker api stabilizes
- #1603: @josegonzalez Add missing plugin triggers
Quiet client #1568

### New Features

- #1490: @vijayviji Add windows-specific vagrant setup
- #1563: @kdomanski Cleanup all dead containers in dokku cleanup
- #1590: @michaelshobbs Trigger docker-args-build for dockerfile deployments
- #1600: @josegonzalez Upgrade to Herokuish 0.3.4
- #1602: @josegonzalez Add pre-receive-app plugin trigger

### Documentation

- #1556: @josegonzalez Use proper cdn link for css asset
- #1557: @josegonzalez Code styling changes
- #1561: @josegonzalez Set dokku-acl compatibility to 0.4.0+
- #1562: @josegonzalez Documentation Overhaul
- #1573: @mateusortiz Add link to license
- #1577: @Flink Add official redirect plugin
- #1587: @josegonzalez Update "reporting other issues" to include `docker inspect`
- #1598: @adamwolf Add missing bootstrap.sh step
- #1599: @ssd532 Fix a few grammatical mistakes
- #1601: @ojacquemart Fix typo
- #1604: @josegonzalez Add every type of favicon to all templates

## 0.4.2

This release was mainly a documentation release, with a few notable features:

- You can now use the commercial version of docker-engine with dokku.
- You can now name your containers using the official `named-containers` plugin

Huge thanks to @Flink for working on our official plugins and adding official [basic auth](https://github.com/dokku/dokku-http-auth), [couchdb](https://github.com/dokku/dokku-couchdb), and [site maintenance](https://github.com/dokku/dokku-maintenance) plugins!

### Bug Fixes

- #1530: @Flink Fix nginx configuration for SSL

### New Features

- #1515: @leonardodino Allow local prebuilt stack sourcing
- #1536 #1537: @josegonzalez Add docker-engine-cs as docker-engine alternative
- #1511: @Flink Add `named-containers` as a core plugin

### Documentation

- #1520: @kimausloos Add `dokku` command-prefix to `plugin:install` command
- #1519: @3onyc Fix typos in documentation
- #1521: @edm00se Use correct url to bootstrap.sh in README.md
- #1522: @josegonzalez Avoid redirects and use raw.githubusercontent.com for github links
- #1523 #1548: @callahad Make room for the longer bootstrap.sh URL
- #1524: @callahad Document idiosyncracies with Linode
- #1529: @pzula Adds helpful information regarding whitespacing when setting config
- #1525 #1550: @josegonzalez Add gratipay shield to readme
- #1544 #1545 #1547 #1551: @josegonzalez @Flink Update compatibility for community plugins
- #1546: @josegonzalez Add missing description to history output in HISTORY.md
- #1552 #1553 #1555: @josegonzalez Various documentation styling tweaks

## 0.4.1

This release is primarily a bugfix and documentation update. In 0.4.0, we incorrectly handled setting environment variables in certain cases, resulting in misconfigured applications. We recommend that all users upgrade from 0.4.1 as soon as possible.

One new feature is colorized logging output, which should make it easier to debug application logging output when deploying multiple processes.

### Bug Fixes

- #1494: @callahad Correctly install packages for DOKKU_TAG=v0.4.0
- #1496: @callahad Don't prompt users when installing dokku package
- #1514: @michaelshobbs Do not use exit 0 in config functions and fix CURL environment variable setting

### New Features

- #1482: @Flink Strip the `dokku-` part from plugins on install
- #1500: @jazzzz Log user name and fingerprint in events
- #1512: @Flink Colorize output from logs

### Documentation

- #1485: @matto1990 Update Slack plugin compatability
- #1488: @josegonzalez Update plugins list with compatibility changes
- #1491: @josegonzalez Promote [maintenance](https://github.com/dokku/dokku-maintenance) and [http basic auth](https://github.com/dokku/dokku-http-auth) plugins to official status
- #1492: @josegonzalez Upgrade hostname plugin to 0.4.0+ compatibility
- #1501: @josegonzalez Clarify bootstrap installation documentation
- #1502: @josegonzalez Update dokku-apt compatibility
- #1504: @michaelshobbs Change plugin install doc to show one-step method

## 0.4.0

This is our first minor release in almost a year. Many new features and removals have occurred, so here is a neat summary:

- Plugins are now triggered via `plugn`. Notably, you'll need add a `plugin.toml` to describe the plugin as well as use `plugn trigger` instead of `pluginhook` to trigger plugin callbacks. Please see the [plugin creation documentation](http://progrium.viewdocs.io/dokku/development/plugin-creation/) for more details.
- A few new official plugins have been added to the core, including [image tagging](http://progrium.viewdocs.io/dokku/application-deployment/), [certificate management](http://progrium.viewdocs.io/dokku/deployment/ssl-configuration/), a tar-based deploy solution, and much more. Check out the *New Features* section for more details.
- We've removed a few deprecated plugin callbacks. Please see the [plugin triggers documentation](http://progrium.viewdocs.io/dokku/development/plugin-triggers/) to see what is available.
- [Official datastorage plugins](https://github.com/dokku) have been created for the most commonly used datastores. If you previously used/maintained a community contributed plugin, please check these out. We'll be adding more official plugins as time goes on.

Thanks to the *many* contributors for making this release our best release so far, and special thanks to both @michaelshobbs and @Flink for pushing along the `0.4.0` release!

### Deprecations and Removals

- #1372: @SonicHedgehog Do no longer force-install a default nginx.conf
- #1415: @tilgovi Remove uses of (un)set-norestart
- #1432: @josegonzalez Delete unmaintained AUTHORS file
- #1450: @michaelshobbs Rename event plugin buildstep hooks to buildpack

### Bug Fixes

- #1344: @AdamVig Add better error checking on nginx:import-ssl
- #1417: @josegonzalez Expose host and port in vagrant build vm
- #1418: @josegonzalez Use cgroupfs-mount as alternative package to cgroup-lite dependency
- #1419: @u2mejc Fix dokku ps <app> over ssh
- #1422: @josegonzalez Guard against missing VHOST files when listing domains
- #1428: @jimeh Use `$PLUGIN_PATH` instead of `$(dirname $0)/..`
- #1430: @lubert Update vagrant box name to `bento/ubuntu-14.04`
- #1439: @michaelshobbs Fix tar tests by sleeping for 5 seconds
- #1447: @alanjds Properly detect app name in the official cli client
- #1449: @josegonzalez Match herokuish deb with released version number
- #1457: @lukechilds Bashstyle fixes
- #1463: @josegonzalez Update `Xenify Distro` option for linode stackscript
- #1464: @josegonzalez Limit number of log lines when calling `dokku logs -t`
- #1466: @josegonzalez Follow bashstyle conventions where possible
- #1471: @michaelshobbs Make the default scaling logic clearer
- #1475: @josegonzalez Fix issue where restart on failure option overrode existing DOCKER_ARGS

### New features

- #1225: @michaelshobbs Add tags plugin to handle image tagging and deployment of tagged app images
- #1228: @michaelshobbs Use plugn instead of pluginhook to trigger plugin hooks
- #1402: @josegonzalez Consolidate configuration management into config plugin
- #1414: @michaelshobbs Add certs plugin for certificate management
- #1420: @josegonzalez Add `dokku enter` for connecting to an app container
- #1421: @basicer Add tar plugin to manage tar-based deployments
- #1423: @josegonzalez Set `DYNO_TYPE_NUMBER` environment variable for each container
- #1431: @josegonzalez Add helper function for inspecting the state of a container
- #1444: @josegonzalez Extract cleanup command into common function
- #1445: @josegonzalez Create CONTRIBUTING.md
- #1455: @michaelshobbs Continue and log an event if/when container retirement fails
- #1458: @michaelshobbs Set Herokuish `TRACE=true` when `DOKKU_TRACE` is set
- #1460: @michaelshobbs Bump herokuish version to 0.3.3
- #1465: @josegonzalez Set DYNO environment variable to heroku-compatible value
- #1467: @josegonzalez Upgrade dokku installation to use docker-engine
- #1468: @michaelshobbs Clean up semver logic and run install-dependencies after package installation
- #1469: @isundaylee Add nginx:access-logs and nginx:error-logs commands
- #1470: @assaf Add nginx configuration for running behind load balancer
- #1472: @michaelshobbs Use processes defined in `Procfile` when generating `DOKKU_SCALE` file
- #1473: @josegonzalez Handle crashing containers by using restart=on-failure policy
- #1476: @michaelshobbs Support static nginx port when deploying without an application VHOST
- #1476: nginx proxy without VHOST
- #1477: @arthurschreiber Support removing config variables that contain `\n`.

### Documentation

- #1407: @ertrzyiks Correct DOKKU_DOCKERFILE_PORT variable name in docs
- #1408: @josegonzalez Add links to official dokku datastorage plugins
- #1426: @henrik Update memcached link to maintained fork
- #1437: @Flink Update compatibility version for several plugins
- #1446: @johnfraney Correct nginx documentation
- #1478: @eljojo Fix naming of herokuish package in installation docs

## 0.3.26

This release has a few new features, the largest of which is switching from buildstep to herokuish for building containers. Going forward, this should help ensure that built containers are as close to heroku containers as possible, and also allow us to be on the cutting edge of heroku buildpack support. Props to @michaelshobbs for his work on herokuish.

### Bug Fixes

- #1401: @josegonzalez Install apt-transport-https before adding https-backed apt sources

### New Features

- #1128: @michaelshobbs Switch from buildstep to herokuish
- #1399: @basicer Make dokku play nice when there are multiple receive-app hooks
- #1413: @michaelshobbs support comments in DOKKU_SCALE and print contents on deploy

### Documentation

- #1400: @josegonzalez Fix HISTORY.md formatting
- #1410: @josegonzalez Clarify DOKKU_SCALE usage
- #1411: @josegonzalez Clarify `ps:scale` example

## 0.3.25

This release is a bugfix release fixing a broken 0.3.25 apt-get installation.

### Bug Fixes

- #1398 @josegonzalez Revert "Remove `dokku plugins-install` from postinst hook

## 0.3.24

This release is a bugfix release covering dokku packaging.

### Bug Fixes

- #1397: @josegonzalez Use https docker apt repo
- #1394: @josegonzalez Remove `dokku plugins-install` from postinst hook

### Documentation

- #1395: @adrianmoya Adding fqdn requirement

## 0.3.23

This release is a bugfix release mostly covering installation and nginx issues. As well, we launched a nicer documentation site [here](http://progrium.viewdocs.io/dokku/). Thanks to all of our contributors for making this a great release!

### Bug Fixes

- #1334: @josegonzalez Fix pluginhook building and update package version
- #1335: @josegonzalez Fix name for michaelshobbs
- #1341: @michaelshobbs Honor $DOKKU_DOCKERFILE_PORT when binding the docker container to an external IP.
- #1357: @michaelshobbs only run domains and nginx config if we have a port and ip
- #1366: @omeid Make bootstrap.sh safe from bad connection
- #1370: @SamuelMarks Switch from /dev/null to -qq, from --silent to -sL, and sudo
- #1380: @emdantrim Removed uses of `sudo` from `bootstrap.sh`
- #1383: @michaelshobbs fix downscaling from 10+

### New Features

- #1292: @michaelshobbs use column to format help output
- #1337: @josegonzalez Update PREBUILT_STACK_URL to latest buildstep release
- #1354: @alessio Log receive-branch pluginhook
- #1359: @michaelshobbs allow DOKKU_WAIT_TO_RETIRE to be defined per app
- #1365: @michaelshobbs support dockerfile images that don't include bash

### Documentation

- #1305: @josegonzalez Updated documentation site
- #1321: @fwolfst Mention alternative to nginx.conf templates: include-dir.
- #1346: @michaelshobbs document dokku cleanup and the potential of compat issues
- #1349: @alexkruegger add plugin dokku-app-configfiles
- #1358: @bkniffler Add suggestion for low memory VMs
- #1377: @vkurup Fix link to docs from README
- #1379: @jezdez Deleted old, unmaintained plugins
- #1381: @lunohodov Instructions for using the bash client in shells other than bash

## 0.3.22

This release is a general bugfix release, with improvements to handling of nginx templates and application configuration.

### Bug Fixes

- #825: @andrewsomething Add support for multiple keys in the installer.
- #1274: @michaelshobbs Parse correct section of path for `dokku ls` container type
- #1289: @michaelshobbs Do not background container cleanup
- #1298: @SonicHedgehog Fix check-deploy skipping the root path
- #1300: @michaelshobbs Fix urls command when NO_VHOST=1
- #1310: @michaelshobbs Use config:get for checks skipping variables
- #1311: @michaelshobbs Ignore protocol of Dockerfile EXPOSE (refs: #1280)
- #1312: @michaelshobbs Use docker inspect fordefault container check. Closes #1270
- #1313: @michaelshobbs Verify we have an app when deploy is called. Closes #1309
- #1319: @MWers Spelling fix: 'comma seperated'=>'comma-separated'
- #1331: @alexkruegger Fix retrieval of nginx.conf.template app

### New Features

- #1149: @mlebkowski Add pluginhook to receive branches different than master
- #1254: @kilianc Add DOKKU_DOCKERFILE_START_CMD support
- #1261: @Flink Add the ability to skip checks (all or default)
- #1277: @krokhale Add gzip to nginx templates by default
- #1278: @assaf Add the ability to retrieve nginx template from app
- #1291: @michaelshobbs Refactored interface for managing global/local app configuration
- #1299: @SonicHedgehog Set X-Forwarded-Proto header if TLS is enabled when running checks

### Documentation

- #1273: @alessio Add docs for the events plugin
- #1276: @josegonzalez Reorder and deprecate a few plugins
- #1279: @josegonzalez Add docs for `receive-branch` hook. Refs #1149
- #1282: @josegonzalez Move primecache to deprecated plugins
- #1285: @josegonzalez Rename dokku-events-logs.md according to index.md
- #1296: @Flink Add docker-auto-volumes to plugins
- #1301: @mixxorz Add reset mtime plugin list
- #1302: @fwolfst Mention where original nginx templates are found by default.
- #1306: @josegonzalez Clarify web/cli installation docs. Closes #1177. Closes #1170
- #1307: @josegonzalez Add release documentation. Closes #1287
- #1324: @michaelshobbs Update docs to reflect default checks

## 0.3.21

This release fixes an issue with installing buildstep and dokku.

### New Features

- #1256: @alessio Log all dokku events to /var/log/dokku/events.log

### Bug Fixes

- #1269: @josegonzalez Peg lxc-docker in buildstep to 1.6.2

## 0.3.20

This release pegs Dokku to Docker 1.6.2. Docker 1.7.0 introduced changes in `docker ps` which cause incompatibilities with many popular dokku plugins.

### New Features

- #1245: @arthurschreiber Support config variables containing `\n`
- #1257: @josegonzalez Split nginx ssl logs by $APP

### Bug Fixes

- #1207: @rockymadden Fixed bug with client commands taking verb, appname, and also arguments.
- #1251: @josegonzalez Fallback to using /etc/init.d/nginx reload directly to restart nginx
- #1264: @josegonzalez Require lxc-docker-1.6.2

### Documentation

- #1252: @josegonzalez Fix ssh port for vagrant installation. Closes #1139. [ci skip]
- #1250: @josegonzalez SSL documentation is misleading

## 0.3.19

### New Features

- #1118: @michaelshobbs Heroku-style Container-Level scaling
- #1210: @cddr Split nginx logs out by $APP
- #1232: @michaelshobbs Allow passing of docker build options and document dockerfile deployment. Closes #1231

### Bug Fixes

- #1179: @follmann Prevent dismissal of URLs in CHECKS file that contain query params
- #1193: @michaelshobbs Handle docker opts over ssh without escaping quotes. closes #1187
- #1198: @3onyc Check web_config before key_file (Fixes #1196)
- #1200: @josegonzalez Fix lintball from #1198
- #1202: @michaelshobbs Filter out literal wildcard when deploying an unrelated domain. Closes #1185
- #1204: @3onyc Fix bootstrap.sh, install curl when it's missing, make curl follow redirects, don't suppress stderr
- #1206: @rockymadden Handle for installs in /usr/local/bin and the like.
- #1212: @michaelshobbs Let circleci dictate docker install (fixes master)
- #1217: @kirushanth-sakthivetpillai Fix broken ssl wildcard redirect
- #1227: @Flink Use --no-cache when building Dockerfile
- #1246: @josegonzalez Ensure we call apt-get before installing packages

### Documentation

- #1168: @cjblomqvist [docs] Update git-rev plugin to point to maintained version
- #1180: @sherbondy [docs] Clarify usage around official dokku `docker-options` plugin
- #1192: @alessio [docs] Add reference to dokku-events plugin
- #1218: @michaelshobbs [docs] add dokku-logspout plugin
- #1224: @lmars [docs] Add link from plugin-creation to pluginhooks
- #1237: @zyegfryed [docs] Typo (at at -> at)

## 0.3.18

- #1111: @michaelshobbs Call pre-build-dockerfile before docker build
- #1119: @joshco Logging info suggesting tuned CHECKS
- #1120: @josegonzalez [docs] Add freenode shield to readme
- #1121: @josegonzalez Prompt users to run the web installer via MOTD. Closes #943
- #1129: @josegonzalez Validate nginx configuration before restarting nginx
- #1137: @YellowApple [docs] Safer installation method
- #1138: @chrisbutcher [docs] Include tip about using sshcommand acl-add
- #1140: @NigelThorne [docs] Replaced reference to gitreceive with sshcommand as per #746
- #1144: @protonet Allow git-remote with different port
- #1145: @michaelshobbs allow docker-options over ssh. plus test. closes #1135
- #1146: @michaelshobbs Don't re-deploy on domains:add. allow multple domains on command line. Closes #1142
- #1147: @michaelshobbs Utilize all 4 free CircleCI containers
- #1148: @TheEmpty [docs] Add information about 444 for nginx in default_sever.
- #1150: @cjblomqvist [docs] Add monit plugin
- #1151: @LTe Do not kill docker container with SIGKILL
- #1153: @econya [docs] Add README-section: how to contribute
- #1058: @josegonzalez Move bootstrap script to use debian package where possible
- #1171: @josegonzalez Use debconf for package configuration
- #1172: @michaelshobbs unify default and custom nginx template processing
- #1173: @josegonzalez [docs] standardize readme badges
- #1178: @jagandecapri [docs] Update plugins.md
- #1189: @vincentfretin wait 30 seconds and not 30 minutes
- #1190: @josegonzalez Fix docker gpg key installation

## 0.3.17

- #1056: @joshco New check retries feature
- #1060: @josegonzalez Add .template suffix to nginx configuration templates. Refs #1054
- #1064: @michaelshobbs [docs] Document test suite
- #1065: @michaelshobbs Minor dev env cleanup
- #1067: @martinAntsy Fix nginx docs wording around config template eg
- #1068: @matiaskorhonen Fix escaping in the rc.local script in the Linode StackScript
- #1074: @Flink Better detection of dokku remote in dokku_client.sh
- #1075: @Flink Use TTY option for SSH
- #1077: @Flink [docs] Add dokku-psql-single-container to plugins
- #1079: @rorykoehler Corrected configuration link in bootstrap.sh
- #1080: @michaelshobbs Include official docker-options plugin. closes #1062
- #1081: @michaelshobbs Force testing .env with no newline. Closes #1025, #1026, #1063
- #1082: @michaelshobbs Test cleanup with slight performance boost
- #1084: @awendt Make port forwarding configurable
- #1087: @michaelshobbs Make docker-options adhere to DOKKU_NOT_IMPLEMENTED_EXIT pattern
- #1088: @michaelshobbs Support dockerfiles without expose command. closes #1083
- #1097: @michaelshobbs Use config:set-norestart in domains plugin. config:get for dockerfile port. closes #1041
- #1102: @kblcuk Source app-specific ENV during check-deploy
- #1107: @Benjamin-Dobell [docs] Added Dokku Graduate to the list of known plugins

## 0.3.16

- #974: @michaelshobbs Don't use set to guard against pipefail
- #975: @michaelshobbs Simplify SSL hostname handling and avoid overwriting variables. refs #971
- #978: @michaelshobbs Add apparmor and cgroup-lite as pre-dependencies for dokku debian package
- #980: @josegonzalez [docs] Add documentation for pluginhooks
- #981: @josegonzalez Remove old files
- #982: @josegonzalez [docs] Add documentation for existing clients. Closes #977
- #983: @josegonzalez [docs] Update installation documentation
- #984: @josegonzalez [docs] Clarify installation instructions
- #988: @josegonzalez [docs] Add deprecated plugins section and where to get help
- #989: @josegonzalez [docs] Add more clients
- #986: @josegonzalez Switch to yabawock's static nginx buildpack for tests
- #987: @techniq Improve Dockerfile example/test
- #967: @alessio Really clean-up containers and images a-la-Docker
- #992: @josegonzalez [docs] Fix digital ocean docs. Closes #991
- #994: @alessio Fix quoting in container termination. Closes #249
- #996: @pmvieira [docs] Minor typo fix in the pluginhooks documentation
- #1003: @michaelshobbs Remove quoting around cleanup and disable lint for those lines
- #1001: @sekjun9878 [docs] Add sekjun9878/dokku-redis-plugin to plugins.md
- #1004: @michaelshobbs Remove quoting from dockerfile env. closes #1002
- #1018: @michaelshobbs Confine arg shifting to between dokku and command. closes #1017
- #1022: @Flink [docs] Add dokku-maintenance to plugins
- #1008: @lmars Support multiple domains using a wildcard TLS certificate
- #1013: @lmars Fix URL schemes in `dokku urls` output
- #1027: @nickstenning [docs] Add webhooks plugin to documentation
- #1026: @michaelshobbs Ensure we have newlines around our config. closes #1025
- #1010: @michaelshobbs Don't run create/destroy twice in tests
- #1028: @Flink [docs] Add rails-logs to plugins
- #1031: @michaelshobbs Upgrade docker in CI to 1.5.0
- #1029: @assaf Added several enhancements for CHECKS file:
  - Specify how long to wait before running first check
  - Specify timeout for each check
  - Check specific hosts, e.g. http://signin.example.com
  - Check both HTTP and HTTPS resources
- #1032: @cameron-martin Updated dokku-installer to use relative path
- #1035: @Flink [docs] Add dokku-http-auth to plugins
- #1040: @ebeigarts [docs] Add dokku-slack plugin information
- #1038: @michaelshobbs Default container check. closes #1020
- #1036: @michaelshobbs Create config set/unset without restart. closes #908
- #1009: @michaelshobbs Extract first port from Dockerfile and set config variable for use in deploy phase. closes #993
- #1042: @michaelshobbs Update to Support xip.io wildcard DNS as a VHOST
- #1043: @michaelshobbs Use upstart config from docker docs. closes #1015
- #1047: @michaelshobbs Show logs on deploy success and failure
- #1049: @ebeigarts [docs] Change Slack Notifications link
- #1051: @Flink [docs] Add dokku-airbrake-deploy to plugins
- #1057: @josegonzalez Updated deb packaging

## 0.3.15

- #950: @michaelshobbs Do not count blank lines in `make count`
- #952: @michaelshobbs Document cli args over ssh. closes #951
- #954: @michaelshobbs Dockerfile support
- #955: @michaelshobbs Quick style refactor
- #956: @michaelshobbs Comment out skipped tests as we pay the cost for setup() and teardown() anyway
- #957: @michaelshobbs Implement dokku shell and ls (by @plietar). refs #312
- #960: @michaelshobbs Use consistent bash shebang. closes #959
- #962: @josegonzalez Update debian package building due to man page generation changes
- #964: @michaelshobbs Only look for long args in global space. allows for short args in plugins. closes #963
- #966: @djelic handle upgrade in debian/preinst script

## 0.3.14

- #891: @josegonzalez Keep existing configuration files when installing nginx. Refs #886
- #892: @josegonzalez Change documentation on where the ssh config PORT is setup
- #894: @josegonzalez Dokku client improvements
- #895: @michaelshobbs Document deploying private git submodules. Closes #644
- #896: @michaelshobbs Add docker-args pluginhook call to build phase. Closes #515
- #897: @michaelshobbs Official PS plugin
- #898: @joliv Update man page for 0.3.13
- #899: @joliv Use help2man to automatically generate man pages
- #900: @michaelshobbs Support extracting SANs from SSL certificates and adding them to nginx config
- #901: @misto Pull new tags when upgrading to update VERSION
- #904: @michaelshobbs Prevent error on restartall when no apps deployed
- #905: @vincentfretin robv/dokku-elasticsearch not compatible with latest version
- #907: @vincentfretin Don't use -o pipefail for plugin
- #914: @michaelshobbs Conditionally set interactive and tty on dokku run. Closes #552. Closes #913
- #915: @michaelshobbs Document default sites in nginx. Closes #650
- #916: @michaelshobbs Document build phase troubleshooting suggestions. Closes #841. Closes #911.
- #917: @michaelshobbs Document resolvconf troubleshooting step. Closes #649
- #922: @michaelshobbs Use tty cmd to detect if we have one. Closes #921
- #925: @michaelshobbs Implement rebuild command that reuses git_archive_all
- #926: @dyson Update Troubleshooting link in README.md.
- #927: @michaelshobbs Support both docker-args PHASE and docker-args-PHASE. Closes #906
- #933: @michaelshobbs Remove references to .pem. Closes #930
- #936: @michaelshobbs Only execute build stack if we have access to /var/run/docker.sock. Closes #929
- #938: @vincentfretin Enable ssl_prefer_server_ciphers
- #940: @michaelshobbs Use valid composer json with specified php runtime
- #941: @michaelshobbs Source global env prior to app env. Closes #931
- #942: @michaelshobbs Test clojure app
- #949: @michaelshobbs Common functions library with simple argument parsing. Closes #932. Closes #945

## 0.3.13

- #815: @abossard Added wordpress installation helper to plugin index
- #858: @josegonzalez Disable server tokens in nginx. Closes #857
- #859: @josegonzalez Pass command being executed when retrieving DOCKER_ARGS via pluginhook.
- #861: @josegonzalez Default DOKKU_ROOT to ~dokku if unspecified. Closes #587
- #863: @josegonzalez Add missing properties to the php composer.json
- #864: @michaelshobbs bind docker container to internal port if using vhosts
- #867: @michaelshobbs silent grep stderr. closes #862
- #868: @michaelshobbs use circleci for automated testing
- #872: @michaelshobbs fix/enable multi buildpack test
- #873: @michaelshobbs support pre deployment usage of domains plugin. fixes interface binding issue
- #874: @josegonzalez Add advanced installation docs that were removed in #706. Closes #869
- #876: @vincentfretin give CACHE_PATH env variable for forward compatibility with herokuish
- #877: @michaelshobbs add MH to AUTHORS
- #880: @michaelshobbs disable VHOST deployment if global VHOST file is missing and an app domain has not been added
- #881: @jomo troubleshooting typo: 64 != 46
- #884: @michaelshobbs IP and PORT are likely to get clobbered. rename them
- #885: @michaelshobbs test deploy node app without procfile

## 0.3.12

- #846: @michaelshobbs add certificate CN to app VHOST if it's not already
- #847: @leonardodino Update bootstrap.sh: new docs url
- #849: @cjoudrey Add docs for CHECKS
- #850: @michaelshobbs test scala deployment

## 0.3.11

- #821: @michaelshobbs use wercker for automated testing
- #833: @michaelshobbs auto remove the cache dir cleanup container
- #835: @michaelshobbs make sure we match the specific string in VHOST
- #838: @michaelshobbs quote build_env vars to allow for spaces in config
- #839: @michaelshobbs notify irc on builds
- #844: @michaelshobbs build app urls based on wildcard ssl or app ssl

## 0.3.10

- #783: @josegonzalez New domains plugin and nginx extension
- #818: @michaelshobbs rebuild nginx config on domain change
- #827: @michaelshobbs Handle IP only access
- #828: @michaelshobbs Display the port for an app when falling back to the ip address

## 0.3.9

- #787: @josegonzalez/@michaelshobbs Official user-env-compile plugin
  - Uses ENV and APP/ENV files
  - Supports old `BUILD_ENV` files (which are likely in wide-use)
  - Allows user's to override globals with app-specific configuration
  - Migrate `$DOKKU_ROOT/BUILD_ENV` to `$DOKKU_ROOT/ENV` if the former exists and the latter does not
  - Drop `BUILD_ENV` support in favor of just `ENV` via a `mv` command
  - Add default ENV with `CURL_TIMEOUT` and `CURL_CONNECT_TIMEOUT`
- #811: @abossard Increased `server_names_hash_bucket_size` in nginx.conf to 512
- #814: @josegonzalez Source files in $DOKKU_ROOT/.dokkurc directory and add `dokku trace` command
- #816: @josegonzalez Add documentation for user-env feature

## 0.3.8

- #796: @josegonzalez Better vagrant documentation
- #801: @joelvh Point users to upgrade guides
- #805: @ademuk Fixed import-ssl server.crt/key check
- #806: @josegonzalez Dokku pushes now happen as the dokku user, not git. Refs #796
- #807: @josegonzalez Write proper nginx conf upon installation. Closes #799
- #808: @josegonzalez Ensure makefiles are properly formatted

## 0.3.7

- #788: @mmerickel fix apps plugin issues in 0.3.6
- #789: @mmerickel do not output message when creating ENV file

## 0.3.6

- #782: @josegonzalez Simplified config checking
- #785: @lsde fix missing semicolon in nginx config

## 0.3.5

- #784: @josegonzalez Fix NO_VHOST check

## 0.3.4

- #780: @josegonzalez Output error message when a command is not found. Closes #778
- #781: @michaelshobbs use DOKKU_IMAGE (i.e. progrium/buildstep)

## 0.3.3

- #659: @Xe contrib: add dokku client shell script
- #669 @ohardy Handle dokku plugins-update command
- #722: @wrboyce Add `git pull` support with git-pre-pull and git-post-pull hooks
- #751: @tboerger Partial openSUSE support
- #776: @joliv Update man page for new commands
- #777: @tboerger Use PLUGINS_PATH env var and persist environment when running dokku with sudo
- #779: @josegonzalez Minor bash formatting changes

## 0.3.2

- #675: @michaelshobbs port wait-to-retire from broadly/dokku
- #765: @josegonzalez Ignore tls directory when listing apps
- #766: @josegonzalez Sort output of apps command
- #771: @josegonzalez Doc updates
- #518 #772: @nickl- Import ssl certificates
- #773: @alex-sherwin Support a way to not create nginx vhost
- #774: @josegonzalez Add the ability to customize an app's hostname using nginx-hostname pluginhook

## 0.3.1

- 647b2157: @josegonzalez Update HISTORY.md for 0.3.0
- #359: @plietar Remove plugins before copying them again
- #573: @eriknomitch Use command instead of which for apt-get existential check in bootstrap.sh
- #579: @motin Plugin nginx-vhosts includes files in folder nginx.conf.d
- #607: @fbochu Use PLUGIN_PATH in dokku default case
- #617: @markstos Document what bootstrap.sh is doing
- #758: @josegonzalez Make ENV file readable only from dokku user. Closes #621
- #699: @tombell Actually suppress the output from `git_archive_all`
- #702: @jazzzz Allows config:set to set parameters values with spaces
- #754: @josegonzalez Remove all references to Ubuntu 12.04. Refs #238
- #755: @josegonzalez Setup dokku-installer within Vagrant VM on first provision
- #759: @josegonzalez Create an `apps` core plugin
- #760: @josegonzalez Formatting
- #761: @josegonzalez Add dokku-registry to list. Closes #716
- #762: @josegonzalez Update template for dokku docs

## 0.3.0

- Added git submodules support
- 969aed87: @jfrazelle  Fix double brackets issue in nginx-vhosts install
- 42fee25f: @rhy-jot Mention 14.04 & 12.10 sunset
- 4f5fc586: @rhy-jot Update cipher list
- #276: @wingrunr21 Add dependencies hook
- #476: @joliv Added man page entry
- #522: @wingrunr21 Improve SSL support and implement SPDY
- #544: @jfrazelle if dokku.conf has not been created, create it
- #555: @jakubholynet Readme fix, Env vars only set at run time
- #562: @assaf Zero down-time deploy and server checks
- #571: @joliv Added man page plugin command
  #601: @jazzzz Restart app only if config changed
- #628: @voronianski Update Vagrant box to trusty because of raring server issues
- #632: @jazzzz Fixed port when docker is start with --ip with an IP other than 0.0.0.0
- #654: @cameron-martin Fixed variable name for RESTART
- #664: @alexernst Don't overwrite $APP/URL modified by plugins in post-deploy hook
- #665: @protomouse Explicitly install man-db
- #698: @tombell Help output formatting
- #701: @jazzzz  Fix issues with single-letter config:set usage
- #703: @jazzzz  Display help when invoking dokku with no parameter
- #706: @josegonzalez Consolidate documentation on viewthedocs
- #709: @rinti Grammar and spelling fixes
- #708: @josegonzalez Simplify vagrant workflow
- #717: @kristofsajdak Add dokku-registry to plugin list
- #718: @Coaxial Use https for installation instructions
- #721: @wrboyce  Fix issue in plugins-install-dependencies
- #723: @ghostbar Let users know they are starting buildstep during installation
- #735: @andrewsomething Redirect to the app deployment docs on success.
- #745: @rcarmo Typo
- #746: @vincentfretin replace gitreceive by sshcommand in example url
- #748: @vincentfretin Link to proper blog url
- #749: @vincentfretin Fix app certificate directory in backup-import/export
- #750: @th4t Remove unintended phrase repetition in installation.md


## 0.2.0 (2013-11-24)

* Added DOKKU_TRACE variable for verbose trace information
* Added an installer (for pre-built images)
* Application config (environment variable management)
* Backup/import plugin
* Basic hooks/plugin system
* Cache dir is preserved across builds
* Command to delete an application
* Exposed commands over SSH using sshcommand
* Git handling is moved to a plugin
* Integration test coverage
* Pulled nginx vhosts out into plugin
* Run command
* Separated dokku and buildstep more cleanly
* Uses latest version of Docker again

## 0.1.0 (2013-06-15)

 * First release
   * Bootstrap script for Ubuntu system
   * Basic push / deploy with git
   * Hostname support with Nginx
   * Support for Java, Ruby, Node.js buildpacks
