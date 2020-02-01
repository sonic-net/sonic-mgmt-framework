#!/usr/bin/python
import os
import subprocess
import sys
import json
import collections
import re
import swsssdk
import cli_client as cc
import socket
import ipaddress
from rpipe_utils import pipestr
from scripts.render_cli import show_cli_output
from swsssdk import ConfigDBConnector
from operator import itemgetter
from collections import OrderedDict

ALLOW_SYSNAME=False
"""
module: ietf-snmp
  +--rw snmp
     +--rw usm
     |  +--rw local
     |  |  +--rw user* [name]
     |  |     +--rw name     snmp:identifier
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
AGENTADDRESS    = 'SNMP_AGENT_ADDRESS_CONFIG'
SYSTEM          = 'SYSTEM'
SNMP_SERVER     = 'SNMP_SERVER'
SNMP_SERVER_GROUP_MEMBER = 'SNMP_SERVER_GROUP_MEMBER'
sysname         = 'sysName'
contact         = 'sysContact'
location        = 'sysLocation'
traps           = 'traps'
context         = 'Default'
SecurityModels = { 'any' : 'any', 'v1': 'v1', 'v2c': 'v2c', 'v3': 'usm' }
SecurityLevels = { 'noauth' : 'no-auth-no-priv', 'auth' : 'auth-no-priv', 'priv' : 'auth-priv' }
ViewOpts       = { 'read' : 'readView', 'write' : 'writeView', 'notify' : 'notifyView'}
SORTED_ORDER   = ['sysName', 'sysLocation','sysContact', 'engineID', 'traps']

config_db = ConfigDBConnector()
if config_db is None:
  sys.exit()
config_db.connect()

aa = cc.ApiClient()

def manageGroupMasterKey(group):
  """ Group table has two sub-tables, Access and Memmber.
      This routine removes the master if it is no longer needed.
  """
  deleteGroup = True

  response = invoke('snmp_group_member_get', None)
  for entry in response.content['group-member']:
    if entry['name'] == group:
      deleteGroup = False

  response = invoke('snmp_group_access_get', None)
  for entry in response.content['group-access']:
    if entry['name'] == group:
      deleteGroup = False

  if deleteGroup == True:
    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}'
    keypath = cc.Path(path, name=group)
    response = aa.delete(keypath)

  return deleteGroup

def entryNotFound(response):
  """ Helper routine to detect entries that are not found. """
  error_data = response.content['ietf-restconf:errors']['error'][0]
  if 'error-message' in error_data:
    err_msg = error_data['error-message']
    if err_msg == 'Entry not found':
      return True
  return False

def findKeyForTargetEntry(ipAddr):
  """ Search the Target Table for ipAddr and return the key
      Keys are of the form targetEntry1, targetEntry2, etc.
  """
  keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target')
  response=aa.get(keypath)
  if response.ok():
    if 'ietf-snmp:target' in response.content.keys():
      for key, table in response.content.items():
        while len(table) > 0:
          data = table.pop(0)
          udp = data['udp']
          if udp['ip'] == ipAddr:
            return data['name']
  return "None"

def findNextKeyForTargetEntry(ipAddr):
  """ Find the next available TargetEntry key """
  key = "None"
  index = 1
  while 1:
    key = "targetEntry{}".format(index)
    index += 1
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target={name}', name=key)
    response=aa.get(keypath)
    if response.ok():
      if len(response.content) == 0:
        break
    else:
      break
  return key

def createYangHexStr(textString):
  """ Convert plain hex string into yang:hex-string """
  data = textString[0:2]
  i = 2
  while i < len(textString):
    data = data + ':' + textString[i:i+2]
    i = i + 2
  return data

def getEngineID():
  """ Construct SNMP engineID from the configured value or from scratch """
  keypath = cc.Path('/restconf/data/ietf-snmp:snmp/engine/engine-id')
  response=aa.get(keypath)

  # First, try to get engineID via rest
  engineID = ''
  if response.ok():
    content = response.content
    if content.has_key('ietf-snmp:engine-id'):
      engineID = content['ietf-snmp:engine-id']
    elif content.has_key('engine-id'):
      engineID = content['engine-id']
    engineID = engineID.encode('ascii')
    engineID = engineID.translate(None, ':')

  # ensure engineID is properly formatted before use. See RFC 3411
  try:
    # must be hex (base 16)
    value = int(engineID, 16)
    # length of 5 - 32 octets
    if len(engineID) < 10:
      engineID = ''
    if len(engineID) > 64:
      engineID = ''
  except:
    # Whoops, not hex
    engineID = ''

  # if the engineID is not configured, construct as per SnmpEngineID 
  # TEXTUAL-CONVENTION in RFC 3411 using the system MAC address.
  if len(engineID) == 0:
    sysmac = config_db.get_entry('DEVICE_METADATA', "localhost").get('mac')
    if sysmac == None:
      # All else fails, something must be used. Fabricated MAC Address
      sysmac = '00:00:00:12:34:56'
    sysmac = sysmac.translate(None, ':')
    # engineID is:
    # 3) The length of the octet string varies.
    #   bit 0 == '1'
    #   The snmpEngineID has a length of 12 octets
    #   The first four octets are set to the binary equivalent of the agent's 
    #     SNMP management private enterprise number as assigned by IANA.
    #     Microsoft = 311 = 0000 0137
    #   The fifth octet indicates how the rest (6th andfollowing octets) are formatted.
    #     3     - MAC address (6 octets)
    #   + System MAC address
    engineID = "8000013703"+sysmac

  return engineID

def set_system(row, data):
  """ Set a system entry using direct write to config_db  """
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

def getIPType(x):
  try: socket.inet_pton(socket.AF_INET6, x)
  except socket.error:
    return False
  return True

def getAgentAddresses():
  """ Read system saved agent addresses.
      This has history in the CLICK config command:
        config snmpagentaddress add [-p <udpPort>] [-v <vrfName>] <IpAddress>
        config snmpagentaddress del [-p <udpPort>] [-v <vrfName>] <IpAddress>
      The key to this table is  ipaddr|udpPort|ifname
  """
  tableData = config_db.get_table(AGENTADDRESS)
  agentAddresses = []
  if len(tableData) > 0:
    for data in tableData.iterkeys():          # the key is the data
      ipAddr, udpPort, ifName = data
      agentAddresses.append({ "ipAddr" : ipAddr, "udpPort" : udpPort, "ifName" : ifName })
  return agentAddresses

def invoke(func, args):
  if func == 'snmp_get':
   keys = config_db.get_keys(SNMP_SERVER)
   datam = {}
   for key in keys:
     datam = config_db.get_entry(SNMP_SERVER, key)
   datam['engineID'] = getEngineID()
   if len(datam) > 0:
     order = []
     for key in SORTED_ORDER:
       if datam.has_key(key):
         order.append(key)
     tuples = [(key, datam[key]) for key in order]
     datam = OrderedDict(tuples) 

   agentAddr = {}
   agentAddresses = getAgentAddresses()
   if len(agentAddresses) > 0:
     agentAddr['agentAddr'] = agentAddresses

   response=aa.cli_not_implemented("global")      # Just to get the proper format to return data and status
   response.content = {}                          # This method is used extensively throughout
   response.status_code = 204
   response.content['system'] = datam
   response.content['global'] = agentAddr
   return response

  elif func == 'snmp_sysname':
    if ALLOW_SYSNAME==False:
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
      # Configure Engine ID
      engineID = createYangHexStr(args[0])
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/engine')
      entry=collections.defaultdict(dict)
      entry['engine']={ "engine-id" : engineID }
      response = aa.patch(keypath, entry)
    else:
      # Remove Engine ID
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/engine/engine-id')
      response = aa.delete(keypath)

    return response

  elif func == 'snmp_agentaddr' or func == 'no_snmp_agentaddr':
    ipAddress = args.pop(0)
    port = '161'                    # standard IPv4 listening UDP port.
    interface = ''
    if 'port' in args:
      index = args.index('port')
      port = args[index+1]
    if 'interface' in args:
      index = args.index('interface')
      interface = args[index+1]
    key = (ipAddress, port, interface)
    entry = None                    # default is to delete the entry
    if func == 'snmp_agentaddr':
      entry = {}                    # if configuring, this tells set_entry to create it
    config_db.set_entry(AGENTADDRESS, key, entry)
    return None

  # Get the configured communities.
  elif func == 'snmp_community_get':
    groupResps = invoke('snmp_group_member_get', None)
    groups = {}
    if groupResps.ok():
      for grpResponse in groupResps.content['group-member']:
        if grpResponse['security-model'] == 'v2c':                # communities only
          comm = grpResponse['security-name']
          grp =  grpResponse['name']
          groups[comm] = grp

    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community')
    response=aa.get(keypath)
    data = []
    if response.ok():
      if 'ietf-snmp:community' in response.content.keys():
        communities = response.content['ietf-snmp:community']
        for community in communities:
          community['group'] = groups[community['index']]
        response.content['community'] = sorted(communities, key=itemgetter('index'))
    return response

  # Configure a new community.
  elif func == 'snmp_community_add':
    invoke('snmp_community_delete', [args[0]])     # delete community config if it already exists
    group="None"
    if (1<len(args)):
      group=args[1]
    entry=collections.defaultdict(dict)
    entry["community"]=[{ "index": args[0],
                          "security-name" : group }]
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community')
    response = aa.patch(keypath, entry)
    if response.ok():
      member = [group, args[0], 'v2c']
      response = invoke('snmp_group_member_add', member)
    return response

  # Remove a community.
  elif func == 'snmp_community_delete':
    group = 'None'
    groupResps = invoke('snmp_group_member_get', None)
    for grpResponse in groupResps.content['group-member']:
      if grpResponse['security-name'] == args[0] and grpResponse['security-model'] == 'v2c':
        group = grpResponse['name']
        break

    member = [group, args[0]]
    response = invoke('snmp_group_member_del', member)

    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/community={index}', index=args[0])
    response = aa.delete(keypath)
    return response

#============================================================================
  # Get the configured member groups.
  elif func == 'snmp_group_member_get':
    groups = []
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/group')
    response = aa.get(keypath)
    if response.ok():
      if 'ietf-snmp:group' in response.content.keys():
        groupDict = response.content['ietf-snmp:group']
        while len(groupDict) > 0:
          row = groupDict.pop(0)
          group = row['name']

          if 'member' in row.keys():
            members = row['member']
            for member in members:
              g = {}
              g['name'] = group
              g['security-name'] = member['security-name']
              secModel = member['security-model']
              g['security-model'] = secModel.pop()
              groups.append(g)

          else:
            path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/member'
            keypath = cc.Path(path, name = group)
            response = aa.get(keypath)
            if response.ok():
              if 'ietf-snmp:member' in response.content.keys():
                data = response.content['ietf-snmp:member']
                while len(data) > 0:
                  entry = data.pop(0)
                  g = {}
                  g['name'] = group
                  g['security-name'] = entry['security-name']
                  g['security-model'] = entry['security-model']
                  groups.append(g)

    response=aa.cli_not_implemented("group")              # just to get the proper format
    response.content = {}
    response.status_code = 200
    response.content['group-member'] = sorted(groups, key=itemgetter('name', 'security-model', 'security-name'))

    return response

  elif func == 'snmp_group_member_add':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/group')
    entry=collections.defaultdict(dict)
    entry["group"]=[{ "name" : args[0] }]
    response = aa.patch(keypath, entry)

    if response.ok():
      #path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/member={secName}/security-model'
      path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/member={secName}'
      keypath = cc.Path(path, name=args[0], secName=args[1])
      entry=collections.defaultdict(dict)
      entry["member"]=[{ "security-name" : args[1],
                         "security-model" : [args[2]]
                         }]
      response = aa.patch(keypath, entry)

    return response

  # Remove an member group.
  elif func == 'snmp_group_member_del':
    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/member={secName}'
    keypath = cc.Path(path, name=args[0], secName=args[1])
    response = aa.delete(keypath)

    # only delete master key if all access and all member entries are removed.
    if response.ok() or entryNotFound(response):
      manageGroupMasterKey(args[0])

    return response
#============================================================================

  # Get the configured access groups.
  elif func == 'snmp_group_access_get':
    groups = []
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/group')
    response = aa.get(keypath)
    if response.ok():
      if 'ietf-snmp:group' in response.content.keys():
        groupDict = response.content['ietf-snmp:group']
        while len(groupDict) > 0:
          row = groupDict.pop(0)
          group = row['name']
          # Simple get request for '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access'
          # returns 'not found' but a search for every possible combo should return one match
          # if an access entry exists. Could be a member entry.
          # path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access'
          for security in [('any', 'no-auth-no-priv'), ('v1', 'no-auth-no-priv'), ('v2c', 'no-auth-no-priv'), ('usm', 'no-auth-no-priv'), ('usm', 'auth-no-priv'), ('usm', 'auth-priv')]:
            secModel, secLevel = security
            path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access=Default,{securityModel},{securityLevel}'
            keypath = cc.Path(path, name = group, securityModel=secModel, securityLevel=secLevel)
            response = aa.get(keypath)
            if response.ok():
              if 'ietf-snmp:access' in response.content.keys():
                data = response.content['ietf-snmp:access']
                while len(data) > 0:
                  entry = data.pop(0)
                  if 'read-view' in entry.keys():
                    g = {}
                    g['name'] = group
                    g['context'] = entry['context']
                    if secModel == "usm":
                      g['model'] = 'v3'
                    else:
                      g['model'] = secModel
                    g['security'] = secLevel
                    g['read-view']   = entry['read-view']
                    g['write-view']  = entry['write-view']
                    g['notify-view'] = entry['notify-view']
                    groups.append(g)

    response=aa.cli_not_implemented("group")              # just to get the proper format
    response.content = {}
    response.status_code = 204
    response.content['group-access'] = sorted(groups, key=itemgetter('name', 'model', 'security'))

    return response

  # Add an access group.
  elif func == 'snmp_group_access_add':
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

    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/vacm/group')
    entry=collections.defaultdict(dict)
    entry["group"]=[{ "name" : args[0] }]
    response = aa.patch(keypath, entry)

    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access={contextName},{securityModel},{securityLevel}/read-view'
    keypath = cc.Path(path, name=args[0], contextName="Default", securityModel=secModel, securityLevel=secLevel)
    entry = { "ietf-snmp:read-view" : viewOpts['read'],
              "ietf-snmp:write-view" : viewOpts['write'],
              "ietf-snmp:notify-view" : viewOpts['notify'] }
    response = aa.patch(keypath, entry)
    return response

  # Remove an access group.
  elif func == 'snmp_group_access_delete':
    secModel = SecurityModels[args[1]]
    if secModel == 'usm':
      secLevel = SecurityLevels[args[2]]
    else:
      secLevel = 'no-auth-no-priv'

    path = '/restconf/data/ietf-snmp:snmp/vacm/group={name}/access={contextName},{securityModel},{securityLevel}'
    keypath = cc.Path(path, name=args[0], contextName="Default", securityModel=secModel, securityLevel=secLevel)
    response = aa.delete(keypath)

    # only delete master key if all access and all member entries are removed.
    if response.ok() or entryNotFound(response):
      manageGroupMasterKey(args[0])

    return response

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
                  response.content = {'view' : sorted(views, key=itemgetter('name', 'oid'))}
    return response

  # Add a view.
  elif func == 'snmp_view_add':
    action = args[2].rstrip('d')      # one of 'exclude' or 'include'
    path = '/restconf/data/ietf-snmp:snmp/vacm/view={name}/%s' %action
    keypath = cc.Path(path, name=args[0])
    row = "ietf-snmp:%s" %action
    entry = { row: [ args[1] ] }
    response = aa.patch(keypath, entry)
    return response

  # Remove a view.
  elif func == 'snmp_view_delete':
    for action in ['exclude', 'include']:
      # though only one exists, extraneous action appears harmless
      path = '/restconf/data/ietf-snmp:snmp/vacm/view={name}/%s={oidtree}' %action
      keypath = cc.Path(path, name=args[0], oidtree=args[1])
      response = aa.delete(keypath)
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
            u['group'] = 'None'
            groupResps = invoke('snmp_group_member_get', None)
            for grpResponse in groupResps.content['group-member']:
              if grpResponse['security-name'] == u['username'] and grpResponse['security-model'] == 'usm':
                u['group'] = grpResponse['name']
                break

            auth = row['auth']
            if auth.has_key('md5'):
              u['auth'] = 'md5'
            elif auth.has_key('sha'):
              u['auth'] = 'sha'
            else:
              u['auth'] = 'None'
            key = auth[u['auth']]
            value = key['key']
            value = value.encode('ascii')
            value = value.translate(None, ':')
            if value == '00000000000000000000000000000000':
              u['auth'] = 'None'

            u['priv'] = 'None'
            if row.has_key('priv'):
              priv = row['priv']
              if priv.has_key('aes'):
                u['priv'] = 'aes'
              elif priv.has_key('des'):
                u['priv'] = 'des'
              else:
                u['priv'] = 'None'
              key = priv[u['priv']]
              value = key['key']
              value = value.encode('ascii')
              value = value.translate(None, ':')
              if value == '00000000000000000000000000000000':
                u['priv'] = 'None'
              if u['priv'] == 'aes':
                u['priv'] = 'aes-128'

            users.append(u)
      response.content = {'user' : users}
    return response

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

    payload = {}
    payload["name"] = user
    payload["auth"] = { authType : { "key": authKey}}
    if not privType == None:
      payload["priv"] = { privType : { "key": privKey}}

    entry=collections.defaultdict(dict)
    entry["user"]=[payload]
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/usm/local/user')
    response = aa.patch(keypath, entry)
    if response.ok():
      member = [group, user, 'usm']
      response = invoke('snmp_group_member_add', member)
    return response

  # Remove a user.
  elif func == 'snmp_user_delete':
    groupResps = invoke('snmp_group_member_get', None)
    group = 'None'
    for grpResponse in groupResps.content['group-member']:
      if grpResponse['security-name'] == args[0] and grpResponse['security-model'] == 'usm':
        group = grpResponse['name']
        break
    member = [group, args[0]]
    groupResps = invoke('snmp_group_member_del', member)

    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/usm/local/user={index}', index=args[0])
    response = aa.delete(keypath)
    return response

  # Get the configured hosts.
  elif func == 'snmp_host_get':
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target')
    response=aa.get(keypath)
    hosts4_c = []
    hosts4_u = []
    hosts6_c = []
    hosts6_u = []
    if response.ok():
      if 'ietf-snmp:target' in response.content.keys():
        for key, table in response.content.items():
          hosts4_c, hosts6_c, hosts4_u, hosts6_u = ([] for i in range(4))
          while len(table) > 0:
            data = table.pop(0)
            h = {}
            h['target'] = data['target-params']
            udp = data['udp']
            h['ipaddr'] = udp['ip']
            h['ip6'] = getIPType(h['ipaddr'])
            for key, value in data.items():
              if key == 'target-params':
                path = cc.Path('/restconf/data/ietf-snmp:snmp/target-params={name}', name=data[key])
                params=aa.get(path)
                if response.ok():
                  if 'ietf-snmp:target-params' in params.content.keys():
                    data = params.content['ietf-snmp:target-params']
                    while len(data) > 0:
                      entry = data.pop(0)
                      if 'v1' in entry:
                        security = entry['v1']
                        h['version'] = 'v1'
                        h['security-name'] = security['security-name']
                      elif 'v2c' in entry:
                        security = entry['v2c']
                        h['version'] = 'v2c'
                        h['security-name'] = security['security-name']
                      elif 'usm' in entry:
                        security = entry['usm']
                        h['version'] = 'usm'
                        h['user-name'] = security['user-name']
                        h['security-level'] = security['security-level']
              if key == 'tag':
                tag = value[0]
                if tag.endswith("Notify"):
                  h['trapOrInform'] = tag[:-6]
                else:
                  h['trapOrInform'] = tag
              h[key] = value
            if 'timeout' in h:
              h['timeout'] = h['timeout']/100  # displayed in seconds
            if "user-name" not in h:
              if not h['ip6']:
                hosts4_c.append(h)
              else:
                hosts6_c.append(h)
            else:
              if not h['ip6']:
                hosts4_u.append(h)
              else:
                hosts6_u.append(h)

    if len(hosts4_c) == 0 and len(hosts6_c) == 0 and len(hosts4_u) == 0 and len(hosts6_u) == 0:
      return None
    else:
      response.content = { "community" : sorted(hosts4_c, key=lambda i: ipaddress.ip_address(i['ipaddr'])) + sorted(hosts6_c, key=lambda i: ipaddress.ip_address(i['ipaddr'])),
                           "user"      : sorted(hosts4_u, key=lambda i: ipaddress.ip_address(i['ipaddr'])) + sorted(hosts6_u, key=lambda i: ipaddress.ip_address(i['ipaddr'])) }
    return response

  # Add a host.
  elif func == 'snmp_host_add':
    key = findKeyForTargetEntry(args[0])
    if key == 'None':
      key = findNextKeyForTargetEntry(args[0])

    type = 'trapNotify'
    if 'user' == args[1]:
      secModel = SecurityModels['v3']
    else:
      secModel = SecurityModels['v2c']                      # v1 is not supported, v2c should be default

    response = invoke('snmp_host_delete', [args[0]])        # delete user config if it already exists
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

    # informs can never be 'v1'
    if type == 'informNotify' and secModel == SecurityModels['v1']:
      secModel = SecurityModels['v2c']

    targetEntry=collections.defaultdict(dict)
    targetEntry["target"]=[{ "name": key,
                             "timeout": 100*int(params['timeout']),
                             "retries": int(params['retries']),
                             "target-params": key,
                             "tag": [ type ],
                             "udp" : { "ip": args[0], "port": 162}
                             }]
    if secModel == 'usm':
      security = { "user-name": args[2],
                   "security-level": secLevel}
    else:
      security = { "security-name": args[2]}

    targetParams=collections.defaultdict(dict)
    targetParams["target-params"]=[{ "name": key,
                                     secModel : security }]


    # since it is the targetParams is a leafref to "target-params",
    # target-params must be written first
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target-params')
    response = aa.patch(keypath, targetParams)

    if response.ok():
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target')
      response = aa.patch(keypath, targetEntry)

    return response

  # Remove a host.
  elif func == 'snmp_host_delete':
    key = findKeyForTargetEntry(args[0])
    keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target={name}', name=key)
    response = aa.delete(keypath)
    if response.ok():
      keypath = cc.Path('/restconf/data/ietf-snmp:snmp/target-params={name}', name=key)
      response = aa.delete(keypath)
    return response

  else:
      print("%Error: %func not implemented "+func)
      exit(1)

  return None

def run(func, args):
  try:
    api_response = invoke(func, args)

    if api_response == None:
      return
    elif api_response.ok():
      if func == 'snmp_get':
        show_cli_output(args[0], api_response.content['system'])
        temp = api_response.content['global']
        if len(temp)>0:
          show_cli_output('show_snmp_agentaddr.j2', temp['agentAddr'])
      elif func == 'snmp_community_get':
        show_cli_output(args[0], api_response.content['community'])
      elif func == 'snmp_view_get':
        show_cli_output(args[0], api_response.content['view'])
      elif func == 'snmp_group_access_get':
        show_cli_output(args[0], api_response.content['group-access'])
      elif func == 'snmp_user_get':
        show_cli_output(args[0], api_response.content['user'])
      elif func == 'snmp_host_get':
        show_cli_output(args[0], api_response.content['community'])
        show_cli_output('show_snmp_host_user.j2', api_response.content['user'])
    #else:
    #  # For some reason, the show commands return a status_code of 500
    #   print(api_response.status_code)
    #   print(api_response.error_message())

  except:
    # system/network error
    print "%Error: Transaction Failure"

if __name__ == '__main__':
  pipestr().write(sys.argv)
  run(sys.argv[1], sys.argv[2:])
