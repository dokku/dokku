Description of problem:


Output of the following commands

- `uname -a`:
- `free -m`:
- `docker version`:
- `docker -D info`:
- `docker run -ti gliderlabs/herokuish:latest herokuish version`:
- `dokku version`:
- `dokku plugin`:
- `docker inspect CONTAINER_ID` (if applicable):
  (BEWARE: `docker inspect` will print environment variables for some commands, be sure you're not exposing any sensitive information when posting issues. You may replace these values with XXXXXXX):
- `cat /home/dokku/<app>/nginx.conf` (if applicable):
- Link to the exact repository being deployed (if possible/applicable):
- If a deploy is failing:
  - the type of application being deployed (node, php, python, ruby, etc.)
  - whether or not you have a `Dockerfile` in the root of your repository
  - if using buildpacks, which custom buildpacks are in use
- Output of failing dokku commands after running `dokku trace on`
  (BEWARE: `trace on` will print environment variables for some commands, be sure you're not exposing any sensitive information when posting issues. You may replace these values with XXXXXXX):

Environment details (AWS, VirtualBox, physical, etc.):

How was dokku installed?:

How reproducible:


Steps to Reproduce:
1.
2.
3.

Actual Results:


Expected Results:


Additional info:
