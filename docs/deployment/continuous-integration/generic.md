# Generic CI/CD Integration

In cases where there is no direct or documented integration available, the Dokku project provides an [Official Docker Image](https://github.com/dokku/ci-docker-image) for use in Continuous Integration/Continuous Deployment (CI/CD) systems.

Assuming a Docker image can be run as a CI task with environment variables injected, the following CI systems will have their variables automatically detected:

- [circleci](https://circleci.com/)
- [cloudbees](https://www.cloudbees.com/)
- [drone](https://www.drone.io/)
- [github actions](https://github.com/features/actions)
- [gitlab-ci](https://about.gitlab.com/stages-devops-lifecycle/continuous-integration/)
- [semaphoreci](https://semaphoreci.com/)
- [travisci](https://travis-ci.com/)

## Simple Usage

The simplest usage of the image is as follows.

```shell
# where the `.env` file contains `GIT_REMOTE_URL` and `SSH_PRIVATE_KEY`

docker run --rm -v="$PWD:/app" --env-file=.env dokku/ci-docker-image dokku-deploy
```

For more configuration examples and further documentation, see the [ci-docker-image](https://github.com/dokku/ci-docker-image) readme.
