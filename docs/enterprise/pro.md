# Dokku Pro

Dokku Pro is a commercial offering that provides a familiar Web UI for all common tasks performed by developers. End users can expect an interface that provides various complex cli commands in an intuitive, app-centric manner, quickly speeding up tasks that might otherwise be difficult for new and old users to perform. Additionally, it provides a way to perform these tasks remotely via a json api, enabling easier, audited remote management of servers. Finally, Dokku Pro provides an alternative, https-based method for deploying code which can be used in environments that lockdown ssh access to servers.

## Purchasing

Dokku Pro may be purchased online by clicking the following button:

<a data-dpd-type="button" data-text="PURCHASE NOW" data-variant="price-right" data-button-size="dpd-large" data-bg-color="469d3d" data-bg-color-hover="5cc052" data-text-color="ffffff" data-pr-bg-color="ffffff" data-pr-color="000000" data-lightbox="1" href="https://dokku.dpdcart.com/cart/add?product_id=217344&amp;method_id=236878">PURCHASE NOW</a>

> Currently, the server must be able to contact the public internet to validate the license, or it will fail to start. For offline support, inquire for enterprise offline licensing.

## Installation

Dokku Pro is shipped as Debian and RPM packages, and depends on the following files:

- `/etc/default/dokku-pro`: Configures certain environment variables for usage by the dokku-pro binary
- `/etc/dokku-pro/license.key`: Contains the downloaded license key
- `/var/lib/dokku/data/pro/db`: Contains the local dokku-pro database

Please refer to the purchase email for details on configuring Dokku Pro.

## Features and Development

Dokku Pro has the following functionality:

- Shipped as a single binary for ease of use
- JSON-API-compatible API with JWT authentication
- Authenticated HTTP(S) endpoints for git push functionality
- Single Page App (SPA) Web UI exposing app, datastore, and ssh key management

While each release is fairly feature complete, individual features and documentation will expand over time. Feature development follows a monthly release cadence, with individual bug fixes released on an as needed basis. 
