DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}
DOKKU_STACK=${DOKKU_STACK:-"https://s3.amazonaws.com/progrium-dokku/progrium_buildstep.tgz"}
DOCKER_PKG=${DOCKER_PKG:-"https://launchpad.net/~dotcloud/+archive/lxc-docker/+files/lxc-docker_0.4.2-1_amd64.deb"}

apt-get update
apt-get install -y linux-image-extra-`uname -r`
apt-get install -y python-software-properties
apt-get install -y git ruby nginx make curl dnsutils

cd /tmp
wget "$DOCKER_PKG"
dpkg -i lxc-docker_0.4.2-1_amd64.deb
apt-get install -f -y
rm lxc-docker_0.4.2-1_amd64.deb

cd ~ && git clone ${DOKKU_REPO}
cd dokku && make install
if [[ $DOKKU_STACK ]]; then
  curl "$DOKKU_STACK" | gunzip -cd | docker import - progrium/buildstep
else
  cd buildstep && make build
fi

if [ -f /etc/nginx/nginx.conf ]; then
  sed -i 's/# server_names_hash_bucket_size/server_names_hash_bucket_size/' /etc/nginx/nginx.conf
fi

/etc/init.d/nginx start
start nginx-reloader

[[ $(dig +short $HOSTNAME) ]] && HOSTFILE=DOMAIN || HOSTFILE=HOSTNAME
echo $HOSTNAME > /home/git/$HOSTFILE

echo
echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""
