#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
  create_app
}

teardown() {
  dokku proxy:redirect:clear "$TEST_APP" >/dev/null 2>&1 || true
  destroy_app
  global_teardown
}

@test "(proxy:redirect) help text mentions redirect commands" {
  run /bin/bash -c "dokku proxy:help"
  echo "output: $output"
  echo "status: $status"
  assert_output_contains "proxy:redirect:set"
  assert_output_contains "proxy:redirect:remove"
  assert_output_contains "proxy:redirect:clear"
  assert_output_contains "proxy:redirect:report"
}

@test "(proxy:redirect:set) stores a redirect with the default code and is idempotent" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Redirect www.$TEST_APP.$DOKKU_DOMAIN -> $TEST_APP.$DOKKU_DOMAIN (301)"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "1"
}

@test "(proxy:redirect:set) --code sets the code and re-setting without it resets to 301" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 307"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects[0].code'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "307"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects[0].code'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "301"
}

@test "(proxy:redirect:set) lowercases hostnames" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP WWW.Example.COM Example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects[0] | .source + \" \" + .destination'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.example.com example.com"
}

@test "(proxy:redirect:set) rejects invalid input" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Usage: dokku proxy:redirect:set"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP example.com example.com"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "source and destination cannot be the same hostname"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP _ example.com"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot be the catch-all hostname"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP '*.example.com' example.com"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "cannot be a wildcard hostname"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com https://example.com"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "is not a valid hostname"

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com --code 200"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "invalid redirect code 200"

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects | length'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "0"
}

@test "(proxy:redirect:remove) removes a single redirect and is idempotent" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  assert_success
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP old.example.com example.com"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:remove $TEST_APP www.example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -r '.redirects[].source'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "old.example.com"

  run /bin/bash -c "dokku proxy:redirect:remove $TEST_APP www.example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(proxy:redirect:clear) removes every redirect and is idempotent" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  assert_success
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP old.example.com example.com"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Redirects: (none)"

  run /bin/bash -c "dokku proxy:redirect:clear $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
}

@test "(proxy:redirect:report) lists redirects sorted by source and rejects unknown formats" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  assert_success
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP old.example.com example.com --code 308"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format json | jq -c '.redirects'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output '[{"source":"old.example.com","destination":"example.com","code":308},{"source":"www.example.com","destination":"example.com","code":301}]'

  run /bin/bash -c "dokku proxy:redirect:report"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "$TEST_APP proxy redirects information"

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP --format yaml"
  echo "output: $output"
  echo "status: $status"
  assert_failure
  assert_output_contains "Invalid format specified"
}

@test "(proxy:redirect) proxy-redirects-list trigger output" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com --code 302"
  assert_success

  run_plugn_trigger proxy-redirects-list "$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.example.com|example.com|302"
}

@test "(proxy:redirect) redirects follow app clone and rename" {
  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  assert_success

  run /bin/bash -c "dokku apps:clone $TEST_APP $TEST_APP-clone"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP-clone --format json | jq -r '.redirects[0].source'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.example.com"

  run /bin/bash -c "dokku --force apps:destroy $TEST_APP-clone"
  assert_success

  run /bin/bash -c "dokku apps:rename $TEST_APP $TEST_APP-renamed"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:report $TEST_APP-renamed --format json | jq -r '.redirects[0].source'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.example.com"

  run /bin/bash -c "dokku apps:rename $TEST_APP-renamed $TEST_APP"
  assert_success
}

@test "(proxy:redirect) warns when the proxy backend needs a rebuild or does not render redirects" {
  run /bin/bash -c "dokku proxy:set $TEST_APP caddy"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "dokku ps:rebuild $TEST_APP"

  run /bin/bash -c "dokku proxy:set $TEST_APP haproxy"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com --code 302"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "always responds to redirects with a 301"

  run /bin/bash -c "dokku proxy:set $TEST_APP unsupported-proxy"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.example.com example.com"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Redirects are stored but not rendered by the unsupported-proxy proxy backend"
}
