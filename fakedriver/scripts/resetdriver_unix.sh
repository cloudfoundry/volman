#!/bin/bash

pkill -f fakedriver

rm /tmp/plugins/fakedriver.*

mkdir -p /tmp/moundir/mount1
mkdir -p /tmp/plugins

$DRIVER_PATH -listenAddr="/tmp/plugins/fakedriver.sock" -transport="unix" -mountDir="/tmp/mountdir/mount1" -driversPath="/tmp/plugins" &