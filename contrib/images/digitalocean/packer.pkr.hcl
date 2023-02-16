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
  snapshot_name = format("dokku-%s-%s", var.dokku_version, formatdate("YYYYMMDDhhmm", timestamp()))
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
    source      = "${path.root}/files/etc/update-motd.d/99-one-click"
    destination = "/etc/update-motd.d/99-one-click"
  }

  provisioner "file" {
    source      = "${path.root}/files/etc/nginx/sites-available/digitalocean"
    destination = "/etc/nginx/sites-available/digitalocean"
  }

  provisioner "file" {
    source      = "${path.root}/files/var/www/html/index.html"
    destination = "/var/www/html/index.html"
  }

  provisioner "file" {
    source      = "${path.root}/files/var/lib/cloud/scripts/per-once/001_setup"
    destination = "/var/lib/cloud/scripts/per-once/001_setup"
  }

  provisioner "file" {
    source      = "${path.root}/files/var/lib/cloud/scripts/per-once/002_enable_ssh"
    destination = "/var/lib/cloud/scripts/per-once/002_enable_ssh"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/011-docker"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/011-ssh-message"
  }

  provisioner "shell" {
    environment_vars = [
      "DOKKU_VERSION=${var.dokku_version}"
    ]
    script = "${path.root}/in_parts/012-dokku-packages"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/012-grub-opts"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/014-docker-dns"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/014-ufw-rules"
  }

  provisioner "shell" {
    environment_vars = [
      "DOKKU_VERSION=${var.dokku_version}"
    ]
    script = "${path.root}/in_parts/099-application_tag"
  }

  provisioner "shell" {
    script = "${path.root}/in_parts/099-cleanup"
  }

  post-processor "manifest" {
    output = "digitalocean-manifest.json"
  }
}
