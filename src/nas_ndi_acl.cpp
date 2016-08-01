
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
 * filename: nas_ndi_acl.cpp
 */


#include "std_assert.h"
#include "dell-base-acl.h"
#include "nas_ndi_int.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_utils.h"
#include "nas_base_utils.h"
#include "nas_ndi_acl.h"
#include "nas_ndi_acl_utl.h"
#include <vector>
#include <unordered_map>
#include <string.h>
#include <list>
#include <inttypes.h>

static inline t_std_error _sai_to_ndi_err (sai_status_t st)
{
    return ndi_utl_mk_std_err (e_std_err_ACL, st);
}

extern "C" {

t_std_error ndi_acl_table_create (npu_id_t npu_id, const ndi_acl_table_t* ndi_tbl_p,
                                  ndi_obj_id_t* ndi_tbl_id_p)
{
    t_std_error                   rc = STD_ERR_OK;

    sai_status_t                  sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t               sai_tbl_id = 0;

    sai_attribute_t               sai_tbl_attr = {0}, nil_attr = {0};
    std::vector<sai_attribute_t>  sai_tbl_attr_list;
    nas_ndi_db_t                  *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    /* Reserve some space to avoid repeated memmoves */
    sai_tbl_attr_list.reserve (2 + ndi_tbl_p->filter_count);

    // Table Stage
    sai_tbl_attr.id = SAI_ACL_TABLE_ATTR_STAGE;
    switch (ndi_tbl_p->stage) {
        case  BASE_ACL_STAGE_INGRESS:
            sai_tbl_attr.value.u32 = SAI_ACL_STAGE_INGRESS;
            break;
        case  BASE_ACL_STAGE_EGRESS:
            sai_tbl_attr.value.u32 = SAI_ACL_STAGE_EGRESS;
            break;
        default:
            NDI_ACL_LOG_ERROR ("Invalid Stage %d", ndi_tbl_p->stage);
            return STD_ERR(ACL, PARAM, 0);
    }
    sai_tbl_attr_list.push_back (sai_tbl_attr);

    // Table Priority
    sai_tbl_attr = nil_attr;
    sai_tbl_attr.id = SAI_ACL_TABLE_ATTR_PRIORITY;
    sai_tbl_attr.value.u32 = ndi_tbl_p->priority;
    sai_tbl_attr_list.push_back (sai_tbl_attr);

    // Set of Filters allowed in Table
    for (uint_t count = 0; count < ndi_tbl_p->filter_count; count++) {
        sai_tbl_attr = nil_attr;
        auto filter_type = ndi_tbl_p->filter_list[count];

        if ((rc = ndi_acl_utl_ndi2sai_tbl_filter_type (filter_type, &sai_tbl_attr))
            != STD_ERR_OK) {
            NDI_ACL_LOG_ERROR("ACL filter type %d is not supported in SAI",
                              filter_type);
            return rc;
        }
        sai_tbl_attr_list.push_back (sai_tbl_attr);
    }

    NDI_ACL_LOG_DETAIL ("Creating ACL Table with %d attributes",
                        sai_tbl_attr_list.size());

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->create_acl_table (&sai_tbl_id,
                                                                   sai_tbl_attr_list.size(),
                                                                   sai_tbl_attr_list.data()))
        != SAI_STATUS_SUCCESS) {
        NDI_ACL_LOG_ERROR ("Create ACL Table failed in SAI %d", sai_ret);
        return _sai_to_ndi_err (sai_ret);
    }

    *ndi_tbl_id_p = ndi_acl_utl_sai2ndi_table_id (sai_tbl_id);
    NDI_ACL_LOG_INFO ("Successfully created ACL Table - Return NDI ID %" PRIx64,
                      *ndi_tbl_id_p);
    return rc;
}

t_std_error ndi_acl_table_delete (npu_id_t npu_id, ndi_obj_id_t ndi_tbl_id)
{
    t_std_error       rc = STD_ERR_OK;
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t     *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_tbl_id = ndi_acl_utl_ndi2sai_table_id (ndi_tbl_id);

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->delete_acl_table (sai_tbl_id))
        != SAI_STATUS_SUCCESS) {

        NDI_ACL_LOG_ERROR ("Delete ACL Table failed %d", sai_ret);
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully deleted ACL Table NDI ID %" PRIx64,
                      ndi_tbl_id);
    return rc;
}

t_std_error ndi_acl_table_set_priority (npu_id_t npu_id,
                                        ndi_obj_id_t ndi_tbl_id,
                                        ndi_acl_priority_t tbl_priority)
{
    t_std_error         rc = STD_ERR_OK;
    sai_status_t        sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t     sai_tbl_attr = {0};
    nas_ndi_db_t       *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_tbl_id = ndi_acl_utl_ndi2sai_table_id (ndi_tbl_id);

    sai_tbl_attr.id = SAI_ACL_TABLE_ATTR_PRIORITY;
    sai_tbl_attr.value.u32 = tbl_priority;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_table_attribute (sai_tbl_id,
                                                                          &sai_tbl_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully set priority for ACL Table NDI ID %" PRIx64,
                      ndi_tbl_id);
    return rc;
}

t_std_error ndi_acl_entry_create (npu_id_t npu_id,
                                  const ndi_acl_entry_t* ndi_entry_p,
                                  ndi_obj_id_t* ndi_entry_id_p)
{
    t_std_error                   rc = STD_ERR_OK;

    sai_status_t                  sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t               sai_entry_id = 0;

    sai_attribute_t               sai_entry_attr = {0}, nil_attr = {0};
    std::vector<sai_attribute_t>  sai_entry_attr_list;
    nas_ndi_db_t                  *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    nas::mem_alloc_helper_t       malloc_tracker;

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    /* Reserve some space to avoid repeated memmoves */
    sai_entry_attr_list.reserve (3 + ndi_entry_p->filter_count + ndi_entry_p->action_count);

    // Table Id to which Entry belongs
    sai_entry_attr.id = SAI_ACL_ENTRY_ATTR_TABLE_ID;
    sai_entry_attr.value.oid = ndi_entry_p->table_id;
    sai_entry_attr_list.push_back (sai_entry_attr);

    // Entry Priority
    sai_entry_attr = nil_attr;
    sai_entry_attr.id = SAI_ACL_ENTRY_ATTR_PRIORITY;
    sai_entry_attr.value.u32 = ndi_entry_p->priority;
    sai_entry_attr_list.push_back (sai_entry_attr);

    // Entry Admin State
    sai_entry_attr = nil_attr;
    sai_entry_attr.id = SAI_ACL_ENTRY_ATTR_ADMIN_STATE;
    sai_entry_attr.value.u8 = 1; // Enabled
    sai_entry_attr_list.push_back (sai_entry_attr);

    // Filter fields and their values
    for (uint_t count = 0; count < ndi_entry_p->filter_count; count++) {
        sai_entry_attr = nil_attr;
        ndi_acl_entry_filter_t *filter_p = &(ndi_entry_p->filter_list[count]);

        if ((rc = ndi_acl_utl_fill_sai_filter (&sai_entry_attr, filter_p,
                                                malloc_tracker)) != STD_ERR_OK) {
            return rc;
        }
        sai_entry_attr.value.aclfield.enable = true;
        sai_entry_attr_list.push_back (sai_entry_attr);
    }

    // Actions and their values
    for (uint_t count = 0; count < ndi_entry_p->action_count; count++) {
        sai_entry_attr = nil_attr;
        ndi_acl_entry_action_t *action_p = &(ndi_entry_p->action_list[count]);

        if ((rc = ndi_acl_utl_fill_sai_action (&sai_entry_attr, action_p,
                                                malloc_tracker)) != STD_ERR_OK) {
            return rc;
        }
        sai_entry_attr.value.aclfield.enable = true;
        sai_entry_attr_list.push_back (sai_entry_attr);
    }

    NDI_ACL_LOG_DETAIL ("Creating ACL Entry with %d attributes",
                        sai_entry_attr_list.size());

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->create_acl_entry (&sai_entry_id,
                                                                   sai_entry_attr_list.size(),
                                                                   sai_entry_attr_list.data()))
        != SAI_STATUS_SUCCESS) {
        NDI_ACL_LOG_ERROR ("Create ACL Entry failed in SAI %d", sai_ret);
        return _sai_to_ndi_err (sai_ret);
    }

    *ndi_entry_id_p = ndi_acl_utl_sai2ndi_entry_id (sai_entry_id);
    NDI_ACL_LOG_INFO ("Successfully created ACL Entry - Return NDI ID %" PRIx64,
                      *ndi_entry_id_p);

    return rc;
}

t_std_error ndi_acl_entry_delete (npu_id_t npu_id, ndi_obj_id_t ndi_entry_id)
{
    t_std_error rc = STD_ERR_OK;
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t     *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->delete_acl_entry (sai_entry_id))
        != SAI_STATUS_SUCCESS) {

        NDI_ACL_LOG_ERROR ("Delete ACL Entry failed in SAI %d", sai_ret);
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully deleted ACL Entry NDI ID %" PRIx64,
                      ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_entry_set_priority (npu_id_t npu_id,
                                        ndi_obj_id_t ndi_entry_id,
                                        ndi_acl_priority_t entry_prio)
{
    t_std_error         rc = STD_ERR_OK;
    sai_status_t        sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t     sai_entry_attr = {0};
    nas_ndi_db_t       *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    sai_entry_attr.id = SAI_ACL_ENTRY_ATTR_PRIORITY;
    sai_entry_attr.value.u32 = entry_prio;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_entry_attribute (sai_entry_id,
                                                                          &sai_entry_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully set priority for ACL Entry NDI ID %" PRIx64,
                      ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_entry_set_filter (npu_id_t npu_id,
                                      ndi_obj_id_t ndi_entry_id,
                                      ndi_acl_entry_filter_t* filter_p)
{
    t_std_error           rc = STD_ERR_OK;
    sai_status_t          sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t       sai_entry_attr = {0};
    nas_ndi_db_t         *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    nas::mem_alloc_helper_t malloc_tracker;

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    if ((rc = ndi_acl_utl_fill_sai_filter (&sai_entry_attr, filter_p,
                                            malloc_tracker)) != STD_ERR_OK) {
        return rc;
    }

    sai_entry_attr.value.aclfield.enable = true;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_entry_attribute (sai_entry_id,
                                                                          &sai_entry_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully set filter type %d for ACL Table NDI ID %" PRIx64,
                      filter_p->filter_type, ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_entry_disable_filter (npu_id_t npu_id,
                                          ndi_obj_id_t ndi_entry_id,
                                          BASE_ACL_MATCH_TYPE_t filter_type)
{
    t_std_error           rc = STD_ERR_OK;
    sai_status_t          sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t       sai_entry_attr = {0};
    nas_ndi_db_t         *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    // Action ID
    if ((rc = ndi_acl_utl_ndi2sai_filter_type (filter_type, &sai_entry_attr))
        != STD_ERR_OK) {
        NDI_ACL_LOG_ERROR ("Filter type %d is not supported by SAI",
                           filter_type);
        return rc;
    }
    sai_entry_attr.value.aclfield.enable = false;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_entry_attribute (sai_entry_id,
                                                                          &sai_entry_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully disabled filter type %d for ACL Table NDI ID %" PRIx64,
                      filter_type, ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_entry_set_action (npu_id_t npu_id,
                                      ndi_obj_id_t ndi_entry_id,
                                      ndi_acl_entry_action_t* action_p)
{
    t_std_error           rc = STD_ERR_OK;
    sai_status_t          sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t       sai_entry_attr = {0};
    nas_ndi_db_t         *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    nas::mem_alloc_helper_t malloc_tracker;

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    if ((rc = ndi_acl_utl_fill_sai_action (&sai_entry_attr, action_p,
                                            malloc_tracker)) != STD_ERR_OK) {
        return rc;
    }

    sai_entry_attr.value.aclfield.enable = true;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_entry_attribute (sai_entry_id,
                                                                          &sai_entry_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully set action type %d for ACL Entry NDI ID %" PRIx64,
                      action_p->action_type, ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_entry_disable_action (npu_id_t npu_id,
                                          ndi_obj_id_t ndi_entry_id,
                                          BASE_ACL_ACTION_TYPE_t action_type)
{
    t_std_error           rc = STD_ERR_OK;
    sai_status_t          sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t       sai_entry_attr = {0};
    nas_ndi_db_t         *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_entry_id = ndi_acl_utl_ndi2sai_entry_id (ndi_entry_id);

    // Action ID
    if ((rc = ndi_acl_utl_ndi2sai_action_type (action_type, &sai_entry_attr))
        != STD_ERR_OK) {
        NDI_ACL_LOG_ERROR ("Action type %d is not supported by SAI",
                           action_type);
        return rc;
    }
    sai_entry_attr.value.aclaction.enable = false;

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->set_acl_entry_attribute (sai_entry_id,
                                                                          &sai_entry_attr))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully disabled action type %d for ACL Table NDI ID %" PRIx64,
                      action_type, ndi_entry_id);
    return rc;
}

t_std_error ndi_acl_counter_create (npu_id_t npu_id,
                                    const ndi_acl_counter_t* ndi_counter_p,
                                    ndi_obj_id_t* ndi_counter_id_p)
{
    sai_status_t                  sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t               sai_counter_id = 0;

    sai_attribute_t               sai_counter_attr = {0}, nil_attr = {0};
    std::vector<sai_attribute_t>  sai_counter_attr_list;
    nas_ndi_db_t                  *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    /* Reserve some space to avoid repeated memmoves */
    sai_counter_attr_list.reserve (3);

    // Table Id to which Entry belongs
    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_TABLE_ID;
    sai_counter_attr.value.oid = ndi_counter_p->table_id;
    sai_counter_attr_list.push_back (sai_counter_attr);

    // Enable/Disable Pkt count mode
    sai_counter_attr = nil_attr;
    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_ENABLE_PACKET_COUNT;
    sai_counter_attr.value.u8 = ndi_counter_p->enable_pkt_count;
    sai_counter_attr_list.push_back (sai_counter_attr);

    // Enable/Disable Byte count mode
    sai_counter_attr = nil_attr;
    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_ENABLE_BYTE_COUNT;
    sai_counter_attr.value.u8 = ndi_counter_p->enable_byte_count; // Enabled
    sai_counter_attr_list.push_back (sai_counter_attr);

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->create_acl_counter (&sai_counter_id,
                                                                     sai_counter_attr_list.size(),
                                                                     sai_counter_attr_list.data()))
        != SAI_STATUS_SUCCESS) {
        return _sai_to_ndi_err (sai_ret);
    }

    *ndi_counter_id_p = ndi_acl_utl_sai2ndi_counter_id (sai_counter_id);

    NDI_ACL_LOG_INFO ("Successfully created counter - Return NDI ID %" PRIx64,
                      *ndi_counter_id_p);

    return STD_ERR_OK;
}

t_std_error ndi_acl_counter_delete (npu_id_t npu_id,
                                    ndi_obj_id_t ndi_counter_id)
{
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t     *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) return STD_ERR(ACL, FAIL, 0);

    sai_object_id_t sai_counter_id = ndi_acl_utl_ndi2sai_counter_id (ndi_counter_id);

    // Call SAI API
    if ((sai_ret = ndi_acl_utl_api_get(ndi_db_ptr)->delete_acl_counter (sai_counter_id))
        != SAI_STATUS_SUCCESS) {

        return _sai_to_ndi_err (sai_ret);
    }

    NDI_ACL_LOG_INFO ("Successfully deleted counter - Return NDI ID %" PRIx64,
                      ndi_counter_id);

    return STD_ERR_OK;
}

t_std_error ndi_acl_counter_enable_pkt_count (npu_id_t npu_id,
                                              ndi_obj_id_t ndi_counter_id,
                                              bool enable)
{
    sai_attribute_t    sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_ENABLE_PACKET_COUNT;
    sai_counter_attr.value.u8 = enable;

    return ndi_acl_utl_set_counter_attr (npu_id, ndi_counter_id, &sai_counter_attr);
}

t_std_error ndi_acl_counter_enable_byte_count (npu_id_t npu_id,
                                               ndi_obj_id_t ndi_counter_id,
                                               bool enable)
{
    sai_attribute_t    sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_ENABLE_BYTE_COUNT;
    sai_counter_attr.value.u8 = enable;

    return ndi_acl_utl_set_counter_attr (npu_id, ndi_counter_id, &sai_counter_attr);
}

t_std_error ndi_acl_counter_set_pkt_count (npu_id_t npu_id,
                                           ndi_obj_id_t ndi_counter_id,
                                           uint64_t  pkt_count)
{
    sai_attribute_t   sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_PACKETS;
    sai_counter_attr.value.u64 = pkt_count;

    return ndi_acl_utl_set_counter_attr (npu_id, ndi_counter_id, &sai_counter_attr);
}

t_std_error ndi_acl_counter_set_byte_count (npu_id_t npu_id,
                                           ndi_obj_id_t ndi_counter_id,
                                           uint64_t  byte_count)
{
    sai_attribute_t   sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_BYTES;
    sai_counter_attr.value.u64 = byte_count;

    return ndi_acl_utl_set_counter_attr (npu_id, ndi_counter_id, &sai_counter_attr);
}

t_std_error ndi_acl_counter_get_pkt_count (npu_id_t npu_id,
                                           ndi_obj_id_t ndi_counter_id,
                                           uint64_t*  pkt_count_p)
{
    sai_attribute_t   sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_PACKETS;

    t_std_error rc = ndi_acl_utl_get_counter_attr (npu_id, ndi_counter_id,
                                                    &sai_counter_attr);
    *pkt_count_p = sai_counter_attr.value.u64;

    return rc;
}

t_std_error ndi_acl_counter_get_byte_count (npu_id_t npu_id,
                                           ndi_obj_id_t ndi_counter_id,
                                           uint64_t*  byte_count_p)
{
    sai_attribute_t   sai_counter_attr = {0};

    sai_counter_attr.id = SAI_ACL_COUNTER_ATTR_BYTES;

    t_std_error rc = ndi_acl_utl_get_counter_attr (npu_id, ndi_counter_id,
                                                    &sai_counter_attr);
    *byte_count_p = sai_counter_attr.value.u64;

    return rc;
}

} // end Extern C
