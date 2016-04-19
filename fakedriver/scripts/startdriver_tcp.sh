#!/bin/bash

mkdir -p /tmp/moundir/mount1
mkdir -p /tmp/plugins

$DRIVER_PATH -listenAddr="http://0.0.0.0:9776" -transport="tcp" -mountDir="/tmp/mountdir/mount1" -driversPath="/tmp/plugins" &