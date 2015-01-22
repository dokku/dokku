# -*- mode: ruby -*-
# vi: set ft=ruby :

BOX_NAME = ENV["BOX_NAME"] || "ubuntu/trusty64"
# BOX_URI = ENV["BOX_URI"] || "https://cloud-images.ubuntu.com/vagrant/trusty/current/trusty-server-cloudimg-amd64-vagrant-disk1.box"
BOX_MEMORY = ENV["BOX_MEMORY"] || "1024"
DOKKU_DOMAIN = ENV["DOKKU_DOMAIN"] || "kunst-dokku"
DOKKU_IP = ENV["DOKKU_IP"] || "10.0.0.2"
PREBUILT_STACK_URL = File.exist?("#{File.dirname(__FILE__)}/stack.tgz") ? 'file:///root/dokku/stack.tgz' : nil

make_cmd = "make install"
if PREBUILT_STACK_URL
  make_cmd = "PREBUILT_STACK_URL='#{PREBUILT_STACK_URL}' #{make_cmd}"
end

Vagrant::configure("2") do |config|
  config.vm.box = BOX_NAME
  config.vm.box_check_update = false;
  if Vagrant.has_plugin?("vagrant-cachier")
    # More info on http://fgrehm.viewdocs.io/vagrant-cachier/usage
    config.cache.scope = :box
  end
  # config.vm.box_url = BOX_URI
  config.ssh.username = 'vagrant';
  config.ssh.password = 'vagrant';
  config.ssh.insert_key = false;
  config.ssh.private_key_path = "#{File.dirname(__FILE__)}/keys/mbp"
  # config.ssh.forward_agent = true
  
  config.vm.synced_folder File.dirname(__FILE__), "/root/dokku"
  
  config.vm.hostname = "#{DOKKU_DOMAIN}"
  config.vm.network "public_network", :mac => "309E399DE47E", :bridge=> 'en0: Ethernet';
  config.vm.network :private_network, ip: DOKKU_IP;

  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    # Ubuntu's Raring 64-bit cloud image is set to a 32-bit Ubuntu OS type by
    # default in Virtualbox and thus will not boot. Manually override that.
    vb.customize ["modifyvm", :id, "--ostype", "Ubuntu_64"]
    vb.customize ["modifyvm", :id, "--memory", BOX_MEMORY]
  end

  config.vm.provision :shell, :inline => "echo 'STACK: '#{PREBUILT_STACK_URL}; apt-get -qq -y install git > /dev/null && cd /root/dokku && #{make_cmd}"
  config.vm.provision :shell, :inline => "cd /root/dokku && make dokku-installer"
end
