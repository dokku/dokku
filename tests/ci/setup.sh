#!/usr/bin/env bash
set -eo pipefail

export DEBIAN_FRONTEND=noninteractive 
apt-get install -y git make curl

cd /tmp

wget http://j.mp/godeb
tar -zxvf ./godeb
./godeb install 1.1.2

export GOPATH=/root/go
git clone https://github.com/flynn/gitreceive-next.git
cd gitreceive-next && make install

ssh-keygen -f $HOME/.ssh/id_rsa -t rsa -N ''

cat<<EOF > /etc/init/gitreceived.conf
start on runlevel [2345]
exec /usr/local/bin/gitreceived -p 2022 -n /root/.ssh/id_rsa /tmp/receiver
EOF

cat<<EOF > /etc/rc.local
curl https://raw.github.com/progrium/dokku/master/tests/ci/receiver -s > /tmp/receiver
chmod +x /tmp/receiver
EOF