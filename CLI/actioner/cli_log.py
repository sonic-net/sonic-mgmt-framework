################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

import os
import syslog
import inspect

syslog.openlog('sonic-cli')
__enable_debug = (os.getenv("DEBUG") is not None)
__enable_print = (os.getenv("LOGTOSCREEN") is not None)
__severity_str = [ 'EMERG', 'ALERT', 'CRIT', 'ERROR', 'WARN', 'NOTICE', 'INFO', 'DEBUG' ]


def log_debug(msg):
    if __enable_debug:
        __write_log(syslog.LOG_DEBUG, msg)

def log_info(msg):
    __write_log(syslog.LOG_INFO, msg)

def log_warning(msg):
    __write_log(syslog.LOG_WARNING, msg)

def log_error(msg):
    __write_log(syslog.LOG_ERR, msg)


def __write_log(severity, msg):
    if not isinstance(msg, str):
        msg = str(msg)

    syslog.syslog(severity, msg)

    if __enable_print:
        caller_frame = inspect.stack()[2][0]
        frame_info = inspect.getframeinfo(caller_frame)
        for line in msg.split('\n'):
            print(('[{}:{}/{}] {}:: {}'.format(
                os.path.basename(frame_info.filename),
                frame_info.lineno,
                frame_info.function,
                __severity_str[severity],
                line)))
