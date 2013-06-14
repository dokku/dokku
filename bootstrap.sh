DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}

apt-get install -y linux-image-extra-`uname -r` software-properties-common
add-apt-repository -y ppa:dotcloud/lxc-docker
apt-get update && apt-get install -y lxc-docker git ruby nginx make

cd ~ && git clone ${DOKKU_REPO}
cd dokku && make install
cd buildstep && make build

/etc/init.d/nginx start
start nginx-reloader

[[ $(dig +short $HOSTNAME) ]] && HOSTFILE=DOMAIN || HOSTFILE=HOSTNAME
echo $HOSTNAME > /home/git/$HOSTFILE

echo
echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""