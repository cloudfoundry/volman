#!/bin/bash

mkdir -p /tmp/moundir/mount1
mkdir -p /tmp/plugins

$DRIVER_PATH -listenAddr="http://0.0.0.0:9876" -transport="tcp-json" -mountDir="/tmp/mountdir/mount1" -driversPath="/tmp/plugins" &