# Source locations - set environment variables to override defaults
#  e.g. DOKKU_REPO=https://github.com/yourusername/dokku.git bootstrap.sh
DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}
GITRECEIVE_URL=${GITRECEIVE_URL:-"https://raw.github.com/progrium/gitreceive/master/gitreceive"}
BUILDSTEP_URL=${BUILDSTEP_URL:-"https://raw.github.com/progrium/buildstep/master/buildstep"}
BUILDSTEP_CONTAINER=${BUILDSTEP_CONTAINER:-"progrium/buildstep"}

# Docker and base dependencies
apt-get install -y linux-image-extra-`uname -r`
apt-get install -y software-properties-common
add-apt-repository -y ppa:dotcloud/lxc-docker
apt-get update
apt-get install -y lxc-docker
apt-get install -y git ruby nginx make

# install and init gitreceive
cd /usr/local/bin
wget ${GITRECEIVE_URL}
chmod +x gitreceive
gitreceive init

# install buildstep script
cd /home/git
wget ${BUILDSTEP_URL}
chmod +x buildstep

# fetch prebuilt buildstep container
docker pull ${BUILDSTEP_CONTAINER}

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