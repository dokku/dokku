# Microsoft Azure Installation Notes

1. If you don't already have one [generate an SSH key pair](https://help.github.com/articles/generating-ssh-keys/).

2. Go to the [Dokku on Azure deployment page](https://github.com/Azure/azure-quickstart-templates/tree/master/application-workloads/dokku/dokku-vm) and click **Deploy to Azure**.

3. You'll be prompted to enter a few parameters, including a unique storage account name and a unique name for the subdomain used for your public IP address. For the `sshKeyData` parameter, copy and paste the contents of the *public* key file you just created. After a few minutes the Dokku instance will be deployed.

4. In your browser of choice, navigate to `http://[dnsNameForPublicIP].[location].cloudapp.azure.com`. Where `[dnsNameForPublicIP]` and `[location]` are template parameters you used to deploy the template.

5. Finish your Dokku setup like you normally would by creating a *new* public/private key pair for your deployments using `ssh-keygen` (don't use the same one as you created in the first step). You should select **Use Virtual Host Naming** and set the **Hostname** to a *public DNS name* that you own such as one you would purchase from [Namecheap](http://namecheap.com). Alternatively thanks to [xip.io](http://xip.io/) you can just use `[yourAzurePublicIP].xip.io` for free. For example, if your public IP is `44.44.44.44` then you would set it to `44.44.44.44.xip.io`.
