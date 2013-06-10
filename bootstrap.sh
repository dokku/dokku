# Repository locations - set environment variables to override defaults
#  e.g. OVERRIDE_DOKKU_REPO=https://github.com/yourusername/dokku.git bootstrap.sh
GITRECEIVE_URL=${GITRECEIVE_URL:-"https://raw.github.com/progrium/gitreceive/master/gitreceive"}
BUILDSTEP_REPO=${BUILDSTEP_REPO:-"https://github.com/progrium/buildstep.git"}
DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}

# Docker and base dependencies
apt-get install -y linux-image-extra-`uname -r`
apt-get install -y software-properties-common
add-apt-repository -y ppa:dotcloud/lxc-docker
apt-get update
apt-get install -y lxc-docker
apt-get install -y git ruby nginx make

# gitreceive
cd /usr/local/bin
wget ${GITRECEIVE_URL}
chmod +x gitreceive
gitreceive init

# buildstep
cd ~
git clone ${BUILDSTEP_REPO}
cd buildstep
cp buildstep /home/git/buildstep
make

# dokku (this!)
cd ~
git clone ${DOKKU_REPO}
cd dokku
cp receiver /home/git/receiver
cp deploystep /home/git/deploystep
cp nginx-app-conf /home/git/nginx-app-conf
cp nginx-reloader.conf /etc/init/nginx-reloader.conf
stop nginx-reloader
start nginx-reloader
echo $HOSTNAME > /home/git/DOMAIN

# configure and start nginx
echo "include /home/git/*/nginx.conf;" > /etc/nginx/conf.d/dokku.conf
/etc/init.d/nginx start

echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""