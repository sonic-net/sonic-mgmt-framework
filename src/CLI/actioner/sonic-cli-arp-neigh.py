#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################

#Commands to support
#--------------------
#  To get link layer address of single IP
#  get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors_neighbor
#  get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors_neighbor
#
#  To get all entries
#  get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors
#  get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors

import sys
import time
import json
import ast
import openconfig_interfaces_client
from rpipe_utils import pipestr
from openconfig_interfaces_client.rest import ApiException
from scripts.render_cli import show_cli_output

import urllib3
urllib3.disable_warnings()


plugins = dict()

def register(func):
    """Register sdk client method as a plug-in"""
    plugins[func.__name__] = func
    return func

def call_method(name, args):
    method = plugins[name]
    return method(args)

def generate_body(func, args):
    body = None
    keypath = []
    if func.__name__ == 'get_openconfig_interfaces_interfaces':
        keypath = []
    elif func.__name__ == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors':
            keypath = [args[1], 0]
    elif func.__name__ == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors':
            keypath = [args[1], 0]
    elif func.__name__ == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv4_neighbors_neighbor':
            keypath = [args[1], 0, args[3]]
    elif func.__name__ == 'get_openconfig_if_ip_interfaces_interface_subinterfaces_subinterface_ipv6_neighbors_neighbor':
            keypath = [args[1], 0, args[3]]
    else:
       body = {}
    return keypath, body

def getId(item):
    prfx = "Ethernet"
    state_dict = item['state']
    ifName = state_dict['name']

    if ifName.startswith(prfx):
        ifId = int(ifName[len(prfx):])
        return ifId
    return ifName

def run(func, args):
    c = openconfig_interfaces_client.Configuration()
    c.verify_ssl = False
    aa = openconfig_interfaces_client.OpenconfigInterfacesApi(api_client=openconfig_interfaces_client.ApiClient(configuration=c))

    # create a body block
    keypath, body = generate_body(func, args)
    neigh_list = []

    try:
        api_response = getattr(aa,func.__name__)(*keypath)
        response = api_response.to_dict()

        if 'openconfig_if_ipneighbor' in response.keys():
                neigh = response['openconfig_if_ipneighbor']

                if neigh[0]['state'] is None:
                    return

                ipAddr = neigh[0]['state']['ip']
                if ipAddr is None:
                    return

                macAddr = neigh[0]['state']['link_layer_address']
                if macAddr is None:
                    return

                neigh_table_entry = {'ipAddr':ipAddr,
                                    'macAddr':macAddr,
                                    'intfName':args[1]
                                  }
                neigh_list.append(neigh_table_entry)
        elif 'openconfig_if_ipneighbors' in response.keys():
                if response['openconfig_if_ipneighbors'] is None:
                    return

                neighs = response['openconfig_if_ipneighbors']['neighbor']
                if neighs is None:
                    return

                for neigh in neighs:
                        ipAddr = neigh['state']['ip']
                        if ipAddr is None:
                            return

                        macAddr = neigh['state']['link_layer_address']
                        if macAddr is None:
                            return

                        neigh_table_entry = {'ipAddr':ipAddr,
                                    'macAddr':macAddr,
                                    'intfName':args[1]
                                  }

                        if (args[2] == "mac") and (args[3] != macAddr):
                                print "%Error: Entry not found"
                        else:
                                neigh_list.append(neigh_table_entry)

        show_cli_output(args[0],neigh_list)
        return

    except ApiException as e:
        #print("Exception when calling OpenconfigInterfacesApi->%s : %s\n" %(func.__name__, e))
        if e.body != "":
            body = json.loads(e.body)
            if "ietf-restconf:errors" in body:
                 err = body["ietf-restconf:errors"]
                 if "error" in err:
                     errList = err["error"]

                     errDict = {}
                     for dict in errList:
                         for k, v in dict.iteritems():
                              errDict[k] = v

                     if "error-message" in errDict:
                         print "%Error: " + errDict["error-message"]
                         return
                     print "%Error: Transaction Failure"
                     return
            print "%Error: Transaction Failure"
        else:
            print "Failed"

if __name__ == '__main__':
    pipestr().write(sys.argv)
    func = eval(sys.argv[1], globals(), openconfig_interfaces_client.OpenconfigInterfacesApi.__dict__)

    run(func, sys.argv[2:])


