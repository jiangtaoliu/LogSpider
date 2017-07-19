#!/bin/bash


#add user
ssh-copy-id -i ../id_rsa logspider@$1
# enable multiplexing in ssh echo MaxSessions 8192 > /etc/ssh/sshd_config; service ssh restart
# chmod g+s <directory>  //set gid
#setfacl -d -m o::r /<directory>
