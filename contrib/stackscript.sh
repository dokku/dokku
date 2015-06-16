#!/bin/bash
# <UDF name="hostname" default="" label="Hostname for dokku instance" example="example.com">
# <UDF name="ssh_key" default="" label="Public SSH Key for root user" example="Sets the root user's public ssh key, which is also automatically imported into the dokku installer">
# <UDF name="notify_email" default="" Label="Send Finish Notification To" example="Email address to send notification to when finished." />

function logit {
  # Simple logging function that prepends an easy-to-find marker '=> ' and a timestamp to a message
  TIMESTAMP=$(date -u +'%m/%d %H:%M:%S')
  MSG="=> ${TIMESTAMP} $1"
  echo ${MSG}
}

function set_ssh_key {
  if [ -n "${SSH_KEY}" ]; then
    logit "Setting root ssh key"
    mkdir -p /root/.ssh
    chmod 700 /root/.ssh
    echo "$SSH_KEY" > /root/.ssh/authorized_keys
    chmod 600 /root/.ssh/authorized_keys
    chown -R root:root /root/.ssh
  fi
}

function set_passwordless_ssh {
  logit "Turn off password authentication and root login for SSH"
  echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config
  service ssh restart
}

function system_primary_ip {
  # returns the primary IP assigned to eth0
  ifconfig eth0 | awk -F: '/inet addr:/ {print $2}' | awk '{ print $1 }'
}

function set_hostname {
  logit "Set up hostname"
  if [ -n "${HOSTNAME}" ]; then
    echo $HOSTNAME > /etc/hostname
    echo $IPADDR $FQDN $HOSTNAME >> /etc/hosts
  else
    system_primary_ip > /etc/hostname
    echo "$(system_primary_ip) localhost" >> /etc/hosts
  fi

  IPADDR=$(/sbin/ifconfig eth0 | awk '/inet / { print $2 }' | sed 's/addr://')
  hostname -F /etc/hostname
}

function postfix_install_loopback_only {
  logit "Installing and configuring Postfix"
  # Installs postfix and configure to listen only on the local interface. Also
  # allows for local mail delivery

  echo "postfix postfix/destinations string localhost.localdomain, localhost" | debconf-set-selections
  echo "postfix postfix/mailname string localhost" | debconf-set-selections
  echo "postfix postfix/main_mailer_type select Internet Site" | debconf-set-selections
  echo "postfix postfix/myhostname string localhost" | debconf-set-selections
  sudo apt-get install -qq -y postfix > /dev/null 2>&1
  /usr/sbin/postconf -e "inet_interfaces = loopback-only"
  #/usr/sbin/postconf -e "local_transport = error:local delivery is disabled"

  touch /tmp/restart-postfix
}

function notify_install_via_email {
  if [ -n "${NOTIFY_EMAIL}" ]; then
    logit "Sending notification email to ${NOTIFY_EMAIL}"
    /usr/sbin/sendmail "${NOTIFY_EMAIL}" <<EOD
To: ${NOTIFY_EMAIL}
Subject: Dokku installation is complete
From: Dokku StackScript <no-reply@${HOSTNAME}>

Your Dokku installation is complete and now ready to be configured: http://$(system_primary_ip) . Please visit this url to complete the setup of your Dokku instance.

Enjoy using Dokku!
EOD
  fi
}

function notify_restart_via_email {
  if [ -n "${NOTIFY_EMAIL}" ]; then
    logit "Sending notification email to ${NOTIFY_EMAIL} of required restart"
    /usr/sbin/sendmail "${NOTIFY_EMAIL}" <<EOD
To: ${NOTIFY_EMAIL}
Subject: Dokku Linode instance must be restarted
From: Dokku StackScript <no-reply@${HOSTNAME}>

The following linode instance must be restarted:

    ${LINODE_LISHUSERNAME}

Before restarting, please go to this url:

    https://manager.linode.com/linodes/dashboard/${LINODE_LISHUSERNAME}

Then click "Edit" next to the selected configuration profile and make the following changes:

- Change the "Kernel" option to the current "pv-grub" release
- Set the "Xenify Distro" option to "no"

Then save your changes. Next, reboot the instance from the Linode Dashboard. You'll receive an email once the instance is available to continue the dokku installation.
EOD
  fi
}

function setup_linode {
  logit "Installing via linode"
  DEBIAN_FRONTEND=noninteractive apt-get install -qq -y linux-virtual
  DEBIAN_FRONTEND=noninteractive apt-get purge -qq -y grub2 grub-pc
  DEBIAN_FRONTEND=noninteractive apt-get install -qq -y grub
  mkdir -p /boot/grub
  update-grub -y
  sed -i 's/kopt=root=UUID=.* ro/kopt=root=\/dev\/xvda console=hvc0 ro quiet/g' /boot/grub/menu.lst
  sed -i 's/# groot=(hd0,0)/# groot=(hd0)/g' /boot/grub/menu.lst
  update-grub

  cp /etc/rc.local /etc/rc.local-bak
  cat << "EOF" > /etc/rc.local
#!/bin/sh -e
sudo apt-get update >> /root/setup_linode.txt 2>&1
sudo apt-get install -y linux-image-extra-"$(uname -r)" >> /root/setup_linode.txt 2>&1
modprobe aufs >> /root/setup_linode.txt 2>&1

sudo apt-get install -y apparmor cgroup-lite >> /root/setup_linode.txt 2>&1

sudo apt-get install -y dokku >> /root/setup_linode.txt 2>&1

# Clean up this script so it only runs once
rm -f /etc/rc.local
mv /etc/rc.local-bak /etc/rc.local
exit 0
EOF
  chmod +x /etc/rc.local
}

function install_prerequisites {
  sudo apt-get install -qq -y curl > /dev/null 2>&1

  logit "Installing docker gpg key"
  curl -sSL https://get.docker.com/gpg 2> /dev/null | apt-key add - > /dev/null 2>&1

  logit "Installing dokku gpg key"
  curl -sSL https://packagecloud.io/gpg.key 2> /dev/null | apt-key add - > /dev/null 2>&1

  logit "Running apt-get update"
  sudo apt-get update > /dev/null

  logit "Installing pre-requisites"
  sudo apt-get install -qq -y apt-transport-https > /dev/null 2>&1

  logit "Setting up apt repositories"
  echo "deb http://get.docker.io/ubuntu docker main" > /etc/apt/sources.list.d/docker.list
  echo "deb https://packagecloud.io/dokku/dokku/ubuntu/ trusty main" > /etc/apt/sources.list.d/dokku.list

  logit "Running apt-get update"
  sudo apt-get update > /dev/null
}

function install_dokku {
  logit "Installing pre-requisites"
  sudo apt-get install -qq -y linux-image-extra-"$(uname -r)" > /dev/null 2>&1

  logit "Installing dokku"
  sudo apt-get install -qq -y dokku > /dev/null 2>&1

  logit "Done!"
}

exec &> /root/stackscript.log

set_ssh_key
set_passwordless_ssh
postfix_install_loopback_only
set_hostname
install_prerequisites

if [ -n "$LINODE_ID" ]; then
  setup_linode
  notify_restart_via_email
else
  install_dokku
  notify_install_via_email
fi
