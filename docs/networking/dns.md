# DNS Configuration

> Note: This is a work in progress.

## DNS Versions

There are many different DNS servers 'in the wild'.  Some of the popular ones on Linux are BIND, dnsmasq, and pdns.  Windows has its own built-in DNS server as well as Unbound, Posadis, and more.  A full list of DNS packages can be found on Wikipedia under [Comparison of DNS Server Software](http://en.wikipedia.org/wiki/Comparison_of_DNS_server_software).  In addition to the various DNS packages, there are tens of thousands of [Managed DNS Providers](http://en.wikipedia.org/wiki/List_of_managed_DNS_providers) out that all have different DNS interfaces.

## Focus

Because there are so many different DNS server packages out there as well as a tremendous number of Managed DNS Providers, we will focus on the concepts of DNS as well as providing examples in the 'BIND' format so you can adapt the information to your own server package or managed DNS provider.

## Assumptions

* We assume you have a passing familiarity with DNS.  If not, you can read an [in-depth article](http://www.diaryofaninja.com/blog/2012/03/03/devops-dns-for-developers-ndash-now-therersquos-no-excuse-not-to-know) on DNS.  But basically you need to know that DNS changes names (like example.tld) into addresses (like 127.0.0.1)
* We assume you already have a domain name registered and pointed to your favorite Managed DNS Provider or have your own BIND DNS server running.
* You have a server on the internet and are about to follow the instructions in the [README](https://github.com/dokku/dokku/blob/master/README.md) to get Dokku installed.  Don't do the install just yet though.

## Caching

Please remember that DNS relies heavily on _caching_.  Changes you make to DNS could take anywhere from a few seconds to a few _days_ to propagate.  If you tried surfing to example.tld, then changed the IP address in DNS, it could be a while before your computer picks up on the changes.

## Getting started

For the examples, we will use the domain name `example.tld` and the IP address `127.0.0.1`.

Dokku uses DNS to differentiate between apps on your dokku-powered server.  If you are using the domain `example.tld`, and you have two apps `node-js-app1` and `node-js-app2`, Dokku will make them available at `node-js-app1.example.tld` and `node-js-app2.example.tld`.

To get started, you need to know the IP address of your Dokku server.  Connect to it and run `ifconfig` or `ip addr` to see the IP address.

Now you have to make a decision about your domain.  Do you want everything and anything at `example.tld` to go to your Dokku server, or would you rather use a 'sub domain' for your Dokku server?

In other words, do you want your applications on your Dokku server accessible via `node-js-app.example.tld` or via `node-js-app.myserver.example.tld`?

### Using a sub-domain (node-js-app.myserver.example.tld)

Using a sub-domain is easy.  When you set up your server, you probably gave it a name like `myserver.example.tld`.

Go in to your Managed DNS provider and create an `A` record named `myserver` and put in the IP address you got from your server a few moments ago.

Hopefully your managed DNS provider also supports wildcards.  Create a second `A` record named `*.myserver` along with the IP address you got from your server a few moments ago.

If you are using BIND, your zone file will look similar to this:

```
$ORIGIN example.tld
$TTL 5m

myserver         IN      A       127.0.0.1
*.myserver       IN      A       127.0.0.1
```

You can verify your changes in Linux by trying one or more of the following commands:

* `host myserver.example.tld`
* `dig -t A myserver.example.tld`
* `nslookup myserver.example.tld`

Now is a good time to remind you that the answers you get MAY BE CACHED.

If everything is working correctly, you should also be able to query for any other name under `myserver.example.tld` and get back the IP address of your server.  Try:

* `host test.myserver.example.tld`
* `host xyzzy.myserver.example.tld`

If they all return your IP address, you have set DNS up properly for dokku.  You should also be able to `ssh root@myserver.example.tld` and access your server.

Proceed with the setup instructions in the [installation documentation](/docs/getting-started/installation.md)

### Using the root of your domain (node-js-app.example.tld)

This section is a work in progress.  It is incomplete.

Using the 'root' of your domain is nearly identical to the previous example.

* hostname is under `example.tld`, still needs `A` record.
* Update your global domain using the [domains plugin](docs/configuration/domains.md).
