# Digital Ocean

On Digital Ocean, there is a pre-made image that can be run for Dokku. You can choose to run this image on any sized droplet, though larger droplets will allow you to run larger applications.

When choosing your Droplet configuration, please disable IPv6 on the droplet. There are known issues with IPv6 on Digital Ocean and Docker, and many have been reported to the Dokku issue tracker.

If you would like to run Dokku on an IPv6 Digital Ocean Droplet, please consult [this guide](https://jeffloughridge.wordpress.com/2015/01/17/native-ipv6-functionality-in-docker/) for modifying Docker to run under the Digital Ocean IPv6 configuration.
