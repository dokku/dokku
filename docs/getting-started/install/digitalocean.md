# Digital Ocean Droplet

[Digital Ocean](https://www.digitalocean.com/products/compute/) offers a pre-installed Dokku image. You can run this image on any sized droplet, although larger droplets will allow you to run larger applications.

When choosing your Droplet configuration please disable IPv6. There are known issues with IPv6 on Digital Ocean and Docker, and many have been reported to the Dokku issue tracker. If you would like to run Dokku on an IPv6 Digital Ocean Droplet, please consult [this guide](https://jeffloughridge.wordpress.com/2015/01/17/native-ipv6-functionality-in-docker/).

## Dokku setup using a Mac

1. Login to your [Digital Ocean](https://m.do.co/c/d716c8c29fb5) account

2. Click **Create a Droplet**

3. Under **Choose an image > One-click apps** and choose **Dokku 0.6.5 on 14.04** _(version numbers may vary)_

4. Under **Choose a size** and select your a machine spec

5. Under **Choose a datacenter region** select your region

6. Under **Add your SSH keys** click **New SSH Key** _(this opens a dialog)_

7. From your terminal, execute `cat ~/.ssh/id_rsa.pub`

8. Copy the output and paste it into the **New SSH Key** dialog, provide a name and click **Add SSH Key**

9. Under **Finalize and create**, give your droplet a hostname _(not required)_ and click **Create**

10. Once created, copy the IP address to your clipboard

11. From your terminal, verify you can connect by executing `ssh root@ip-address` - then disconnect using `exit`

12. From your terminal execute `cat ~/.ssh/id_rsa.pub` to get your public key and copy it's output to your clipboard

13. Open a text editor and build the following line: `echo "paste ssh key from clipboard" | sshcommand acl-add dokku user` _(removing line breaks)_

14. From terminal execute `ssh root@ip-address` to reconnect to your server

15. From your server, execute `echo "your ssh key without line breaks should be in here" | sshcommand acl-add dokku user`

16. Then execute `exit` to disconnect from your server

17. Now try to connect as the dokku user `ssh dokku@ip-address` _(this should return dokku help and disconnect)_

18. Now, go into your git project and add git remote host `git remote add dokku dokku@ip-of-host:appname`

19. You should now be able to execute `git push dokku master`

## Troubleshooting

Should you have any issues, please visit the [freenode chat](https://webchat.freenode.net/?channels=dokku)

## Unattended Installation

It is possible to setup Dokku using an unattended installation method using the following [guide](http://dokku.viewdocs.io/dokku/getting-started/install/debian/#unattended-installation)
