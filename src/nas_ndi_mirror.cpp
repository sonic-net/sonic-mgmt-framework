
/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 *  LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */


/*
 * nas_ndi_mirror.cpp
 */

#include "nas_ndi_mirror.h"
#include "std_error_codes.h"
#include "saimirror.h"
#include "dell-base-mirror.h"
#include "dell-base-common.h"
#include "event_log.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_common.h"

#include <new>
#include <utility>
#include <vector>
#include <unordered_map>
#include <map>
#include <unordered_set>
#include <stdlib.h>
#include <inttypes.h>
#include <algorithm>

#define MAX_MIRROR_SAI_ATTR 15

#define NDI_MIRROR_LOG(type,lvl, msg, ...)\
                       EV_LOG(type,NAS_L2,lvl,"NDI-MIRROR",msg, ##__VA_ARGS__)

// Key for maintaining port and mirroring direction to ndi mirror id
struct port_to_mirror_id_map_key{
    npu_port_t port;
    sai_port_attr_t dir;

    //Comparison operator for map
    bool operator () (port_to_mirror_id_map_key const &A, port_to_mirror_id_map_key const &B) const {
        return (A.port < B.port || (A.port == B.port && A.dir < B.dir));
    }

};

// List of Mirror ids
typedef std::vector<ndi_mirror_id_t> mirror_ids;

// Map which maintains port and mirroring direction to ndi mirror id mapping
static std::map<port_to_mirror_id_map_key,mirror_ids,port_to_mirror_id_map_key > port_to_mirror_id_map;
typedef std::pair<port_to_mirror_id_map_key,mirror_ids > port_to_mirror_id_pair;


static std::unordered_map<BASE_CMN_TRAFFIC_PATH_t, sai_port_attr_t, std::hash<int>>
ndi_mirror_dir_to_sai_map = {
    {BASE_CMN_TRAFFIC_PATH_INGRESS,SAI_PORT_ATTR_INGRESS_MIRROR_SESSION},
    {BASE_CMN_TRAFFIC_PATH_EGRESS,SAI_PORT_ATTR_EGRESS_MIRROR_SESSION}
};


static std::unordered_map<BASE_MIRROR_MODE_t, sai_mirror_type_t, std::hash<int>>
ndi_mirror_type_to_sai_map = {
    {BASE_MIRROR_MODE_SPAN,SAI_MIRROR_TYPE_LOCAL},
    {BASE_MIRROR_MODE_RSPAN,SAI_MIRROR_TYPE_REMOTE},
    {BASE_MIRROR_MODE_ERSPAN,SAI_MIRROR_TYPE_ENHANCED_REMOTE}
};


static std::unordered_map<sai_mirror_type_t, BASE_MIRROR_MODE_t, std::hash<int>>
ndi_mirror_type_from_sai_map = {
    {SAI_MIRROR_TYPE_LOCAL,BASE_MIRROR_MODE_SPAN},
    {SAI_MIRROR_TYPE_REMOTE,BASE_MIRROR_MODE_RSPAN},
    {SAI_MIRROR_TYPE_ENHANCED_REMOTE,BASE_MIRROR_MODE_ERSPAN}
};


static inline  sai_mirror_api_t * ndi_mirror_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_mirror_api_tbl);
}


static inline  sai_port_api_t * ndi_port_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_port_api_tbl);
}


static bool ndi_mirror_fill_common_attr(ndi_mirror_entry_t * entry, sai_attribute_t * attr_list,
                                           unsigned int & attr_ix){

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_TYPE;
    auto it = ndi_mirror_type_to_sai_map.find(entry->mode);

    if(it == ndi_mirror_type_to_sai_map.end() ){
        NDI_MIRROR_LOG(ERR,0,"Not a valid Mirror type %d",entry->mode);
        return false;
    }

    attr_list[attr_ix++].value.s32 = it->second;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_MONITOR_PORT;
    if(entry->is_dest_lag){
        attr_list[attr_ix++].value.oid = (sai_object_id_t)entry->ndi_lag_id;
    }else{
        if(!ndi_port_to_sai_oid(&entry->dst_port,&attr_list[attr_ix++].value.oid)) return false;
    }

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_TC;
    attr_list[attr_ix++].value.u8 = 0;

    return true;
}


static void ndi_mirror_fill_rspan_attr(ndi_mirror_entry_t * entry, sai_attribute_t * attr_list,
                                                unsigned int & attr_ix){

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_VLAN_TPID;
    attr_list[attr_ix++].value.u16 =NDI_VLAN_TPID;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_VLAN_ID;
    attr_list[attr_ix++].value.u16 = entry->vlan_id;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_VLAN_PRI;
    attr_list[attr_ix++].value.u8 = 0;

}


static void ndi_mirror_fill_erspan_attr(ndi_mirror_entry_t * entry, sai_attribute_t * attr_list,
                                                unsigned int & attr_ix){

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_SRC_IP_ADDRESS;
    attr_list[attr_ix].value.ipaddr.addr.ip4 = entry->src_ip.u.v4_addr;
    attr_list[attr_ix++].value.ipaddr.addr_family = SAI_IP_ADDR_FAMILY_IPV4;


    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_DST_IP_ADDRESS;
    attr_list[attr_ix].value.ipaddr.addr.ip4 = entry->dst_ip.u.v4_addr;
    attr_list[attr_ix++].value.ipaddr.addr_family = SAI_IP_ADDR_FAMILY_IPV4;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_SRC_MAC_ADDRESS;
    memcpy( attr_list[attr_ix++].value.mac,entry->src_mac,sizeof(entry->src_mac));

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_DST_MAC_ADDRESS;
    memcpy( attr_list[attr_ix++].value.mac,entry->dst_mac,sizeof(entry->dst_mac));

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_TOS;
    attr_list[attr_ix++].value.u8 = 0;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_IPHDR_VERSION;
    attr_list[attr_ix++].value.u8 = NDI_IPV4_VERSION;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_TTL;
    attr_list[attr_ix++].value.u8 = NDI_TTL;

    attr_list[attr_ix].id = SAI_MIRROR_SESSION_ATTR_GRE_PROTOCOL_TYPE;
    attr_list[attr_ix++].value.u16 = NDI_GRE_TYPE;

    attr_list[attr_ix].id= SAI_MIRROR_SESSION_ATTR_ENCAP_TYPE;
    attr_list[attr_ix++].value.s32 = SAI_MIRROR_L3_GRE_TUNNEL;

}


t_std_error ndi_mirror_create_session(ndi_mirror_entry_t * entry){

    if(entry == NULL ){
        NDI_MIRROR_LOG(ERR,0,"NDI Mirror entry passed to create Mirror session is NULL");
        return STD_ERR(MIRROR,PARAM,0);
    }

    sai_status_t sai_ret;
    sai_attribute_t sai_mirror_attr_list[MAX_MIRROR_SAI_ATTR];
    unsigned int ndi_mirror_attr_count = 0;

    if(!ndi_mirror_fill_common_attr(entry,sai_mirror_attr_list,ndi_mirror_attr_count)){
        return STD_ERR(MIRROR,PARAM,0);
    }

    if(entry->mode == BASE_MIRROR_MODE_RSPAN || entry->mode == BASE_MIRROR_MODE_ERSPAN){
        ndi_mirror_fill_rspan_attr(entry,sai_mirror_attr_list,ndi_mirror_attr_count);
    }

    if(entry->mode == BASE_MIRROR_MODE_ERSPAN){
        ndi_mirror_fill_erspan_attr(entry,sai_mirror_attr_list,ndi_mirror_attr_count);
    }


    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(entry->dst_port.npu_id);

    if ((sai_ret = ndi_mirror_api_get(ndi_db_ptr)->create_mirror_session((sai_object_id_t *)
                &entry->ndi_mirror_id,ndi_mirror_attr_count,sai_mirror_attr_list))
                != SAI_STATUS_SUCCESS) {
        NDI_MIRROR_LOG(ERR,0,"Failed to create a new Mirroring Session");
        return STD_ERR(MIRROR, FAIL, sai_ret);
    }

    NDI_MIRROR_LOG(INFO,3,"Created new mirroring session with Id %" PRIu64 " ",entry->ndi_mirror_id);

    return STD_ERR_OK;
}


t_std_error ndi_mirror_delete_session(ndi_mirror_entry_t * entry){

    if(entry == NULL ){
        NDI_MIRROR_LOG(ERR,0,"NDI Mirror entry passed to delete Mirror session is NULL");
        return STD_ERR(MIRROR,PARAM,0);
    }

    sai_status_t sai_ret;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(entry->dst_port.npu_id);

    if ((sai_ret = ndi_mirror_api_get(ndi_db_ptr)->remove_mirror_session((sai_object_id_t)
                            entry->ndi_mirror_id))!= SAI_STATUS_SUCCESS) {
        NDI_MIRROR_LOG(ERR,0,"Failed to delete Mirroring Session %d",entry->ndi_mirror_id);
        return STD_ERR(MIRROR, FAIL, sai_ret);
    }

    NDI_MIRROR_LOG(INFO,3,"Deleted mirroring session with Id %d",entry->ndi_mirror_id);

    return STD_ERR_OK;
}


static bool ndi_mirror_get_port_to_id_list(port_to_mirror_id_map_key & key,sai_attribute_t *attr,
                                           ndi_mirror_id_t id,bool enable){

    auto port_it = port_to_mirror_id_map.find(key);

    /* If have to enable Mirroring on a source port, then check if there is already a mirroring session
     * on that port in the same direction, if so add the new mirror id to list and pass new mirror
     * id list for that port to NPU. Otherwise create a new map with give port and its direction
     * as its key and add mirror id as its value and pass to NPU
     *
     * In case of removing mirroring from source port, get the map entry, remove the mirror id from
     * mirror id list and update the NPU
     */

    if(enable){
        if(port_it != port_to_mirror_id_map.end()){
            mirror_ids & ndi_mirror_ids = port_it->second;
            if(std::find(ndi_mirror_ids.begin(),ndi_mirror_ids.end(),id)
                                                == ndi_mirror_ids.end()){
                ndi_mirror_ids.push_back(id);
            }
            attr->value.objlist.count = ndi_mirror_ids.size();
            attr->value.objlist.list = (sai_object_id_t *)&ndi_mirror_ids[0];

        }else{
            mirror_ids ndi_mirror_ids;
            ndi_mirror_ids.push_back(id);
            port_to_mirror_id_map.insert(port_to_mirror_id_pair(key,std::move(ndi_mirror_ids)));
            attr->value.objlist.count = port_to_mirror_id_map[key].size();
            attr->value.objlist.list = (sai_object_id_t *)port_to_mirror_id_map[key].data();

        }
    } else {
        if(port_it == port_to_mirror_id_map.end()){
            NDI_MIRROR_LOG(ERR,0,"No port has mirror id %" PRIu64 " configured on port %d in direction %d"
                    ,id,key.port,key.dir);
            return false;
        }
        mirror_ids & ndi_mirror_ids = port_it->second;
        auto it = std::find(ndi_mirror_ids.begin(),ndi_mirror_ids.end(),id);
        if(it == ndi_mirror_ids.end()){
            NDI_MIRROR_LOG(ERR,0,"No port has mirror id %d configured on port %d in direction %d"
                        ,id,key.port,key.dir);
            return false;
        }
        ndi_mirror_ids.erase(it);
        attr->value.objlist.count = ndi_mirror_ids.size();
        if(ndi_mirror_ids.size() > 0){
            attr->value.objlist.list = (sai_object_id_t *)&ndi_mirror_ids[0];
        }
    }
    return true;
}


t_std_error ndi_mirror_update_direction(ndi_mirror_entry_t *entry, ndi_mirror_src_port_t port,
                                        bool enable){

    if(entry == NULL ){
        NDI_MIRROR_LOG(ERR,0,"NDI Mirror entry passed to update Mirror session is NULL");
        return STD_ERR(MIRROR,PARAM,0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(port.src_port.npu_id);

    auto it = ndi_mirror_dir_to_sai_map.find(port.direction);
    if(it == ndi_mirror_dir_to_sai_map.end()){
        NDI_MIRROR_LOG(ERR,0,"Invalid Direction %d passed to updated entry %d",port.direction
                                                                    ,entry->ndi_mirror_id);
        return STD_ERR(MIRROR,PARAM,0);
    }

    sai_attribute_t mirror_attr;
    mirror_attr.id = it->second;

    port_to_mirror_id_map_key key;
    key.dir = it->second;
    key.port = port.src_port.npu_port;
    if(!ndi_mirror_get_port_to_id_list(key,&mirror_attr,entry->ndi_mirror_id,enable)){
        return false;
    }

    sai_status_t  sai_ret;
    sai_object_id_t port_oid;
    if(!ndi_port_to_sai_oid(&port.src_port,&port_oid)) return false;

    if ((sai_ret = ndi_port_api_get(ndi_db_ptr)->set_port_attribute(port_oid,
                                                 &mirror_attr))!= SAI_STATUS_SUCCESS) {
        NDI_MIRROR_LOG(ERR,0,"Failed to update the Mirror Direction to %d for entry "
                                "%" PRIu64 " ",port.direction,entry->ndi_mirror_id);
        /*
         * IF new source port can not be added because of resource constraints
         * remove the port to mirror id mapping
         */
        if(!ndi_mirror_get_port_to_id_list(key,&mirror_attr,entry->ndi_mirror_id,!enable)){
            return STD_ERR(MIRROR,FAIL,sai_ret);
        }

        return STD_ERR(MIRROR, FAIL,sai_ret);
    }

    NDI_MIRROR_LOG(INFO,0,"Updated Mirror Direction to %d for entry %" PRIu64 " ",
                                            port.direction,entry->ndi_mirror_id);
    return STD_ERR_OK;
}


static bool ndi_mirror_set_attr(ndi_mirror_entry_t *entry, sai_attribute_t *attr){

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(entry->dst_port.npu_id);
    if (ndi_mirror_api_get(ndi_db_ptr)->set_mirror_session_attribute((sai_object_id_t)
                                  entry->ndi_mirror_id,attr)!= SAI_STATUS_SUCCESS) {
        NDI_MIRROR_LOG(ERR,0,"Failed to set the Mirroring attribute %d in entry %" PRIu64 " ",
                                                                attr->id,entry->ndi_mirror_id);
        return false;
    }
    NDI_MIRROR_LOG(ERR,0,"Updated the attribute %d in entry %" PRIu64 " ",attr->id,
                                                              entry->ndi_mirror_id);
    return true;
}


t_std_error ndi_mirror_update_session(ndi_mirror_entry_t * entry, BASE_MIRROR_ENTRY_t attr_id){

    if(entry == NULL ){
       NDI_MIRROR_LOG(ERR,0,"NDI Mirror entry passed to update Mirror session is NULL");
       return STD_ERR(MIRROR,PARAM,0);
    }

    sai_attribute_t mirror_attr;

    switch(attr_id){

        case BASE_MIRROR_ENTRY_DST_INTF:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_MONITOR_PORT;
            if(!ndi_port_to_sai_oid(&entry->dst_port,&mirror_attr.value.oid)) return false;
            break;

        case BASE_MIRROR_ENTRY_VLAN:
        case BASE_MIRROR_ENTRY_ERSPAN_VLAN_ID:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_VLAN_ID;
            mirror_attr.value.u32 = entry->vlan_id;
            break;

        case BASE_MIRROR_ENTRY_SOURCE_IP:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_SRC_IP_ADDRESS;
            mirror_attr.value.ipaddr.addr.ip4 = entry->src_ip.u.v4_addr;
            mirror_attr.value.ipaddr.addr_family = SAI_IP_ADDR_FAMILY_IPV4;
            break;

        case BASE_MIRROR_ENTRY_DESTINATION_IP:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_DST_IP_ADDRESS;
            mirror_attr.value.ipaddr.addr.ip4 = entry->dst_ip.u.v4_addr;
            mirror_attr.value.ipaddr.addr_family = SAI_IP_ADDR_FAMILY_IPV4;
            break;

        case BASE_MIRROR_ENTRY_SOURCE_MAC:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_SRC_MAC_ADDRESS;
            memcpy(mirror_attr.value.mac,entry->src_mac,sizeof(entry->src_mac));
            break;

        case BASE_MIRROR_ENTRY_DEST_MAC:
            mirror_attr.id = SAI_MIRROR_SESSION_ATTR_DST_MAC_ADDRESS;
            memcpy(mirror_attr.value.mac,entry->dst_mac,sizeof(entry->dst_mac));
            break;

        default:
            NDI_MIRROR_LOG(ERR,0,"Invalid Attribute Id passed to update Mirror Session"
                    " %" PRIu64 "",attr_id,entry->ndi_mirror_id);
            return STD_ERR(MIRROR,PARAM,0);
    }

    if(!ndi_mirror_set_attr(entry,&mirror_attr)){
        return STD_ERR(MIRROR,FAIL,0);
    }

    return STD_ERR_OK;
}

/*@TODO implement it later
 *
 */
t_std_error ndi_mirror_get_session(ndi_mirror_entry_t * entry,npu_id_t npu_id,ndi_mirror_id_t ndi_mirror_id){
    return STD_ERR_OK;
}

