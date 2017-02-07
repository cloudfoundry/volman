# DRAFT DRAFT DRAFT DRAFT DRAFT

# Troubleshooting Problems with Cloud Foundry Volume Services

## When the application does not start

- If you have pushed an app with `--no-start` and then bound the app to a service, the cf cli will tell you to `cf restage` to start the app with the new binding.  This is incorrect.  You must use `cf start` on an app that is not running.  It the app is already running, then `cf restage` is OK.
- If your application still won't start, try unbinding it from the volume service and see if it starts when it is *not* bound.  (Most of our test applications like [pora](https://github.com/cloudfoundry-incubator/persi-acceptance-tests/tree/master/assets/pora) and [kitty](https://github.com/EMC-Dojo/Kitty) will start up even when no volume is available.)
  If the application starts up, then that indicates the problem is volume services related.  I you still see the same error regardless, then that indicates that the problem is elsewhere, and you should go through the [Troubleshooting Application Deployment and Health](https://docs.cloudfoundry.org/devguide/deploy-apps/troubleshoot-app-health.html) steps.
- 

## When the application starts, but data is missing

## When BOSH deployment fails

### Broker deployment (for bosh deployed brokers)

When broker deployment fails, assuming that Bosh has successfully parsed the manifest and created a vm for your broker, you will normally find any errors that occurred during startup by looking in the bosh logs.
Although you can gather the logs from your bosh vm using the `bosh logs` command, that command creates a big zip file with all the logs in it that muust be unpacked, so it is usually easier and faster to `bosh ssh` onto the vm and look at the logs in a shell.
Instructions for bosh ssh are [here](https://bosh.io/docs/sysadmin-commands.html#ssh).

Once you are ssh'd into the vm, switch to root with `sudo su` and then type `monit summary` to make sure that your broker job is really not running.
Assuming that the broker is not showing as running, you should see some type of error in one of three places:
- `/var/vcap/sys/log/monit/` contains monit script output for the various bosh logs.  Errors that occur in outer monit scripts will appear here.
- `/var/vcap/sys/log/packages/<broker name>` contains package installation logs for your broker source.  Some packaging errors end up here  
- `/var/vcap/sys/log/jobs/<broker name>` contains logs for your actual broker process.  Any errors from the running executable or pre-start script will appear in this directory.

### Driver deployment

Diagnosing failures in driver deployment is quite similar to bosh deployed broker diagnosis as described above.  The principal difference is that the driver is deployed alongside diego, so you must use the diego deployment manifest when calling `bosh ssh` and you must ssh into the diego cell vm to gather logs.  
In a multi-cell deployment, sometimes it is necessary to try different cell vms to find the failed one, but most of the time if configuration is not right, all cells will fail in the same way.

## When the service broker cannot be registered with `cf create-service-broker`

* Check to make sure that the service broker is reachable at the URL you are passing to the `create-service-broker` call:
```curl http://user:password@yourbroker.your.app.domain.com/v2/catalog```
* Check to make sure that your cloudfoundry manifest has `properties.cc.volume_services_enabled` set to `true`.  If not, change your manifest and redeploy cloudfoundry.

