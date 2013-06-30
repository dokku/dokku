DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}
DOKKU_STACK=${DOKKU_STACK:-"https://s3.amazonaws.com/progrium-dokku/progrium_buildstep.tgz"}
DOCKER_PKG=${DOCKER_PKG:-"https://launchpad.net/~dotcloud/+archive/lxc-docker/+files/lxc-docker_0.4.2-1_amd64.deb"}
set -e

DEBIAN_FRONTEND=noninteractive apt-get install -y linux-image-extra-`uname -r`
apt-get update && apt-get install -y git make curl

wget -qO- "$DOCKER_PKG" > /tmp/lxc-docker_0.4.2-1_amd64.deb
dpkg --force-depends -i /tmp/lxc-docker_0.4.2-1_amd64.deb && apt-get install -f -y
rm /tmp/lxc-docker_0.4.2-1_amd64.deb

cd ~ && git clone ${DOKKU_REPO}
cd dokku && make install

curl "$DOKKU_STACK" | gunzip -cd | docker import - progrium/buildstep

PLUGINPATH=/home/git/.plugins
pluginhook install

echo
echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""
