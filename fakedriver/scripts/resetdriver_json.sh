#!/bin/bash

pkill -f startdriver
pkill -f fakedriver

rm /tmp/plugins/fakedriver.*

mkdir -p /tmp/moundir/mount1
mkdir -p /tmp/plugins

$DRIVER_PATH -listenAddr="0.0.0.0:9876" -transport="tcp-json" -mountDir="/tmp/mountdir/mount1" -driversPath="/tmp/plugins" &
