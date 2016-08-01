
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
 * filename: nas_ndi_qos_map.cpp
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


typedef std::unordered_map<ndi_qos_map_type_t, sai_qos_map_type_t, std::hash<int>> ndi_2_sai_qos_map_type_mapping;
ndi_2_sai_qos_map_type_mapping    NDI_2_SAI_QOS_MAP_TYPE = {
        {NDI_QOS_MAP_DOT1P_TO_TC,       SAI_QOS_MAP_DOT1P_TO_TC},
        {NDI_QOS_MAP_DOT1P_TO_COLOR,    SAI_QOS_MAP_DOT1P_TO_COLOR},
        {NDI_QOS_MAP_DOT1P_TO_TC_COLOR, SAI_QOS_MAP_DOT1P_TO_TC_AND_COLOR},
        {NDI_QOS_MAP_DSCP_TO_TC,        SAI_QOS_MAP_DSCP_TO_TC},
        {NDI_QOS_MAP_DSCP_TO_COLOR,     SAI_QOS_MAP_DSCP_TO_COLOR},
        {NDI_QOS_MAP_DSCP_TO_TC_COLOR,  SAI_QOS_MAP_DSCP_TO_TC_AND_COLOR},
        {NDI_QOS_MAP_TC_TO_QUEUE,       SAI_QOS_MAP_TC_TO_QUEUE},
        {NDI_QOS_MAP_TC_TO_DSCP,        SAI_QOS_MAP_TC_TO_DSCP},
        {NDI_QOS_MAP_TC_TO_DOT1P,       SAI_QOS_MAP_TC_TO_DOT1P},
        {NDI_QOS_MAP_TC_COLOR_TO_DSCP,  SAI_QOS_MAP_TC_AND_COLOR_TO_DSCP},
        {NDI_QOS_MAP_TC_COLOR_TO_DOT1P, SAI_QOS_MAP_TC_AND_COLOR_TO_DOT1P},
        {NDI_QOS_MAP_TC_TO_PG,          SAI_QOS_MAP_TC_TO_PRIORITY_GROUP},
        {NDI_QOS_MAP_PG_TO_PFC,         SAI_QOS_MAP_PRIORITY_GROUP_TO_PFC_PRIORITY},
        {NDI_QOS_MAP_PFC_TO_QUEUE,      SAI_QOS_MAP_PFC_PRIORITY_TO_QUEUE},

};

sai_packet_color_t ndi2sai_qos_packet_color(BASE_QOS_PACKET_COLOR_t color)
{
    if (color == BASE_QOS_PACKET_COLOR_RED)
        return SAI_PACKET_COLOR_RED;

    if (color == BASE_QOS_PACKET_COLOR_YELLOW)
        return SAI_PACKET_COLOR_YELLOW;

    return SAI_PACKET_COLOR_GREEN;
}

typedef std::unordered_map<sai_qos_map_type_t, ndi_qos_map_type_t, std::hash<int>> sai_2_ndi_qos_map_type_mapping;
sai_2_ndi_qos_map_type_mapping    SAI_2_NDI_QOS_MAP_TYPE = {
        {SAI_QOS_MAP_DOT1P_TO_TC,           NDI_QOS_MAP_DOT1P_TO_TC},
        {SAI_QOS_MAP_DOT1P_TO_COLOR,        NDI_QOS_MAP_DOT1P_TO_COLOR},
        {SAI_QOS_MAP_DOT1P_TO_TC_AND_COLOR, NDI_QOS_MAP_DOT1P_TO_TC_COLOR},
        {SAI_QOS_MAP_DSCP_TO_TC,            NDI_QOS_MAP_DSCP_TO_TC},
        {SAI_QOS_MAP_DSCP_TO_COLOR,         NDI_QOS_MAP_DSCP_TO_COLOR},
        {SAI_QOS_MAP_DSCP_TO_TC_AND_COLOR,  NDI_QOS_MAP_DSCP_TO_TC_COLOR},
        {SAI_QOS_MAP_TC_TO_QUEUE,           NDI_QOS_MAP_TC_TO_QUEUE},
        {SAI_QOS_MAP_TC_TO_DSCP,            NDI_QOS_MAP_TC_TO_DSCP},
        {SAI_QOS_MAP_TC_TO_DOT1P,           NDI_QOS_MAP_TC_TO_DOT1P},
        {SAI_QOS_MAP_TC_AND_COLOR_TO_DSCP,  NDI_QOS_MAP_TC_COLOR_TO_DSCP},
        {SAI_QOS_MAP_TC_AND_COLOR_TO_DOT1P, NDI_QOS_MAP_TC_COLOR_TO_DOT1P},
        {SAI_QOS_MAP_TC_TO_PRIORITY_GROUP,  NDI_QOS_MAP_TC_TO_PG},
        {SAI_QOS_MAP_PFC_PRIORITY_TO_QUEUE, NDI_QOS_MAP_PFC_TO_QUEUE},
        {SAI_QOS_MAP_PRIORITY_GROUP_TO_PFC_PRIORITY, NDI_QOS_MAP_PG_TO_PFC},
};

BASE_QOS_PACKET_COLOR_t sai2ndi_qos_packet_color(sai_packet_color_t color)
{
    if (color == SAI_PACKET_COLOR_RED)
        return BASE_QOS_PACKET_COLOR_RED;

    if (color == SAI_PACKET_COLOR_YELLOW)
        return BASE_QOS_PACKET_COLOR_YELLOW;

    return BASE_QOS_PACKET_COLOR_GREEN;
}

static void _fill_sai_request_map_entries(sai_qos_map_t * entry_list,
        uint_t map_entry_count,
        const ndi_qos_map_struct_t *map_entry)
{
    for (uint_t i = 0; i<map_entry_count; i++) {
        entry_list[i].key.color = ndi2sai_qos_packet_color(map_entry[i].key.color);
        entry_list[i].key.dot1p = map_entry[i].key.dot1p;
        entry_list[i].key.dscp = map_entry[i].key.dscp;
        entry_list[i].key.queue_index = ndi2sai_map_qid(map_entry[i].key.qid);
        entry_list[i].key.tc = map_entry[i].key.tc;
        entry_list[i].key.prio = map_entry[i].key.prio;
        entry_list[i].key.pg = map_entry[i].key.pg;

        entry_list[i].value.color = ndi2sai_qos_packet_color(map_entry[i].value.color);
        entry_list[i].value.dot1p = map_entry[i].value.dot1p;
        entry_list[i].value.dscp = map_entry[i].value.dscp;
        entry_list[i].value.queue_index = ndi2sai_map_qid(map_entry[i].value.qid);
        entry_list[i].value.tc = map_entry[i].value.tc;
        entry_list[i].value.pg = map_entry[i].value.pg;
        entry_list[i].value.prio = map_entry[i].value.prio;
    }
}

static void _fill_sai_request_map_key(sai_qos_map_t * entry_list,
        uint_t map_entry_count,
        const ndi_qos_map_struct_t *map_entry)
{
    for (uint_t i = 0; i<map_entry_count; i++) {
        entry_list[i].key.color = ndi2sai_qos_packet_color(map_entry[i].key.color);
        entry_list[i].key.dot1p = map_entry[i].key.dot1p;
        entry_list[i].key.dscp = map_entry[i].key.dscp;
        entry_list[i].key.queue_index = ndi2sai_map_qid(map_entry[i].key.qid);
        entry_list[i].key.tc = map_entry[i].key.tc;
        entry_list[i].key.prio = map_entry[i].key.prio;
        entry_list[i].key.pg = map_entry[i].key.pg;
    }
}

static void _fill_ndi_response_map_entries(sai_qos_map_t * entry_list,
        uint_t map_entry_count,
        ndi_qos_map_struct_t *map_entry)
{
    for (uint_t i = 0; i<map_entry_count; i++) {
        map_entry[i].value.color = sai2ndi_qos_packet_color(entry_list[i].value.color);
        map_entry[i].value.dot1p = entry_list[i].value.dot1p;
        map_entry[i].value.dscp = entry_list[i].value.dscp;
        map_entry[i].value.qid = sai2ndi_map_qid(entry_list[i].value.queue_index);
        map_entry[i].value.tc = entry_list[i].value.tc;
        map_entry[i].value.prio = entry_list[i].value.prio;
        map_entry[i].value.pg = entry_list[i].value.pg;
    }
}


/**
 * This function creates a map ID in the NPU.
 * @param npu id
 * @param type of qos map
 * @param number of qos_map_struct to follow
 * @param key-to-value mappings
 * @param[out] ndi_map_id
 * @return standard error
 */
t_std_error ndi_qos_create_map(npu_id_t npu_id,
                                ndi_qos_map_type_t type,
                                uint_t map_entry_count,
                                const ndi_qos_map_struct_t *map_entry,
                                ndi_obj_id_t *ndi_map_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    uint_t attr_count = 1;  // map-type only
    sai_attribute_t attr_list[2];

    if (map_entry_count > 0)
        attr_count = 2; // Plus map-list

    try {

        attr_list[0].id = SAI_QOS_MAP_ATTR_TYPE;
        attr_list[0].value.s32 = NDI_2_SAI_QOS_MAP_TYPE.at(type);

    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "Unexpected error.\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_qos_map_id;
    if ((sai_ret = ndi_sai_qos_map_api(ndi_db_ptr)->
                        create_qos_map(&sai_qos_map_id,
                                attr_count,
                                attr_list))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d map creation failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }
    *ndi_map_id = sai2ndi_qos_map_id(sai_qos_map_id);

    return STD_ERR_OK;

}

 /**
  * This function updates one key-to-value mapping for a map.
  * @param npu id
  * @param ndi_map_id
  * @param number of map entries
  * @param key-to-value mapping
  * @return standard error
  */
t_std_error ndi_qos_set_map_attr(npu_id_t npu_id,
                     ndi_obj_id_t ndi_map_id,
                     uint_t map_entry_count,
                     const ndi_qos_map_struct_t *map_entry)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_qos_map_t> entry_list(map_entry_count);
    sai_attribute_t attr_list[1];

    if (map_entry_count == 0)
        return STD_ERR(QOS, CFG, 0);

    try {
        _fill_sai_request_map_entries(&entry_list[0], map_entry_count, map_entry);

        attr_list[0].id = SAI_QOS_MAP_ATTR_MAP_TO_VALUE_LIST;
        attr_list[0].value.qosmap.count = map_entry_count;
        attr_list[0].value.qosmap.list = &entry_list[0];
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "Unexpected error.\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_qos_map_id = ndi2sai_qos_map_id(ndi_map_id);
    if ((sai_ret = ndi_sai_qos_map_api(ndi_db_ptr)->
                        set_qos_map_attribute(sai_qos_map_id, attr_list))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d map set failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}


/**
 * This function deletes a map in the NPU.
 * @param npu_id npu id
 * @param ndi_map_id
 * @return standard error
 */
t_std_error ndi_qos_delete_map(npu_id_t npu_id,
                               ndi_obj_id_t ndi_map_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_qos_map_id = ndi2sai_qos_map_id(ndi_map_id);
    if ((sai_ret = ndi_sai_qos_map_api(ndi_db_ptr)->
                        remove_qos_map(sai_qos_map_id))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

/**
 * This function get a map from the NPU.
 * @param npu id
 * @param ndi_map_id
 * @param[out] type of qos map
 * @param number of qos_map_struct to be queried
 * @param[out] key-to-value mappings
 * @return standard error
 */
t_std_error ndi_qos_get_map(npu_id_t npu_id,
                            ndi_obj_id_t ndi_map_id,
                            ndi_qos_map_type_t *type,
                            uint_t map_entry_count,
                            ndi_qos_map_struct_t *map_entry)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    uint_t attr_count = 1; // map-type
    std::vector<sai_qos_map_t> entry_list(map_entry_count);
    sai_attribute_t attr_list[2];

    if (map_entry_count > 0)
        attr_count = 2; //plus map-list


    try {
        // Fill the request key
        attr_list[0].id = SAI_QOS_MAP_ATTR_TYPE;

        if (attr_count == 2) {
            _fill_sai_request_map_key(&entry_list[0], map_entry_count, map_entry);

            attr_list[1].id = SAI_QOS_MAP_ATTR_MAP_TO_VALUE_LIST;
            attr_list[1].value.qosmap.count = map_entry_count;
            attr_list[1].value.qosmap.list = &entry_list[0];
        }
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "Unexpected error.\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_object_id_t sai_qos_map_id = ndi2sai_qos_map_id(ndi_map_id);
    if ((sai_ret = ndi_sai_qos_map_api(ndi_db_ptr)->
                        get_qos_map_attribute(sai_qos_map_id,
                                attr_count,
                                attr_list))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d map get fails\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // fill the outgoing parameters
    if (type != NULL)
        *type = SAI_2_NDI_QOS_MAP_TYPE[(sai_qos_map_type_t)(attr_list[0].value.s32)];

    _fill_ndi_response_map_entries(&entry_list[0], map_entry_count, map_entry);

    return STD_ERR_OK;


}

