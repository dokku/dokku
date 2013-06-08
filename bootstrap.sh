apt-get install -y linux-image-extra-`uname -r`
apt-get install -y software-properties-common
add-apt-repository -y ppa:dotcloud/lxc-docker
apt-get update
apt-get install -y lxc-docker
apt-get install -y git ruby nginx make

cd /usr/local/bin
wget https://raw.github.com/progrium/gitreceive/master/gitreceive
chmod +x gitreceive
gitreceive init

cd ~
git clone https://github.com/progrium/buildstep.git
cd buildstep
cp buildstep /home/git/buildstep
make


cd ~
git clone https://github.com/progrium/dokku.git
cd dokku
cp receiver /home/git/receiver
cp nginx-app-conf /home/git/nginx-app-conf
cp nginx-reloader.conf /etc/init/nginx-reloader.conf

start nginx-reloader

echo $HOSTNAME > /home/git/DOMAIN

echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOStNAME \"gitreceive upload-key progrium\""