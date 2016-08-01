
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
 * filename: nas_ndi_qos_scheduler_group.cpp
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

#define MAX_SCHEDULER_LEVEL 4

static t_std_error ndi_qos_fill_sg_attr(nas_attr_id_t attr_id,
                        const ndi_qos_scheduler_group_struct_t *p,
                        sai_attribute_t &sai_attr)
{
    // Only the following attributes are settable
    if (attr_id == BASE_QOS_SCHEDULER_GROUP_PORT_ID) {
        sai_attr.id = SAI_SCHEDULER_GROUP_ATTR_PORT_ID;
        sai_object_id_t sai_port;
        if (ndi_sai_port_id_get(p->ndi_port.npu_id, p->ndi_port.npu_port, &sai_port) != STD_ERR_OK) {
            return STD_ERR(NPU, PARAM, 0);
        }
        sai_attr.value.oid = sai_port;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_GROUP_LEVEL) {
        sai_attr.id = SAI_SCHEDULER_GROUP_ATTR_LEVEL;
        sai_attr.value.u32 = p->level;
    }
    else if (attr_id == BASE_QOS_SCHEDULER_GROUP_SCHEDULER_PROFILE_ID) {
        sai_attr.id = SAI_SCHEDULER_GROUP_ATTR_SCHEDULER_PROFILE_ID;
        sai_attr.value.oid = ndi2sai_scheduler_profile_id(p->scheduler_profile_id);
    }

    return STD_ERR_OK;
}


static t_std_error ndi_qos_fill_sg_attr_list(const nas_attr_id_t *nas_attr_list,
                                    uint_t num_attr,
                                    const ndi_qos_scheduler_group_struct_t *p,
                                    std::vector<sai_attribute_t> &attr_list,
                                    uint_t &count)
{
    sai_attribute_t sai_attr = {0};
    t_std_error      rc = STD_ERR_OK;

    count = 0;

    for (uint_t i = 0; i < num_attr; i++) {
        if ((rc = ndi_qos_fill_sg_attr(nas_attr_list[i], p, sai_attr)) != STD_ERR_OK)
            return rc;

        attr_list.push_back(sai_attr);

        count++;
    }

    return STD_ERR_OK;
}




/**
 * This function creates a Scheduler group ID in the NPU.
 * @param npu id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param p scheduler group structure to be modified
 * @param[out] ndi_scheduler_group_id
 * @return standard error
 */
t_std_error ndi_qos_create_scheduler_group(npu_id_t npu_id,
                                const nas_attr_id_t *nas_attr_list,
                                uint_t num_attr,
                                const ndi_qos_scheduler_group_struct_t *p,
                                ndi_obj_id_t *ndi_scheduler_group_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    uint_t attr_count = 0;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    std::vector<sai_attribute_t>  attr_list;

    if (ndi_qos_fill_sg_attr_list(nas_attr_list, num_attr, p,
                                attr_list, attr_count) != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    sai_object_id_t sai_qos_sg_id;
    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            create_scheduler_group(&sai_qos_sg_id,
                                attr_count,
                                &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group creation failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }
    *ndi_scheduler_group_id = sai2ndi_scheduler_group_id(sai_qos_sg_id);
    return STD_ERR_OK;

}

 /**
  * This function sets the scheduler_group attributes in the NPU.
  * @param npu id
  * @param ndi_scheduler_group_id
  * @param attr_id based on the CPS API attribute enumeration values
  * @param p scheduler_group structure to be modified
  * @return standard error
  */
t_std_error ndi_qos_set_scheduler_group_attr(npu_id_t npu_id,
                                    ndi_obj_id_t ndi_scheduler_group_id,
                                    BASE_QOS_SCHEDULER_GROUP_t attr_id,
                                    const ndi_qos_scheduler_group_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    sai_attribute_t sai_attr;
    if (ndi_qos_fill_sg_attr(attr_id, p, sai_attr) != STD_ERR_OK)
        return STD_ERR(QOS, CFG, 0);

    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            set_scheduler_group_attribute(
                    ndi2sai_scheduler_group_id(ndi_scheduler_group_id),
                    &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group set failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

/**
 * This function deletes a scheduler_group in the NPU.
 * @param npu_id npu id
 * @param ndi_scheduler_group_id
 * @return standard error
 */
t_std_error ndi_qos_delete_scheduler_group(npu_id_t npu_id,
                                    ndi_obj_id_t ndi_scheduler_group_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            remove_scheduler_group(ndi2sai_scheduler_group_id(ndi_scheduler_group_id)))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group deletion failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

const static std::unordered_map<nas_attr_id_t, sai_attr_id_t, std::hash<int>>
    ndi2sai_scheduler_group_attr_id_map = {
        {BASE_QOS_SCHEDULER_GROUP_CHILD_COUNT,    SAI_SCHEDULER_GROUP_ATTR_CHILD_COUNT},
        {BASE_QOS_SCHEDULER_GROUP_CHILD_LIST,     SAI_SCHEDULER_GROUP_ATTR_CHILD_LIST},
        {BASE_QOS_SCHEDULER_GROUP_PORT_ID,        SAI_SCHEDULER_GROUP_ATTR_PORT_ID},
        {BASE_QOS_SCHEDULER_GROUP_LEVEL,          SAI_SCHEDULER_GROUP_ATTR_LEVEL},
        {BASE_QOS_SCHEDULER_GROUP_SCHEDULER_PROFILE_ID, SAI_SCHEDULER_GROUP_ATTR_SCHEDULER_PROFILE_ID},
};

static t_std_error _fill_ndi_qos_scheduler_group_struct(sai_attribute_t *attr_list,
                        uint_t num_attr, ndi_qos_scheduler_group_struct_t *p)
{

    t_std_error rc;
    for (uint_t i = 0 ; i< num_attr; i++ ) {
        sai_attribute_t *attr = &attr_list[i];
        switch(attr->id) {
        case SAI_SCHEDULER_GROUP_ATTR_MAX_CHILDS:
            p->max_child = attr->value.u32;
            break;
        case SAI_SCHEDULER_GROUP_ATTR_CHILD_COUNT:
            p->child_count = attr->value.u32;
            break;
        case SAI_SCHEDULER_GROUP_ATTR_CHILD_LIST:
            for (uint_t j = 0; j < attr->value.objlist.count; j ++) {
                if (p->level >= MAX_SCHEDULER_LEVEL - 2) {
                    p->child_list[j] =
                        sai2ndi_queue_id(attr->value.objlist.list[j]);
                } else {
                    p->child_list[j] =
                        sai2ndi_scheduler_group_id(attr->value.objlist.list[j]);
                }
            }
            break;
        case SAI_SCHEDULER_GROUP_ATTR_PORT_ID:
            rc = ndi_npu_port_id_get(attr->value.oid,
                                     &p->ndi_port.npu_id,
                                     &p->ndi_port.npu_port);
            if (rc != STD_ERR_OK) {
                EV_LOG_TRACE(ev_log_t_QOS, ev_log_s_MAJOR, "QOS",
                             "Invalid port_id attribute");
                return rc;
            }
            break;
        case SAI_SCHEDULER_GROUP_ATTR_LEVEL:
            p->level = attr->value.u32;
            break;
        case SAI_SCHEDULER_GROUP_ATTR_SCHEDULER_PROFILE_ID:
            p->scheduler_profile_id = sai2ndi_scheduler_profile_id(attr->value.oid);
            break;
        default:
            EV_LOG_TRACE(ev_log_t_QOS, ev_log_s_MAJOR, "QOS",
                         "SAI attribute %lu is not supported by NDI and will be ignored",
                         attr->id);
            continue;
        }
    }

    return STD_ERR_OK;
}

/**
 * This function get a scheduler_group from the NPU.
 * @param npu id
 * @param ndi_scheduler_group_id
 * @param nas_attr_list based on the CPS API attribute enumeration values
 * @param num_attr number of attributes in attr_list array
 * @param[out] ndi_qos_scheduler_group_struct_t filled if success
 * @return standard error
 */
t_std_error ndi_qos_get_scheduler_group(npu_id_t npu_id,
                            ndi_obj_id_t ndi_scheduler_group_id,
                            const nas_attr_id_t *nas_attr_list,
                            uint_t num_attr,
                            ndi_qos_scheduler_group_struct_t *p)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_attribute_t> attr_list;
    std::vector<sai_object_id_t> child_id_list;
    sai_attribute_t sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    try {
        for (uint_t i = 0; i < num_attr; i++) {
            sai_attr.id = ndi2sai_scheduler_group_attr_id_map.at(nas_attr_list[i]);
            if (sai_attr.id == SAI_SCHEDULER_GROUP_ATTR_CHILD_LIST) {
                sai_attr.value.objlist.count = p->child_count;
                child_id_list.resize(p->child_count);
                sai_attr.value.objlist.list = &child_id_list[0];
            }
            attr_list.push_back(sai_attr);
        }
    }
    catch(...) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                    "Unexpected error.\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            get_scheduler_group_attribute(
                    ndi2sai_scheduler_group_id(ndi_scheduler_group_id),
                    num_attr,
                    &attr_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group get failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    // convert sai result to NAS format
    _fill_ndi_qos_scheduler_group_struct(&attr_list[0], num_attr, p);

    return STD_ERR_OK;

}

/**
 * This function adds a list of child node to a scheduler group
 * @param npu_id
 * @param ndi_scheduler_group_id
 * @param child_count number of childs in ndi_child_list
 * @param ndi_child_list list of the childs to be added
 */
t_std_error ndi_qos_add_child_to_scheduler_group(npu_id_t npu_id,
                            ndi_obj_id_t ndi_scheduler_group_id,
                            uint32_t child_count,
                            const ndi_obj_id_t *ndi_child_list)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_object_id_t> sai_child_list(child_count);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    for (uint_t i= 0; i< child_count; i++) {
        // no special conversion necessary for ndi-queue-id or ndi-scheduler-group-id
        sai_child_list[i] = ndi_child_list[i];
    }

    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            add_child_object_to_group(
                    ndi2sai_scheduler_group_id(ndi_scheduler_group_id),
                    child_count,
                    &sai_child_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group add child failed:"
                      "ndi_scheduler_group_id: 0x%016lx,"
                      "child_count: %u\n",
                      npu_id, ndi_scheduler_group_id, child_count);
        for (uint i=0; i< child_count; i++)
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                        "child [%u]: 0x%016lx\n",
                        i, ndi_child_list[i]);

        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

/**
 * This function deletes a list of child node to a scheduler group
 * @param npu_id
 * @param ndi_scheduler_group_id
 * @param child_count number of childs in ndi_child_list
 * @param ndi_child_list list of the childs to be deleted
 */
t_std_error ndi_qos_delete_child_from_scheduler_group(npu_id_t npu_id,
                            ndi_obj_id_t ndi_scheduler_group_id,
                            uint32_t child_count,
                            const ndi_obj_id_t *ndi_child_list)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    std::vector<sai_object_id_t> sai_child_list(child_count);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", npu_id);
        return STD_ERR(QOS, CFG, 0);
    }

    for (uint_t i= 0; i< child_count; i++) {
        // no conversion necessary for ndi-queue-id or ndi-scheduler-group-id
        sai_child_list[i] = ndi_child_list[i];
    }

    if ((sai_ret = ndi_sai_qos_scheduler_group_api(ndi_db_ptr)->
            remove_child_object_from_group(
                    ndi2sai_scheduler_group_id(ndi_scheduler_group_id),
                    child_count,
                    &sai_child_list[0]))
                         != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d scheduler group add child failed\n", npu_id);
        return STD_ERR(QOS, CFG, sai_ret);
    }

    return STD_ERR_OK;

}

/**
 * This function gets the total number of scheduler-groups on a port
 * @param ndi_port_id
 * @Return number of scheduler-group
 */
uint_t ndi_qos_get_number_of_scheduler_groups(ndi_port_t ndi_port_id)
{
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);
        return 0;
    }

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_QOS_NUMBER_OF_SCHEDULER_GROUPS;

    sai_object_id_t sai_port;
    if (ndi_sai_port_id_get(ndi_port_id.npu_id, ndi_port_id.npu_port, &sai_port) != STD_ERR_OK) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "Failed to find SAI port for npu_id %d port %d\n",
                      ndi_port_id.npu_id, ndi_port_id.npu_port);
        return 0;
    }

    sai_status_t sai_ret = ndi_sai_qos_port_api(ndi_db_ptr)->get_port_attribute(sai_port,
                                                                                1, &sai_attr);

    if (sai_ret != SAI_STATUS_SUCCESS) {
         NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                       "Failed to get port attribute\n");
         return 0;
    }

    return sai_attr.value.u32;
}

/**
 * This function gets the list of scheduler-groups of a port
 * @param ndi_port_id
 * @param count size of the scheduler-group list
 * @param[out] ndi_sg_id_list[] to be filled with either the number of scheduler-groups
 *            that the port owns or the size of array itself, whichever is less.
 * @Return Number of scheduler-groups that the port owns.
 */
uint_t ndi_qos_get_scheduler_group_id_list(ndi_port_t ndi_port_id,
                                           uint_t count,
                                           ndi_obj_id_t *ndi_sg_id_list)
{
    /* get scheduler-group list */
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(ndi_port_id.npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-QOS",
                      "npu_id %d not exist\n", ndi_port_id.npu_id);

        return 0;
    }

    sai_attribute_t sai_attr;
    std::vector<sai_object_id_t> sai_sg_id_list(count);

    sai_attr.id = SAI_PORT_ATTR_QOS_SCHEDULER_GROUP_LIST;
    sai_attr.value.objlist.count = count;
    sai_attr.value.objlist.list = &sai_sg_id_list[0];

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
    } else if (sai_ret == SAI_STATUS_BUFFER_OVERFLOW) {
        // caller didn't provide enough buffer to store all IDs,
        // only returns the count of IDs
        return sai_attr.value.objlist.count;
    }

    // copy out sai-returned scheduler-group ids to nas
    if (ndi_sg_id_list) {
        for (uint_t i = 0; i < sai_attr.value.objlist.count; i++) {
            ndi_sg_id_list[i] = sai2ndi_scheduler_group_id(sai_attr.value.objlist.list[i]);
        }
    }

    return sai_attr.value.objlist.count;
}
