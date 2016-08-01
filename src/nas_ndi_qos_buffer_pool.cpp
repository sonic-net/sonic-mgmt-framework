
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
 * filename: nas_ndi_qos_buffer_pool.cpp
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
    ndi2sai_buffer_pool_attr_id_map = {
    {BASE_QOS_BUFFER_POOL_SHARED_SIZE,        SAI_BUFFER_POOL_ATTR_SHARED_SIZE},
    {BASE_QOS_BUFFER_POOL_POOL_TYPE,          SAI_BUFFER_POOL_ATTR_TYPE},
    {BASE_QOS_BUFFER_POOL_SIZE,               SAI_BUFFER_POOL_ATTR_SIZE},
    {BASE_QOS_BUFFER_POOL_THRESHOLD_MODE,     SAI_BUFFER_POOL_ATTR_TH_MODE},

};


static t_std_error ndi_qos_fill_buffer_pool_attr(nas_attr_id_t attr_id,
                        const qos_buffer_pool_struct_t *p,
                        sai_attribute_t &sai_attr)
{
    // Only the settable attributes are included
    try {
        sai_attr.id = ndi2sai_buffer_pool_attr_id_map.at(attr_id);
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "attr_id %u not supported\n", attr_id);
        return STD_ERR(QOS, CFG, 0);
    }

    if (attr_id == BASE_QOS_BUFFER_POOL_SHARED_SIZE)
        sai_attr.value.u32 = p->shared_size;
    else if (attr_id == BASE_QOS_BUFFER_POOL_POOL_TYPE)
        sai_attr.value.s32 = (p->type == BASE_QOS_BUFFER_POOL_TYPE_INGRESS?
                                SAI_BUFFER_POOL_INGRESS: SAI_BUFFER_POOL_EGRESS);
    else if (attr_id == BASE_QOS_BUFFER_POOL_SIZE)
        sai_attr.value.u32 = p->size;
    else if (attr_id == BASE_QOS_BUFFER_POOL_THRESHOLD_MODE)
        sai_attr.value.s32 = (p->threshold_mode == BASE_QOS_BUFFER_THRESHOLD_MODE_STATIC?
                                SAI_BUFFER_THRESHOLD_MODE_STATIC: SAI_BUFFER_THRESHOLD_MODE_DYNAMIC);

    return STD_ERR_OK;
}


static t_std_error ndi_qos_fill_buffer_pool_attr_list(const nas_attr_id_t *nas_attr_list,
                                    uint_t num_attr,
                                    const qos_buffer_pool_struct_t *p,
                                    std::vector<sai_attribute_t> &attr_list)
{
    sai_attribute_t sai_attr = {0};
    t_std_error      rc = STD_ERR_OK;

    for (uint_t i = 0; i < num_attr; i++) {
        if ((rc = ndi_qos_fill_buffer_pool_attr(nas_attr_list[i], p, sai_attr)) != STD_ERR_OK)
            return rc;

        attr_list.push_back(sai_attr);

    }

    return STD_ERR_OK;
}



/**
 * This function creates a buffer_pool profile in the NPU.
 * @param npu id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param p buffer_pool structure to be modified
 * @param[out] ndi_buffer_pool_id
 * @return standard error
 */
t_std_error ndi_qos_create_buffer_pool(npu_id_t npu_id,
                                const nas_attr_id_t *nas_attr_list,
                                uint_t num_attr,
                                const qos_buffer_pool_struct_t *p,
                                ndi_obj_id_t *ndi_buffer_pool_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    std::vector<sai_attribute_t>  attr_list;

    if (ndi_qos_fill_buffer_pool_attr_list(nas_attr_list, num_attr, p, attr_list)
            != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    sai_object_id_t sai_qos_buffer_pool_id;
    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            create_buffer_pool(&sai_qos_buffer_pool_id,
                                attr_list.size(),
                                &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d buffer_pool creation failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }
    *ndi_buffer_pool_id = sai2ndi_buffer_pool_id(sai_qos_buffer_pool_id);

    return STD_ERR_OK;
}

 /**
  * This function sets the buffer_pool profile attributes in the NPU.
  * @param npu id
  * @param ndi_buffer_pool_id
  * @param attr_id based on the CPS API attribute enumeration values
  * @param p buffer_pool structure to be modified
  * @return standard error
  */
t_std_error ndi_qos_set_buffer_pool_attr(npu_id_t npu_id, ndi_obj_id_t ndi_buffer_pool_id,
                                  BASE_QOS_BUFFER_POOL_t attr_id, const qos_buffer_pool_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attribute_t sai_attr;
    if (ndi_qos_fill_buffer_pool_attr(attr_id, p, sai_attr) != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            set_buffer_pool_attr(
                    ndi2sai_buffer_pool_id(ndi_buffer_pool_id),
                    &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d buffer_pool profile set failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

/**
 * This function deletes a buffer_pool profile in the NPU.
 * @param npu_id npu id
 * @param ndi_buffer_pool_id
 * @return standard error
 */
t_std_error ndi_qos_delete_buffer_pool(npu_id_t npu_id, ndi_obj_id_t ndi_buffer_pool_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            remove_buffer_pool(ndi2sai_buffer_pool_id(ndi_buffer_pool_id)))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d buffer_pool profile deletion failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

static t_std_error _fill_ndi_qos_buffer_pool_struct(sai_attribute_t *attr_list,
                        uint_t num_attr, qos_buffer_pool_struct_t *p)
{

    for (uint_t i = 0 ; i< num_attr; i++ ) {
        sai_attribute_t *attr = &attr_list[i];
        if (attr->id == SAI_BUFFER_POOL_ATTR_SHARED_SIZE)
            p->shared_size = attr->value.u32;
        else if (attr->id == SAI_BUFFER_POOL_ATTR_TYPE)
            p->type = (attr->value.s32 == SAI_BUFFER_POOL_INGRESS?
                          BASE_QOS_BUFFER_POOL_TYPE_INGRESS: BASE_QOS_BUFFER_POOL_TYPE_EGRESS);
        else if (attr->id == SAI_BUFFER_POOL_ATTR_SIZE)
            p->size = attr->value.u32;
        else if (attr->id == SAI_BUFFER_POOL_ATTR_TH_MODE)
            p->threshold_mode = (attr->value.s32 == SAI_BUFFER_THRESHOLD_MODE_STATIC?
                                     BASE_QOS_BUFFER_THRESHOLD_MODE_STATIC: BASE_QOS_BUFFER_THRESHOLD_MODE_DYNAMIC);
    }

    return STD_ERR_OK;
}


/**
 * This function get a buffer_pool profile from the NPU.
 * @param npu id
 * @param ndi_buffer_pool_id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param[out] qos_buffer_pool_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_buffer_pool(npu_id_t npu_id,
                            ndi_obj_id_t ndi_buffer_pool_id,
                            const nas_attr_id_t *nas_attr_list,
                            uint_t num_attr,
                            qos_buffer_pool_struct_t *p)

{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list;
    sai_attribute_t sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    try {
        for (uint_t i = 0; i < num_attr; i++) {
            sai_attr.id = ndi2sai_buffer_pool_attr_id_map.at(nas_attr_list[i]);
            attr_list.push_back(sai_attr);
        }
    }
    catch(...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "Unexpected error.\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
            get_buffer_pool_attr(
                    ndi2sai_buffer_pool_id(ndi_buffer_pool_id),
                    num_attr,
                    &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d buffer_pool get failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // convert sai result to NAS format
    _fill_ndi_qos_buffer_pool_struct(&attr_list[0], num_attr, p);


    return STD_ERR_OK;

}


static const std::unordered_map<BASE_QOS_BUFFER_POOL_STAT_t, sai_buffer_pool_stat_counter_t,
                                    std::hash<int>>
    nas2sai_buffer_pool_counter_type = {
        {BASE_QOS_BUFFER_POOL_STAT_CURRENT_OCCUPANCY_BYTES, SAI_BUFFER_POOL_STAT_CURR_OCCUPANCY_BYTES},
        {BASE_QOS_BUFFER_POOL_STAT_WATERMARK_BYTES, SAI_BUFFER_POOL_STAT_WATERMARK_BYTES},
    };

static void _fill_counter_stat_by_type(sai_buffer_pool_stat_counter_t type, uint64_t val,
        nas_qos_buffer_pool_stat_counter_t *stat )
{
    switch(type) {

    case SAI_BUFFER_POOL_STAT_CURR_OCCUPANCY_BYTES:
        stat->current_occupancy_bytes = val;
        break;
    case SAI_BUFFER_POOL_STAT_WATERMARK_BYTES:
        stat->watermark_bytes = val;
        break;
    default:
        break;
    }
}


/**
 * This function gets the buffer_pool statistics
 * @param npu_id
 * @param ndi_buffer_pool_id
 * @param list of buffer_pool counter types to query
 * @param number of buffer_pool counter types specified
 * @param[out] counter stats
  * return standard error
 */
t_std_error ndi_qos_get_buffer_pool_stats(npu_id_t npu_id,
                                ndi_obj_id_t ndi_buffer_pool_id,
                                BASE_QOS_BUFFER_POOL_STAT_t *counter_ids,
                                uint_t number_of_counters,
                                nas_qos_buffer_pool_stat_counter_t *stats)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    std::vector<sai_buffer_pool_stat_counter_t> counter_id_list;
    std::vector<uint64_t> counters(number_of_counters);

    for (uint_t i= 0; i<number_of_counters; i++) {
        counter_id_list.push_back(nas2sai_buffer_pool_counter_type.at(counter_ids[i]));
    }
    if ((sai_ret = ndi_sai_qos_buffer_api(ndi_db_ptr)->
                        get_buffer_pool_stats(ndi2sai_buffer_pool_id(ndi_buffer_pool_id),
                                &counter_id_list[0],
                                number_of_counters,
                                &counters[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                "buffer_pool get stats fails: buffer pool id %u\n",
                ndi_buffer_pool_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // copy the stats out
    for (uint i= 0; i<number_of_counters; i++) {
        _fill_counter_stat_by_type(counter_id_list[i], counters[i], stats);
    }

    return STD_ERR_OK;
}

