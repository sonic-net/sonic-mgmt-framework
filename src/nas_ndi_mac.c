/*
 * Copyright (c) 2016 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License. You may obtain
 * a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * THIS CODE IS PROVIDED ON AN  *AS IS* BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
 * LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS
 * FOR A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
 *
 * See the Apache Version 2.0 License for specific language governing
 * permissions and limitations under the License.
 */

/*
 * filename: nas_ndi_mac.c
 */

#include "std_error_codes.h"
#include "ds_common_types.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_mac.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_mac_utl.h"
#include "sai.h"
#include "saistatus.h"
#include "saitypes.h"
#include "std_mac_utils.h"
#include <stdio.h>
#include <string.h>

#define MAC_STR_LEN 20

static inline t_std_error sai_to_ndi_err_translate (sai_status_t sai_err)
{
    return ndi_utl_mk_std_err(e_std_err_MAC, sai_err);
}

/*  NDI L2 MAC specific APIs  */

static inline  sai_fdb_api_t *ndi_mac_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_fdb_api_tbl);
}

t_std_error ndi_update_mac_entry(ndi_mac_entry_t *p_mac_entry, ndi_mac_attr_flags attr_changed)
{
    t_std_error               rc = STD_ERR_OK;
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_fdb_entry_t           sai_mac_entry;
    sai_attribute_t           sai_attr;
    sai_object_id_t           sai_port;
    char mac_string[MAC_STR_LEN];

    EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", ": Entered");

    if (p_mac_entry == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "NULL parameter passed");
        return (STD_ERR(MAC, FAIL, 0));
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_mac_entry->npu_id);

    if (ndi_db_ptr == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "Not able to find NDI DB node for npu_id: %d", p_mac_entry->npu_id);
        return (STD_ERR(MAC, FAIL, 0));
    }

    memcpy(&(sai_mac_entry.mac_address), p_mac_entry->mac_addr, HAL_MAC_ADDR_LEN);
    sai_mac_entry.vlan_id = p_mac_entry->vlan_id;

    if (p_mac_entry->ndi_lag_id != 0) {
        sai_port = p_mac_entry->ndi_lag_id;

    } else {
        if ((rc = ndi_sai_port_id_get(p_mac_entry->port_info.npu_id,
                    p_mac_entry->port_info.npu_port, &sai_port)) != STD_ERR_OK) {
            EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Not able to find SAI port id for npu:%d port:%d",
                p_mac_entry->port_info.npu_id, p_mac_entry->port_info.npu_port);
            return sai_to_ndi_err_translate(rc);
        }
    }

    switch (attr_changed) {
        case NDI_MAC_ENTRY_ATTR_PORT_ID:
            EV_LOG(TRACE, NDI, ev_log_s_MINOR, "NDI-MAC",
                    " updating mac entry %s with new port npu_id = 0x%x, npu_port = 0x%x",
                    std_mac_to_string((const hal_mac_addr_t *)&p_mac_entry->mac_addr, mac_string, MAC_STR_LEN),
                    p_mac_entry->port_info.npu_id, p_mac_entry->port_info.npu_port);
            sai_attr.id = SAI_FDB_ENTRY_ATTR_PORT_ID;
            sai_attr.value.oid = sai_port;
            break;
        case NDI_MAC_ENTRY_ATTR_PKT_ACTION:
            EV_LOG(TRACE, NDI, ev_log_s_MINOR, "NDI-MAC",
                    " updating mac entry %s with new action 0x%x",
                    std_mac_to_string((const hal_mac_addr_t *)&p_mac_entry->mac_addr, mac_string, MAC_STR_LEN),
                    p_mac_entry->action);
            sai_attr.id = SAI_FDB_ENTRY_ATTR_PACKET_ACTION;
            sai_attr.value.s32 = ndi_mac_sai_packet_action_get(p_mac_entry->action);
            break;
        default:
            EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC",
                    "unsupported attribute value for update , attr 0x%x ", attr_changed);
            return  STD_ERR_OK;
    }


    if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->set_fdb_entry_attribute(&sai_mac_entry, &sai_attr))
                          != SAI_STATUS_SUCCESS) {
        EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC",
                "Failed to update mac entry for vlan:%d: ret:%d",
                p_mac_entry->vlan_id, sai_ret);
        return sai_to_ndi_err_translate(sai_ret);
    }

    EV_LOG(TRACE, NDI, ev_log_s_MINOR, "NDI-MAC", "Successfully returned");
    return STD_ERR_OK;
}


t_std_error ndi_create_mac_entry(ndi_mac_entry_t *p_mac_entry)
{
    t_std_error               rc = STD_ERR_OK;
    uint32_t                  attr_idx = 0;
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_fdb_entry_t           sai_mac_entry;
    sai_attribute_t           sai_attr[NDI_MAC_ENTRY_ATTR_MAX -1];
    sai_object_id_t           sai_port;

    EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "%s: Entered", __FUNCTION__);

    if (p_mac_entry == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "NULL parameter passed");
        return (STD_ERR(MAC, FAIL, 0));
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_mac_entry->npu_id);

    if (ndi_db_ptr == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "Not able to find NDI DB node for npu_id: %d", p_mac_entry->npu_id);
        return (STD_ERR(MAC, FAIL, 0));
    }

    memcpy(&(sai_mac_entry.mac_address), p_mac_entry->mac_addr, HAL_MAC_ADDR_LEN);
    sai_mac_entry.vlan_id = p_mac_entry->vlan_id;

    if (p_mac_entry->ndi_lag_id != 0) {
        sai_port = p_mac_entry->ndi_lag_id;
    } else {
        if ((rc = ndi_sai_port_id_get(p_mac_entry->port_info.npu_id,
                    p_mac_entry->port_info.npu_port, &sai_port)) != STD_ERR_OK) {
            EV_LOG(ERR, NDI, ev_log_s_CRITICAL,
                "NDI-MAC", "Not able to find SAI port id for npu:%d port:%d",
                p_mac_entry->port_info.npu_id, p_mac_entry->port_info.npu_port);
            return sai_to_ndi_err_translate(rc);
        }
    }

    sai_attr[attr_idx].id = SAI_FDB_ENTRY_ATTR_TYPE;
    sai_attr[attr_idx++].value.s32 = (p_mac_entry->is_static) ? SAI_FDB_ENTRY_STATIC : SAI_FDB_ENTRY_DYNAMIC;

    sai_attr[attr_idx].id = SAI_FDB_ENTRY_ATTR_PORT_ID;
    sai_attr[attr_idx++].value.oid = sai_port;

    sai_attr[attr_idx].id = SAI_FDB_ENTRY_ATTR_PACKET_ACTION;
    sai_attr[attr_idx++].value.s32 = ndi_mac_sai_packet_action_get(p_mac_entry->action);

    if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->create_fdb_entry(&sai_mac_entry, attr_idx, sai_attr))
                          != SAI_STATUS_SUCCESS) {
        EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to configure mac entry for vlan:%d: ret:%d",
                p_mac_entry->vlan_id, sai_ret);
        return sai_to_ndi_err_translate(sai_ret);
    }

    EV_LOG(TRACE, NDI, ev_log_s_MINOR, "NDI-MAC", "Successfully returned from %s", __FUNCTION__);
    return STD_ERR_OK;
}

t_std_error ndi_delete_mac_entry(ndi_mac_entry_t *p_mac_entry, ndi_mac_delete_type_t delete_type, bool type_set)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    uint32_t                  attr_idx = 0;
    sai_fdb_entry_t           sai_mac_entry;
    sai_object_id_t           sai_port;
    sai_attribute_t           fdb_flush_attr[3];

    EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "%s: Entered", __FUNCTION__);

    if (p_mac_entry == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "NULL parameter passed");
        return (STD_ERR(MAC, FAIL, 0));
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_mac_entry->npu_id);

    if (ndi_db_ptr == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_MAJOR, "NDI-MAC", "Not able to find NDI DB node");
        return (STD_ERR(MAC, FAIL, 0));
    }

    memset(fdb_flush_attr, 0, sizeof(fdb_flush_attr));

    switch(delete_type) {

        case NDI_MAC_DEL_SINGLE_ENTRY:
            memcpy(&(sai_mac_entry.mac_address), p_mac_entry->mac_addr, HAL_MAC_ADDR_LEN);
            sai_mac_entry.vlan_id = p_mac_entry->vlan_id;
            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_SINGLE_ENTRY:"
                    "Before SAI call vlan=%d", sai_mac_entry.vlan_id);
            if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->remove_fdb_entry(&sai_mac_entry))
                    != SAI_STATUS_SUCCESS) {
                EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to remove mac entry for"
                        "vlan:%d ret=%d",
                        p_mac_entry->vlan_id, sai_ret);
                return sai_to_ndi_err_translate(sai_ret);
            }
            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_SINGLE_ENTRY: "
                    "Success vlan=%d", sai_mac_entry.vlan_id);
            break;

        case NDI_MAC_DEL_BY_PORT:
            if(p_mac_entry->ndi_lag_id != 0)
            {
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_PORT_ID;
                fdb_flush_attr[attr_idx++].value.oid = p_mac_entry->ndi_lag_id;
                EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT: "
                                                  "on Lag Intf 0x%llx", p_mac_entry->ndi_lag_id);
            }
            else
            {
                if (ndi_sai_port_id_get(p_mac_entry->port_info.npu_id,
                            p_mac_entry->port_info.npu_port, &sai_port) != STD_ERR_OK) {
                    EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to retrieve SAI port id "
                            "for npu:%d port:%d: ret:%d",
                            p_mac_entry->port_info.npu_id, p_mac_entry->port_info.npu_port, sai_ret);
                    return (STD_ERR(MAC, FAIL, 0));
                }
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_PORT_ID;
                fdb_flush_attr[attr_idx++].value.oid = sai_port;
            }

            /* to flush all entries for this port below attribute setting needs to be skipped */
            if (type_set) {
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_ENTRY_TYPE;
                if (p_mac_entry->is_static)
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_STATIC;
                else
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_DYNAMIC;
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT: "
                    "Before SAI Call- port id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.oid, fdb_flush_attr[1].id, fdb_flush_attr[1].value.s32);
            if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->flush_fdb_entries(attr_idx, (const sai_attribute_t *)fdb_flush_attr))
                    != SAI_STATUS_SUCCESS) {
                EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to remove mac entry by port:%d ret:%d",
                        p_mac_entry->port_info.npu_port, sai_ret);
                return sai_to_ndi_err_translate(sai_ret);
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT: "
                    "Success - port id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.oid, fdb_flush_attr[1].id, fdb_flush_attr[1].value.s32);
            break;

        case NDI_MAC_DEL_BY_VLAN:

            fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_VLAN_ID;
            fdb_flush_attr[attr_idx++].value.u16 = p_mac_entry->vlan_id;

            /* to flush all entries for this vlan below attribute setting needs to be skipped */
            if (type_set) {
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_ENTRY_TYPE;
                if (p_mac_entry->is_static)
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_STATIC;
                else
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_DYNAMIC;
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_VLAN Before SAI Call - "
                    "vlan_id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.u16, fdb_flush_attr[1].id, fdb_flush_attr[1].value.s32);

            if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->flush_fdb_entries(attr_idx, (const sai_attribute_t *)fdb_flush_attr))
                    != SAI_STATUS_SUCCESS) {
                EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to remove mac entry by vlan:%d ret:%d",
                        p_mac_entry->vlan_id, sai_ret);
                return sai_to_ndi_err_translate(sai_ret);
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_VLAN Sucsess - "
                    "vlan_id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.u16, fdb_flush_attr[1].id, fdb_flush_attr[1].value.s32);
            break;

        case NDI_MAC_DEL_BY_PORT_VLAN:

            fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_VLAN_ID;
            fdb_flush_attr[attr_idx++].value.u16 = p_mac_entry->vlan_id;
            if(p_mac_entry->ndi_lag_id != 0)
            {
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_PORT_ID;
                fdb_flush_attr[attr_idx++].value.oid = p_mac_entry->ndi_lag_id;
                EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT_VLAN: "
                                                  "on Lag Intf 0x%llx", p_mac_entry->ndi_lag_id);
            }
            else
            {
                if (ndi_sai_port_id_get(p_mac_entry->port_info.npu_id,
                            p_mac_entry->port_info.npu_port, &sai_port) != STD_ERR_OK) {
                    EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to retrieve SAI "
                            "port id - npu:%d port:%d ret:%d",
                       p_mac_entry->port_info.npu_id, p_mac_entry->port_info.npu_port, sai_ret);
                    return (STD_ERR(MAC, FAIL, sai_ret));
                }
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_PORT_ID;
                fdb_flush_attr[attr_idx++].value.oid = sai_port;
            }

            /* to flush all entries for port/vlan below attribute setting needs to be skipped */
            if (type_set) {
                fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_ENTRY_TYPE;
                if (p_mac_entry->is_static)
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_STATIC;
                else
                    fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_DYNAMIC;
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT_VLAN Before Calling SAI - "
                    "vlan_id=%d port_id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.u16, fdb_flush_attr[1].value.oid,
                    fdb_flush_attr[2].id, fdb_flush_attr[2].value.s32);

            if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->flush_fdb_entries(attr_idx, (const sai_attribute_t *)fdb_flush_attr))
                    != SAI_STATUS_SUCCESS) {
                EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to remove mac entry "
                        "by port:%d and vlan:%d ret:%d",
                        p_mac_entry->port_info.npu_port, p_mac_entry->vlan_id, sai_ret);
                return sai_to_ndi_err_translate(sai_ret);
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_BY_PORT_VLAN Success -  "
                    "vlan_id=%d port_id=%d entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].value.oid, fdb_flush_attr[1].value.oid,
                    fdb_flush_attr[2].id, fdb_flush_attr[2].value.s32);
            break;

        case NDI_MAC_DEL_ALL_ENTRIES:

            fdb_flush_attr[attr_idx].id = SAI_FDB_FLUSH_ATTR_ENTRY_TYPE;
            if (p_mac_entry->is_static)
                fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_STATIC;
            else
                fdb_flush_attr[attr_idx++].value.s32 = SAI_FDB_FLUSH_ENTRY_DYNAMIC;

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_ALL_ENTRIES Before "
                    "calling SAI - entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].id, fdb_flush_attr[0].value.s32);
            if ((sai_ret = ndi_mac_api_get(ndi_db_ptr)->flush_fdb_entries(attr_idx, (const sai_attribute_t *)fdb_flush_attr))
                    != SAI_STATUS_SUCCESS) {
                EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to remove all mac entries "
                        "for npu %d ret:%d",
                        p_mac_entry->npu_id, sai_ret);
                return sai_to_ndi_err_translate(sai_ret);
            }

            EV_LOG(INFO, NDI, ev_log_s_MINOR, "NDI-MAC", "NDI_MAC_DEL_ALL_ENTRIES Success -"
                    "entry_id=%d entry_type=%d",
                    fdb_flush_attr[0].id, fdb_flush_attr[0].value.s32);

            break;

        default:
            EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Not all delete cases are handled yet.");
            break;
    }

    EV_LOG(TRACE, NDI, ev_log_s_MINOR, "NDI-MAC", "Successfully returned from %s", __FUNCTION__);
    return STD_ERR_OK;
}

t_std_error ndi_mac_event_notify_register(ndi_mac_event_notification_fn reg_fn)
{
    t_std_error ret_code = STD_ERR_OK;
    npu_id_t npu_id = ndi_npu_id_get();
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        EV_LOG(ERR, NDI, ev_log_s_CRITICAL, "NDI-MAC", "Failed to retrive NDI DB pointer for npu %d", npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }
    STD_ASSERT(reg_fn != NULL);

    ndi_db_ptr->switch_notification->mac_event_notify_cb = reg_fn;

    return ret_code;
}

t_std_error ndi_get_mac_entry_attr(ndi_mac_entry_t *mac_entry)
{
    return STD_ERR_OK;
}

t_std_error ndi_set_mac_entry_attr(ndi_mac_entry_t *mac_entry)
{
    return STD_ERR_OK;
}
