# Woodpecker CI

[Woodpecker CI](https://woodpecker-ci.org/docs/intro) can be used to automatically deploy a Dokku application via the official [dokku/ci-docker-image](https://github.com/dokku/ci-docker-image). The simplest example is as follows:

```yaml
when:
  - event: push
    branch: main

clone:
  git:
    image: woodpeckerci/plugin-git
    settings:
      partial: false
      depth: 0

steps:
  - name: deploy
    image: dokku/ci-docker-image
    environment:
      GIT_REMOTE_URL: ssh://dokku@dokku.me:22/appname
      SSH_PRIVATE_KEY:
          from_secret: DOKKU_ME_SSH_DEPLOY_KEY
    commands:
      - dokku-deploy
```

- The standard `clone` stage is replaced with one that performs a complete clone of the repository (the default is a shallow clone, which dokku would reject).
- A `GIT_REMOTE_URL` and a `SSH_PRIVATE_KEY` (with help of the `from_secret` directive) suffice to carry out a deployment.