#!/usr/bin/env python3
'''
Copyright 2019 Broadcom. The term "Broadcom" refers to Broadcom Inc.
and/or its subsidiaries.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
'''

import sys
import os
import subprocess
import errno
import time
import datetime
import signal
from ztp.JsonReader import JsonReader
from ztp.ZTPLib import getCfg, setCfg, getFeatures, getTimestamp
from ztp.ZTPCfg import ZTPCfg
import host_service
import json
import syslog 

"""ZTP command handler"""

MOD_NAME = 'ztp'

class ZTP(host_service.HostModule):
    """DBus endpoint that executes ZTP related commands
    """
    @staticmethod
    def _run_command(commands):
        """Run a ZTP command"""
        cmd = ['/usr/bin/ztp']
        if isinstance(commands, list):
            cmd.extend(commands)
        else:
            cmd.append(commands)

        try:
            rc = 0
            output = subprocess.check_output(cmd)
        except subprocess.CalledProcessError as err:
            rc = err.returncode
            output = ""

        return rc, output

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='')
    def enable(self):
        self._run_command("enable")

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='s')
    def getcfg(self):
        ztp_cfg=ZTPCfg()
        ad = getCfg('admin-mode', ztp_cfg=ztp_cfg)
        if ad == False:
            return "disabled"
        else:
            return "enabled"


    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='')
    def disable(self):
        self._run_command(["disable", "-y"])

    @host_service.method(host_service.bus_name(MOD_NAME), in_signature='', out_signature='s')
    def status(self):
        print("Calling ZTP Status")       
        b = ztp_status()
        print("Returning", b)
        return b
        
def register():
    """Return the class name"""
    return ZTP, MOD_NAME



ztp_cfg = None
## Helper API to modify status variable to a user friendly string.
def getStatusString(val):
    if val == 'BOOT':
        return 'Not Started'
    else:
        return val

## Administratively enable ZTP.
#  Changes only the configuration file.
def ztp_enable():
    if getCfg('admin-mode', ztp_cfg=ztp_cfg) is False:
        # Set ZTP configuration
        setCfg('admin-mode', True, ztp_cfg=ztp_cfg)

## Helper API to check if ztp systemd service is active or not.
def ztp_active():
    try:
        rc = subprocess.check_call(['systemctl', 'is-active', '--quiet', 'ztp'])
    except subprocess.CalledProcessError as e:
        rc = e.returncode
        pass
    return rc

## Convert seconds to hh:mm:ss format
def formatTime(seconds):
    if seconds < 60:
        format_str = '%Ss'
    elif seconds < 3600:
        format_str = '%Mm %Ss'
    else:
        format_str = '%Hh %Mm %Ss'
    return time.strftime(format_str, time.gmtime(seconds))

## Calculate time diff
def timeDiff(startTimeStamp, endTimeStamp):
    try:
        endTime = time.strptime(endTimeStamp.strip(), "%Y-%m-%d %H:%M:%S %Z")
        startTime = time.strptime(startTimeStamp.strip(), "%Y-%m-%d %H:%M:%S %Z")
        time_diff = int(time.mktime(endTime) - time.mktime(startTime))
        timeStr = formatTime(time_diff)
    except:
        timeStr = None
    return timeStr

## Calcluate and return runtime
def getRuntime(status, startTimeStamp, endTimeStamp):
    if status == 'IN-PROGRESS':
        return timeDiff(startTimeStamp, getTimestamp())
    elif status == 'SUCCESS' or status == 'FAILED':
        return timeDiff(startTimeStamp, endTimeStamp)
    else:
        return None

## Calculate time string
def getTimeString(msg):
   split_msg = msg.split('|', 1)
   if len(split_msg) == 2:
       time_str = split_msg[0]
       try:
           return "({})".format(timeDiff(time_str.strip(), getTimestamp())) + split_msg[1]
       except:
           ret = msg
   else:
       return msg

## Return duration since ZTP service has been active
def ztpServiceRuntime():
    try:
        output = subprocess.check_output(['systemctl', 'show', 'ztp', '-p', 'ActiveEnterTimestampMonotonic'])
        startTime = int(output.decode("utf-8").split('=')[1])/1000000
        currentTime = int(time.clock_gettime(time.CLOCK_MONOTONIC))
        time_diff = currentTime - startTime
    except:
        return None
    return formatTime(int(time_diff))
       

## Get ZTP Server activity
def getActivityString():
    if ztp_active() != 0:
       return 'ZTP Service is not running'

    activity_str = None
    f = getCfg('ztp-activity')
    if os.path.isfile(f):
       fh = open(f, 'r')
       activity_str = fh.readline().strip()
       fh.close()

    if activity_str is not None and activity_str != '':
       return getTimeString(activity_str)

## Display list of ZTP features available in the image
def ztp_features(verboseFlag=False):
    features = getFeatures()
    
    for feat in features:
       if  verboseFlag:
            print('%s: %s: %s' %(feat, getCfg('info-'+feat, ztp_cfg=ztp_cfg), getCfg(feat, ztp_cfg=ztp_cfg)))
       else:
            if getCfg(feat) is True:
               print(feat)


## Display current ztp status in brief format.
#  Overall ZTP status, ZTP admin mode and list of configuration sections
#  and their results are displayed.
def ztp_status():
    # Print overall ZTP status
    statusdict = {}
    configdict = {}
    ztp_cfg=ZTPCfg()
    statusdict['admin_mode'] = getCfg('admin-mode', ztp_cfg=ztp_cfg)

    if os.path.isfile(getCfg('ztp-json', ztp_cfg=ztp_cfg)):
        objJson, jsonDict = JsonReader(getCfg('ztp-json', ztp_cfg=ztp_cfg), indent=4)
        ztpDict = jsonDict.get('ztp')
        if ztp_active() != 0:
            statusdict['service'] = 'Inactive'
        else:
            statusdict['service'] = 'Processing'
        statusdict['status'] = getStatusString(ztpDict.get('status'))
        if ztpDict.get('error') is not None:
            statusdict['error'] =  ztpDict.get('error')
        else:
            statusdict['error'] = 'NONE'
        statusdict['source'] = ztpDict.get('ztp-json-source')
        runtime = getRuntime(ztpDict.get('status'), ztpDict.get('start-timestamp'), ztpDict.get('timestamp'))
        if runtime is not None:
            statusdict['runtime'] = runtime
        statusdict['timestamp'] = ztpDict.get('timestamp')
        if ztpDict.get('ignore-result'):
            statusdict['ignore result'] = ztpDict.get('ignore-result')
        statusdict['json_version'] = ztpDict.get('ztp-json-version')

        statusdict['activity_string'] = getActivityString()
    
    # Print individual section ZTP status
        keys = sorted(ztpDict.keys())
        for k in keys:
            v = ztpDict.get(k)
            if isinstance(v, dict):
                configdict[k] = {}
                configdict[k]['cfg_sectionname'] = k
                configdict[k]['cfg_status'] = getStatusString(v.get('status'))
                runtime = getRuntime(v.get('status'), v.get('start-timestamp'), v.get('timestamp'))
                if runtime is not None:
                    configdict[k]['cfg_runtime'] = runtime
                configdict[k]['cfg_timestamp'] = v.get('timestamp')
                if v.get('exit-code') is not None:
                    configdict[k]['cfg_exitcode'] = v.get('exit-code')
                if v.get('error') is not None:
                    configdict[k]['error'] = v.get('error')
                else:
                    configdict[k]['error'] = 'NONE'
                configdict[k]['cfg_ignoreresult'] = v.get('ignore-result')
                if v.get('halt-on-failure') is not None and v.get('halt-on-failure'):
                    print('Halt on Failure : %r' % v.get('halt-on-failure'))
                print (' ')

    else:
        if ztp_active() == 0:
            statusdict['service'] = 'Active Discovery'
            runtime = ztpServiceRuntime()
            if runtime:
                statusdict['runtime'] = runtime
        else:
            statusdict['service'] = 'Inactive'
        statusdict['status'] = getStatusString('BOOT')
        statusdict['activty_string'] = getActivityString()

    statusdict['config_section_list'] = configdict
    retVal = json.dumps(statusdict)
    return retVal

