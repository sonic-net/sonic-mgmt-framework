
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
  * filename: nas_ndi_qos_port.cpp
  */


#include "std_error_codes.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_qos_utl.h"
#include "sai.h"
#include "dell-base-qos.h" //from yang model
#include "nas_ndi_qos.h"

#include <stdio.h>
#include <vector>
#include <unordered_map>


static std::unordered_map<BASE_QOS_FLOW_CONTROL_t, sai_port_flow_control_mode_t, std::hash<int>> \
    ndi2sai_flow_control_map = {
        {BASE_QOS_FLOW_CONTROL_DISABLE,        SAI_PORT_FLOW_CONTROL_DISABLE},
        {BASE_QOS_FLOW_CONTROL_TX_ONLY,        SAI_PORT_FLOW_CONTROL_TX_ONLY},
        {BASE_QOS_FLOW_CONTROL_RX_ONLY,        SAI_PORT_FLOW_CONTROL_RX_ONLY},
        {BASE_QOS_FLOW_CONTROL_BOTH_ENABLE,    SAI_PORT_FLOW_CONTROL_BOTH_ENABLE},
    };

static std::unordered_map<sai_port_flow_control_mode_t, BASE_QOS_FLOW_CONTROL_t, std::hash<int>> \
    sai2ndi_flow_control_map = {
        {SAI_PORT_FLOW_CONTROL_DISABLE,        BASE_QOS_FLOW_CONTROL_DISABLE        },
        {SAI_PORT_FLOW_CONTROL_TX_ONLY,        BASE_QOS_FLOW_CONTROL_TX_ONLY        },
        {SAI_PORT_FLOW_CONTROL_RX_ONLY,        BASE_QOS_FLOW_CONTROL_RX_ONLY        },
        {SAI_PORT_FLOW_CONTROL_BOTH_ENABLE,    BASE_QOS_FLOW_CONTROL_BOTH_ENABLE    },
    };

static  std::unordered_map<BASE_QOS_PORT_INGRESS_t, sai_port_attr_t, std::hash<int>>
    ndi2sai_port_ing_attr_id_map = {
    {BASE_QOS_PORT_INGRESS_POLICER_ID,              SAI_PORT_ATTR_POLICER_ID},
    {BASE_QOS_PORT_INGRESS_FLOOD_STORM_CONTROL,     SAI_PORT_ATTR_FLOOD_STORM_CONTROL_POLICER_ID},
    {BASE_QOS_PORT_INGRESS_BROADCAST_STORM_CONTROL, SAI_PORT_ATTR_BROADCAST_STORM_CONTROL_POLICER_ID},
    {BASE_QOS_PORT_INGRESS_MULTICAST_STORM_CONTROL, SAI_PORT_ATTR_MULTICAST_STORM_CONTROL_POLICER_ID},
    {BASE_QOS_PORT_INGRESS_FLOW_CONTROL,            SAI_PORT_ATTR_GLOBAL_FLOW_CONTROL},
    {BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_NUMBER,   SAI_PORT_ATTR_NUMBER_OF_PRIORITY_GROUPS},
    {BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_ID_LIST,  SAI_PORT_ATTR_PRIORITY_GROUP_LIST},
    {BASE_QOS_PORT_INGRESS_PER_PRIORITY_FLOW_CONTROL, SAI_PORT_ATTR_PRIORITY_FLOW_CONTROL},
    {BASE_QOS_PORT_INGRESS_DEFAULT_TRAFFIC_CLASS,   SAI_PORT_ATTR_QOS_DEFAULT_TC},
    {BASE_QOS_PORT_INGRESS_DOT1P_TO_TC_MAP,         SAI_PORT_ATTR_QOS_DOT1P_TO_TC_MAP},
    {BASE_QOS_PORT_INGRESS_DOT1P_TO_COLOR_MAP,      SAI_PORT_ATTR_QOS_DOT1P_TO_COLOR_MAP},
    {BASE_QOS_PORT_INGRESS_DOT1P_TO_TC_COLOR_MAP,   SAI_PORT_ATTR_QOS_DOT1P_TO_TC_AND_COLOR_MAP},
    {BASE_QOS_PORT_INGRESS_DSCP_TO_TC_MAP,          SAI_PORT_ATTR_QOS_DSCP_TO_TC_MAP},
    {BASE_QOS_PORT_INGRESS_DSCP_TO_COLOR_MAP,       SAI_PORT_ATTR_QOS_DSCP_TO_COLOR_MAP},
    {BASE_QOS_PORT_INGRESS_DSCP_TO_TC_COLOR_MAP,    SAI_PORT_ATTR_QOS_DSCP_TO_TC_AND_COLOR_MAP},
    {BASE_QOS_PORT_INGRESS_TC_TO_QUEUE_MAP,         SAI_PORT_ATTR_QOS_TC_TO_QUEUE_MAP},
    {BASE_QOS_PORT_INGRESS_TC_TO_PRIORITY_GROUP_MAP,SAI_PORT_ATTR_QOS_TC_TO_PRIORITY_GROUP_MAP},
    {BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_TO_PFC_PRIORITY_MAP, SAI_PORT_ATTR_QOS_PRIORITY_GROUP_TO_PFC_PRIORITY_MAP},
    };

static t_std_error _fill_port_qos_ing_attr(BASE_QOS_PORT_INGRESS_t attr_id,
                                 const qos_port_ing_struct_t *p,
                                 sai_attribute_t *attr)
{
    try {
        attr->id = ndi2sai_port_ing_attr_id_map[attr_id];
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "attr_id %d out of range\n", attr_id);
        return STD_ERR(QOS, CFG, 0);
    }

    switch (attr_id) {
    case BASE_QOS_PORT_INGRESS_POLICER_ID:
        attr->value.oid = ndi2sai_policer_id(p->policer_id);
        break;
    case BASE_QOS_PORT_INGRESS_FLOOD_STORM_CONTROL:
        attr->value.oid = ndi2sai_policer_id(p->flood_storm_control);
        break;
    case BASE_QOS_PORT_INGRESS_BROADCAST_STORM_CONTROL:
        attr->value.oid = ndi2sai_policer_id(p->bcast_storm_control);
        break;
    case BASE_QOS_PORT_INGRESS_MULTICAST_STORM_CONTROL:
        attr->value.oid = ndi2sai_policer_id(p->mcast_storm_control);
        break;
    case BASE_QOS_PORT_INGRESS_FLOW_CONTROL:
        attr->value.u8 = ndi2sai_flow_control_map[p->flow_control];
        break;
    case BASE_QOS_PORT_INGRESS_DEFAULT_TRAFFIC_CLASS:
        attr->value.u8 = p->default_tc;
        break;
    case BASE_QOS_PORT_INGRESS_DOT1P_TO_TC_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dot1p_to_tc_map);
        break;
    case BASE_QOS_PORT_INGRESS_DOT1P_TO_COLOR_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dot1p_to_color_map);
        break;
    case BASE_QOS_PORT_INGRESS_DOT1P_TO_TC_COLOR_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dot1p_to_tc_color_map);
        break;
    case BASE_QOS_PORT_INGRESS_DSCP_TO_TC_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dscp_to_tc_map);
        break;
    case BASE_QOS_PORT_INGRESS_DSCP_TO_COLOR_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dscp_to_color_map);
        break;
    case BASE_QOS_PORT_INGRESS_DSCP_TO_TC_COLOR_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->dscp_to_tc_color_map);
        break;
    case BASE_QOS_PORT_INGRESS_TC_TO_QUEUE_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_to_queue_map);
        break;
    case BASE_QOS_PORT_INGRESS_TC_TO_PRIORITY_GROUP_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_to_priority_group_map);
        break;
    case BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_TO_PFC_PRIORITY_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->priority_group_to_pfc_priority_map);
        break;
    case BASE_QOS_PORT_INGRESS_PER_PRIORITY_FLOW_CONTROL:
        attr->value.u8 = p->per_priority_flow_control;
        break;
    case BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_NUMBER:
    case BASE_QOS_PORT_INGRESS_PRIORITY_GROUP_ID_LIST:
        // READ-only
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS", "Read only attributes: %d", attr_id);
        break;
    default:
        return STD_ERR(QOS, CFG, 0);
    }

    return STD_ERR_OK;
}

/**
 * This function sets port ingress profile attributes in the NPU.
 * @param npu id
 * @param port id
 * @param attr_id based on the CPS API attribute enumeration values
 * @param p port ingress structure to be modified
 * @return standard error
 */
t_std_error ndi_qos_set_port_ing_profile_attr(npu_id_t npu_id,
                                 npu_port_t port_id,
                                 BASE_QOS_PORT_INGRESS_t attr_id,
                                 const qos_port_ing_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_port;

    if (ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_attribute_t attr;

    _fill_port_qos_ing_attr(attr_id, p, &attr);
    if ((sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->
                        set_port_attribute(sai_port, &attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "port ingress qos set fails: npu_id %u, port_id %u, sai attr_id %u\n",
                npu_id, port_id, attr.id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;
}


static t_std_error _fill_ndi_qos_port_ing_struct(sai_attribute_t *attr_list,
                        uint_t num_attr, qos_port_ing_struct_t*p)
{
    for (uint_t i = 0 ; i< num_attr; i++ ) {
        sai_attribute_t *attr = &attr_list[i];
        switch (attr->id) {
        case SAI_PORT_ATTR_POLICER_ID:
            p->policer_id = sai2ndi_policer_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_FLOOD_STORM_CONTROL_POLICER_ID:
            p->flood_storm_control = sai2ndi_policer_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_BROADCAST_STORM_CONTROL_POLICER_ID:
            p->bcast_storm_control = sai2ndi_policer_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_MULTICAST_STORM_CONTROL_POLICER_ID:
            p->mcast_storm_control = sai2ndi_policer_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_GLOBAL_FLOW_CONTROL:
            p->flow_control = sai2ndi_flow_control_map[(sai_port_flow_control_mode_t)(attr->value.u8)];
            break;
        case SAI_PORT_ATTR_QOS_DEFAULT_TC:
            p->default_tc = attr->value.u8;
            break;
        case SAI_PORT_ATTR_QOS_DOT1P_TO_TC_MAP:
            p->dot1p_to_tc_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_DOT1P_TO_COLOR_MAP:
            p->dot1p_to_color_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_DOT1P_TO_TC_AND_COLOR_MAP:
            p->dot1p_to_tc_color_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_DSCP_TO_TC_MAP:
            p->dscp_to_tc_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_DSCP_TO_COLOR_MAP:
            p->dscp_to_color_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_DSCP_TO_TC_AND_COLOR_MAP:
            p->dscp_to_tc_color_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_TO_QUEUE_MAP:
            p->tc_to_queue_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_TO_PRIORITY_GROUP_MAP:
            p->tc_to_priority_group_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_PRIORITY_GROUP_TO_PFC_PRIORITY_MAP:
            p->priority_group_to_pfc_priority_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_NUMBER_OF_PRIORITY_GROUPS:
            p->priority_group_number = attr->value.u32;
            break;
        case SAI_PORT_ATTR_PRIORITY_GROUP_LIST:
            p->num_priority_group_id = attr->value.objlist.count;
            for (uint_t j = 0; j < p->num_priority_group_id; j ++) {
                p->priority_group_id_list[j] = sai2ndi_priority_group_id(attr->value.objlist.list[j]);
            }
            break;
        case SAI_PORT_ATTR_PRIORITY_FLOW_CONTROL:
            p->per_priority_flow_control = attr->value.u8;
            break;
        default:
            return STD_ERR(QOS, CFG, 0);
        }
    }
    return STD_ERR_OK;

}

/**
 * This function get a port ingress profile from the NPU.
 * @param npu id
 * @param port id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param[out] qos_port_ing_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_port_ing_profile(npu_id_t npu_id,
                            npu_port_t port_id,
                            const BASE_QOS_PORT_INGRESS_t *nas_attr_list,
                            uint_t num_attr,
                            qos_port_ing_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list(num_attr);
    for (uint_t i = 0; i< num_attr; i++) {
        try {
            attr_list[i].id = ndi2sai_port_ing_attr_id_map[nas_attr_list[i]];
        }
        catch (...) {
            return STD_ERR(QOS, CFG, 0);
        }
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_port;

    if (ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->
                        get_port_attribute(sai_port,
                                num_attr,
                                &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "port ingress qos get fails: npu_id %u, port_id %u, sai attr_id[0] %u\n",
                npu_id, port_id, attr_list[0].id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // fill the outgoing parameters
    _fill_ndi_qos_port_ing_struct(&attr_list[0], num_attr, p);

    return STD_ERR_OK;

}

static  std::unordered_map<BASE_QOS_PORT_EGRESS_t, sai_port_attr_t, std::hash<int>>
    ndi2sai_port_egr_attr_id_map = {
        {BASE_QOS_PORT_EGRESS_TC_TO_QUEUE_MAP,      SAI_PORT_ATTR_QOS_TC_TO_QUEUE_MAP},
        {BASE_QOS_PORT_EGRESS_TC_TO_DOT1P_MAP,      SAI_PORT_ATTR_QOS_TC_TO_DOT1P_MAP},
        {BASE_QOS_PORT_EGRESS_TC_TO_DSCP_MAP,       SAI_PORT_ATTR_QOS_TC_TO_DSCP_MAP},
        {BASE_QOS_PORT_EGRESS_TC_COLOR_TO_DOT1P_MAP,SAI_PORT_ATTR_QOS_TC_AND_COLOR_TO_DOT1P_MAP},
        {BASE_QOS_PORT_EGRESS_TC_COLOR_TO_DSCP_MAP, SAI_PORT_ATTR_QOS_TC_AND_COLOR_TO_DSCP_MAP},
        {BASE_QOS_PORT_EGRESS_WRED_PROFILE_ID,      SAI_PORT_ATTR_QOS_WRED_PROFILE_ID},
        {BASE_QOS_PORT_EGRESS_SCHEDULER_PROFILE_ID, SAI_PORT_ATTR_QOS_SCHEDULER_PROFILE_ID},
        {BASE_QOS_PORT_EGRESS_QUEUE_ID_LIST,        SAI_PORT_ATTR_QOS_QUEUE_LIST},
        {BASE_QOS_PORT_EGRESS_PFC_PRIORITY_TO_QUEUE_MAP,SAI_PORT_ATTR_QOS_PFC_PRIORITY_TO_QUEUE_MAP},
        /** @todo: not supported in SAI api yet
        {BASE_QOS_PORT_EGRESS_NUM_QUEUE,             0},
        {BASE_QOS_PORT_EGRESS_NUM_UNICAST_QUEUE,     0},
        {BASE_QOS_PORT_EGRESS_NUM_MULTICAST_QUEUE,     0},
        {BASE_QOS_PORT_EGRESS_BUFFER_LIMIT,         0},
        */
};

static t_std_error _fill_port_qos_egr_attr(BASE_QOS_PORT_EGRESS_t attr_id,
                                 const qos_port_egr_struct_t *p,
                                 sai_attribute_t *attr)
{
    try {
        attr->id = ndi2sai_port_egr_attr_id_map[attr_id];
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "attr_id %d out of range\n", attr_id);
        return STD_ERR(QOS, CFG, 0);
    }

    switch (attr_id) {
    case BASE_QOS_PORT_EGRESS_TC_TO_QUEUE_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_to_queue_map);
        break;
    case BASE_QOS_PORT_EGRESS_TC_TO_DOT1P_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_to_dot1p_map);
        break;
    case BASE_QOS_PORT_EGRESS_TC_TO_DSCP_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_to_dscp_map);
        break;
    case BASE_QOS_PORT_EGRESS_TC_COLOR_TO_DOT1P_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_color_to_dot1p_map);
        break;
    case BASE_QOS_PORT_EGRESS_TC_COLOR_TO_DSCP_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->tc_color_to_dscp_map);
        break;
    case BASE_QOS_PORT_EGRESS_WRED_PROFILE_ID:
        attr->value.oid = ndi2sai_wred_profile_id(p->wred_profile_id);
        break;
    case BASE_QOS_PORT_EGRESS_SCHEDULER_PROFILE_ID:
        attr->value.oid = ndi2sai_scheduler_profile_id(p->scheduler_profile_id);
        break;
    case BASE_QOS_PORT_EGRESS_QUEUE_ID_LIST:
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "attr_id %u: QUEUE_ID_LIST not settable\n", attr_id);
        return STD_ERR(QOS, CFG, 0);
    case BASE_QOS_PORT_EGRESS_PFC_PRIORITY_TO_QUEUE_MAP:
        attr->value.oid = ndi2sai_qos_map_id(p->pfc_priority_to_queue_map);
        break;

    default:
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "attr_id %u:  not supported\n", attr_id);
        return STD_ERR(QOS, CFG, 0);
    }
    return STD_ERR_OK;
}

/**
 * This function sets port egress profile attributes in the NPU.
 * @param npu id
 * @param port id
 * @param attr_id based on the CPS API attribute enumeration values
 * @param p port egress structure to be modified
 * @return standard error
 */
t_std_error ndi_qos_set_port_egr_profile_attr(npu_id_t npu_id,
                                 npu_port_t port_id,
                                 BASE_QOS_PORT_EGRESS_t attr_id,
                                 const qos_port_egr_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_port;

    if (ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_attribute_t attr;

    _fill_port_qos_egr_attr(attr_id, p, &attr);
    if ((sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->
                        set_port_attribute(sai_port, &attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "port egress qos set fails: npu_id %u, port_id %u, sai attr_id %u\n",
                npu_id, port_id, attr.id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

static t_std_error _fill_ndi_qos_port_egr_struct(const sai_attribute_t *attr_list,
                        uint_t num_attr, qos_port_egr_struct_t*p)
{

    for (uint_t i = 0 ; i< num_attr; i++ ) {
        const sai_attribute_t *attr = &attr_list[i];
        switch (attr->id) {
        case SAI_PORT_ATTR_QOS_TC_TO_QUEUE_MAP:
            p->tc_to_queue_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_TO_DOT1P_MAP:
            p->tc_to_dot1p_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_TO_DSCP_MAP:
            p->tc_to_dscp_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_AND_COLOR_TO_DOT1P_MAP:
            p->tc_color_to_dot1p_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_TC_AND_COLOR_TO_DSCP_MAP:
            p->tc_color_to_dscp_map  = sai2ndi_qos_map_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_WRED_PROFILE_ID:
            p->wred_profile_id = sai2ndi_wred_profile_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_SCHEDULER_PROFILE_ID:
            p->scheduler_profile_id = sai2ndi_scheduler_profile_id(attr->value.oid);
            break;
        case SAI_PORT_ATTR_QOS_QUEUE_LIST:
            p->num_queue_id = attr->value.objlist.count;
            for (uint_t j = 0; j < p->num_queue_id; j ++) {
                p->queue_id_list[j] = sai2ndi_queue_id(attr->value.objlist.list[j]);
            }
            break;
        case SAI_PORT_ATTR_QOS_PFC_PRIORITY_TO_QUEUE_MAP:
            p->pfc_priority_to_queue_map = sai2ndi_qos_map_id(attr->value.oid);
            break;
        default:
            return STD_ERR(QOS, CFG, 0);
        }
    }
    return STD_ERR_OK;

}

static t_std_error ndi_qos_update_port_queue_number(npu_id_t npu_id,
                                        npu_port_t port_id,
                                        qos_port_egr_struct_t *port_egr)
{
    uint_t idx;
    int rc, queue_cnt = 0, ucast_cnt = 0, mcast_cnt = 0;
    ndi_port_t ndi_port;
    ndi_qos_queue_attribute_t queue_info;

    if (!port_egr || port_egr->num_queue_id <= 0 ||
        !port_egr->queue_id_list) {
        return STD_ERR_OK;
    }
    ndi_port.npu_id = npu_id;
    ndi_port.npu_port = port_id;
    for (idx = 0; idx < port_egr->num_queue_id; idx ++) {
        rc = ndi_qos_get_queue_attribute(ndi_port,
                                    port_egr->queue_id_list[idx],
                                    &queue_info);
        if (rc != STD_ERR_OK) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "failed to get queue attribute: npu_id %u port_id %u queue %lx\n",
                npu_id, port_id, port_egr->queue_id_list[idx]);
            return rc;
        }
        if (queue_info.type == BASE_QOS_QUEUE_TYPE_UCAST) {
            ucast_cnt ++;
        } else if (queue_info.type == BASE_QOS_QUEUE_TYPE_MULTICAST) {
            mcast_cnt ++;
        }
        queue_cnt ++;
    }
    port_egr->num_queue = queue_cnt;
    port_egr->num_ucast_queue = ucast_cnt;
    port_egr->num_mcast_queue = mcast_cnt;

    return STD_ERR_OK;
}

/**
* This function get a port egress profile from the NPU.
* @param npu id
* @param port id
* @param nas_attr_list based on the CPS API attribute enumeration values
* @param num_attr number of attributes in attr_list array
* @param[out] qos_port_eg_struct_t filled if success
* @return standard error
*/
t_std_error ndi_qos_get_port_egr_profile(npu_id_t npu_id,
                           npu_port_t port_id,
                           const BASE_QOS_PORT_EGRESS_t  *nas_attr_list,
                           uint_t num_attr,
                           qos_port_egr_struct_t *p)
{

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_object_id_t> queue_id_list;
    std::vector<sai_attribute_t> attr_list(num_attr);
    for (uint_t i = 0; i< num_attr; i++) {
        try {
            attr_list[i].id = ndi2sai_port_egr_attr_id_map[nas_attr_list[i]];
            if (attr_list[i].id == SAI_PORT_ATTR_QOS_QUEUE_LIST) {
                attr_list[i].value.objlist.count = p->num_queue_id;
                queue_id_list.resize(p->num_queue_id);
                attr_list[i].value.objlist.list = &queue_id_list[0];
            }
        }
        catch (...) {
            return STD_ERR(QOS, CFG, 0);
        }
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_port;

    if (ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->
                        get_port_attribute(sai_port,
                                num_attr,
                                &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "port egress qos get fails: npu_id %u, port_id %u, sai attr_id[0] %u\n",
                npu_id, port_id, attr_list[0].id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // fill the outgoing parameters
    _fill_ndi_qos_port_egr_struct(&attr_list[0], num_attr, p);

    ndi_qos_update_port_queue_number(npu_id, port_id, p);

    return STD_ERR_OK;

}
