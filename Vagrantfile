# -*- mode: ruby -*-
# vi: set ft=ruby :

BOX_NAME = ENV["BOX_NAME"] || "bento/ubuntu-18.04"
BOX_CPUS = ENV["BOX_CPUS"] || "1"
BOX_MEMORY = ENV["BOX_MEMORY"] || "1024"
DOKKU_DOMAIN = ENV["DOKKU_DOMAIN"] || "dokku.me"
DOKKU_IP = ENV["DOKKU_IP"] || "10.0.0.2"
FORWARDED_PORT = (ENV["FORWARDED_PORT"] || '8080').to_i
PREBUILT_STACK_URL = File.exist?("#{File.dirname(__FILE__)}/stack.tgz") ? 'file:///root/dokku/stack.tgz' : nil
PUBLIC_KEY_PATH = "#{Dir.home}/.ssh/id_rsa.pub"

make_cmd = "DEBIAN_FRONTEND=noninteractive make -e install"
if PREBUILT_STACK_URL
  make_cmd = "PREBUILT_STACK_URL='#{PREBUILT_STACK_URL}' #{make_cmd}"
end

Vagrant::configure("2") do |config|
  config.ssh.forward_agent = true

  config.vm.box = BOX_NAME

  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    # Ubuntu's Raring 64-bit cloud image is set to a 32-bit Ubuntu OS type by
    # default in Virtualbox and thus will not boot. Manually override that.
    vb.customize ["modifyvm", :id, "--ostype", "Ubuntu_64"]
    vb.customize ["modifyvm", :id, "--cpus", BOX_CPUS]
    vb.customize ["modifyvm", :id, "--memory", BOX_MEMORY]
  end

  config.vm.provider :vmware_fusion do |v, override|
    v.vmx["memsize"] = BOX_MEMORY
    v.ssh_info_public = true
  end

  config.vm.provider :vmware_desktop do |v, override|
    v.vmx["memsize"] = BOX_MEMORY
    v.ssh_info_public = true
  end

  config.vm.define "empty", autostart: false

  config.vm.define "dokku", primary: true do |vm|
    vm.vm.synced_folder File.dirname(__FILE__), "/root/dokku"
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP

    # Use the same nameserver as the host machine in order to avoid the "too many redirects" problem.
    vm.vm.provider :virtualbox do |vb|
      vb.customize ["modifyvm", :id, "--natdnshostresolver1", "off"]
      vb.customize ["modifyvm", :id, "--natdnsproxy1", "off"]
      # enable NAT adapter cable https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=838999
      vb.customize ["modifyvm", :id, "--cableconnected1", "on"]
    end

    vm.vm.provision :shell, :inline => "export DEBIAN_FRONTEND=noninteractive && apt-get update -qq >/dev/null && apt-get -qq -y --no-install-recommends install git build-essential >/dev/null && cd /root/dokku && #{make_cmd}"
    vm.vm.provision :shell do |s|
      s.inline = <<-EOT
        echo '"\e[5~": history-search-backward' > /root/.inputrc
        echo '"\e[6~": history-search-forward' >> /root/.inputrc
        echo 'set show-all-if-ambiguous on' >> /root/.inputrc
        echo 'set completion-ignore-case on' >> /root/.inputrc
      EOT
    end

    if Pathname.new(PUBLIC_KEY_PATH).exist?
      vm.vm.provision :shell, :inline => "echo 'Importing ssh key into dokku' && cat /root/.ssh/authorized_keys | dokku ssh-keys:add admin"
    end
  end

  # For windows users. Sharing folder from windows creates problem with sym links and so, sync the repo instead from GOS.
  config.vm.define "dokku-windows", autostart: false do |vm|
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP
    vm.vm.provision :shell, :inline => "export DEBIAN_FRONTEND=noninteractive && apt-get update -qq >/dev/null && apt-get -qq -y --no-install-recommends install git dos2unix >/dev/null"
    vm.vm.provision :shell, :inline => "cd /vagrant/ && export DOKKU_BRANCH=`git symbolic-ref -q --short HEAD 2>/dev/null` && export DOKKU_TAG=`git describe --tags --exact-match 2>/dev/null` && cd /root/ && cp /vagrant/bootstrap.sh ./ && dos2unix bootstrap.sh && bash bootstrap.sh"
  end

  config.vm.define "dokku-deb", autostart: false do |vm|
    vm.vm.synced_folder File.dirname(__FILE__), "/root/dokku"
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP
    vm.vm.provision :shell, :inline => "cd /root/dokku && make install-from-deb"
  end

  config.vm.define "dokku-rpm", autostart: false do |vm|
    vm.vm.box = "centos/7"
    vm.vm.synced_folder File.dirname(__FILE__), "/root/dokku"
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP
    vm.vm.provision :shell, :inline => "cd /root/dokku && bash bootstrap.sh"
  end

  config.vm.define "build", autostart: false do |vm|
    vm.vm.synced_folder File.dirname(__FILE__), "/root/dokku"
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP
    vm.vm.provision :shell, :inline => "export DEBIAN_FRONTEND=noninteractive && apt-get update -qq >/dev/null && apt-get -qq -y --no-install-recommends install git >/dev/null && cd /root/dokku && #{make_cmd}"
    vm.vm.provision :shell, :inline => "export IS_RELEASE=true && cd /root/dokku && make deb-all rpm-all"
  end

  config.vm.define "build-arch", autostart: false do |vm|
    vm.vm.box = "bugyt/archlinux"
    vm.vm.synced_folder File.dirname(__FILE__), "/dokku"
    if Pathname.new("#{File.dirname(__FILE__)}/../dokku-arch").exist?
      vm.vm.synced_folder "#{File.dirname(__FILE__)}/../dokku-arch", "/dokku-arch"
    end
    vm.vm.network :forwarded_port, guest: 80, host: FORWARDED_PORT
    vm.vm.hostname = "#{DOKKU_DOMAIN}"
    vm.vm.network :private_network, ip: DOKKU_IP
    vm.vm.provision :shell, :inline => "cd /dokku && make arch-all", privileged: false
  end

  if Pathname.new(PUBLIC_KEY_PATH).exist?
    config.vm.provision :file, source: PUBLIC_KEY_PATH, destination: '/tmp/id_rsa.pub'
    config.vm.provision :shell, :inline => "echo 'Copying ssh key into vm' && rm -f /root/.ssh/authorized_keys && mkdir -p /root/.ssh && sudo cp /tmp/id_rsa.pub /root/.ssh/authorized_keys"
  end
end
