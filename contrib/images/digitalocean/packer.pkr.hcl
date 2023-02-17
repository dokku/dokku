packer {
  required_plugins {
    digitalocean = {
      version = ">= 1.0.4"
      source  = "github.com/digitalocean/digitalocean"
    }
  }
}

variable "dokku_version" {
  type    = string
}

source "digitalocean" "ubuntu" {
  image         = "ubuntu-22-04-x64"
  region        = "nyc1"
  size          = "s-1vcpu-512mb-10gb"
  ssh_username  = "root"
  snapshot_name = "dokku-${var.dokku_version}-snapshot-{{timestamp}}"
}

build {
  name = "dokku"
  sources = [
    "source.digitalocean.ubuntu"
  ]

  provisioner "shell" {
    inline = [
      "echo '--> Waiting until cloud-init is complete'",
      "/usr/bin/cloud-init status --wait",
    ]
  }

  provisioner "file" {
    source      = "${path.root}/files/etc/"
    destination = "/etc/"
  }

  provisioner "file" {
    source      = "${path.root}/files/var/"
    destination = "/var/"
  }

  provisioner "shell" {
    environment_vars = [
      "DOKKU_VERSION=${var.dokku_version}",
      "DEBIAN_FRONTEND=noninteractive",
      "LC_ALL=C",
      "LANG=en_US.UTF-8",
      "LC_CTYPE=en_US.UTF-8",
    ]

    scripts = [
      "${path.root}/in_parts/011-docker",
      "${path.root}/in_parts/011-ssh-message",
      "${path.root}/in_parts/012-dokku-packages",
      "${path.root}/in_parts/012-grub-opts",
      "${path.root}/in_parts/014-docker-dns",
      "${path.root}/in_parts/014-ufw-rules",
      "${path.root}/in_parts/099-application_tag",
      "${path.root}/in_parts/099-cleanup",
      "${path.root}/in_parts/100-image-check",
    ]
  }

  post-processor "manifest" {
    output = "digitalocean-manifest.json"
  }
}
