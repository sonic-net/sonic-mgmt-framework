
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
 * filename: nas_ndi_qos_priority_group.cpp
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


const static std::unordered_map<nas_attr_id_t, sai_attr_id_t, std::hash<int>>
    ndi2sai_priority_group_attr_id_map = {
    {BASE_QOS_PRIORITY_GROUP_BUFFER_PROFILE_ID,            SAI_INGRESS_PRIORITY_GROUP_ATTR_BUFFER_PROFILE},

};

 /**
  * This function sets the priority_group profile attributes in the NPU.
  * @param ndi_port_id
  * @param ndi_priority_group_id
  * @param buffer_profile_id
  * @return standard error
  */
t_std_error ndi_qos_set_priority_group_buffer_profile_id(ndi_port_t ndi_port_id,
                                        ndi_obj_id_t ndi_priority_group_id,
                                        ndi_obj_id_t buffer_profile_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attribute_t sai_attr = {0};
    sai_attr.id = SAI_INGRESS_PRIORITY_GROUP_ATTR_BUFFER_PROFILE;
    sai_attr.value.oid = ndi2sai_buffer_profile_id(buffer_profile_id);

    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            set_ingress_priority_group_attr(
                    ndi2sai_priority_group_id(ndi_priority_group_id),
                    &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d priority_group profile set failed\n", ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

static t_std_error _fill_ndi_qos_priority_group_struct(sai_attribute_t *attr_list,
                        uint_t num_attr, ndi_qos_priority_group_attribute_t *p)
{

    for (uint_t i = 0 ; i< num_attr; i++ ) {
        sai_attribute_t *attr = &attr_list[i];
        if (attr->id == SAI_INGRESS_PRIORITY_GROUP_ATTR_BUFFER_PROFILE)
            p->buffer_profile = sai2ndi_buffer_profile_id(attr->value.u64);
    }

    return STD_ERR_OK;
}

/**
 * This function get a priority_group from the NPU.
 * @param ndi_port_id
 * @param ndi_priority_group_id
 * @param[out] ndi_qos_priority_group_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_priority_group_attribute(ndi_port_t ndi_port_id,
                            ndi_obj_id_t ndi_priority_group_id,
                            ndi_qos_priority_group_attribute_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list;
    sai_attribute_t  sai_attr = {0};

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attr.id = SAI_INGRESS_PRIORITY_GROUP_ATTR_BUFFER_PROFILE;
    attr_list.push_back(sai_attr);

    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            get_ingress_priority_group_attr(
                    ndi2sai_priority_group_id(ndi_priority_group_id),
                    attr_list.size(),
                    &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d priority_group 0x%016lx get failed\n",
                      ndi_port_id.npu_id, ndi_priority_group_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // convert sai result to NAS format
    _fill_ndi_qos_priority_group_struct(&attr_list[0], attr_list.size(), p);


    return STD_ERR_OK;

}

/**
 * This function gets the total number of priority_groups on a port
 * @param ndi_port_id
 * @Return standard error code
 */
uint_t ndi_qos_get_number_of_priority_groups(ndi_port_t ndi_port_id)
{
    return ndi_qos_get_priority_group_id_list(ndi_port_id, 1, NULL);
}

/**
 * This function gets the list of priority_groups of a port
 * @param ndi_port_id
 * @param count size of the priority_group_list
 * @param[out] ndi_priority_group_id_list[] to be filled with either the number of priority_groups
 *            that the port owns or the size of array itself, whichever is less.
 * @Return Number of priority_groups that the port owns.
 */
uint_t ndi_qos_get_priority_group_id_list(ndi_port_t ndi_port_id,
                                uint_t count,
                                ndi_obj_id_t *ndi_priority_group_id_list)
{
    /* get priority_group list */
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);

        return 0;
    }

    sai_attribute_t sai_attr;
    std::vector<sai_object_id_t> sai_priority_group_id_list(count);

    sai_attr.id = SAI_PORT_ATTR_PRIORITY_GROUP_LIST;
    sai_attr.value.objlist.count = count;
    sai_attr.value.objlist.list = &sai_priority_group_id_list[0];

    sai_object_id_t sai_port;
    if (ndi_sai_port_id_get(ndi_port_id.npu_id, ndi_port_id.npu_port, &sai_port) != STD_ERR_OK) {
        return 0;
    }

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->
                        get_port_attribute(sai_port,
                                    1, &sai_attr);

    if (sai_ret != SAI_STATUS_SUCCESS  &&
        sai_ret != SAI_STATUS_BUFFER_OVERFLOW) {
        return 0;
    }

    // copy out sai-returned priority_group ids to nas
    if (ndi_priority_group_id_list) {
        for (uint_t i = 0; (i< sai_attr.value.objlist.count) && (i < count); i++) {
            ndi_priority_group_id_list[i] = sai2ndi_priority_group_id(sai_attr.value.objlist.list[i]);
        }
    }

    return sai_attr.value.objlist.count;

}


static const std::unordered_map<BASE_QOS_PRIORITY_GROUP_STAT_t, sai_ingress_priority_group_stat_counter_t, std::hash<int>>
    nas2sai_priority_group_counter_type = {
        {BASE_QOS_PRIORITY_GROUP_STAT_PACKETS, SAI_INGRESS_PRIORITY_GROUP_STAT_PACKETS},
        {BASE_QOS_PRIORITY_GROUP_STAT_BYTES, SAI_INGRESS_PRIORITY_GROUP_STAT_BYTES},
        {BASE_QOS_PRIORITY_GROUP_STAT_CURRENT_OCCUPANCY_BYTES, SAI_INGRESS_PRIORITY_GROUP_STAT_CURR_OCCUPANCY_BYTES},
        {BASE_QOS_PRIORITY_GROUP_STAT_WATERMARK_BYTES, SAI_INGRESS_PRIORITY_GROUP_STAT_WATERMARK_BYTES},
    };

static void _fill_counter_stat_by_type(sai_ingress_priority_group_stat_counter_t type, uint64_t val,
        nas_qos_priority_group_stat_counter_t *stat )
{
    switch(type) {
    case SAI_INGRESS_PRIORITY_GROUP_STAT_PACKETS:
        stat->packets = val;
        break;
    case SAI_INGRESS_PRIORITY_GROUP_STAT_BYTES:
        stat->bytes = val;
        break;
    case SAI_INGRESS_PRIORITY_GROUP_STAT_CURR_OCCUPANCY_BYTES:
        stat->current_occupancy_bytes = val;
        break;
    case SAI_INGRESS_PRIORITY_GROUP_STAT_WATERMARK_BYTES:
        stat->watermark_bytes = val;
        break;
    default:
        break;
    }
}


/**
 * This function gets the priority_group statistics
 * @param ndi_port_id
 * @param ndi_priority_group_id
 * @param list of priority_group counter types to query
 * @param number of priority_group counter types specified
 * @param[out] counter stats
  * return standard error
 */
t_std_error ndi_qos_get_priority_group_stats(ndi_port_t ndi_port_id,
                                ndi_obj_id_t ndi_priority_group_id,
                                BASE_QOS_PRIORITY_GROUP_STAT_t *counter_ids,
                                uint_t number_of_counters,
                                nas_qos_priority_group_stat_counter_t *stats)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    std::vector<sai_ingress_priority_group_stat_counter_t> counter_id_list;
    std::vector<uint64_t> counters(number_of_counters);

    for (uint_t i= 0; i<number_of_counters; i++) {
        counter_id_list.push_back(nas2sai_priority_group_counter_type.at(counter_ids[i]));
    }
    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
                        get_ingress_priority_group_stats(ndi2sai_priority_group_id(ndi_priority_group_id),
                                &counter_id_list[0],
                                number_of_counters,
                                &counters[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "priority_group get stats fails: npu_id %u\n",
                ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // copy the stats out
    for (uint i= 0; i<number_of_counters; i++) {
        _fill_counter_stat_by_type(counter_id_list[i], counters[i], stats);
    }

    return STD_ERR_OK;
}

/**
 * This function clears the priority_group statistics
 * @param ndi_port_id
 * @param ndi_priority_group_id
 * @param list of priority_group counter types to clear
 * @param number of priority_group counter types specified
 * return standard error
 */
t_std_error ndi_qos_clear_priority_group_stats(ndi_port_t ndi_port_id,
                                ndi_obj_id_t ndi_priority_group_id,
                                BASE_QOS_PRIORITY_GROUP_STAT_t *counter_ids,
                                uint_t number_of_counters)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    std::vector<sai_ingress_priority_group_stat_counter_t> counter_id_list;
    std::vector<uint64_t> counters(number_of_counters);

    for (uint_t i= 0; i<number_of_counters; i++) {
        counter_id_list.push_back(nas2sai_priority_group_counter_type.at(counter_ids[i]));
    }
    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
                        clear_ingress_priority_group_stats(ndi2sai_priority_group_id(ndi_priority_group_id),
                                &counter_id_list[0],
                                number_of_counters))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "priority_group clear stats fails: npu_id %u\n",
                ndi_port_id.npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

