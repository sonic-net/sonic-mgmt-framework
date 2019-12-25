#!/usr/bin/env python3
import os
import pwd

container = 'mgmt-framework'
user = pwd.getpwuid(os.getuid())[0]
os.execvp('/usr/bin/docker', ['', 'exec', '--user', user, '-it', container, '/bin/klish'])


