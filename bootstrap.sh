set -e
export DEBIAN_FRONTEND=noninteractive
export DOKKU_REPO=${DOKKU_REPO:-"https://github.com/progrium/dokku.git"}

apt-get update
apt-get install -y git make curl

cd ~ && test -d dokku || git clone $DOKKU_REPO
cd dokku && test $DOKKU_BRANCH && git checkout origin/$DOKKU_BRANCH || true
make all

echo
echo "Be sure to upload a public key for your user:"
echo "  cat ~/.ssh/id_rsa.pub | ssh root@$HOSTNAME \"gitreceive upload-key progrium\""
