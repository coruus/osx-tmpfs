#!/bin/sh
sudo rm osx-tmpfs&&
go build&&
sudo chown root:wheel osx-tmpfs&&
sudo chmod go-r,u+s osx-tmpfs
