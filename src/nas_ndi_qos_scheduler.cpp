
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
 * filename: nas_ndi_qos_scheduler.cpp
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


static t_std_error ndi_qos_fill_scheduler_attr(nas_attr_id_t attr_id,
                        const qos_scheduler_struct_t *p,
                        sai_attribute_t &sai_attr)
{
    // Only the following attributes are settable
    if (attr_id == BASE_QOS_SCHEDULER_PROFILE_ALGORITHM) {
        sai_attr.id = SAI_SCHEDULER_ATTR_SCHEDULING_ALGORITHM;
        sai_attr.value.s32 = (p->algorithm == BASE_QOS_SCHEDULING_TYPE_SP?
                                SAI_SCHEDULING_STRICT:
                                (p->algorithm == BASE_QOS_SCHEDULING_TYPE_WRR?
                                    SAI_SCHEDULING_WRR:    SAI_SCHEDULING_DWRR));
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_WEIGHT) {
        sai_attr.id = SAI_SCHEDULER_ATTR_SCHEDULING_WEIGHT;
        sai_attr.value.u8 = p->weight;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_METER_TYPE) {
        sai_attr.id = SAI_SCHEDULER_ATTR_SHAPER_TYPE;
        sai_attr.value.s32 = (p->meter_type == BASE_QOS_METER_TYPE_PACKET?
                SAI_METER_TYPE_PACKETS: SAI_METER_TYPE_BYTES);
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_MIN_RATE) {
        sai_attr.id = SAI_SCHEDULER_ATTR_MIN_BANDWIDTH_RATE;
        sai_attr.value.u64 = p->min_rate;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_MIN_BURST) {
        sai_attr.id = SAI_SCHEDULER_ATTR_MIN_BANDWIDTH_BURST_RATE;
        sai_attr.value.u64 = p->min_burst;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_MAX_RATE) {
        sai_attr.id = SAI_SCHEDULER_ATTR_MAX_BANDWIDTH_RATE;
        sai_attr.value.u64 = p->max_rate;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_PROFILE_MAX_BURST) {
        sai_attr.id = SAI_SCHEDULER_ATTR_MAX_BANDWIDTH_BURST_RATE;
        sai_attr.value.u64 = p->max_burst;
    }

    return STD_ERR_OK;
}



static void _fill_ndi_qos_scheduler_info(const std::vector<sai_attribute_t> attr_list,
                                        qos_scheduler_struct_t *p)
{
    for (auto attr: attr_list) {
        switch (attr.id) {
        case SAI_SCHEDULER_ATTR_SCHEDULING_ALGORITHM:
            p->algorithm = (attr.value.s32 == SAI_SCHEDULING_STRICT?
                            BASE_QOS_SCHEDULING_TYPE_SP:
                             (attr.value.s32 == SAI_SCHEDULING_WRR?
                                 BASE_QOS_SCHEDULING_TYPE_WRR: BASE_QOS_SCHEDULING_TYPE_WDRR));
            break;
        case SAI_SCHEDULER_ATTR_SCHEDULING_WEIGHT:
            p->weight = attr.value.u8;
            break;
        case SAI_SCHEDULER_ATTR_SHAPER_TYPE:
            p->meter_type = (attr.value.s32 == SAI_METER_TYPE_BYTES?
                            BASE_QOS_METER_TYPE_BYTE: BASE_QOS_METER_TYPE_PACKET);
            break;
        case SAI_SCHEDULER_ATTR_MIN_BANDWIDTH_RATE:
            p->min_rate = attr.value.u64;
            break;
        case SAI_SCHEDULER_ATTR_MIN_BANDWIDTH_BURST_RATE:
            p->min_burst = attr.value.u64;
            break;
        case SAI_SCHEDULER_ATTR_MAX_BANDWIDTH_RATE:
            p->max_rate = attr.value.u64;
            break;
        case SAI_SCHEDULER_ATTR_MAX_BANDWIDTH_BURST_RATE:
            p->max_burst = attr.value.u64;
            break;
        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "unknown SAI attribute id: %u", attr.id);
            break;
        }
    }
}


/**
 * This function creates a Scheduler profile ID in the NPU.
 * @param npu id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param p Scheduler structure to be modified
 * @param[out] ndi_scheduler_id
 * @return standard error
 */
t_std_error ndi_qos_create_scheduler_profile(npu_id_t npu_id,
                                const nas_attr_id_t *nas_attr_list,
                                uint_t num_attr,
                                const qos_scheduler_struct_t *p,
                                ndi_obj_id_t *ndi_scheduler_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    std::vector<sai_attribute_t>  sai_scheduler_attr_list;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attribute_t sai_attr = {0};
    std::vector<int32_t> list;
    for (uint_t i=0; i< num_attr; i++ ) {
        if (nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_SWITCH_ID ||
            nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_ID ||
            nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_NPU_ID_LIST)
            continue; // these attributes are not interpreted at ndi level

        if (ndi_qos_fill_scheduler_attr(nas_attr_list[i], p, sai_attr)
                != STD_ERR_OK)
            return STD_ERR(QOS, CFG, 0);

        sai_scheduler_attr_list.push_back(sai_attr);
    }

    sai_object_id_t sai_scheduler_id;
    if ((sai_ret = ndi_sai_qos_scheduler_api(ndi_db_ptr)->
                    create_scheduler_profile(&sai_scheduler_id,
                                sai_scheduler_attr_list.size(),
                                &sai_scheduler_attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }
    *ndi_scheduler_id = sai2ndi_scheduler_profile_id(sai_scheduler_id);

    return STD_ERR_OK;
}

 /**
  * This function sets the scheduler profile attributes in the NPU.
  * @param npu id
  * @param ndi_scheduler_id
  * @param attr_id based on the CPS API attribute enumeration values
  * @param p scheduler structure to be modified
  * @return standard error
  */
t_std_error ndi_qos_set_scheduler_profile_attr(npu_id_t npu_id,
                                    ndi_obj_id_t ndi_scheduler_id,
                                    BASE_QOS_SCHEDULER_PROFILE_t attr_id,
                                    const qos_scheduler_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);

        return STD_ERR(QOS, CFG, 0);

    }

    sai_attribute_t sai_attr = {0};
    if (ndi_qos_fill_scheduler_attr(attr_id, p, sai_attr) != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    /* send to SAI */
    sai_object_id_t sai_scheduler_id = ndi2sai_scheduler_profile_id(ndi_scheduler_id);
    if ((sai_ret = ndi_sai_qos_scheduler_api(ndi_db_ptr)->
                    set_scheduler_attribute(sai_scheduler_id, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

/**
 * This function deletes a scheduler profile in the NPU.
 * @param npu_id npu id
 * @param ndi_scheduler_id
 * @return standard error
 */
t_std_error ndi_qos_delete_scheduler_profile(npu_id_t npu_id, ndi_obj_id_t ndi_scheduler_id)
{

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t sai_scheduler_id = ndi2sai_scheduler_profile_id(ndi_scheduler_id);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);

        return STD_ERR(QOS, CFG, 0);

    }

    if ((sai_ret = ndi_sai_qos_scheduler_api(ndi_db_ptr)->
                    remove_scheduler_profile(sai_scheduler_id))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

/**
 * This function get a scheduler profile from the NPU.
 * @param npu id
 * @param ndi_scheduler_id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param[out] qos_scheduler_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_scheduler_profile(npu_id_t npu_id,
                            ndi_obj_id_t ndi_scheduler_id,
                            const nas_attr_id_t *nas_attr_list,
                            uint_t num_attr,
                            qos_scheduler_struct_t *p)

{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list;
    sai_attribute_t sai_attr;

    // Fill the attribute flags
    for (uint_t i=0; i< num_attr; i++ ) {
        if (nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_SWITCH_ID ||
            nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_ID ||
            nas_attr_list[i] == BASE_QOS_SCHEDULER_PROFILE_NPU_ID_LIST)
            continue; // these attributes are not interpreted at ndi level

        if (ndi_qos_fill_scheduler_attr(nas_attr_list[i], p, sai_attr)
                != STD_ERR_OK)
            return STD_ERR(QOS, CFG, 0);

        attr_list.push_back(sai_attr);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_qos_scheduler_api(ndi_db_ptr)->
                    get_scheduler_attribute(
                            ndi2sai_scheduler_profile_id(ndi_scheduler_id),
                            attr_list.size(), &attr_list[0]))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    //convert attr_list[] to qos_scheduler_struct_t
    _fill_ndi_qos_scheduler_info(attr_list, p);

    return STD_ERR_OK;

}


