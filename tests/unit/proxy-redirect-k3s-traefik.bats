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

@test "(proxy-redirect:k3s-traefik) redirects a source domain and keeps the request uri" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

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

  run /bin/bash -c "dokku ps:rebuild $TEST_APP"
  echo "output: $output"
  echo "status: $status"
  assert_success

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/some/path?query=1" "301" "http://$TEST_APP.$DOKKU_DOMAIN/some/path?query=1"
  assert_http_localhost_response_contains "http" "$TEST_APP.$DOKKU_DOMAIN" "80" "/" "python/http.server"
}

@test "(proxy-redirect:k3s-traefik) temporary codes render a non-permanent redirect" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

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

  run /bin/bash -c "kubectl get middleware.traefik.io $TEST_APP-web-redirect-www-$TEST_APP-dokku-me -o jsonpath='{.spec.redirectRegex.permanent}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "false"

  assert_http_localhost_redirect "http" "www.$TEST_APP.$DOKKU_DOMAIN" "80" "/" "302" "http://$TEST_APP.$DOKKU_DOMAIN/"
}

@test "(proxy-redirect:k3s-traefik) a source that is also an app domain moves from the web route to the redirect and stays on the certificate" {
  if [[ -z "$DOCKERHUB_USERNAME" ]] || [[ -z "$DOCKERHUB_TOKEN" ]]; then
    skip "skipping due to missing docker.io credentials DOCKERHUB_USERNAME:DOCKERHUB_TOKEN"
  fi

  INGRESS_CLASS=traefik install_k3s

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

  run /bin/bash -c "kubectl get ingressroute.traefik.io $TEST_APP-web-http-80-5000 -o json | jq -r '.spec.routes[].match' | sort -u | xargs -0"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output_contains "Host(\`$TEST_APP.$DOKKU_DOMAIN\`)"
  assert_output_not_contains "www.$TEST_APP.$DOKKU_DOMAIN"

  run /bin/bash -c "kubectl get ingressroute.traefik.io $TEST_APP-web-redirect-www-$TEST_APP-dokku-me -o jsonpath='{.spec.routes[0].match}'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "Host(\`www.$TEST_APP.$DOKKU_DOMAIN\`)"

  run /bin/bash -c "kubectl get certificate $TEST_APP-web -o json | jq -r '.spec.dnsNames | sort | join(\" \")'"
  echo "output: $output"
  echo "status: $status"
  assert_success
  assert_output "$TEST_APP.$DOKKU_DOMAIN www.$TEST_APP.$DOKKU_DOMAIN"
}
