#!/usr/bin/env bash
# watchdog process 10 sec timeout to set the known hosts
mainpid=$$
(sleep 10; kill $mainpid) &
watchdogpid=$!

SSH_ENV="/home/dokku/.ssh/environment"

function start_agent {
echo "Initialising new SSH agent..."
/usr/bin/ssh-agent | sed 's/^echo/#echo/' > "${SSH_ENV}"
echo succeeded
chmod 600 "${SSH_ENV}"
. "${SSH_ENV}" > /dev/null
/usr/bin/ssh-add;
}

echo "DOKU USER KEY:  $(ssh-keygen -lf /home/dokku/.ssh/id_rsa.pub)";
start_agent

#set know_hosts in background
keyscans=()
ssh-keyscan github.com >> /home/dokku/.ssh/known_hosts &
keyscans+=($!)
ssh-keyscan gitlab.com >> /home/dokku/.ssh/known_hosts &
keyscans+=($!)

wait "${keyscans[@]}" #test connection to github (if it adding hosts hasnt timedout)
printf "[GITHUB GREETING] %s\n" "$(ssh -T git@github.com 2>&1 >/dev/null | sed -n 's/.*\(Hi .*\!\).*/\1/gp' || true)"

# if the whole script executed, end the watchdog
kill $watchdogpid