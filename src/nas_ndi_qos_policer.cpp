
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
 * filename: nas_ndi_qos_policer.cpp
 */


#include "std_assert.h"
#include "nas_ndi_int.h"
#include "nas_base_utils.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_qos.h"
#include "nas_ndi_qos_utl.h"
#include <vector>
#include <unordered_map>

static t_std_error ndi_qos_utl_fill_policer_attr (sai_attribute_t *sai_attr_p,
                                         BASE_QOS_METER_t attr_id,
                                         const qos_policer_struct_t * p);

static t_std_error ndi_qos_fill_policer_attr_list(const nas_attr_id_t *nas_attr_list,
                                    uint_t num_attr,
                                    const qos_policer_struct_t *p,
                                    std::vector<sai_attribute_t> &attr_list);

static void ndi_qos_utl_fill_policer_info(const std::vector<sai_attribute_t> attr_list,
                                        qos_policer_struct_t *p);


typedef void (*fill_sai_policer_fn) (sai_attribute_t* s,
                                     const qos_policer_struct_t* p);
static void _fill_sai_meter_type(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_mode(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_color_source(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_green_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_yellow_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_red_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_commited_burst(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_commited_rate(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_peak_burst(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_peak_rate(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);
static void _fill_sai_meter_stat_list(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p);

static const
    std::unordered_map<BASE_QOS_METER_t, fill_sai_policer_fn, std::hash<int>>
    _fill_sai_policer_fn_map = {
        {BASE_QOS_METER_TYPE,                    _fill_sai_meter_type},
        {BASE_QOS_METER_MODE,                    _fill_sai_meter_mode},
        {BASE_QOS_METER_COLOR_SOURCE,            _fill_sai_meter_color_source},
        {BASE_QOS_METER_GREEN_PACKET_ACTION,     _fill_sai_meter_green_packet_action},
        {BASE_QOS_METER_YELLOW_PACKET_ACTION,    _fill_sai_meter_yellow_packet_action},
        {BASE_QOS_METER_RED_PACKET_ACTION,       _fill_sai_meter_red_packet_action},
        {BASE_QOS_METER_COMMITTED_BURST,         _fill_sai_meter_commited_burst},
        {BASE_QOS_METER_COMMITTED_RATE,          _fill_sai_meter_commited_rate},
        {BASE_QOS_METER_PEAK_BURST,              _fill_sai_meter_peak_burst},
        {BASE_QOS_METER_PEAK_RATE,               _fill_sai_meter_peak_rate},
        {BASE_QOS_METER_STAT_LIST,               _fill_sai_meter_stat_list},
    };


static const
    std::unordered_map<BASE_QOS_METER_t, sai_policer_attr_t, std::hash<int>>
        _nas2sai_policer_attr_id_map = {
        {BASE_QOS_METER_TYPE,                    SAI_POLICER_ATTR_METER_TYPE},
        {BASE_QOS_METER_MODE,                    SAI_POLICER_ATTR_MODE},
        {BASE_QOS_METER_COLOR_SOURCE,            SAI_POLICER_ATTR_COLOR_SOURCE},
        {BASE_QOS_METER_GREEN_PACKET_ACTION,     SAI_POLICER_ATTR_GREEN_PACKET_ACTION},
        {BASE_QOS_METER_YELLOW_PACKET_ACTION,    SAI_POLICER_ATTR_YELLOW_PACKET_ACTION},
        {BASE_QOS_METER_RED_PACKET_ACTION,       SAI_POLICER_ATTR_RED_PACKET_ACTION},
        {BASE_QOS_METER_COMMITTED_BURST,         SAI_POLICER_ATTR_CBS},
        {BASE_QOS_METER_COMMITTED_RATE,          SAI_POLICER_ATTR_CIR},
        {BASE_QOS_METER_PEAK_BURST,              SAI_POLICER_ATTR_PBS},
        {BASE_QOS_METER_PEAK_RATE,               SAI_POLICER_ATTR_PIR},
        {BASE_QOS_METER_STAT_LIST,               SAI_POLICER_ATTR_ENABLE_COUNTER_LIST},
       };

static sai_policer_attr_t   ndi_qos_utl_ndi2sai_policer_attr_id (
                                    BASE_QOS_METER_t ndi_policer_attr_id)
{
    return _nas2sai_policer_attr_id_map.at(ndi_policer_attr_id);
}

static void _fill_sai_meter_type(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.s32 = (p->meter_type == BASE_QOS_METER_TYPE_PACKET?
            SAI_METER_TYPE_PACKETS: SAI_METER_TYPE_BYTES);
}

static void _fill_sai_meter_mode(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    switch (p->meter_mode) {
    case BASE_QOS_METER_MODE_SR_TCM:
        sai_attr_p->value.s32 = SAI_POLICER_MODE_Sr_TCM;
        break;
    case BASE_QOS_METER_MODE_TR_TCM:
        sai_attr_p->value.s32 = SAI_POLICER_MODE_Tr_TCM;
        break;
        /** @todo
    case BASE_QOS_METER_MODE_SR_TWO_COLOR:
        sai_attr_p->value.s32 = SAI_POLICER_MODE_Sr_TWO_COLOR;
        break;
        */
    case BASE_QOS_METER_MODE_STORM_CONTROL:
        sai_attr_p->value.s32 = SAI_POLICER_MODE_STORM_CONTROL;
        break;
    default:
        break;
    }
}

static void _fill_sai_meter_color_source(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.s32 = (p->color_source == BASE_QOS_METER_COLOR_SOURCE_BLIND?
            SAI_POLICER_COLOR_SOURCE_BLIND: SAI_POLICER_COLOR_SOURCE_AWARE);

}

static void _fill_sai_meter_green_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.s32 = (p->green_packet_action == BASE_QOS_POLICER_ACTION_FORWARD?
            SAI_PACKET_ACTION_FORWARD: SAI_PACKET_ACTION_DROP);

}

static void _fill_sai_meter_yellow_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.s32 = (p->yellow_packet_action == BASE_QOS_POLICER_ACTION_FORWARD?
            SAI_PACKET_ACTION_FORWARD: SAI_PACKET_ACTION_DROP);

}

static void _fill_sai_meter_red_packet_action(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.s32 = (p->red_packet_action == BASE_QOS_POLICER_ACTION_FORWARD?
            SAI_PACKET_ACTION_FORWARD: SAI_PACKET_ACTION_DROP);

}

static void _fill_sai_meter_commited_burst(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.u64 = p->committed_burst ;

}

static void _fill_sai_meter_commited_rate(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.u64 = p->committed_rate;

}

static void _fill_sai_meter_peak_burst(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.u64 = p->peak_burst ;

}

static void _fill_sai_meter_peak_rate(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    sai_attr_p->value.u64 = p->peak_rate;

}

static const std::unordered_map<BASE_QOS_POLICER_STAT_TYPE_t, sai_policer_stat_counter_t, std::hash<int>>
    nas2ndi_policer_stat_type =
    {
      {BASE_QOS_POLICER_STAT_TYPE_PACKETS,          SAI_POLICER_STAT_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_BYTES,            SAI_POLICER_STAT_ATTR_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_GREEN_PACKETS,    SAI_POLICER_STAT_GREEN_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_GREEN_BYTES,      SAI_POLICER_STAT_GREEN_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_YELLOW_PACKETS,   SAI_POLICER_STAT_YELLOW_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_YELLOW_BYTES,     SAI_POLICER_STAT_YELLOW_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_RED_PACKETS,      SAI_POLICER_STAT_RED_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_RED_BYTES,        SAI_POLICER_STAT_RED_BYTES},
    };

static void _fill_sai_meter_stat_list(sai_attribute_t *sai_attr_p,
                                 const qos_policer_struct_t* p)
{
    try {
        sai_attr_p->value.s32list.count = p->stat_list_count;
        for (uint_t i= 0; i < p->stat_list_count; i++) {
            BASE_QOS_POLICER_STAT_TYPE_t type = p->stat_list[i];
            sai_attr_p->value.s32list.list[i] = nas2ndi_policer_stat_type.at(type);
        }
    }
    catch (std::out_of_range e) {
        NDI_LOG_ERROR(1, "NDI-QOS",
                      "%s: stat type out of range: %s",
                      __FUNCTION__, e.what());
    }
}


//////////////////////////////////////////////////////////////////////////////////////
// Map NAS-NDI QoS Policer values to SAI values and populate the SAI attribute
/////////////////////////////////////////////////////////////////////////////////////

static t_std_error ndi_qos_utl_fill_policer_attr (sai_attribute_t *sai_attr_p,
                                         BASE_QOS_METER_t attr_id,
                                         const qos_policer_struct_t * p)
{
    try {

        sai_attr_p->id = ndi_qos_utl_ndi2sai_policer_attr_id(attr_id);

        // value
        auto fn_set_policer = _fill_sai_policer_fn_map.at (attr_id);
        fn_set_policer (sai_attr_p, p);

    } catch (std::out_of_range e) {
        NDI_LOG_ERROR(1, "NDI-QOS",
                      "%s: Invalid Policer attr_id %d or value parameter %s",
                      __FUNCTION__, attr_id, e.what());
        return STD_ERR(QOS, PARAM, 0);
    }

    return STD_ERR_OK;
}



static const std::unordered_map<sai_policer_stat_counter_t, BASE_QOS_POLICER_STAT_TYPE_t, std::hash<int>>
    sai2nas_stat_type = {
        {SAI_POLICER_STAT_PACKETS, BASE_QOS_POLICER_STAT_TYPE_PACKETS},
        {SAI_POLICER_STAT_ATTR_BYTES, BASE_QOS_POLICER_STAT_TYPE_BYTES},
        {SAI_POLICER_STAT_GREEN_PACKETS, BASE_QOS_POLICER_STAT_TYPE_GREEN_PACKETS},
        {SAI_POLICER_STAT_GREEN_BYTES, BASE_QOS_POLICER_STAT_TYPE_GREEN_BYTES},
        {SAI_POLICER_STAT_YELLOW_PACKETS, BASE_QOS_POLICER_STAT_TYPE_YELLOW_PACKETS},
        {SAI_POLICER_STAT_YELLOW_BYTES, BASE_QOS_POLICER_STAT_TYPE_YELLOW_BYTES},
        {SAI_POLICER_STAT_RED_PACKETS, BASE_QOS_POLICER_STAT_TYPE_RED_PACKETS},
        {SAI_POLICER_STAT_RED_BYTES, BASE_QOS_POLICER_STAT_TYPE_RED_BYTES}
    };


static void ndi_qos_utl_fill_policer_info(const std::vector<sai_attribute_t> attr_list,
                                        qos_policer_struct_t *p)
{
    for (auto attr: attr_list) {
        switch (attr.id) {
        case SAI_POLICER_ATTR_METER_TYPE:
            p->meter_type = (attr.value.s32 == SAI_METER_TYPE_BYTES?
                            BASE_QOS_METER_TYPE_BYTE: BASE_QOS_METER_TYPE_PACKET);
            break;

        case SAI_POLICER_ATTR_MODE:
            p->meter_mode = (attr.value.s32 == SAI_POLICER_MODE_STORM_CONTROL? BASE_QOS_METER_MODE_STORM_CONTROL:
                             (attr.value.s32 == SAI_POLICER_MODE_Tr_TCM? BASE_QOS_METER_MODE_TR_TCM:
                                     BASE_QOS_METER_MODE_SR_TCM));
            break;

        case SAI_POLICER_ATTR_COLOR_SOURCE:
            p->color_source = (attr.value.s32 == SAI_POLICER_COLOR_SOURCE_AWARE?
                    BASE_QOS_METER_COLOR_SOURCE_AWARE: BASE_QOS_METER_COLOR_SOURCE_BLIND);
            break;

        case SAI_POLICER_ATTR_GREEN_PACKET_ACTION:
            p->green_packet_action = (attr.value.s32 == SAI_PACKET_ACTION_FORWARD?
                    BASE_QOS_POLICER_ACTION_FORWARD: BASE_QOS_POLICER_ACTION_DROP);
            break;

        case SAI_POLICER_ATTR_YELLOW_PACKET_ACTION:
            p->yellow_packet_action = (attr.value.s32 == SAI_PACKET_ACTION_FORWARD?
                    BASE_QOS_POLICER_ACTION_FORWARD: BASE_QOS_POLICER_ACTION_DROP);
            break;

        case SAI_POLICER_ATTR_RED_PACKET_ACTION:
            p->red_packet_action = (attr.value.s32 == SAI_PACKET_ACTION_FORWARD?
                    BASE_QOS_POLICER_ACTION_FORWARD: BASE_QOS_POLICER_ACTION_DROP);
            break;

        case SAI_POLICER_ATTR_CBS:
            p->committed_burst = attr.value.u64;
            break;

        case SAI_POLICER_ATTR_CIR:
            p->committed_rate = attr.value.u64;
            break;

        case SAI_POLICER_ATTR_PBS:
            p->peak_burst = attr.value.u64;
            break;

        case SAI_POLICER_ATTR_PIR:
            p->peak_rate = attr.value.u64;
            break;

        case SAI_POLICER_ATTR_ENABLE_COUNTER_LIST:
            if (p->stat_list_count > attr.value.s32list.count)
                p->stat_list_count = attr.value.s32list.count;

            for (uint i= 0; i< p->stat_list_count; i++) {
                try {
                    p->stat_list[i] = sai2nas_stat_type.at((sai_policer_stat_counter_t)attr.value.s32list.list[i]);
                }
                catch (...) {
                    NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS", "statistics types unmapped to NAS\n");
                }
            }
            break;

        default:
            break;
        }
    }
}

static t_std_error ndi_qos_fill_policer_attr_list(const nas_attr_id_t *nas_attr_list,
                                    uint_t num_attr,
                                    const qos_policer_struct_t *p,
                                    std::vector<sai_attribute_t> &attr_list)
{
    sai_attribute_t sai_attr = {0};
    t_std_error      rc = STD_ERR_OK;

    for (uint_t i = 0; i < num_attr; i++) {
        if (nas_attr_list[i] == BASE_QOS_METER_SWITCH_ID ||
            nas_attr_list[i] == BASE_QOS_METER_ID ||
            nas_attr_list[i] == BASE_QOS_METER_NPU_ID_LIST)
            continue; // these attributes are not interpreted at ndi level

        rc = ndi_qos_utl_fill_policer_attr(&sai_attr, (BASE_QOS_METER_t)nas_attr_list[i], p);
        if (rc != STD_ERR_OK)
            return rc;

        attr_list.push_back(sai_attr);

    }

    return rc;
}


/******************* QoS Policer ************************/


/**
 * This function sets the policer attributes in the NPU.
 * @param  npu id
 * @param  ndi_policer_id
 * @param attr_id based on the CPS API attribute enumeration values
 * @param p policer structure to be modified
 * @return standard error
 */
t_std_error ndi_qos_set_policer_attr(npu_id_t npu_id,
                                     ndi_obj_id_t ndi_policer_id,
                                     BASE_QOS_METER_t attr_id,
                                     const qos_policer_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);

        return STD_ERR(QOS, CFG, 0);

    }

    sai_attribute_t sai_attr = {0};
    std::vector<int32_t> list;
    if (attr_id == BASE_QOS_METER_STAT_LIST) {
        list.resize(p->stat_list_count);
        sai_attr.value.s32list.count = p->stat_list_count;
        sai_attr.value.s32list.list = &list[0];
    }
    if (ndi_qos_utl_fill_policer_attr(&sai_attr, attr_id, p) != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    /* send to SAI */
    sai_object_id_t sai_policer_id = ndi2sai_policer_id(ndi_policer_id);
    if ((sai_ret = ndi_sai_qos_policer_api(ndi_db_ptr)->
                    set_policer_attribute(sai_policer_id, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;
}


/**
 * This function creates a policer in the NPU.
 * @param npu_id npu id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param p policer structure to be modified
 * @param[out] ndi_policer_id
 * @return standard error
 */
t_std_error ndi_qos_create_policer(npu_id_t npu_id,
                                   const nas_attr_id_t *nas_attr_list,
                                   uint_t num_attr,
                                   const qos_policer_struct_t *p,
                                   ndi_obj_id_t *ndi_policer_id)
{


    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    std::vector<sai_attribute_t>  sai_policer_attr_list;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attribute_t sai_attr = {0};
    std::vector<int32_t> list;
    for (uint_t i=0; i< num_attr; i++ ) {
        if (nas_attr_list[i] == BASE_QOS_METER_SWITCH_ID ||
            nas_attr_list[i] == BASE_QOS_METER_ID ||
            nas_attr_list[i] == BASE_QOS_METER_NPU_ID_LIST)
            continue; // these attributes are not interpreted at ndi level

        if (nas_attr_list[i] == BASE_QOS_METER_STAT_LIST) {
            list.resize(p->stat_list_count);
            sai_attr.value.s32list.count = p->stat_list_count;
            sai_attr.value.s32list.list = &list[0];
        }

        if (ndi_qos_utl_fill_policer_attr(&sai_attr, (BASE_QOS_METER_t)nas_attr_list[i], p)
                != STD_ERR_OK)
            return STD_ERR(QOS, CFG, 0);

        sai_policer_attr_list.push_back(sai_attr);
    }

    sai_object_id_t sai_policer_id;
    if ((sai_ret = ndi_sai_qos_policer_api(ndi_db_ptr)->
                    create_policer(&sai_policer_id,
                                sai_policer_attr_list.size(),
                                &sai_policer_attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }
    *ndi_policer_id = sai2ndi_policer_id(sai_policer_id);

    return ret_code;
}

/**
 * This function deletes a policer in the NPU.
 * @param npu_id npu id
 * @param ndi_policer_id
 * @return standard error
 */
t_std_error ndi_qos_delete_policer(npu_id_t npu_id, ndi_obj_id_t ndi_policer_id)
{
    t_std_error ret_code = STD_ERR_OK;

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t sai_policer_id = ndi2sai_policer_id(ndi_policer_id);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_qos_policer_api(ndi_db_ptr)->
                    remove_policer(sai_policer_id))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return ret_code;
}

/**
 * This function get a policer from the NPU.
 * @param npu_id npu id
 * @param ndi_policer_id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param[out] qos_policer_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_policer(npu_id_t npu_id,
                                ndi_obj_id_t ndi_policer_id,
                                const nas_attr_id_t *nas_attr_list,
                                uint_t num_attr,
                                qos_policer_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list;
    qos_policer_struct_t dummy;

    // Fill the attribute flags
    ndi_qos_fill_policer_attr_list(nas_attr_list, num_attr, &dummy, attr_list);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_qos_policer_api(ndi_db_ptr)->
                    get_policer_attribute(ndi2sai_policer_id(ndi_policer_id),
                            attr_list.size(), &attr_list[0]))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    //convert attr_list[] to qos_policer_struct_t
    ndi_qos_utl_fill_policer_info(attr_list, p);

    return STD_ERR_OK;
}

static const std::unordered_map<BASE_QOS_POLICER_STAT_TYPE_t, sai_policer_stat_counter_t, std::hash<int>>
    ndi2sai_policer_stat_type = {
      {BASE_QOS_POLICER_STAT_TYPE_PACKETS,         SAI_POLICER_STAT_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_BYTES,         SAI_POLICER_STAT_ATTR_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_GREEN_PACKETS, SAI_POLICER_STAT_GREEN_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_GREEN_BYTES,     SAI_POLICER_STAT_GREEN_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_YELLOW_PACKETS, SAI_POLICER_STAT_YELLOW_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_YELLOW_BYTES, SAI_POLICER_STAT_YELLOW_BYTES},
      {BASE_QOS_POLICER_STAT_TYPE_RED_PACKETS,     SAI_POLICER_STAT_RED_PACKETS},
      {BASE_QOS_POLICER_STAT_TYPE_RED_BYTES,     SAI_POLICER_STAT_RED_BYTES}
    };

/**
 * This function get a policer from the NPU.
 * @param npu_id npu id
 * @param ndi_policer_id
 * @param stat_list_count number of statistics types to retrieve
 * @param *stat_list list of statistics types to retrieve
 * @param[out] *counters statistics
 * @return standard error
 */
t_std_error ndi_qos_get_policer_stat(npu_id_t npu_id,
                                ndi_obj_id_t ndi_policer_id,
                                uint_t     stat_list_count,
                                const BASE_QOS_POLICER_STAT_TYPE_t * stat_list,
                                policer_stats_struct_t * stats)
{
    if (stat_list_count > BASE_QOS_POLICER_STAT_TYPE_MAX) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "sai_id: %d, ndi_policer_id %u, too many statistics types!\n",
                      npu_id, ndi_policer_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_policer_stat_counter_t> counter_ids(stat_list_count);
    std::vector<uint64_t> counters(stat_list_count);

    try {
        for (uint_t i = 0; i< stat_list_count; i++) {
            counter_ids[i] = ndi2sai_policer_stat_type.at(stat_list[i]);
        }
    }
    catch (...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "Unknown statistics types!\n");
        return STD_ERR(QOS, CFG, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_sai_qos_policer_api(ndi_db_ptr)->
                    get_policer_statistics(ndi2sai_policer_id(ndi_policer_id),
                            &counter_ids[0], stat_list_count,
                            &counters[0]))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(QOS, CFG, sai_ret);
    }

    /* copy out SAI-returned value  */
    for (uint_t i = 0; i<stat_list_count; i++) {
        switch (counter_ids[i]) {
        case SAI_POLICER_STAT_PACKETS:
            stats->packets = counters[i];
            break;
        case SAI_POLICER_STAT_ATTR_BYTES:
            stats->bytes = counters[i];
            break;
        case SAI_POLICER_STAT_GREEN_PACKETS:
            stats->green_packets = counters[i];
            break;
        case SAI_POLICER_STAT_GREEN_BYTES:
            stats->green_bytes = counters[i];
            break;
        case SAI_POLICER_STAT_YELLOW_PACKETS:
            stats->yellow_packets = counters[i];
            break;
        case SAI_POLICER_STAT_YELLOW_BYTES:
            stats->yellow_bytes = counters[i];
            break;
        case SAI_POLICER_STAT_RED_PACKETS:
            stats->red_packets = counters[i];
            break;
        case SAI_POLICER_STAT_RED_BYTES:
            stats->red_bytes = counters[i];
            break;
        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                          "Unknown statistics type: %u!\n", counter_ids[i]);
            break;
        }
    }
    return STD_ERR_OK;
}


