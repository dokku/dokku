#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(domains) domains:help" {
  run /bin/bash -c "dokku domains"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage domains used by the proxy"
  help_output="$output"

  run /bin/bash -c "dokku domains:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "Manage domains used by the proxy"
  assert_output "$help_output"
}

@test "(domains:report) --global --format json" {
  run /bin/bash -c "dokku domains:report --global --format json | jq -e ."
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report --global --format json | jq -r 'keys[]' | grep -v 'global-' | wc -l"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"

  run /bin/bash -c "dokku domains:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "global domains information"

  run /bin/bash -c "dokku domains:report --global --domains-global-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku domains:report --global --domains-invalid-flag"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid flag passed"
}

@test "(domains) domains" {
  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null | grep ${TEST_APP}.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "${TEST_APP}.${DOKKU_DOMAIN}"
}

@test "(domains) domains:add" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP 2.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP .${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP _"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.test.app.${DOKKU_DOMAIN}"
  assert_output_contains "2.app.${DOKKU_DOMAIN}"
  assert_output_contains a--domain.with--hyphens
  assert_output_contains _
}

@test "(domains) domains:add (multiple)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN} 2.app.${DOKKU_DOMAIN} a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.test.app.${DOKKU_DOMAIN}"
  assert_output_contains "2.app.${DOKKU_DOMAIN}"
  assert_output_contains a--domain.with--hyphens
}

@test "(domains) domains:add (duplicate)" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add $TEST_APP test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:add (invalid)" {
  run /bin/bash -c "dokku domains:add $TEST_APP http://test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) domains:enable and domains:disable (missing app)" {
  run /bin/bash -c "dokku domains:disable missing-app"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App missing-app does not exist"

  run /bin/bash -c "dokku domains:enable missing-app"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "App missing-app does not exist"
}

@test "(domains) domains:remove" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove $TEST_APP test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test.app.${DOKKU_DOMAIN}" 0
}

@test "(domains) domains:remove (multiple)" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN} test.app.${DOKKU_DOMAIN} 2.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove $TEST_APP www.test.app.${DOKKU_DOMAIN} test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.test.app.${DOKKU_DOMAIN}" 0
  assert_output_contains "test.app.${DOKKU_DOMAIN}" 0
  assert_output_contains "2.app.${DOKKU_DOMAIN}"
}

@test "(domains) domains:remove (wildcard domain)" {
  run /bin/bash -c "dokku domains:add $TEST_APP '*.${DOKKU_DOMAIN}'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.${DOKKU_DOMAIN} *.${DOKKU_DOMAIN}"

  run /bin/bash -c "dokku domains:remove $TEST_APP '*.${DOKKU_DOMAIN}'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.${DOKKU_DOMAIN}"
}

@test "(domains) domains:set" {
  run /bin/bash -c "dokku domains:add $TEST_APP www.test.app.${DOKKU_DOMAIN} test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP 2.app.${DOKKU_DOMAIN} a--domain.with--hyphens"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "www.test.app.${DOKKU_DOMAIN}" 0
  assert_output_contains "test.app.${DOKKU_DOMAIN}" 0
  assert_output_contains "2.app.${DOKKU_DOMAIN}"
  assert_output_contains a--domain.with--hyphens
}

@test "(domains) domains:reset" {
  run /bin/bash -c "dokku domains:add $TEST_APP test.app.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:reset $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "test.app.${DOKKU_DOMAIN}" 0
}

@test "(domains) domains:add-global" {
  run /bin/bash -c "dokku domains:add-global global.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report --global"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -qw 'global.${DOKKU_DOMAIN}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:add-global (multiple)" {
  run /bin/bash -c "dokku domains:add-global global1.${DOKKU_DOMAIN} global2.${DOKKU_DOMAIN} global3.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global1.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report 2>/dev/null | grep -q global3.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:clear-global" {
  run /bin/bash -c "dokku domains:add-global global.dokku.invalid"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:add-global global.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:clear-global"
  echo "output: $output"
  echo "status: $status"
  assert_success
  refute_line "global.dokku.invalid"
  refute_line "global.${DOKKU_DOMAIN}"
}

@test "(domains) domains:remove-global" {
  run /bin/bash -c "dokku domains:add-global global.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:remove-global global.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  refute_line "global.${DOKKU_DOMAIN}"
}

@test "(domains) domains (multiple global domains)" {
  run /bin/bash -c "dokku domains:add-global global1.${DOKKU_DOMAIN} global2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "${TEST_APP}.global1.${DOKKU_DOMAIN}"
  assert_output_contains "${TEST_APP}.global2.${DOKKU_DOMAIN}"
}

@test "(domains) domains:set-global" {
  run /bin/bash -c "dokku domains:add-global global1.${DOKKU_DOMAIN} global2.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set-global global3.${DOKKU_DOMAIN} global4.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run destroy_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run create_app
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP 2>/dev/null"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "${TEST_APP}.global1.${DOKKU_DOMAIN}" 0
  assert_output_contains "${TEST_APP}.global2.${DOKKU_DOMAIN}" 0
  assert_output_contains "${TEST_APP}.global3.${DOKKU_DOMAIN}"
  assert_output_contains "${TEST_APP}.global4.${DOKKU_DOMAIN}"
}

@test "(domains) app name overlaps with global domain.tld" {
  run /bin/bash -c "dokku domains:set-global dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --quiet apps:create test.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  # run domains:reset in order to invoke default vhost creation
  run /bin/bash -c "dokku --quiet domains:reset test.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report test.dokku.test --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test.dokku.test"

  run /bin/bash -c "dokku --force apps:destroy test.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) app rename only renames domains associated with global domains" {
  run /bin/bash -c "dokku domains:set-global ${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.${DOKKU_DOMAIN} $TEST_APP.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP other-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report other-name --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "other-name.${DOKKU_DOMAIN} $TEST_APP.dokku.test"

  run /bin/bash -c "dokku apps:rename other-name $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) app rename preserves all subdomains across multiple global vhosts" {
  run /bin/bash -c "dokku domains:set-global ${DOKKU_DOMAIN} dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.${DOKKU_DOMAIN} $TEST_APP.dokku.test custom.example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP other-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report other-name --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "other-name.${DOKKU_DOMAIN} other-name.dokku.test custom.example.com"

  run /bin/bash -c "dokku apps:rename other-name $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) app rename preserves custom domains when no global vhost is set" {
  run /bin/bash -c "dokku domains:clear-global"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP custom.example.com other.example.org"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP other-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report other-name --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "custom.example.com other.example.org"

  run /bin/bash -c "dokku apps:rename other-name $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) verify warning on ipv4/ipv6 domain name" {
  touch /etc/nginx/sites-enabled/default
  rm "$DOKKU_ROOT/VHOST"
  echo "127.0.0.1" >"$DOKKU_ROOT/VHOST"
  run deploy_app
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Detected IPv4 domain name with nginx proxy enabled."

  rm -f /etc/nginx/sites-enabled/default
  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Detected IPv4 domain name with nginx proxy enabled." 0
}

@test "(domains) https:443" {
  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.${DOKKU_DOMAIN}"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku git:from-image $TEST_APP cockpithq/cockpit:core-latest"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "http://$TEST_APP.${DOKKU_DOMAIN}" 2
  assert_output_contains "http://$TEST_APP.${DOKKU_DOMAIN}:2019"
  assert_output_contains "udp://$TEST_APP.${DOKKU_DOMAIN}:443"
  assert_output_contains "http://$TEST_APP.${DOKKU_DOMAIN}:443" 0

  run /bin/bash -c "dokku ports:report $TEST_APP --ports-map-detected"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "http:2019:2019 http:80:80 https:443:443 udp:443:443"
}

@test "(domains) urls do not glob expand a catch-all domain" {
  local tmp_dir
  tmp_dir="$(mktemp -d)"
  touch "$tmp_dir/glob-canary"
  chmod 755 "$tmp_dir"

  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP '*'"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "cd $tmp_dir && dokku urls $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  rm -rf "$tmp_dir"
  assert_success
  assert_output_contains "glob-canary" 0
  assert_output_contains "http://*"
}

@test "(domains) domains:disable stores the disabled state outside the app environment" {
  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "cat $DOKKU_LIB_ROOT/config/domains/$TEST_APP/disabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku config:get $TEST_APP NO_VHOST"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku config:export $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "NO_VHOST" 0

  run /bin/bash -c "dokku domains:enable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "test -f $DOKKU_LIB_ROOT/config/domains/$TEST_APP/disabled"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) the disabled state survives config:clear and a NO_VHOST config var" {
  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku config:clear --no-restart $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku config:set --no-restart $TEST_APP NO_VHOST=0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"
}

@test "(domains) install migrates NO_VHOST to the disabled property" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP NO_VHOST=1"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku plugin:install --core"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Migrating deprecated NO_VHOST to domains disabled property for $TEST_APP"

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku config:get $TEST_APP NO_VHOST"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) install unsets an enabled NO_VHOST" {
  run /bin/bash -c "dokku config:set --no-restart $TEST_APP NO_VHOST=0"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku plugin:install --core"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Migrating deprecated NO_VHOST to domains disabled property for $TEST_APP" 0

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku config:get $TEST_APP NO_VHOST"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) apps:clone and apps:rename keep the disabled state" {
  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:clone --skip-deploy $TEST_APP $TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP-clone --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "dokku apps:rename --skip-deploy $TEST_APP-clone $TEST_APP-renamed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report $TEST_APP-renamed --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/config/domains/$TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_failure
}

@test "(domains) apps:rename keeps an enabled app enabled without global domains" {
  run /bin/bash -c "dokku domains:clear-global"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.dokku.test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku apps:rename --skip-deploy $TEST_APP other-name"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku domains:report other-name --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"

  run /bin/bash -c "dokku domains:report other-name --domains-app-vhosts"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.dokku.test"
}

@test "(domains) apps:destroy removes the domains properties" {
  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/config/domains/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "test -d $DOKKU_LIB_ROOT/config/domains/$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_failure

  run /bin/bash -c "dokku apps:create $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(domains) domains:disable and domains:enable do not restart a deployed app" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "docker ps -q --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web"
  echo "output: $output"
  echo "status: $status"
  assert_success
  local container_id="$output"

  run /bin/bash -c "dokku domains:disable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Restarting app" 0

  run /bin/bash -c "dokku domains:enable $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Restarting app" 0

  run /bin/bash -c "docker ps -q --filter label=com.dokku.app-name=$TEST_APP --filter label=com.dokku.process-type=web"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$container_id"

  run /bin/bash -c "dokku domains:report $TEST_APP --domains-app-enabled"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "true"
}
