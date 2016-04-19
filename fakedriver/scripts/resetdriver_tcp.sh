#!/bin/bash

pkill -f fakedriver

rm /tmp/plugins/fakedriver.*

$DRIVER_PATH -listenAddr="0.0.0.0:9776" -transport="tcp" -mountDir="/tmp/mountdir/mount1" -driversPath="/tmp/plugins" &
