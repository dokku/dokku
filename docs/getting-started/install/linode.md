# Installing on Linode

## Enable AUFS Storage Driver

When installing Dokku on a newly-created Linode server, you'll likely see an error during Docker installation:

> Warning: current kernel is not supported by the linux-image-extra-virtual package.  We have no AUFS support.  Consider installing the packages linux-image-virtual kernel and linux-image-extra-virtual for AUFS support.

After 10 seconds, the installation will continue as normal.

This warning is the result of Linode using its own kernel, which does not support AUFS, instead of using the kernel supplied by Debian/Ubuntu. If you ignore the warning, Docker will fall back to using the DeviceMapper storage backend and Dokku will work normally. However, AUFS is better tested and will perform better.

To avoid this error message and use AUFS, follow the steps below to select the GRUB 2 kernel.


### Select the GRUB 2 kernel instead of the default Linode kernel:

1. Login to your [Linode Manager](http://manager.linode.com), open the `Dashboard` associated with your new server.
2. Click `Edit` for the `Configuration Profile`. You should now be on the `Edit Configuration Profile` page.
3. Scroll down to `Boot Settings` > `Kernel menu`. Change the "Kernel" option to "GRUB 2" and save your changes.
4. If you have already pushed the `Boot` button and your server is running, you will need to reboot it before continuing. Otherwise, you can now push the `Boot` button to start your server and proceed with the normal Dokku installation.

Once your server comes back online, you'll be running Ubuntu's default kernel. You can now follow Dokku's [normal installation instructions](/dokku/getting-started/installation/) and `bootstrap.sh` will take care of everything else.


### Verify that you are using AUFS:

Once you have fully installed Dokku and rebooted your server, you can verify that AUFS is being used with the terminal command: `docker info`. If AUFS is being used, you should see `Storage Driver: aufs` in the output.
