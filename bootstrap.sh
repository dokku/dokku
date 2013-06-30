DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}; DOKKU_BRANCH=${DOKKU_BRANCH:-"master"}
DOKKU_STACK=${DOKKU_STACK:-"https://s3.amazonaws.com/progrium-dokku/progrium_buildstep.tgz"}
set -e

DEBIAN_FRONTEND=noninteractive apt-get install -y linux-image-extra-`uname -r`
apt-get install -y python-software-properties
add-apt-repository -y ppa:dotcloud/lxc-docker
apt-get update && apt-get install -y lxc-docker git nginx make curl dnsutils

cd ~ && git clone ${DOKKU_REPO}
cd dokku && git pull origin $DOKKU_BRANCH && make install

curl "$DOKKU_STACK" | gunzip -cd | docker import - progrium/buildstep

sed -i 's/docker -d/docker -d -r=true/' /etc/init/docker.conf
sed -i 's/# server_names_hash_bucket_size/server_names_hash_bucket_size/' /etc/nginx/nginx.conf

[[ $(dig +short $HOSTNAME) ]] && hostfile="DOMAIN" || hostfile="HOSTNAME"
echo $HOSTNAME > /home/git/$hostfile

/etc/init.d/nginx start
start nginx-reloader

echo
echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""
