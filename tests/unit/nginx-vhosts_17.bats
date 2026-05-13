#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  [[ -f "$DOKKU_ROOT/VHOST" ]] && cp -fp "$DOKKU_ROOT/VHOST" "$DOKKU_ROOT/VHOST.bak"
  create_app
}

teardown() {
  sudo systemctl start coredns-docker >/dev/null 2>&1 || true
  dokku --quiet nginx:set --global dns-resolver
  dokku --quiet nginx:set --global dns-zone
  destroy_app
  [[ -f "$DOKKU_ROOT/VHOST.bak" ]] && mv "$DOKKU_ROOT/VHOST.bak" "$DOKKU_ROOT/VHOST" && chown dokku:dokku "$DOKKU_ROOT/VHOST"
  global_teardown
}

@test "(nginx-vhosts:set) dns-resolver and dns-zone defaults report 127.0.0.1:1053 and docker" {
  run /bin/bash -c "dokku nginx:report $TEST_APP --nginx-computed-dns-resolver"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.1:1053"

  run /bin/bash -c "dokku nginx:report $TEST_APP --nginx-computed-dns-zone"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "docker"
}

@test "(nginx-vhosts:set) dns-resolver and dns-zone can be overridden globally" {
  run /bin/bash -c "dokku nginx:set --global dns-resolver 127.0.0.2:5353"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:set --global dns-zone test"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:report $TEST_APP --nginx-computed-dns-resolver"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "127.0.0.2:5353"

  run /bin/bash -c "dokku nginx:report $TEST_APP --nginx-computed-dns-zone"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "test"
}

@test "(nginx-vhosts) generated nginx.conf uses resolver and variable-based proxy_pass by default" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "resolver 127.0.0.1:1053 valid=10s ipv6=off"
  assert_output_contains "set \$dokku_upstream"
  assert_output_contains "${TEST_APP}.web.docker:"
  assert_output_contains "proxy_pass http://\$dokku_upstream"
  assert_output_not_contains "upstream ${TEST_APP}-"
}

@test "(nginx-vhosts) generated nginx.conf falls back to IP-based upstream when dns-resolver=off" {
  run /bin/bash -c "dokku nginx:set --global dns-resolver off"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "resolver 127.0.0.1:1053"
  assert_output_not_contains "\$dokku_upstream"
  assert_output_contains "upstream ${TEST_APP}-"
}

@test "(nginx-vhosts) setting dns-resolver to empty re-engages the default resolver path" {
  run /bin/bash -c "dokku nginx:set --global dns-resolver off"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  assert_output_contains "upstream ${TEST_APP}-"

  run /bin/bash -c "dokku nginx:set --global dns-resolver"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:build-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "resolver 127.0.0.1:1053 valid=10s ipv6=off"
  assert_output_not_contains "upstream ${TEST_APP}-"
}

@test "(nginx-vhosts) vendored nginx.conf.sigil renders IP-based upstream regardless of dns-resolver" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP custom_nginx_template
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku nginx:show-config $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_not_contains "resolver 127.0.0.1:1053"
  assert_output_not_contains "\$dokku_upstream"
  assert_output_contains "upstream ${TEST_APP}-"
}

@test "(nginx-vhosts) nginx returns 502 when coredns-docker is stopped" {
  run deploy_app python dokku@$DOKKU_DOMAIN:$TEST_APP
  echo "output: $output"
  echo "status: $status"
  assert_success

  run sudo systemctl stop coredns-docker
  echo "output: $output"
  echo "status: $status"
  assert_success

  run curl --connect-to "$TEST_APP.$DOKKU_DOMAIN:80:localhost:80" -kSso /dev/null -w "%{http_code}" -m 5 "http://$TEST_APP.$DOKKU_DOMAIN/"
  echo "output: $output"
  echo "status: $status"
  assert_output "502"

  run sudo systemctl start coredns-docker
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_response http "$TEST_APP.$DOKKU_DOMAIN" 80 / "" 200
}
