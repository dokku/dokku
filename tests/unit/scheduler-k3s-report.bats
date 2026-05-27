#!/usr/bin/env bats

load test_helper

setup() {
  global_setup
}

teardown() {
  for prop in deploy-timeout image-pull-secrets ingress-class kubeconfig-path kube-context kustomize-root-path namespace network-interface rollback-on-failure shm-size; do
    dokku scheduler-k3s:set --global "$prop" >/dev/null 2>/dev/null || true
  done
  global_teardown
}

# assert_k3s_global_unset_set verifies that a global property cycles correctly
# through the raw/computed split: when unset, the global key is empty and the
# computed key returns the built-in default; after setting, both pair members
# match the set value. Excludes letsencrypt-* properties whose set path applies
# cluster issuers via helm and requires a live k3s cluster.
assert_k3s_global_unset_set() {
  declare prop="$1" expected_default="$2" set_value="$3"

  run /bin/bash -c "dokku scheduler-k3s:set --global $prop"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-$prop\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-$prop\"'"
  assert_success
  assert_output "$expected_default"

  run /bin/bash -c "dokku scheduler-k3s:set --global $prop '$set_value'"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-$prop\"'"
  assert_success
  assert_output "$set_value"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-$prop\"'"
  assert_success
  assert_output "$set_value"

  run /bin/bash -c "dokku scheduler-k3s:set --global $prop"
  assert_success
}

@test "(scheduler-k3s:report --global) defaulted properties raw/computed cycle" {
  assert_k3s_global_unset_set "deploy-timeout" "300s" "600s"
  assert_k3s_global_unset_set "kustomize-root-path" "config/kustomize" "deploy/kustomize"
  assert_k3s_global_unset_set "namespace" "default" "production"
  assert_k3s_global_unset_set "rollback-on-failure" "false" "true"
  assert_k3s_global_unset_set "ingress-class" "nginx" "traefik"
  assert_k3s_global_unset_set "network-interface" "eth0" "eth1"
  assert_k3s_global_unset_set "kubeconfig-path" "/etc/rancher/k3s/k3s.yaml" "/tmp/custom-kubeconfig.yaml"
}

@test "(scheduler-k3s:report --global) empty-default properties expose computed sibling" {
  for prop in image-pull-secrets shm-size kube-context; do
    run /bin/bash -c "dokku scheduler-k3s:set --global $prop"
    assert_success

    run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-$prop\"'"
    assert_success
    assert_output ""

    run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-$prop\"'"
    assert_success
    assert_output ""
  done

  run /bin/bash -c "dokku scheduler-k3s:set --global image-pull-secrets my-secret"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:set --global shm-size 64Mi"
  assert_success
  run /bin/bash -c "dokku scheduler-k3s:set --global kube-context my-context"
  assert_success

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-image-pull-secrets\"'"
  assert_success
  assert_output "my-secret"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-image-pull-secrets\"'"
  assert_success
  assert_output "my-secret"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-shm-size\"'"
  assert_success
  assert_output "64Mi"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-shm-size\"'"
  assert_success
  assert_output "64Mi"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-kube-context\"'"
  assert_success
  assert_output "my-context"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-kube-context\"'"
  assert_success
  assert_output "my-context"
}

@test "(scheduler-k3s:report --global) letsencrypt properties unset state" {
  for prop in letsencrypt-server letsencrypt-email-prod letsencrypt-email-stag; do
    run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-global-$prop\"'"
    assert_success
    assert_output ""
  done

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-letsencrypt-server\"'"
  assert_success
  assert_output "prod"

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-letsencrypt-email-prod\"'"
  assert_success
  assert_output ""

  run /bin/bash -c "dokku scheduler-k3s:report --global --format json | jq -r '.\"scheduler-k3s-computed-letsencrypt-email-stag\"'"
  assert_success
  assert_output ""
}
