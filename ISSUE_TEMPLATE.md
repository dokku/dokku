> If you need help figuring out how to use a specific buildpack, or are having issues using multiple buildpacks, please see our [irc or slack channels](http://dokku.viewdocs.io/dokku/getting-started/where-to-get-help/#the-irc-and-slack-channels). Issues pertaining to buildpacks may be closed and locked.

Description of problem:


Output of the following commands

- `dokku report APP_NAME`
- `docker inspect CONTAINER_ID` (if applicable):
  (BEWARE: `docker inspect` will print environment variables for some commands, be sure you're not exposing any sensitive information when posting issues. You may replace these values with XXXXXXX):
- `cat /home/dokku/<app>/nginx.conf` (if applicable):
- Link to the exact repository being deployed (if possible/applicable):
- If a deploy is failing or behaving unexpectedly:
  - Application name
  - The type of application being deployed (node, php, python, ruby, etc.)
  - If using buildpacks, which custom buildpacks are in use
  - If using a `Dockerfile`, the contents of that file
  - If it exists, the contents of your `Procfile`.
- Output of failing Dokku commands after running `dokku trace on`
  (BEWARE: `trace on` will print environment variables for some commands, be sure you're not exposing any sensitive information when posting issues. You may replace these values with XXXXXXX):

Environment details (AWS, VirtualBox, physical, etc.):

How was Dokku installed?:

How reproducible:


Steps to Reproduce:
1.
2.
3.

Actual Results:


Expected Results:


Additional info:
