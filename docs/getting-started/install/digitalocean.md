# DigitalOcean Droplet Installation Notes

[DigitalOcean](https://marketplace.digitalocean.com/apps/dokku?refcode=fe06b043a083) offers a pre-installed Dokku image. You can run this image on any sized Droplet, although larger Droplets will allow you to run larger applications.

> **Please disable IPv6**. There are known issues with IPv6 on DigitalOcean and Docker. If you would like to run Dokku on an IPv6 DigitalOcean Droplet, please consult [this guide](https://jeffloughridge.wordpress.com/2015/01/17/native-ipv6-functionality-in-docker/).

1. Login to your [DigitalOcean](https://m.do.co/c/fe06b043a083) account.
2. Click **Create a Droplet**.
3. Under **Choose an image > Marketplace**, search latest **Dokku** release for Ubuntu 24.04 _(version numbers may vary)_.
4. Under **Choose a size**, select your machine spec.
5. Under **Choose a datacenter region**, select your region.
6. Add an SSH Key.
   - New Keys
     1. Under **Add your SSH keys** click **New SSH Key** _(this opens a dialog)_.
     2. From your terminal, execute `cat $HOME/.ssh/id_rsa.pub`.
     3. Copy the output and paste it into the **New SSH Key** dialog, provide a name and click **Add SSH Key**.
   - Existing Keys
     1. Simply add a checkmark next to the existing keys you'd like to add.
7. Under **Finalize and create**, give your Droplet a hostname _(not required)_ and click **Create**.
8. Once created, copy the IP address to your clipboard.
9. In a terminal, ssh onto the server by running `ssh root@$SERVER_IP`, where `$SERVER_IP` is your server's IP address.
10. Remove the default nginx site via `rm /etc/nginx/sites-enabled/default`
11. If you added more than one key, use the `dokku ssh-keys:add` to add an ssh key to the dokku user:

    ```shell
    echo "$CONTENTS_OF_YOUR_PUBLIC_SSH_KEY_HERE" | dokku ssh-keys:add KEY_NAME
    ```

12. Ensure your server has a correct global hostname via `dokku domains:set-global`

    ```shell
    # replace dokku.me with a hostname pointed at your server
    dokku domains:set-global dokku.me
    ```

13. Once the web UI has been submitted, you will be redirected to our [application deployment tutorial](/docs/deployment/application-deployment.md), which will guide you through deploying a sample application to your Dokku server.
