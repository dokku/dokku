# History

## 0.3.16

- #974: @michaelhobbs Don't use set to guard against pipefail
- #975: @michaelhobbs Simplify SSL hostname handling and avoid overwriting variables. refs #971
- #978: @michaelhobbs Add apparmor and cgroup-lite as pre-dependencies for dokku debian package
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
- #1003: @michaelhobbs Remove quoting around cleanup and disable lint for those lines
- #1001: @sekjun9878 [docs] Add sekjun9878/dokku-redis-plugin to plugins.md
- #1004: @michaelhobbs Remove quoting from dockerfile env. closes #1002
- #1018: @michaelhobbs Confine arg shifting to between dokku and command. closes #1017
- #1022: @Flink [docs] Add dokku-maintenance to plugins
- #1008: @lmars Support multiple domains using a wildcard TLS certificate
- #1013: @lmars Fix URL schemes in `dokku urls` output
- #1027: @nickstenning [docs] Add webhooks plugin to documentation
- #1026: @michaelhobbs Ensure we have newlines around our config. closes #1025
- #1010: @michaelhobbs Don't run create/destroy twice in tests
- #1028: @Flink [docs] Add rails-logs to plugins
- #1031: @michaelhobbs Upgrade docker in CI to 1.5.0
- #1029: @assaf Added several enhancements for CHECKS file:
  - Specify how long to wait before running first check
  - Specify timeout for each check
  - Check specific hosts, e.g. http://signin.example.com
  - Check both HTTP and HTTPS resources
- #1032: @cameron-martin Updated dokku-installer to use relative path
- #1035: @Flink [docs] Add dokku-http-auth to plugins
- #1040: @ebeigarts [docs] Add dokku-slack plugin information
- #1038: @michaelhobbs Default container check. closes #1020
- #1036: @michaelhobbs Create config set/unset without restart. closes #908
- #1009: @michaelhobbs Extract first port from Dockerfile and set config variable for use in deploy phase. closes #993
- #1042: @michaelhobbs Update to Support xip.io wildcard DNS as a VHOST
- #1043: @michaelhobbs Use upstart config from docker docs. closes #1015
- #1047: @michaelhobbs Show logs on deploy success and failure
- #1049: @ebeigarts [docs] Change Slack Notifications link
- #1051: @Flink [docs] Add dokku-airbrake-deploy to plugins
- #1057: @josegonzalez Updated deb packaging

## 0.3.15

- #950: @michaelhobbs Do not count blank lines in `make count`
- #952: @michaelhobbs Document cli args over ssh. closes #951
- #954: @michaelhobbs Dockerfile support
- #955: @michaelhobbs Quick style refactor
- #956: @michaelhobbs Comment out skipped tests as we pay the cost for setup() and teardown() anyway
- #957: @michaelhobbs Implement dokku shell and ls (by @plietar). refs #312
- #960: @michaelhobbs Use consistent bash shebang. closes #959
- #962: @josegonzalez Update debian package building due to man page generation changes
- #964: @michaelhobbs Only look for long args in global space. allows for short args in plugins. closes #963
- #966: @djelic handle upgrade in debian/preinst script

## 0.3.14

- #891: @josegonzalez Keep existing configuration files when installing nginx. Refs #886
- #892: @josegonzalez Change documentation on where the ssh config PORT is setup
- #894: @josegonzalez Dokku client improvements
- #895: @michaelhobbs Document deploying private git submodules. Closes #644
- #896: @michaelhobbs Add docker-args pluginhook call to build phase. Closes #515
- #897: @michaelhobbs Official PS plugin
- #898: @joliv Update man page for 0.3.13
- #899: @joliv Use help2man to automatically generate man pages
- #900: @michaelhobbs Support extracting SANs from SSL certificates and adding them to nginx config
- #901: @misto Pull new tags when upgrading to update VERSION
- #904: @michaelhobbs Prevent error on restartall when no apps deployed
- #905: @vincentfretin robv/dokku-elasticsearch not compatible with latest version
- #907: @vincentfretin Don't use -o pipefail for plugin
- #914: @michaelhobbs Conditionally set interactive and tty on dokku run. Closes #552. Closes #913
- #915: @michaelhobbs Document default sites in nginx. Closes #650
- #916: @michaelhobbs Document build phase troubleshooting suggestions. Closes #841. Closes #911.
- #917: @michaelhobbs Document resolvconf troubleshooting step. Closes #649
- #922: @michaelhobbs Use tty cmd to detect if we have one. Closes #921
- #925: @michaelhobbs Implement rebuild command that reuses git_archive_all
- #926: @dyson Update Troubleshooting link in README.md.
- #927: @michaelhobbs Support both docker-args PHASE and docker-args-PHASE. Closes #906
- #933: @michaelhobbs Remove references to .pem. Closes #930
- #936: @michaelhobbs Only execute build stack if we have access to /var/run/docker.sock. Closes #929
- #938: @vincentfretin Enable ssl_prefer_server_ciphers
- #940: @michaelhobbs Use valid composer json with specified php runtime
- #941: @michaelhobbs Source global env prior to app env. Closes #931
- #942: @michaelhobbs Test clojure app
- #949: @michaelhobbs Common functions library with simple argument parsing. Closes #932. Closes #945

## 0.3.13

- #815: @abossard Added wordpress installation helper to plugin index
- #858: @josegonzalez Disable server tokens in nginx. Closes #857
- #859: @josegonzalez Pass command being executed when retrieving DOCKER_ARGS via pluginhook.
- #861: @josegonzalez Default DOKKU_ROOT to ~dokku if unspecified. Closes #587
- #863: @josegonzalez Add missing properties to the php composer.json
- #864: @michaelhobbs bind docker container to internal port if using vhosts
- #867: @michaelhobbs silent grep stderr. closes #862
- #868: @michaelhobbs use circleci for automated testing
- #872: @michaelhobbs fix/enable multi buildpack test
- #873: @michaelhobbs support pre deployment usage of domains plugin. fixes interface binding issue
- #874: @josegonzalez Add advanced installation docs that were removed in #706. Closes #869
- #876: @vincentfretin give CACHE_PATH env variable for forward compatibility with herokuish
- #877: @michaelhobbs add MH to AUTHORS
- #880: @michaelhobbs disable VHOST deployment if global VHOST file is missing and an app domain has not been added
- #881: @jomo troubleshooting typo: 64 != 46
- #884: @michaelhobbs IP and PORT are likely to get clobbered. rename them
- #885: @michaelhobbs test deploy node app without procfile

## 0.3.12

- #846: @michaelhobbs add certificate CN to app VHOST if it's not already
- #847: @leonardodino Update bootstrap.sh: new docs url
- #849: @cjoudrey Add docs for CHECKS
- #850: @michaelhobbs test scala deployment

## 0.3.11

- #821: @michaelhobbs use wercker for automated testing
- #833: @michaelhobbs auto remove the cache dir cleanup container
- #835: @michaelhobbs make sure we match the specific string in VHOST
- #838: @michaelhobbs quote build_env vars to allow for spaces in config
- #839: @michaelhobbs notify irc on builds
- #844: @michaelhobbs build app urls based on wildcard ssl or app ssl

## 0.3.10

- #783: @josegonzalez New domains plugin and nginx extension
- #818: @michaelhobbs rebuild nginx config on domain change
- #827: @michaelhobbs Handle IP only access
- #828: @michaelhobbs Display the port for an app when falling back to the ip address

## 0.3.9

- #787: @josegonzalez/@michaelhobbs Official user-env-compile plugin
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
- #781: @michaelhobbs use DOKKU_IMAGE (i.e. progrium/buildstep)

## 0.3.3

- #659: @Xe contrib: add dokku client shell script
- #669 @ohardy Handle dokku plugins-update command
- #722: @wrboyce Add `git pull` support with git-pre-pull and git-post-pull hooks
- #751: @tboerger Partial openSUSE support
- #776: @joliv Update man page for new commands
- #777: @tboerger Use PLUGINS_PATH env var and persist environment when running dokku with sudo
- #779: @josegonzalez Minor bash formatting changes

## 0.3.2

- #675: @michaelhobbs port wait-to-retire from broadly/dokku
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
- #276: @joliv  Add dependencies hook - wingrunr21- #476: Added man page entry
- #522: @wingrunr21 Improve SSL support and implement SPDY
- #544: @jfrazelle  if dokku.conf has not been created, create it
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
