# Gitlab CI
----

Gitlab-CI can be used to automatically deploy a Dokku application via the official the [dokku/ci-docker-image](https://github.com/dokku/ci-docker-image). The simplest example is as follows:

```
---
image: dokku/ci-docker-image

stages:
  - deploy

variables:
  GIT_DEPTH: 0

deploy:
  stage: deploy
  only:
    - master
  variables:
    GIT_REMOTE_URL: ssh://dokku@dokku.me:22/appname
  script: dokku-deploy
  after_script: dokku-unlock
```

For further usage documentation and other advanced examples, see Dokku's [gitlab-ci](https://github.com/dokku/gitlab-ci) repository.
