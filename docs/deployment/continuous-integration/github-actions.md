# GitHub Actions

The Dokku project has an official GitHub Action available on the [GitHub Marketplace](https://github.com/marketplace/actions/dokku).

## Simple example

This example assumes that the GitHub repository default branch is `main` and not `master` (GitHub's default for new repos since Oct 2020), and Dokku is configured to match this by deploying the `main` branch of all apps i.e. `dokku git:set --global deploy-branch main` has previously been run.

The main branch will be configured to deploy to an app in Dokku which is used as the production environment.

This workflow will run on every push to the `main` branch (including `--force` pushes), and sensitive information is stored in the GitHub repo settings rather than directly in the repo code.

The workflow can be duplicated to create a workflow for other branches e.g `staging` deploys to a different Dokku app.

### Add the workflow to the repo

Create a file in your repository `.github/workflows/deploy-production` and push to GitHub.

```
yaml
---
name: 'deploy-production'

on:
  push:
    branches:
      - main

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Cloning repo
        uses: actions/checkout@v2
        with:
          fetch-depth: 0

      - name: Push to dokku
        uses: dokku/github-action@master
        with:
          branch: 'main'
          git_push_flags: '--force'
          git_remote_url: ${{ secrets.GIT_REMOTE_URL }}
          ssh_private_key: ${{ secrets.SSH_PRIVATE_KEY }}
```


### Generate an SSH key pair for the GitHub action and Dokku to use

To generate a new pair of keys refer to [GitHub's own docs](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent).

Once deploments to the Dokku server via a GitHub action are configured and working, the keys can be deleted from your local machine for security, as a new pair can be generated easily if they're ever needed.

```
ssh-keygen -t ed25519 -C "your_email@example.com"
# choose a directory to save the key pair in e.g. desktop
# do not overwrite any of your own keys
# leave the passphrase blank (passphrases require human intervention or a keychain store)
```

### Configure the repo's "actions secrets"

Instead of hard coding the SSH url and the key in the repo, create two new actions secrets in the Github repo's settings e.g. [https://github.com/user/repo/settings/secrets/actions]([https://github.com/user/repo/settings/secrets/actions]) for the workflow to use.

- `GIT_REMOTE_URL`, in SSH format e.g. `ssh://dokku@dokku.myhost.ca:22/myappname`
- `SSH_PRIVATE_KEY`, the contents of the private key which was generated

```
# to copy the contents of the private key to your clipboard on MacOS
cat /path-to-keys/id_ed25519 | pbcopy 
```

### Add the public key to Dokku

Refer to Dokku's [user management documentation](https://dokku.com/docs/deployment/user-management/).

Pipe the contents of the key to a Dokku command over SSH. Note that the root user is required to manage keys, as within Dokku these keys represent users.

For security the key name used for each should be unique to each repo so that each repo's access to the Dokku server can be revoked at a later date if necessary.

```
cat /path-to-keys/id_ed25519.pub | ssh root@dokku.myhost.ca dokku ssh-keys:add myappname-github-action-key
```

Confirm that the public key has been added.

```
dokku ssh-keys:list
```

### Test the workflow

Enable event logging and follow logs.

```
dokku events:on
dokku events -t # tail events
# or
dokku logs -t # tail log
```

Create a temporary file and commit it to the repo.

```
cd /path-to-repo/
touch deploy-test.txt
git add -A && git commit -m 'chore: Test deployment with a temporary file'
git push origin head
```

## Advanced examples

For further usage documentation and other advanced examples, refer to the official Dokku GitHub action [repo](https://github.com/dokku/github-action), the action in the [marketplace](https://github.com/marketplace/actions/dokku) and the [Dokku documentation](https://dokku.com/docs/deployment/continuous-integration/github-actions/).
