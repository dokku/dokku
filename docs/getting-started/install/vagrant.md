# Install Dokku using Vagrant

- Download and install [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
- Download and install [Vagrant](http://www.vagrantup.com/downloads.html)
- Clone Dokku

    ```shell
    git clone https://github.com/dokku/dokku.git
    ```

- Create VM

    ```shell
    # Optional ENV arguments:
    # - `BOX_NAME`
    # - `BOX_URI`
    # - `BOX_MEMORY`
    # - `DOKKU_DOMAIN`
    # - `DOKKU_IP`
    # - `FORWARDED_PORT`.
    cd path/to/dokku
    
    # for most users
    vagrant up
    
    # windows users must instead use the following in an elevated command prompt
    vagrant up dokku-windows
    ```

- Setup SSH Config in `~/.ssh/config`.

    ```ini
    Host dokku.me
        Port 22
    ```

    > For users that have customized the IP address of their VM - either in a custom `Vagrantfile` or via the `DOKKU_IP` environment variable - and are not using `10.0.0.2` for the Vagrant IP, you'll need to instead use the output of `vagrant ssh-config dokku` for your `~/.ssh/config` entry. 

- Copy your SSH key via `cat ~/.ssh/id_rsa.pub | pbcopy` and paste it into the dokku-installer at http://dokku.me . Change the `Hostname` field on the Dokku Setup screen to your domain and then check the box that says **Use virtualhost naming**. Then click **Finish Setup** to install your key. You'll be directed to application deployment instructions from here.

Please note, the `dokku.me` domain is setup to point to `10.0.0.2` along with all subdomains (i.e. `yourapp.dokku.me`). If you change the `DOKKU_IP` in your Vagrant setup you'll need to update your `/etc/hosts` file to point your reconfigured ip address.

You are now ready to deploy an app or install plugins.
