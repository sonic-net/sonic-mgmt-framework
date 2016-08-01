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
 * filename: nas_ndi_router_interface.c
 * Contains NAS NDI Virtual Router and Router Interface functionality
 */

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include "stdarg.h"
#include "std_error_codes.h"
#include <string.h>
#include "std_error_codes.h"
#include "std_assert.h"
#include "ds_common_types.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_int.h"
#include "nas_ndi_port.h"
#include "nas_ndi_router_interface.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "sairouterintf.h"
#include "sairouter.h"
#include "saistatus.h"
#include "saitypes.h"


/*  Router Interface APIs  */
static inline  sai_router_interface_api_t *ndi_rif_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_route_interface_api_tbl);
}


t_std_error ndi_rif_create (ndi_rif_entry_t *rif_entry, ndi_rif_id_t *rif_id)
{
    uint32_t                  attr_idx = 0;
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t           sai_attr[NDI_MAX_RIF_ATTR];
    sai_object_id_t           sai_port;
    t_std_error rc;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(rif_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);


    sai_attr[attr_idx].value.oid = (rif_entry->vrf_id);
    sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_VIRTUAL_ROUTER_ID;
    attr_idx++;

    switch(rif_entry->rif_type) {
        case NDI_RIF_TYPE_PORT:
        case NDI_RIF_TYPE_LAG:
            sai_attr[attr_idx].value.s32 = SAI_ROUTER_INTERFACE_TYPE_PORT;
            sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_TYPE;
            attr_idx++;

            if (rif_entry->rif_type == NDI_RIF_TYPE_PORT) {
                if ((rc = ndi_sai_port_id_get(rif_entry->attachment.port_id.npu_id,
                                                    rif_entry->attachment.port_id.npu_port,
                                                    &sai_port)) != STD_ERR_OK) {
                NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "SAI port id get "
                              "failed for NPU-id:%d NPU-port:%d",
                              rif_entry->attachment.port_id.npu_id,
                              rif_entry->attachment.port_id.npu_port);
                }
                sai_attr[attr_idx].value.oid = sai_port;
            } else {
                sai_attr[attr_idx].value.oid = rif_entry->attachment.lag_id;
            }

            sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_PORT_ID;
            attr_idx++;
            break;

        case NDI_RIF_TYPE_VLAN:
            sai_attr[attr_idx].value.s32 = SAI_ROUTER_INTERFACE_TYPE_VLAN;
            sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_TYPE;
            attr_idx++;

            sai_attr[attr_idx].value.u16 = rif_entry->attachment.vlan_id;
            sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_VLAN_ID;
            attr_idx++;
            break;
        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "Invalid attribute");
            return STD_ERR(ROUTE, FAIL, 0);
    }

    if (rif_entry->flags & NDI_RIF_ATTR_ADMIN_V4_STATE) {
        sai_attr[attr_idx].value.booldata = (rif_entry->v4_admin_state);
        sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_ADMIN_V4_STATE;
        attr_idx++;
    }
    if (rif_entry->flags & NDI_RIF_ATTR_ADMIN_V6_STATE) {
        sai_attr[attr_idx].value.booldata = (rif_entry->v6_admin_state);
        sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_ADMIN_V6_STATE;
        attr_idx++;
    }
    if (rif_entry->flags & NDI_RIF_ATTR_SRC_MAC_ADDRESS) {
        sai_attr[attr_idx].value.booldata = (rif_entry->v6_admin_state);
        memcpy (sai_attr[attr_idx].value.mac, rif_entry->src_mac,
                HAL_MAC_ADDR_LEN);
        sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_SRC_MAC_ADDRESS;
        attr_idx++;
    }
    if (rif_entry->flags & NDI_RIF_ATTR_MTU) {
        sai_attr[attr_idx].value.u32 = (rif_entry->mtu);
        sai_attr[attr_idx].id = SAI_ROUTER_INTERFACE_ATTR_MTU;
        attr_idx++;
    }


    if ((sai_ret = ndi_rif_api_get(ndi_db_ptr)->create_router_interface
                                                (rif_id, attr_idx, sai_attr))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_rif_delete(npu_id_t npu_id, ndi_rif_id_t rif_id)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_rif_api_get(ndi_db_ptr)->remove_router_interface (rif_id))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_rif_set_attribute (ndi_rif_entry_t *rif_entry)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t           sai_attr;
    sai_object_id_t           sai_port;
    t_std_error               rc;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(rif_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);


   switch (rif_entry->flags) {
        case NDI_RIF_ATTR_VIRTUAL_ROUTER_ID:
            sai_attr.value.oid = rif_entry->vrf_id;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_VIRTUAL_ROUTER_ID;
            break;

        case NDI_RIF_ATTR_TYPE:
            sai_attr.value.s32 = rif_entry->rif_type;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_TYPE;
            break;

        case NDI_RIF_ATTR_PORT_ID:
            if ((rc = ndi_sai_port_id_get(rif_entry->attachment.port_id.npu_id,
                                                rif_entry->attachment.port_id.npu_port,
                                                &sai_port)) != STD_ERR_OK) {
                NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "SAI port id get "
                              "failed for NPU-id:%d NPU-port:%d",
                              rif_entry->attachment.port_id.npu_id,
                              rif_entry->attachment.port_id.npu_port);
            }

            sai_attr.value.oid = sai_port;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_PORT_ID;
            break;

        case NDI_RIF_ATTR_VLAN_ID:
            sai_attr.value.u16 = rif_entry->attachment.vlan_id;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_VLAN_ID;
            break;

        case NDI_RIF_ATTR_SRC_MAC_ADDRESS:
            memcpy (sai_attr.value.mac, rif_entry->src_mac,
                    HAL_MAC_ADDR_LEN);
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_SRC_MAC_ADDRESS;
            break;

        case NDI_RIF_ATTR_ADMIN_V4_STATE:
            sai_attr.value.booldata = rif_entry->v4_admin_state;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_ADMIN_V4_STATE;
            break;

        case NDI_RIF_ATTR_ADMIN_V6_STATE:
            sai_attr.value.booldata = rif_entry->v6_admin_state;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_ADMIN_V6_STATE;
            break;

        case NDI_RIF_ATTR_MTU:
            sai_attr.value.u32 = rif_entry->mtu;
            sai_attr.id = SAI_ROUTER_INTERFACE_ATTR_MTU;
            break;

        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "Invalid attribute");
            return STD_ERR(ROUTE, FAIL, 0);
    }

    if ((sai_ret = ndi_rif_api_get(ndi_db_ptr)->set_router_interface_attribute(
                       rif_entry->rif_id, &sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}


t_std_error ndi_rif_get_attribute (ndi_rif_entry_t *rif_entry)
{

    sai_attribute_t           sai_attr[NDI_MAX_RIF_ATTR];
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    ndi_port_t                ndi_port;
    uint32_t                  attr_count = 1;
    uint32_t                  attr_idx = 0;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(rif_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_rif_api_get(ndi_db_ptr)->get_router_interface_attribute(
                       rif_entry->rif_id, attr_count, sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    for(attr_idx = 0; attr_idx < attr_count; attr_idx++) {
        switch(sai_attr[attr_idx].id) {
            case SAI_ROUTER_INTERFACE_ATTR_VIRTUAL_ROUTER_ID:
                 rif_entry->vrf_id = sai_attr[attr_idx].value.oid;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_TYPE:
                 rif_entry->rif_type = sai_attr[attr_idx].value.u32;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_PORT_ID:
                 ndi_npu_port_id_get(sai_attr[attr_idx].value.oid,
                                      &ndi_port.npu_id, &ndi_port.npu_port);
                 rif_entry->attachment.port_id.npu_id = ndi_port.npu_id;
                 rif_entry->attachment.port_id.npu_port = ndi_port.npu_port;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_VLAN_ID:
                 rif_entry->attachment.vlan_id = sai_attr[attr_idx].value.u16;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_SRC_MAC_ADDRESS:
                 memcpy (rif_entry->src_mac, sai_attr[attr_idx].value.mac,
                         HAL_MAC_ADDR_LEN);
                break;
            case  SAI_ROUTER_INTERFACE_ATTR_ADMIN_V4_STATE:
                 rif_entry->v4_admin_state = sai_attr[attr_idx].value.booldata;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_ADMIN_V6_STATE:
                 rif_entry->v6_admin_state = sai_attr[attr_idx].value.booldata;
                 break;
            case SAI_ROUTER_INTERFACE_ATTR_MTU:
                 rif_entry->mtu = sai_attr[attr_idx].value.u32;
                 break;
            default:
                NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "Invalid get attribute");
                return STD_ERR(ROUTE, FAIL, 0);
        }
    }
    return STD_ERR_OK;
}


/*  Virtual Router specific APIs  */
static inline  sai_virtual_router_api_t *ndi_route_vr_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_virtual_router_api_tbl);
}


t_std_error ndi_route_vr_create (ndi_vr_entry_t *vr_entry, ndi_vrf_id_t *vrf_id)
{
    sai_status_t     sai_ret = SAI_STATUS_FAILURE;
    unsigned int     attr_idx = 0;
    sai_attribute_t  sai_attr[NDI_MAX_RIF_ATTR];

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(vr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if (vr_entry->flags & NDI_VR_ATTR_ADMIN_V4_STATE) {
        sai_attr[attr_idx].value.booldata = (vr_entry->v4_admin_state);
        sai_attr[attr_idx].id = SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V4_STATE;
        attr_idx++;
    }

    if (vr_entry->flags & NDI_VR_ATTR_ADMIN_V6_STATE) {
        sai_attr[attr_idx].value.booldata = (vr_entry->v6_admin_state);
        sai_attr[attr_idx].id = SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V6_STATE;
        attr_idx++;
    }

    if (vr_entry->flags & NDI_VR_ATTR_SRC_MAC_ADDRESS) {
            memcpy(sai_attr[attr_idx].value.mac, vr_entry->src_mac,
                    HAL_MAC_ADDR_LEN);
            sai_attr[attr_idx].id = SAI_VIRTUAL_ROUTER_ATTR_SRC_MAC_ADDRESS;
            attr_idx++;
    }

    if (vr_entry->flags & NDI_VR_ATTR_VIOLATION_TTL1_ACTION) {
        sai_attr[attr_idx].value.s32 = (vr_entry->ttl1_action);
        sai_attr[attr_idx].id = SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_TTL1_ACTION;
        attr_idx++;
    }

    if (vr_entry->flags & NDI_VR_ATTR_VIOLATION_IP_OPTIONS) {
        sai_attr[attr_idx].value.s32 = (vr_entry->violation_ip_options);
        sai_attr[attr_idx].id = SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_IP_OPTIONS;
        attr_idx++;
    }

    if ((sai_ret = ndi_route_vr_api_get(ndi_db_ptr)->create_virtual_router
                       (vrf_id, attr_idx, sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_vr_delete(npu_id_t npu_id, ndi_vrf_id_t vrf_id)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_route_vr_api_get(ndi_db_ptr)->remove_virtual_router (vrf_id))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_vr_set_attribute (ndi_vr_entry_t *vr_entry)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t           sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(vr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);


   switch (vr_entry->flags) {
        case NDI_VR_ATTR_ADMIN_V4_STATE:
            sai_attr.value.booldata = vr_entry->v4_admin_state;
            sai_attr.id = SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V4_STATE;
            break;

        case NDI_VR_ATTR_ADMIN_V6_STATE:
            sai_attr.value.booldata = vr_entry->v6_admin_state;
            sai_attr.id = SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V6_STATE;
            break;

        case NDI_VR_ATTR_SRC_MAC_ADDRESS:
            memcpy (sai_attr.value.mac, vr_entry->src_mac,
                    HAL_MAC_ADDR_LEN);
            sai_attr.id = SAI_VIRTUAL_ROUTER_ATTR_SRC_MAC_ADDRESS;
            break;

        case NDI_VR_ATTR_VIOLATION_TTL1_ACTION:
            sai_attr.value.s32 = vr_entry->ttl1_action;
            sai_attr.id = SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_TTL1_ACTION;
            break;
        case NDI_VR_ATTR_VIOLATION_IP_OPTIONS:
            sai_attr.value.s32 = vr_entry->violation_ip_options;
            sai_attr.id = SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_IP_OPTIONS;
            break;

        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "Invalid attribute");
            return STD_ERR(ROUTE, FAIL, 0);
    }

    if ((sai_ret = ndi_route_vr_api_get(ndi_db_ptr)->set_virtual_router_attribute(
                                vr_entry->vrf_id, &sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}


t_std_error ndi_route_vr_get_attribute (ndi_vr_entry_t *vr_entry)
{

    sai_attribute_t           sai_attr[NDI_MAX_VR_ATTR];
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    int                       attr_count = 1;
    uint32_t                  attr_idx = 0;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(vr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_route_vr_api_get(ndi_db_ptr)->get_virtual_router_attribute(
                                vr_entry->vrf_id, attr_count, sai_attr)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    for(attr_idx = 0; attr_idx < attr_count; attr_idx++) {
        switch(sai_attr[attr_idx].id) {
            case SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V4_STATE:
                 vr_entry->v4_admin_state = sai_attr[attr_idx].value.booldata;
                 break;
            case SAI_VIRTUAL_ROUTER_ATTR_ADMIN_V6_STATE:
                 vr_entry->v6_admin_state = sai_attr[attr_idx].value.booldata;
                 break;
            case SAI_VIRTUAL_ROUTER_ATTR_SRC_MAC_ADDRESS:
                 memcpy (vr_entry->src_mac, sai_attr[attr_idx].value.mac,
                         HAL_MAC_ADDR_LEN);
                 break;
            case SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_TTL1_ACTION:
                 vr_entry->ttl1_action = sai_attr[attr_idx].value.s32;
                 break;
            case SAI_VIRTUAL_ROUTER_ATTR_VIOLATION_IP_OPTIONS:
                 vr_entry->violation_ip_options = sai_attr[attr_idx].value.s32;
                 break;
            default:
                 NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-RIF", "Invalid get attribute");
                 return STD_ERR(ROUTE, FAIL, 0);
        }
    }
    return STD_ERR_OK;
}
