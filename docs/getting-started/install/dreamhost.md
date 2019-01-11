# Install Dokku on DreamHost Cloud Server with cloud-init

Cloud-init script can be used to automate installation of Dokku on
Dreamhost (or any other OpenStack-compatible cloud with minimal
changes).

A new server instance can be created on DreamHost Cloud from the command line
using OpenStack client or from the web UI and with the same command
use a cloud-init script to install Dokku. Install the [OpenStack
CLI](https://help.dreamhost.com/hc/en-us/articles/216185658-How-to-Install-the-OpenStack-command-line-clients),
download the [DreamHost Cloud credentials
file](https://iad2.dreamcompute.com/project/access_and_security/api_access/openrc/)
before proceeding and make sure your public SSH key is added to the
cloud.

```sh
source openrc.sh # Set the environment variables for DreamHost Cloud
```

This allows OpenStack client to connect to DreamHost API endpoints.
The command below creates a new server instance named `my-dokku-instance`
based on Ubuntu 14.04, with 2GB RAM and 1CPU (the flavor called
`supersonic`), opening network port access to HTTP and SSH (the
`default` security group), and the name of the chosen SSH key. This
key will be automatically added to the new server in the
`authorized_keys` for the default SSH user (`ubuntu`), and it will
be reused by Dokku.

```sh
openstack server create \
  --image Ubuntu-14.04 \
  --flavor gp1.supersonic \
  --security-group default \
  --key-name $YOUR_SSH_KEYNAME \
  --user-data dokku-cloudinit.sh \
  my-dokku-instance
```

The content of dokku-cloudinit.sh script contains instructions to add
Docker and Dokku's apt repositories and install Dokku with the proper
debconf options set. Don't forget to add the FQDN for your application
server:

```yaml
#cloud-config
apt_upgrade: true

apt_sources:
 - source: "deb https://packagecloud.io/dokku/dokku/ubuntu/ trusty main"
   key: |
    -----BEGIN PGP PUBLIC KEY BLOCK-----
    Version: SKS 1.1.5
    Comment: Hostname: pgpkeys.eu

    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
    =H60S
    -----END PGP PUBLIC KEY BLOCK-----
   filename: dokku.list
 - source: "deb https://apt.dockerproject.org/repo ubuntu-trusty main"
   key: |
    -----BEGIN PGP PUBLIC KEY BLOCK-----
    Version: SKS 1.1.5
    Comment: Hostname: pgpkeys.co.uk

    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
    =Yhur
    -----END PGP PUBLIC KEY BLOCK-----
   filename: docker.list

package_upgrade: true

debconf_selections: |

    dokku dokku/web_config boolean false
    dokku dokku/vhost_enable boolean true
    # set the domain name of the new Dokku server
    dokku dokku/hostname string $YOUR_FULL_QUALIFIED_DOMAIN
    # this copies over the public SSH key assigned to the server
    dokku dokku/key_file string /home/ubuntu/.ssh/authorized_keys

packages:
   - dokku
```

Shortly after running the create command you will get a confirmation that the
instance has been created, and after about a minute it should be ready to login.
Check the IP of the instance through the web UI or by running:

```sh
nova list
```

SSH with the `ubuntu` username and the public key previously added.
Keep in mind that if you logged in quick enough dokku might still be installing
in the background, and not be ready. The installation takes a few minutes.
