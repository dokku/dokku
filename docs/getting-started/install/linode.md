# Installing on Linode

## Using StackScript

Deploy using the following StackScript:
* https://www.linode.com/stackscripts/view/8552

## Without StackScript

* Build a Ubuntu 13.04 instance

* Follow these instructions: https://www.linode.com/wiki/index.php/PV-GRUB#Ubuntu_12.04_Precise

* If `apt-get update` no longer works:

    * Verify if apt-get is trying to use ipv6 instead of ipv4 (e.g. you read something like "[Connecting to us.archive.ubuntu.com (2001:67c:1562::14)]" and apt-get would not proceed). In that case, follow these instructions: http://unix.stackexchange.com/questions/9940/convince-apt-get-not-to-use-ipv6-method (append "precedence ::ffff:0:0/96  100" to /etc/gai.conf)

    * OR: change `/etc/apt/sources.list` to one mentioned in http://mirrors.ubuntu.com/mirrors.txt

* Run the following commands:

    ```shell
    apt-get update

    apt-get install lxc wget bsdtar linux-image-extra-$(uname -r)

    modprobe aufs
    ```
* After this, you can install dokku the default way:

    ```shell
    wget -qO- https://raw.github.com/progrium/dokku/master/bootstrap.sh | sudo bash
    ```
