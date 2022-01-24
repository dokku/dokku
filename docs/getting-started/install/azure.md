# Microsoft Azure Installation Notes

1. If you don't already have one [generate an SSH key pair](https://help.github.com/articles/generating-ssh-keys/).

2. Go to the [Dokku on Azure deployment page](https://github.com/Azure/azure-quickstart-templates/tree/master/application-workloads/dokku/dokku-vm) and click **Deploy to Azure**.

3. You'll be prompted to enter a few parameters, including a unique storage account name and a unique name for the subdomain used for your public IP address. For the `sshKeyData` parameter, copy and paste the contents of the *public* key file you just created. After a few minutes the Dokku instance will be deployed.

4. Once the installation is complete, you should configure an ssh key and set your global domain.

    ```shell
    # usually your key is already available under the current user's `~/.ssh/authorized_keys` file
    cat ~/.ssh/authorized_keys | dokku ssh-keys:add admin

    # you can use any domain you already have access to
    # this domain should have an A record or CNAME pointing at your server's IP
    dokku domains:set-global dokku.me

    # you can also use the ip of your server
    dokku domains:set-global 10.0.0.2

    # finally, you can use sslip.io to get subdomain support
    # as you would with a regular domain name
    # this would be done by appending `.sslip.io` to your ip address
    dokku domains:set-global 10.0.0.2.sslip.io
    ```

   See the [user management](/docs/deployment/user-management.md#adding-ssh-keys) and [domains documentation](/docs/configuration/domains.md#customizing-hostnames) for more information.
