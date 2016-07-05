# Installing on Linode

When installing Dokku on a Linode server, you'll likely see an error during Docker installation:

> Warning: current kernel is not supported by the linux-image-extra-virtual package.  We have no AUFS support.  Consider installing the packages linux-image-virtual kernel and linux-image-extra-virtual for AUFS support.

After 10 seconds, the installation will continue as normal.

This warning is the result of Linode using its own kernel, which does not support AUFS, instead of using the kernel supplied by Ubuntu. If you ignore the warning, Docker will fall back to using the DeviceMapper storage backend and Dokku will work normally. However, AUFS is better tested and will perform better.

If you would like to use AUFS, follow the steps below to install Ubuntu's kernel and configure your server to boot it instead of Linode's.

## Preparing your Linode for AUFS

__Warning__: These steps will delete *everything* on your Linode.

1. Open your server's dashboard in the [Linode Manager](https://manager.linode.com/).

2. Make sure your Linode is [using KVM](https://www.linode.com/docs/platform/kvm#how-to-enable-kvm), not Xen, for virtualization.

3. In the "Rebuild" tab, select "Ubuntu 14.04 LTS", set a root password, and rebuild.

4. Once your Linode has been created, click "Boot" and wait for it to complete.

5. SSH into your Linode as root and run the following commands:

    ```shell
    apt-get update
    apt-get -qq upgrade
    apt-get -qq install linux-image-virtual linux-image-extra-virtual
    ```

6. When prompted, install Grub onto the first hard drive.

7. Back in your server's dashboard, click "Edit" on its Configuration Profile

8. Change the "Kernel" option to "GRUB 2" and save your changes.

9. Lastly, reboot the Linode.

Once your server comes back online, you'll be running Ubuntu's default kernel. You can now follow Dokku's [normal installation instructions](/dokku/getting-started/installation/) and `bootstrap.sh` will take care of everything else.
