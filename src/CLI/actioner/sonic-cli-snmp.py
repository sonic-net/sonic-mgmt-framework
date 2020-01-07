#!/usr/bin/python
import os
import subprocess
import sys
import json
import collections
import re
import swsssdk
import cli_client as cc
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector

"""
module: ietf-snmp
  +--rw snmp
     +--rw engine
     |  +--rw enabled?               boolean
     |  +--rw listen* [name]
     |  |  +--rw name          snmp:identifier
     |  |  +--rw (transport)
     |  |     +--:(udp)
     |  |     |  +--rw udp
     |  |     |     +--rw ip      inet:ip-address
     |  |     |     +--rw port?   inet:port-number
     |  |     +--:(tls) {tlstm}?
     |  |     |  +--rw tls
     |  |     |     +--rw ip      inet:ip-address
     |  |     |     +--rw port?   inet:port-number
     |  |     +--:(dtls) {tlstm}?
     |  |     |  +--rw dtls
     |  |     |     +--rw ip      inet:ip-address
     |  |     |     +--rw port?   inet:port-number
     |  |     +--:(ssh) {sshtm}?
     |  |        +--rw ssh
     |  |           +--rw ip      inet:ip-address
     |  |           +--rw port?   inet:port-number
     |  +--rw version
     |  |  +--rw v1?    empty
     |  |  +--rw v2c?   empty
     |  |  +--rw v3?    empty
     |  +--rw engine-id?             snmp:engine-id
     |  +--rw enable-authen-traps?   boolean
     +--rw target* [name]
     |  +--rw name             snmp:identifier
     |  +--rw (transport)
     |  |  +--:(udp)
     |  |  |  +--rw udp
     |  |  |     +--rw ip               inet:ip-address
     |  |  |     +--rw port?            inet:port-number
     |  |  |     +--rw prefix-length?   uint8
     |  |  +--:(tls) {tlstm}?
     |  |  |  +--rw tls
     |  |  |     +--rw ip                    inet:host
     |  |  |     +--rw port?                 inet:port-number
     |  |  |     +--rw client-fingerprint?   x509c2n:tls-fingerprint
     |  |  |     +--rw server-fingerprint?   x509c2n:tls-fingerprint
     |  |  |     +--rw server-identity?      snmp:admin-string
     |  |  +--:(dtls) {tlstm}?
     |  |  |  +--rw dtls
     |  |  |     +--rw ip                    inet:host
     |  |  |     +--rw port?                 inet:port-number
     |  |  |     +--rw client-fingerprint?   x509c2n:tls-fingerprint
     |  |  |     +--rw server-fingerprint?   x509c2n:tls-fingerprint
     |  |  |     +--rw server-identity?      snmp:admin-string
     |  |  +--:(ssh) {sshtm}?
     |  |     +--rw ssh
     |  |        +--rw ip          inet:host
     |  |        +--rw port?       inet:port-number
     |  |        +--rw username?   string
     |  +--rw tag*             snmp:tag-value
     |  +--rw timeout?         uint32
     |  +--rw retries?         uint8
     |  +--rw target-params    snmp:identifier
     |  +--rw mms?             union
     +--rw target-params* [name]
     |  +--rw name                     snmp:identifier
     |  +--rw (params)?
     |  |  +--:(tsm) {tsm}?
     |  |  |  +--rw tsm
     |  |  |     +--rw security-name     snmp:security-name
     |  |  |     +--rw security-level    snmp:security-level
     |  |  +--:(usm)
     |  |  |  +--rw usm
     |  |  |     +--rw user-name         snmp:security-name
     |  |  |     +--rw security-level    snmp:security-level
     |  |  +--:(v1)
     |  |  |  +--rw v1
     |  |  |     +--rw security-name    snmp:security-name
     |  |  +--:(v2c)
     |  |     +--rw v2c
     |  |        +--rw security-name    snmp:security-name
     |  +--rw notify-filter-profile?   -> /snmp/notify-filter-profile/name {snmp:notification-filter}?
     +--rw tlstm {tlstm}?
     |  +--rw cert-to-name* [id]
     |     +--rw id             uint32
     |     +--rw fingerprint    x509c2n:tls-fingerprint
     |     +--rw map-type       identityref
     |     +--rw name           string
     +--rw vacm
     |  +--rw group* [name]
     |  |  +--rw name      snmp:group-name
     |  |  +--rw member* [security-name]
     |  |  |  +--rw security-name     snmp:security-name
     |  |  |  +--rw security-model*   snmp:security-model
     |  |  +--rw access* [context security-model security-level]
     |  |     +--rw context           snmp:context-name
     |  |     +--rw context-match?    enumeration
     |  |     +--rw security-model    snmp:security-model-or-any
     |  |     +--rw security-level    snmp:security-level
     |  |     +--rw read-view?        snmp:view-name
     |  |     +--rw write-view?       view-name
     |  |     +--rw notify-view?      view-name
     |  +--rw view* [name]
     |     +--rw name       view-name
     |     +--rw include*   snmp:wildcard-object-identifier
     |     +--rw exclude*   snmp:wildcard-object-identifier
     +--rw proxy* [name] {snmp:proxy}?
     |  +--rw name                   snmp:identifier
     |  +--rw type                   enumeration
     |  +--rw context-engine-id      snmp:engine-id
     |  +--rw context-name?          snmp:context-name
     |  +--rw target-params-in?      snmp:identifier
     |  +--rw single-target-out?     snmp:identifier
     |  +--rw multiple-target-out?   snmp:tag-value
     +--rw tsm {tsm}?
     |  +--rw use-prefix?   boolean
     +--rw usm
     |  +--rw local
     |  |  +--rw user* [name]
     |  |     +--rw name     snmp:identifier
     |  |     +--rw group?   snmp:group-name
     |  |     +--rw auth!
     |  |     |  +--rw (protocol)
     |  |     |     +--:(md5)
     |  |     |     |  +--rw md5
     |  |     |     |     +--rw key    yang:hex-string
     |  |     |     +--:(sha)
     |  |     |        +--rw sha
     |  |     |           +--rw key    yang:hex-string
     |  |     +--rw priv!
     |  |        +--rw (protocol)
     |  |           +--:(des)
     |  |           |  +--rw des
     |  |           |     +--rw key    yang:hex-string
     |  |           +--:(aes)
     |  |              +--rw aes
     |  |                 +--rw key    yang:hex-string
     |  +--rw remote* [engine-id]
     |     +--rw engine-id    snmp:engine-id
     |     +--rw user* [name]
     |        +--rw name     snmp:identifier
     |        +--rw group?   snmp:group-name
     |        +--rw auth!
     |        |  +--rw (protocol)
     |        |     +--:(md5)
     |        |     |  +--rw md5
     |        |     |     +--rw key    yang:hex-string
     |        |     +--:(sha)
     |        |        +--rw sha
     |        |           +--rw key    yang:hex-string
     |        +--rw priv!
     |           +--rw (protocol)
     |              +--:(des)
     |              |  +--rw des
     |              |     +--rw key    yang:hex-string
     |              +--:(aes)
     |                 +--rw aes
     |                    +--rw key    yang:hex-string
     +--rw notify* [name]
     |  +--rw name    snmp:identifier
     |  +--rw tag     snmp:tag-value
     |  +--rw type?   enumeration
     +--rw notify-filter-profile* [name] {snmp:notification-filter}?
     |  +--rw name       snmp:identifier
     |  +--rw include*   snmp:wildcard-object-identifier
     |  +--rw exclude*   snmp:wildcard-object-identifier
     +--rw community* [index]
        +--rw index                snmp:identifier
        +--rw (name)?
        |  +--:(text-name)
        |  |  +--rw text-name?     string
        |  +--:(binary-name)
        |     +--rw binary-name?   binary
        +--rw security-name        snmp:security-name
        +--rw engine-id?           snmp:engine-id {snmp:proxy}?
        +--rw context?             snmp:context-name
        +--rw target-tag?          snmp:tag-value
"""
DEVICE_METADATA = 'DEVICE_METADATA'
SYSTEM          = 'SYSTEM'
SNMP_SERVER     = 'SNMP_SERVER'
SNMP_GROUP      = 'SNMP_SERVER_GROUP'
sysname         = 'sysName'
contact         = 'sysContact'
location        = 'sysLocation'
traps           = 'traps'
context         = 'Default'
SecurityModels = { 'any' : 'any', 'v1': 'v1', 'v2c': 'v2c', 'v3': 'usm' }
SecurityLevels = { 'noauth' : 'no-auth-no-priv', 'auth' : 'auth-no-priv', 'priv' : 'auth-priv' }
ViewOpts       = { 'read' : 'readView', 'write' : 'writeView', 'notify' : 'notifyView'}

config_db = ConfigDBConnector()
if config_db is None:
  sys.exit()
config_db.connect()

aa = cc.ApiClient()

def LevelsSecurity(snmpSecLevel):
  """ Reverse lookup to convert SNMP security Level to CLI security option. """
  for key, value in SecurityLevels.items():
    if value == snmpSecLevel:
      return key
  return none

def createYangHexStr(textString):
  """ Convert plain hex string into yang:hex-string """
  data = textString[0:2]
  i = 2
  while i < len(textString):
    data = data + ':' + textString[i:i+2]
    i = i + 2
  return data

redis_kwargs = {'unix_socket_path': '/var/run/redis/redis.sock'}
from swsssdk import SonicV2Connector
def getEngineID():
  """ Construct SNMP engineID from the configured value or from scratch """
  # keypath = cc.Path('/restconf/data/ietf-snmp:snmp/engine/')
  # response=aa.get(keypath)
  # data = []
  # if response.ok():
  #   if 'ietf-snmp:engine' in response.content.keys():
  #     data = response.content['ietf-snmp:engine']
  #     print data
  # else:
  #   print ("cannot get engineID")
  #
  # config_db = ConfigDBConnector()
  # config_db.connect()
  # curr_host_name = config_db.get_entry('DEVICE_METADATA', "localhost").get('hostname')
  # print curr_host_name
  # mac = config_db.get_entry('DEVICE_METADATA', "localhost").get('mac')
  # print mac
  sysmac = '00:00:00:00:12:34'
  datam = config_db.get_entry(DEVICE_METADATA, 'localhost')
  if len(datam):
    sysmac = datam['mac']
  sysmac = sysmac.translate(None, ':')
  engineID = "8000013703"+sysmac
  return engineID

def set_system(row, data):
  key = SYSTEM
  entry = config_db.get_entry(SNMP_SERVER, key)
  if entry:
    if entry.has_key(row):
      del entry[row]
  config_db.delete_entry(SNMP_SERVER, key)
  newentry = {}
  if (len(data)>0):
    newentry[row] = data
  for row, data in entry.iteritems():
    newentry[row] = data
  if len(newentry):
    config_db.mod_entry(SNMP_SERVER, key, newentry)
  return None

def invoke(func, args):

  if func == 'snmp_get':
    keys = config_db.get_keys(SNMP_SERVER)
    datam = {}
    for key in keys:
      datam = config_db.get_entry(SNMP_SERVER, key)
    datam['engineID'] = getEngineID()
    if len(datam) > 0:
      return datam
    else:
      return None

  elif func == 'snmp_sysname':
    row = sysname
    data = ''
    if (len(args)>0):
      data = args[0]
    set_system(row, data)
    return None

  elif func == 'snmp_location':
    row = location
    data = ''
    if (len(args)>0):
      data = args[0]
    set_system(row, data)
    return None

  elif func == 'snmp_contact':
    row = contact
    data = ''
    if (len(args)>0):
      data = args[0]
    set_system(row, data)
    return None

  elif func == 'snmp_trap':
    row = traps
    data = ''
    if (len(args)>0) and (args[0] == 'enable'):
      data = args[0]
    set_system(row, data)
    return None

  elif func == 'snmp_engine':
    data = ''
    if (len(args) == 1):
      data = args[0]
    index = 2
    engineId = data[:2]
    while index < len(data):
      engineId += ':' + data[index:index+2]
      index += 2
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/engine')
    entry=collections.defaultdict(dict)
    entry["engine"]=[{ "engine-id" : engineId }]
    print keypath
    print entry
    response = aa.patch(keypath, entry)
    if response.ok():
      return None
    else:
      return response

  # Get the configured communities.
  elif func == 'snmp_community_get':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community')
    response=aa.get(keypath)
    data = []
    if response.ok():
      if 'ietf-snmp:community' in response.content.keys():
        data = response.content['ietf-snmp:community']
    if len(data) == 0:
      return None
    else:
      return data

  # Configure a new community.
  elif func == 'snmp_community_add':
    group="None"
    if (1<len(args)):
      group=args[1]
    entry=collections.defaultdict(dict)
    entry["community"]=[{ "index": args[0],
                          "security-name" : group }]
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community')
    response = aa.patch(keypath, entry)
    if response.ok():
      return None
    else:
      return response

  # Remove a community.
  elif func == 'snmp_community_delete':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community={index}', index=args[0])
    response = aa.delete(keypath)
    if response.ok():
      return None
    else:
      return response

  # Get the configured groups.
  elif func == 'snmp_group_get':
    # keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/group')
    # response = aa.get(keypath)
    # print keypath, response.content
    # return response

    # temporary code in place until key-transformer is written
    keys = config_db.get_keys(SNMP_GROUP)
    groups = []
    for key in keys:
      g = {}
      g['name'] = str(key[0])
      g['context'] = str(key[1])
      if key[2] == "usm":
        g['model'] = 'v3'
      else:
        g['model'] = str(key[2])
      g['security'] = str(key[3])
      datam = config_db.get_entry(SNMP_GROUP, key)
      for row, data in datam.items():
        g[row] = data
      groups.append(g)
    if len(groups) == 0:
      return None
    else:
      return groups

  elif func == 'snmp_group_add':
    secModel = '1'
    secModel = SecurityModels[args[1]]
    if secModel == 'usm':
      secLevel = SecurityLevels[args[2]]
      index = 3
    else:
      secLevel = 'no-auth-no-priv'
      index = 2
    argsList = []
    if len(args) >  index:
      argsList = args[index:]
    viewOpts = { 'read' : 'None', 'write' : 'None', 'notify' : 'None'}
    argsDict = dict(zip(*[iter(argsList)]*2))
    for key in argsDict:
      viewOpts[key] = argsDict[key]

    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access={contextName},{securityModel},{securityLevel}/read-view'
    keypath = cc.Path(path, name=args[0], contextName="Default", securityModel=secModel, securityLevel=secLevel)
    entry = { "ietf-snmp:read-view" : viewOpts['read'],
              "ietf-snmp:write-view" : viewOpts['write'],
              "ietf-snmp:notify-view" : viewOpts['notify'] }
    response = aa.patch(keypath, entry)
    if response.ok():
      return None
    else:
      return response

  # Remove a group.
  elif func == 'snmp_group_delete':
    secModel = '1'
    secModel = SecurityModels[args[1]]
    if secModel == 'usm':
      secLevel = SecurityLevels[args[2]]
    else:
      secLevel = 'no-auth-no-priv'

    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access={contextName},{securityModel},{securityLevel}'
    keypath = cc.Path(path, name=args[0], contextName="Default", securityModel=secModel, securityLevel=secLevel)
    response = aa.delete(keypath)

    return None

  # Get the configured views.
  elif func == 'snmp_view_get':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/view')
    response=aa.get(keypath)
    views = []
    if response.ok():
      content = response.content
      if 'ietf-snmp:view' in response.content.keys():
        for key, data in response.content.items():
          while len(data) > 0:
            row = data.pop(0)
            for action in ['include', 'exclude']:
              if row.has_key(action):
                for oidTree in row[action]:
                  v = {}
                  v['name'] = row['name']
                  v['type'] = action+'d'
                  v['oid'] = oidTree
                  views.append(v)

    if len(views) == 0:
      return None
    else:
      return views

  # Add a view.
  elif func == 'snmp_view_add':
    action = args[2].rstrip('d')      # one of 'exclude' or 'include'
    path = '/restconf/data/ietf-snmp:snmp/vacm/view={name}/%s' %action
    keypath = cc.Path(path, name=args[0])
    row = "ietf-snmp:%s" %action
    entry = { row: [ args[1] ] }
    response = aa.patch(keypath, entry)
    if response.ok():
      return None
    else:
      return response

  # Remove a view.
  elif func == 'snmp_view_delete':
    for action in ['exclude', 'include']:
      # though only one exists, extraneous action appears harmless
      path = '/restconf/data/ietf-snmp:snmp/vacm/view={name}/%s={oidtree}' %action
      keypath = cc.Path(path, name=args[0], oidtree=args[1])
      response = aa.delete(keypath)
    if response.ok():
      return None
    else:
      return response

  # Get the configured users.
  elif func == 'snmp_user_get':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/usm/local/user')
    response=aa.get(keypath)
    users = []
    if response.ok():
      if 'ietf-snmp:user' in response.content.keys():
        for key, data in response.content.items():
          while len(data) > 0:
            row = data.pop(0)
            u = {}
            u['username'] = row['name']
            u['group'] = row['group']
            # The following two cannot yet be set
            u['auth'] = 'None'
            u['priv'] = 'None'
            users.append(u)

    if len(users) == 0:
      return None
    else:
      return users

  elif func == 'snmp_user_add':
    user = args.pop(0)
    invoke('snmp_user_delete', [user])        # delete user config if it already exists

    engineID = getEngineID()
    group = 'None'
    if args.count("group") > 0:
      index = args.index("group")
      group = args.pop(index+1)
      args.pop(index)

    encrypted = False
    if len(args) > 0 and args[0].lower() == 'encrypted':
      # check if authentication is encrypted
      encrypted = True
      args.pop(0)

    authType = None
    authPassword = '00000000000000000000000000000000'
    authKey = None
    if len(args) > 0 and args[0].lower() == 'auth':
    # At this point, only authentication and privacy information remain in args[]. Parse that.
    # 'auth' will always be the first argument. Don't need it.
      args.pop(0)
      authType = args.pop(0).lower()
      if authType in ['md5', 'sha']:
        args.pop(0)   # remove 'auth-password'
        authPassword = args.pop(0)
      elif authType == 'noauth':
        authType = None

    privType = None
    privPassword = '00000000000000000000000000000000'
    privKey = None
    if len(args) > 0 and args[0].lower() == 'priv':
    # At this point, only privacy information remains in args[]. Parse that.
    # 'priv' will always be the first argument. Don't need it.
      args.pop(0)
      privType = args.pop(0).lower()
      if privType == 'aes-128':
        privType = 'aes'
      if privType in ['des', 'aes']:
        args.pop(0)   # remove 'priv-password'
        privPassword = args.pop(0)

    if authType == None:
      authType = "md5"
      privType = "des"
    elif not (encrypted):
      privacyType = privType
      if privType == None:
        privacyType = "des"
      try:
        rc = subprocess.check_output(["snmpkey", authType, authPassword, engineID, privacyType, privPassword])
      except:
        response = aa.cli_not_implemented("None")
        response.set_error_message("Cannot compute md5 key for user %s" %user)
        return response

      authStr, crlf, privStr = rc.partition('\n')
      securityDict = {}
      if crlf == '\n':        # split was good
        for element in authStr, privStr:
          key, space, data = element.partition(' ')
          if space == ' ':          # split was good
            securityDict[key.rstrip(':')] = data[2:].rstrip() # trim prepended '0x' from the encrypted value and trailing colon from key
      if len(securityDict) > 0:    # good authentication and privacy key are found
        authPassword = securityDict['authKey']
        if not privType == None:
          privPassword = securityDict['privKey']

    authKey = createYangHexStr(authPassword)
    privKey = createYangHexStr(privPassword)

    entry=collections.defaultdict(dict)
    entry["user"]=[{ "name": user,
                     "group" : group,
# the following two crash the rest server
#                     "auth" : { authType : { "key": authKey}},
#                     "priv" : { privType : { "key": privKey}}
                    }]
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/usm/local/user')
    response = aa.patch(keypath, entry)

    if response.ok():
      return None
    else:
      return response.content

  # Remove a user.
  elif func == 'snmp_user_delete':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/usm/local/user={index}', index=args[0])
    response = aa.delete(keypath)

    if response.ok():
      return None
    else:
      return response

  # Get the configured hosts.
  elif func == 'snmp_host_get':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target')
    response=aa.get(keypath)
    hosts_c = []
    hosts_u = []
    if response.ok():
      if 'ietf-snmp:target' in response.content.keys():
        for key, table in response.content.items():
          hosts_c, hosts_u = ([] for i in range(2))
          while len(table) > 0:
            data = table.pop(0)
            h = {}
            h['target'] = data['name']
            for key, value in data.items():
              if key == 'tag':
                h['trapOrInform'] = value[0]
#                if value[1] == 'usm':              # User Entry
#                  h['user'] = value[3]
#                  h['security'] = LevelsSecurity(value[2])
#                else:
#                  h['community'] = value[3]
#                  h['security'] = value[1]
              h[key] = value
            if "user" not in h:
              hosts_c.append(h)
            else:
              hosts_u.append(h)

    if len(hosts_c) == 0 and len(hosts_u) == 0:
      return None
    else:
      return hosts_c, hosts_u

  # Add a host.
  elif func == 'snmp_host_add':
    type = 'trap'
    if 'user' == args[1]:
      secModel = SecurityModels['v3']
    else:
      secModel = SecurityModels['v1']

    invoke('snmp_host_delete', [args[0]])        # delete user config if it already exists
    secLevel = SecurityLevels['noauth']
    params = { 'timeout': '15', 'retries': '3' }
    if len(args) > 3:
      type = (args[3].rstrip('s'))+'Notify'      # one of 'trapNotify' or 'informNotify'
      index = 4
      if secModel == SecurityModels['v3']:
        secLevel = SecurityLevels[args[4]]
        index = 5
      if len(args) > index:
        if type == 'trapNotify':
          secModel = args[index]
        else:
          params[args[index]] = args[index+1]
          if len(args) > (index+2):
            params[args[index+2]] = args[index+3]
    targetEntry=collections.defaultdict(dict)
    targetEntry["target"]=[{ "name": args[0],
                             "timeout": 100*int(params['timeout']),
                             "retries": int(params['retries']),
                             "target-params": args[0],
                             "tag": [ type ],
                             "udp" : { "ip": args[0], "port": 162}
                             }]
    if secModel == 'usm':
      security = { "user-name": args[2],
                   "security-level": secLevel}
    else:
      security = { "security-name": args[2]}

    targetParams=collections.defaultdict(dict)
    targetParams["target-params"]=[{ "name": args[0],
                                     secModel : security }]
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target')
    response = aa.patch(keypath, targetEntry)

    if response.ok():
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target-params')
      response = aa.patch(keypath, targetParams)

    if response.ok():
      return None
    else:
      return response

  # Remove a host.
  elif func == 'snmp_host_delete':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target={name}', name=args[0])
    response = aa.delete(keypath)
    if response.ok():
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target-params={name}', name=args[0])
      response = aa.delete(keypath)
    if response.ok():
      return None
    else:
      return response

  else:
      print("%Error: %func not implemented")
      exit(1)

  return None

def run(func, args):
  try:
    api_response = invoke(func, args)

    if api_response == None:
      return

    if api_response is not None:
      if func == 'snmp_get':
        show_cli_output(args[0], api_response)
      elif func == 'snmp_community_get':
        show_cli_output(args[0], api_response)
      elif func == 'snmp_view_get':
        show_cli_output(args[0], api_response)
      elif func == 'snmp_group_get':
        show_cli_output(args[0], api_response)
      elif func == 'snmp_user_get':
        show_cli_output(args[0], api_response)
      elif func == 'snmp_host_get':
        list_c, list_u = api_response
        show_cli_output(args[0], list_c)
        show_cli_output('show_snmp_host_user.j2', list_u)
      else:
        print api_response.error_message()

  except:
    # system/network error
    print "%Error: Transaction Failure"

if __name__ == '__main__':
  pipestr().write(sys.argv)
  run(sys.argv[1], sys.argv[2:])
