packer {
  required_plugins {
    digitalocean = {
      version = ">= 1.0.4"
      source  = "github.com/digitalocean/digitalocean"
    }
  }
}

variable "source_image" {
  type    = string
  default = "ubuntu-22-04-x64"
}

variable "dokku_version" {
  type    = string
}

source "digitalocean" "ubuntu" {
  image         = var.source_image
  region        = "nyc3"
  size          = "s-2vcpu-2gb"
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
