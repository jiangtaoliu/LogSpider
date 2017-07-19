#!/bin/bash


ssh root@$1 adduser --disabled-password --gecos "" logspider
#ssh-copy-id -i ../id_rsa logspider@$1
cat ../id_rsa.pub | ssh root@$1 "sudo -H -u logspider bash -c 'mkdir -p ~/.ssh; cat >>  ~/.ssh/authorized_keys'"

ssh root@$1 "echo MaxSessions 8192 >> /etc/ssh/sshd_config; service ssh restart; chmod g+s /var/log; setfacl -d -m o::r /var/log"
