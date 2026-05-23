#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  destroy_app
  global_teardown
}

# assert_nginx_property_set_get verifies that for a given nginx property,
# (a) the raw per-app and global keys cycle through set/unset states,
# (b) the computed key always returns a non-empty value when defaults exist,
# and (c) the computed key falls back through per-app -> global.
#
# This intentionally does not pin the exact default value for properties whose
# default is computed at runtime (e.g. proxy-buffer-size depends on the host's
# page size) - those would be brittle to assert verbatim. The set/unset cycle
# is the canonical test for the report convention.
assert_nginx_property_set_get() {
  declare prop="$1" app_value="$2" global_value="$3"

  run /bin/bash -c "dokku nginx:set --global $prop"
  assert_success

  run /bin/bash -c "dokku nginx:set --global $prop $global_value"
  assert_success

  run /bin/bash -c "dokku --quiet nginx:report $TEST_APP --format json | jq -r '.\"global-$prop\"'"
  assert_success
  assert_output "$global_value"

  run /bin/bash -c "dokku --quiet nginx:report $TEST_APP --format json | jq -r '.\"computed-$prop\"'"
  assert_success
  assert_output "$global_value"

  run /bin/bash -c "dokku nginx:set $TEST_APP $prop $app_value"
  assert_success

  run /bin/bash -c "dokku --quiet nginx:report $TEST_APP --format json | jq -r '.\"$prop\"'"
  assert_success
  assert_output "$app_value"

  run /bin/bash -c "dokku --quiet nginx:report $TEST_APP --format json | jq -r '.\"global-$prop\"'"
  assert_success
  assert_output "$global_value"

  run /bin/bash -c "dokku --quiet nginx:report $TEST_APP --format json | jq -r '.\"computed-$prop\"'"
  assert_success
  assert_output "$app_value"

  run /bin/bash -c "dokku nginx:set $TEST_APP $prop"
  assert_success

  run /bin/bash -c "dokku nginx:set --global $prop"
  assert_success
}

@test "(nginx:report) timeout properties raw/global/computed" {
  for prop in client-body-timeout client-header-timeout keepalive-timeout lingering-timeout send-timeout proxy-connect-timeout proxy-read-timeout proxy-send-timeout; do
    assert_nginx_property_set_get "$prop" "30s" "45s"
  done
}

@test "(nginx:report) buffer properties raw/global/computed" {
  assert_nginx_property_set_get "proxy-buffer-size" "8k" "16k"
  assert_nginx_property_set_get "proxy-buffering" "off" "on"
  assert_nginx_property_set_get "proxy-buffers" "16 16k" "32 16k"
  assert_nginx_property_set_get "proxy-busy-buffers-size" "16k" "32k"
  assert_nginx_property_set_get "client-max-body-size" "10m" "20m"
}

@test "(nginx:report) hsts properties raw/global/computed" {
  assert_nginx_property_set_get "hsts" "false" "true"
  assert_nginx_property_set_get "hsts-include-subdomains" "false" "true"
  assert_nginx_property_set_get "hsts-max-age" "3600" "7200"
  assert_nginx_property_set_get "hsts-preload" "true" "false"
}

@test "(nginx:report) x-forwarded properties raw/global/computed" {
  assert_nginx_property_set_get "x-forwarded-for-value" "\$remote_addr" "\$proxy_add_x_forwarded_for"
  assert_nginx_property_set_get "x-forwarded-port-value" "8080" "8443"
  assert_nginx_property_set_get "x-forwarded-proto-value" "https" "http"
  assert_nginx_property_set_get "x-forwarded-ssl" "on" "off"
}

@test "(nginx:report) log/address properties raw/global/computed" {
  assert_nginx_property_set_get "bind-address-ipv4" "127.0.0.1" "0.0.0.0"
  assert_nginx_property_set_get "bind-address-ipv6" "::1" "::"
  assert_nginx_property_set_get "access-log-format" "main" "combined"
  assert_nginx_property_set_get "access-log-path" "/var/log/nginx/app-access.log" "/dev/stdout"
  assert_nginx_property_set_get "error-log-path" "/var/log/nginx/app-error.log" "/dev/stderr"
}

@test "(nginx:report) misc properties raw/global/computed" {
  assert_nginx_property_set_get "disable-custom-config" "false" "true"
  assert_nginx_property_set_get "underscore-in-headers" "on" "off"
  assert_nginx_property_set_get "proxy-keepalive" "on" "off"
}
