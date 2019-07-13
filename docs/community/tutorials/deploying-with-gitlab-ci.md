# Deploying with Gitlab CI

Gitlab-CI can be used to automatically deploy a Dokku application using the [ilyasemenov/gitlab-ci-git-push image](https://github.com/IlyaSemenov/gitlab-ci-git-push) docker image.

## Prerequisites

Make sure you have a Gitlab account and a Dokku project hosted on Gitlab. This method works whether if you are using buildpacks or Dockerfile.

Make sure you have set up an app on the remote machine following [these instructions](http://dokku.viewdocs.io/dokku/deployment/application-deployment/) and can successfully deploy to it from the local machine.

## Deploy automatically to production

### Add a secret variable

Browse to the repository in question and visit the following path: `the Gitlab project > Settings > CI/CD.`

Click on `Secret variables > Expand` and fill in the blanks.

- Key: `SSH_PRIVATE_KEY`
- Value: paste in an SSH private key registered in Dokku:

    -----BEGIN RSA PRIVATE KEY-----
    ...
    -----END RSA PRIVATE KEY-----

- Environment scope: `production` (This make sure that `SSH_PRIVATE_KEY` is not available on merge requests or tests)
- Protected: Do not check this checkbox unless you know what you are doing

### Add CI script

Create a file named `.gitlab-ci.yml` at the root directory of the repository with the following contents:

```yaml
stages:
  - deploy

variables:
  APP_NAME: node-js-app

deploy:
  image: ilyasemenov/gitlab-ci-git-push
  stage: deploy
  environment:
    name: production
    url: https://$APP_NAME.dokku.me/
  only:
    - master
  script:
    - git-push ssh://dokku@dokku.me:22/$APP_NAME
```

You will need to modify the `APP_NAME` variable to the correct value for your application name.

Running `git push origin master` will now trigger a gitlab-ci pipeline that will automatically deploy your application to your Dokku server. Go to your project on Gitlab and visit "Project > Pipelines" to see the deployment log.


## Review Applications

One useful feature of gitlab is to be able to create review applications for non-production branches. This allows teams to review changes _before_ they are pushed to production. First, recreate the `SSH_PRIVATE_KEY` secret - do not delete the existing secret - but scoped to the `review/*` environment. This will allow non-production gitlab environments access to the secret.

Next, we'll need to create the `review_app` gitlab job in our `.gitlab-ci.yaml`.

```yaml
review_app:
  image: ilyasemenov/gitlab-ci-git-push
  stage: deploy
  environment:
    name: review/$CI_COMMIT_REF_NAME
    url: https://$CI_ENVIRONMENT_SLUG.dokku.me/
    on_stop: stop_review_app
  only:
    - branches
  except:
    - master
  script:
    - mkdir -p ~/.ssh && echo "$SSH_PRIVATE_KEY" | tr -d '\r' > ~/.ssh/id_rsa && chmod 600 ~/.ssh/id_rsa
    - ssh-keyscan -H 22 "dokku.me" >> ~/.ssh/known_hosts
    - ssh -t dokku@dokku.me -- apps:clone --ignore-existing --skip-deploy "$APP_NAME" "$CI_ENVIRONMENT_SLUG"
    - git-push ssh://dokku@dokku.me:22/$CI_ENVIRONMENT_SLUG
```

The first two lines ensure that Gitlab can talk to the Dokku server. Next, we take advantage of the `apps:clone` command, and clone the existing application and all of it's configuration to a new location. If the new application already exists, the third line of the `script` step will be ignored. We also ignore the first deploy script to speed up the cloning process. Finally, the fourth line of the `script` step will deploy the code as normal.

The above only runs for non-master branches, and will _also_ trigger an `on_stop` job called `stop_review_app`. When the branch is deleted or the code is merged, the `stop_review_app` job will be triggered.

```yaml
stop_review_app:
  image: ilyasemenov/gitlab-ci-git-push
  stage: deploy
  variables:
    GIT_STRATEGY: none
  environment:
    name: review/$CI_COMMIT_REF_NAME
    action: stop
  when: manual
  script:
    - mkdir -p ~/.ssh && echo "$SSH_PRIVATE_KEY" | tr -d '\r' > ~/.ssh/id_rsa && chmod 600 ~/.ssh/id_rsa
    - ssh-keyscan -H 22 "dokku.me" >> ~/.ssh/known_hosts
    - ssh -t dokku@dokku.me -- --force apps:destroy "$CI_ENVIRONMENT_SLUG"
```

The `stop_review_app` step will delete the temporary application, cleaning up now unused server resources.

Going further, the following modifications can be done:

- Notify via chat whenever a review application is created/destroyed.
- Push to a separate, staging server as opposed to a production server.
- Clone a "staging" application so that review applications do not affect production datasets.
