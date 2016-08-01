
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
 * filename: nas_ndi_switch.c
 */

extern "C" {
#include "sai.h"
#include "saiswitch.h"
}

#include "nas_ndi_switch.h"

#include "std_error_codes.h"
#include "ds_common_types.h"
#include "event_log_types.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_event_logs.h"



#include <unordered_map>
#include <algorithm>
#include <vector>

static t_std_error ndi_switch_attr_get(npu_id_t npu,  sai_attribute_t *attr, size_t count) {
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    if (count==0) return STD_ERR_OK;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->get_switch_attribute((uint32_t)count,attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-SWITCH", "Error from SAI:%d in get attrs (at least id:%d) for NPU:%d",
                      sai_ret,attr->id, (int)npu);
         return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

static t_std_error ndi_switch_attr_set(npu_id_t npu,  const sai_attribute_t *attr) {
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->set_switch_attribute(attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-SAI", "Error from SAI:%d for attr:%d in default set for NPU:%d",
                      sai_ret,attr->id,npu);
         return STD_ERR(NPU, CFG, sai_ret);
    }
    return STD_ERR_OK;
}


//Expectation is that the left side is the SAI value, while the right side is the NDI type
typedef std::unordered_map<uint32_t,uint32_t> _enum_map;

static bool to_sai_type(_enum_map &ens, sai_attribute_t *param ) {
    auto it = std::find_if(ens.begin(),ens.end(),
            [param](decltype(*ens.begin()) &a) { return a.second == param->value.u32; });
    if (it==ens.end()) return false;
    param->value.u32 = it->first;
    return true;
}

static bool from_sai_type(_enum_map &ens, sai_attribute_t *param ) {
    auto it = ens.find(param->value.u32);
    if (it==ens.end()) return false;
    param->value.u32 = it->second;
    return true;
}

static _enum_map _algo_stoy  = {
    {SAI_HASH_XOR, BASE_SWITCH_HASH_ALGORITHM_XOR },
    {SAI_HASH_CRC, BASE_SWITCH_HASH_ALGORITHM_CRC },
    {SAI_HASH_RANDOM, BASE_SWITCH_HASH_ALGORITHM_RANDOM },
};

static bool to_sai_type_hash_algo(sai_attribute_t *param ) {
    return to_sai_type(_algo_stoy,param);
}

static bool from_sai_type_hash_algo(sai_attribute_t *param ) {
    return from_sai_type(_algo_stoy,param);
}

static _enum_map _mode_stoy = {
        {SAI_SWITCHING_MODE_CUT_THROUGH , BASE_SWITCH_SWITCHING_MODE_CUT_THROUGH},
        {SAI_SWITCHING_MODE_STORE_AND_FORWARD, BASE_SWITCH_SWITCHING_MODE_STORE_AND_FORWARD }
};

static bool to_sai_type_switch_mode(sai_attribute_t *param ) {
    return to_sai_type(_mode_stoy,param);
}

static bool from_sai_type_switch_mode(sai_attribute_t *param ) {
    return from_sai_type(_mode_stoy,param);
}

static _enum_map _hash_fields_val= {
        {SAI_HASH_SRC_IP, BASE_SWITCH_HASH_FIELDS_SRC_IP},
        {SAI_HASH_DST_IP,BASE_SWITCH_HASH_FIELDS_DEST_IP},
        {SAI_HASH_VLAN_ID,BASE_SWITCH_HASH_FIELDS_VLAN_ID},
        {SAI_HASH_IP_PROTOCOL,BASE_SWITCH_HASH_FIELDS_IP_PROTOCOL},
        {SAI_HASH_ETHERTYPE,BASE_SWITCH_HASH_FIELDS_ETHERTYPE},
        {SAI_HASH_L4_SOURCE_PORT,BASE_SWITCH_HASH_FIELDS_L4_SRC_PORT},
        {SAI_HASH_L4_DEST_PORT,BASE_SWITCH_HASH_FIELDS_L4_DEST_PORT},
        {SAI_HASH_SOURCE_MAC,BASE_SWITCH_HASH_FIELDS_SRC_MAC},
        {SAI_HASH_DEST_MAC,BASE_SWITCH_HASH_FIELDS_DEST_MAC},
        {SAI_HASH_IN_PORT,BASE_SWITCH_HASH_FIELDS_IN_PORT},
};

static bool to_sai_type_hash_fields(sai_attribute_t *param ) {
    sai_attribute_t tmp;
    size_t ix = 0;
    for ( ; ix < param->value.s32list.count; ++ix ) {
        tmp.value.u32 = param->value.s32list.list[ix];
        if (!to_sai_type(_hash_fields_val,&tmp)) return false;
        param->value.s32list.list[ix] = tmp.value.u32;
    }
    return true;
}

static bool from_sai_type_hash_fields(sai_attribute_t *param ) {
    sai_attribute_t tmp;
    size_t ix = 0;
    for ( ; ix < param->value.s32list.count; ++ix ) {
        tmp.value.s32 = param->value.s32list.list[ix];
        if (!from_sai_type(_hash_fields_val,&tmp)) {
            NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-SAI", "Hash value - %d can't be converted",tmp.value.u32);
            return false;
        }
        param->value.s32list.list[ix] = tmp.value.s32;
    }
    return true;
}

enum nas_ndi_switch_attr_op_type {
    SW_ATTR_U16,
    SW_ATTR_U32,
    SW_ATTR_S32,
    SW_ATTR_LST,
    SW_ATTR_MAC,
};

struct _sai_op_table {
    nas_ndi_switch_attr_op_type type;
    bool (*to_sai_type)( sai_attribute_t *param );
    bool (*from_sai_type)(sai_attribute_t *param );
    sai_attr_id_t id;
};

static std::unordered_map<uint32_t,_sai_op_table> _attr_to_op = {
    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_DEFAULT_MAC_ADDRESS, {
            SW_ATTR_MAC, NULL, NULL, SAI_SWITCH_ATTR_SRC_MAC_ADDRESS } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_LAG_HASH_ALGORITHM, {
            SW_ATTR_U32, to_sai_type_hash_algo, from_sai_type_hash_algo, SAI_SWITCH_ATTR_LAG_HASH_ALGO } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ECMP_HASH_ALGORITHM, {
            SW_ATTR_U32, to_sai_type_hash_algo, from_sai_type_hash_algo, SAI_SWITCH_ATTR_ECMP_HASH_ALGO } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_SWITCH_MODE, {
            SW_ATTR_U32, to_sai_type_switch_mode, from_sai_type_switch_mode, SAI_SWITCH_ATTR_SWITCHING_MODE } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_BRIDGE_TABLE_SIZE, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_FDB_TABLE_SIZE}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_MAX_ECMP_ENTRY_PER_GROUP, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_ECMP_MEMBERS}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_LAG_HASH_FIELDS, {
            SW_ATTR_LST, to_sai_type_hash_fields, from_sai_type_hash_fields, SAI_SWITCH_ATTR_LAG_HASH_FIELDS } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ECMP_HASH_FIELDS, {
            SW_ATTR_LST, to_sai_type_hash_fields, from_sai_type_hash_fields, SAI_SWITCH_ATTR_ECMP_HASH_FIELDS } },

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_MAC_AGE_TIMER, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_FDB_AGING_TIME}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_TEMPERATURE, {
            SW_ATTR_S32, NULL,NULL,SAI_SWITCH_ATTR_MAX_TEMP}},
    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ACL_TABLE_MIN_PRIORITY, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_ACL_TABLE_MINIMUM_PRIORITY}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ACL_TABLE_MAX_PRIORITY, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_ACL_TABLE_MAXIMUM_PRIORITY}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ACL_ENTRY_MIN_PRIORITY, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_ACL_ENTRY_MINIMUM_PRIORITY}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_ACL_ENTRY_MAX_PRIORITY, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_ACL_ENTRY_MAXIMUM_PRIORITY}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_MAX_MTU, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_PORT_MAX_MTU}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_TOTAL_BUFFER_SIZE, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_TOTAL_BUFFER_SIZE}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_INGRESS_BUFFER_POOL_NUM, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_INGRESS_BUFFER_POOL_NUM}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_EGRESS_BUFFER_POOL_NUM, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_EGRESS_BUFFER_POOL_NUM}},

    {BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_COUNTER_REFRESH_INTERVAL, {
            SW_ATTR_U32, NULL,NULL,SAI_SWITCH_ATTR_COUNTER_REFRESH_INTERVAL}},

};

extern "C" t_std_error ndi_switch_set_attribute(npu_id_t npu, BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_t attr,
        const nas_ndi_switch_param_t *param) {
    auto it = _attr_to_op.find(attr);
    if (it==_attr_to_op.end()) {
        NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-SAI", "Invalid operation type for NDI (%d)",attr);
        return STD_ERR(NPU,FAIL,0);
    }
    sai_attribute_t sai_attr;
    sai_attr.id = it->second.id;
    std::vector<sai_int32_t> tmp_lst;

    switch(it->second.type) {
    case SW_ATTR_S32:
        sai_attr.value.s32 = param->s32;
        break;
    case SW_ATTR_U32:
        sai_attr.value.u32 = param->u32;
        break;
    case SW_ATTR_LST:
        sai_attr.value.s32list.count = param->list.len;
        tmp_lst.resize(param->list.len);
        memcpy(&tmp_lst[0],param->list.vals,param->list.len*sizeof(tmp_lst[0]));
        sai_attr.value.s32list.list = &tmp_lst[0];
        break;
    case SW_ATTR_U16:
        sai_attr.value.u16 = param->u16;
        break;
    case SW_ATTR_MAC:
        memcpy(sai_attr.value.mac, param->mac, sizeof(sai_attr.value.mac));
        break;
    }
    if (it->second.to_sai_type!=NULL) {
        if (!(it->second.to_sai_type)(&sai_attr)) {
            NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-SAI", "Values are invalid - can't be converted to SAI types (func:%d)",attr);
            return STD_ERR(NPU,PARAM,0);
        }
    }

    return ndi_switch_attr_set(npu,&sai_attr);
}

extern "C" t_std_error ndi_switch_get_attribute(npu_id_t npu, BASE_SWITCH_SWITCHING_ENTITIES_SWITCHING_ENTITY_t attr,
        nas_ndi_switch_param_t *param) {
    auto it = _attr_to_op.find(attr);
    if (it==_attr_to_op.end()) {
        NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-SAI", "Invalid operation type for NDI (%d)",attr);
        return STD_ERR(NPU,FAIL,0);
    }
    sai_attribute_t sai_attr;
    sai_attr.id = it->second.id;

    switch(it->second.type) {
    case SW_ATTR_LST:
        sai_attr.value.s32list.count = param->list.len;
        sai_attr.value.s32list.list = (int32_t*)param->list.vals;
        break;
    default:
        break;
    }

    t_std_error rc = ndi_switch_attr_get(npu,&sai_attr,1);

    if(rc!=STD_ERR_OK) return rc;
    if (it->second.from_sai_type!=NULL) {
        if (!(it->second.from_sai_type)(&sai_attr)) {
            NDI_LOG_ERROR(ev_log_s_MAJOR, "NDI-SAI", "Values are invalid - can't be converted to SAI types (func:%d)",attr);
            return STD_ERR(NPU,PARAM,0);
        }
    }

    switch(it->second.type) {
    case SW_ATTR_S32:
        param->s32 = sai_attr.value.s32;
        break;
    case SW_ATTR_U32:
        param->u32 = sai_attr.value.u32;
        break;
    case SW_ATTR_LST:
        param->list.len = sai_attr.value.s32list.count ;
        break;
    case SW_ATTR_U16:
        param->u16 = sai_attr.value.u16;
        break;
    case SW_ATTR_MAC:
        memcpy(param->mac, sai_attr.value.mac,sizeof(param->mac));
        break;
    }

    return STD_ERR_OK;
}

extern "C" t_std_error ndi_switch_mac_age_time_set(npu_id_t npu_id, uint32_t timeout_value)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Invalid npu id %d to set mac age timeout value",
                      npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    memset(&sai_attr, 0, sizeof(sai_attribute_t));

    sai_attr.id = SAI_SWITCH_ATTR_FDB_AGING_TIME;
    sai_attr.value.u32 = (sai_uint32_t)timeout_value;

    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->set_switch_attribute(&sai_attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Error from SAI %d to set mac age timeout value %d",
                      sai_ret, timeout_value);
         return STD_ERR(MAC, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

extern "C" t_std_error ndi_switch_mac_age_time_get(npu_id_t npu_id, uint32_t *timeout_value)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_attr;
    uint32_t attr_count = 1;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Invalid npu id %d to get mac age timeout value",
                      npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    memset(&sai_attr, 0, sizeof(sai_attribute_t));

    sai_attr.id = SAI_SWITCH_ATTR_FDB_AGING_TIME;

    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->get_switch_attribute(attr_count, &sai_attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Error from  SAI %d to get mac age timeout value",
                      sai_ret);
         return STD_ERR(MAC, CFG, sai_ret);
    }

    *timeout_value = (sai_uint32_t) sai_attr.value.u32;

    return STD_ERR_OK;
}

extern "C" t_std_error ndi_switch_set_sai_log_level(BASE_SWITCH_SUBSYSTEM_t api_id,
                                                    BASE_SWITCH_LOG_LEVEL_t level)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    if(api_id == BASE_SWITCH_SUBSYSTEM_ALL){
        for(size_t ix = 0 ; ix < BASE_SWITCH_SUBSYSTEM_ALL ; ++ix){
            if ((sai_ret = sai_log_set((sai_api_t)ix,(sai_log_level_t)level))!= SAI_STATUS_SUCCESS){
                NDI_LOG_TRACE(0, "NDI-DIAG", "Error from  SAI %d to set log level %d"
                              "for sai_module %d",sai_ret,level,api_id);
                return STD_ERR(DIAG, PARAM, sai_ret);
            }
        }
    }
    else{
        if ((sai_ret = sai_log_set((sai_api_t)api_id,(sai_log_level_t)level))!= SAI_STATUS_SUCCESS){
            NDI_LOG_TRACE(0, "NDI-DIAG", "Error from  SAI %d to set log level %d"
                          "for sai_module %d",sai_ret,level,api_id);
            return STD_ERR(DIAG, PARAM, sai_ret);
        }
    }
    return STD_ERR_OK;
}

extern "C" t_std_error ndi_switch_get_queue_numbers(npu_id_t npu_id,
                        uint32_t *ucast_queues, uint32_t *mcast_queues,
                        uint32_t *total_queues, uint32_t *cpu_queues)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_attr[4];
    uint32_t attr_count = 4;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Invalid npu id %d to get queue value",
                      npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    memset(&sai_attr, 0, sizeof(sai_attr));

    sai_attr[0].id = SAI_SWITCH_ATTR_NUMBER_OF_UNICAST_QUEUES;
    sai_attr[1].id = SAI_SWITCH_ATTR_NUMBER_OF_MULTICAST_QUEUES;
    sai_attr[2].id = SAI_SWITCH_ATTR_NUMBER_OF_QUEUES;
    sai_attr[3].id = SAI_SWITCH_ATTR_NUMBER_OF_CPU_QUEUES;

    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->get_switch_attribute(attr_count, sai_attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_s_MAJOR, "NDI-MAC", "Error from  SAI %d to get queue value",
                      sai_ret);
         return STD_ERR(NPU, CFG, sai_ret);
    }

    if (ucast_queues)
        *ucast_queues = (sai_uint32_t) sai_attr[0].value.u32;

    if (mcast_queues)
        *mcast_queues = (sai_uint32_t) sai_attr[1].value.u32;

    if (total_queues)
        *total_queues = (sai_uint32_t) sai_attr[2].value.u32;

    if (cpu_queues)
        *cpu_queues = (sai_uint32_t) sai_attr[3].value.u32;

    return STD_ERR_OK;
}
