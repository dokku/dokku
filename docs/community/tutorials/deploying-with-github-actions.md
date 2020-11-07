# Deploying with GitHub Actions

GitHub actions can be used to build the app image and deploy it.

## Prerequisites

You'll need a SSH keypair created (without a passphrase) to allow the GitHub Action runner to SSH to the dokku server.

On your dokku server:

```bash
ssh-keygen -N "" -f /root/.ssh/githubactions
```

Add the `githubactions` public key to authorized_keys:

```bash
cat /root/.ssh/githubactions.pub >> /root/.ssh/authorized_keys
```

## Setting up the GitHub Action workflow

### Secrets

Create the following secrets in your GitHub repo:

- `SSH_PRIVATE_KEY` (the contents of `/root/.ssh/githubactions`)
- `DOKKU_HOST` (the hostname of your dokku server)
- `DOKKU_APP_NAME` (the name of the dokku app you want to deploy)

### Workflow

Create a file called `.github/workflows/publish.yml` in your repo with the following contents:

```yml
name: Publish
on:
  release:
    types: [published]

jobs:
  publish:
    name: Publish
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.SSH_PRIVATE_KEY }}" | tr -d '\r' > ~/.ssh/id_rsa
          chmod 600 ~/.ssh/id_rsa
          ssh-keyscan ${{ secrets.DOKKU_HOST }} >> ~/.ssh/known_hosts
      - name: Build docker image
        run: |
          docker build -t dokku/${{ secrets.DOKKU_APP_NAME }}:latest .
      - name: Deploy app
        run: |
          docker save dokku/${{ secrets.DOKKU_APP_NAME }}:latest | ssh root@${{ secrets.DOKKU_HOST }} "docker load | dokku tags:deploy ${{ secrets.DOKKU_APP_NAME }} latest"
```

The workflow is triggered whenever you create a new GitHub Release. Once triggered, it builds the docker image, deploys it to the dokku server, then deploys the app.
