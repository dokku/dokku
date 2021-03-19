# Run Dokku on External Volume
----

In order to leverage cloud-provider facilities like _attachable volumes_, (_a.k.a. block storage_)
the following is an easy tutorial to achieve Dokku runs on them.

!!! warning
    If the block storage is not available and attached on boot it is possible that containers will not correctly start. Please keep this in mind when considering moving Dokku and/or Docker to network attached storage.

## Tutorial

The following is intended to be executed on the dokku host machine as `root`.
Say, _for instance_, that our volume is mapped into the systems as `/dev/vdb1`.

Stop docker daemon

```shell
systemctl stop docker
```

Prepare the filesystem:

```shell
mkfs -t ext4 /dev/vdb1
mkdir /mnt/volume
mount /dev/vdb1 /mnt/volume
```

Move the old data directories:

```shell
mv /home/dokku /home/dokku.OLD
mv /var/lib/docker /var/lib/docker.OLD
mv /var/lib/dokku /var/lib/dokku.OLD
```

Move the data on the volume

```shell
mkdir /mnt/volume/home/
mkdir -p /mnt/volume/var/lib/
mv /home/dokku.OLD /mnt/volume/home/dokku
mv /var/lib/dokku.OLD /mnt/volume/var/lib/dokku
mv /var/lib/docker.OLD /mnt/volume/var/lib/docker
```

Prepare the mountpoints

```shell
mkdir /home/dokku
mkdir /var/lib/dokku
mkdir /var/lib/docker
chown dokku:dokku /home/dokku   # respect the original ownership
chmod 711 /var/lib/docker       # respect the original permissions
```

Mount bind

```shell
mount -o bind /mnt/volume/home/dokku /home/dokku
mount -o bind /mnt/volume/var/lib/dokku /var/lib/dokku
mount -o bind /mnt/volume/var/lib/docker /var/lib/docker
```

Start docker daemon

```shell
systemctl start docker
```

At this point all should be working fine, please check it out.

Then, let the changes be reboot-persistent

```shell
echo '/dev/vdb1 /mnt/volume ext4 defaults 0 2' | sudo tee -a /etc/fstab
echo '/mnt/volume/home/dokku /home/dokku none defaults,bind 0 0' | sudo tee -a /etc/fstab
echo '/mnt/volume/var/lib/dokku /var/lib/dokku none defaults,bind 0 0' | sudo tee -a /etc/fstab
echo '/mnt/volume/var/lib/docker /var/lib/docker none defaults,bind 0 0' | sudo tee -a /etc/fstab
```
