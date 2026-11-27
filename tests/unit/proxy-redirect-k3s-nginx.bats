#!/usr/bin/env bats

load test_helper

TEST_APP="rdmtestapp"

setup() {
  uninstall_k3s || true
  global_setup
  dokku nginx:stop
  export KUBECONFIG="/etc/rancher/k3s/k3s.yaml"
}

teardown() {
  dokku scheduler-k3s:set --global letsencrypt-email-stag >/dev/null 2>&1 || true
  dokku scheduler-k3s:set --global letsencrypt-server >/dev/null 2>&1 || true
  global_teardown
  dokku nginx:start
  uninstall_k3s || true
}

@test "(proxy-redirect:k3s-nginx) redirects a source domain and keeps the request uri" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "deploy the app or run \`dokku ps:rebuild $TEST_APP\`"

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "301" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(proxy-redirect:k3s-nginx) non-default codes are used for the redirect" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN --code 307"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get ingress $TEST_APP-web-redirect-www-$TEST_APP-dokku-me -o jsonpath='{.metadata.annotations.nginx\.ingress\.kubernetes\.io/configuration-snippet}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "return 307 http://$TEST_APP.$DOKKU_DOMAIN\$request_uri;"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "307" "http://$TEST_APP.$DOKKU_DOMAIN/"
}

@test "(proxy-redirect:k3s-nginx) a source that is also an app domain moves from the web ingress to the redirect and stays on the certificate" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=nginx install_k3s

  run /bin/bash -c "dokku scheduler-k3s:set --global letsencrypt-email-stag test@dokku.me"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:set --global letsencrypt-server staging"
  assert_success

  run /bin/bash -c "dokku apps:create $TEST_APP"
  assert_success

  run /bin/bash -c "dokku domains:set $TEST_APP $TEST_APP.$DOKKU_DOMAIN www.$TEST_APP.$DOKKU_DOMAIN"
  assert_success

  run /bin/bash -c "dokku proxy:redirect:set $TEST_APP www.$TEST_APP.$DOKKU_DOMAIN $TEST_APP.$DOKKU_DOMAIN"
  assert_success

  run deploy_app python "dokku@$DOKKU_DOMAIN:$TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  run /bin/bash -c "kubectl get ingress -o json | jq -r '.items[] | select(.metadata.annotations[\"dokku.com/ingress-method\"] == \"domains\") | .spec.rules[].host' | sort | xargs"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.$DOKKU_DOMAIN"

  run /bin/bash -c "kubectl get ingress -o json | jq -r '.items[] | select(.metadata.annotations[\"dokku.com/ingress-method\"] == \"redirects\") | .spec.rules[].host'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "www.$TEST_APP.$DOKKU_DOMAIN"

  run /bin/bash -c "kubectl get certificate $TEST_APP-web -o json | jq -r '.spec.dnsNames | sort | join(\" \")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.$DOKKU_DOMAIN www.$TEST_APP.$DOKKU_DOMAIN"
}
