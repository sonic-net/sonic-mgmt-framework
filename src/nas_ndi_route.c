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
 * filename: nas_ndi_route.c
 */

#include <stdio.h>
#include <string.h>
#include "std_error_codes.h"
#include "std_assert.h"
#include "std_ip_utils.h"
#include "ds_common_types.h"
#include "nas_ndi_event_logs.h"
#include "nas_ndi_int.h"
#include "nas_ndi_route.h"
#include "nas_ndi_utils.h"
#include "sai.h"
#include "saistatus.h"
#include "saitypes.h"


/*  NDI Route/Neighbor specific APIs  */

static inline  sai_route_api_t *ndi_route_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_route_api_tbl);
}

static inline  sai_neighbor_api_t *ndi_neighbor_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_neighbor_api_tbl);
}

static inline  sai_next_hop_api_t *ndi_next_hop_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_next_hop_api_tbl);
}

static inline void ndi_sai_ip_address_copy(sai_ip_address_t *sai_ip_addr,const hal_ip_addr_t *ip_addr)
{
    if (STD_IP_IS_AFINDEX_V4(ip_addr->af_index)) {
        sai_ip_addr->addr_family = SAI_IP_ADDR_FAMILY_IPV4;
        sai_ip_addr->addr.ip4 = ip_addr->u.v4_addr;
    } else {
        sai_ip_addr->addr_family = SAI_IP_ADDR_FAMILY_IPV6;
        memcpy (sai_ip_addr->addr.ip6, ip_addr->u.v6_addr, sizeof (sai_ip6_t));
    }
}

static void ndi_route_params_copy(sai_unicast_route_entry_t *p_sai_route,
                                  ndi_route_t *p_route_entry)
{
    sai_ip_prefix_t *p_ip_prefix = NULL;
    hal_ip_addr_t    ip_mask;
    uint32_t         af_index;

    p_sai_route->vr_id = p_route_entry->vrf_id;

    p_ip_prefix = &p_sai_route->destination;

    af_index = p_route_entry->prefix.af_index;
    p_ip_prefix->addr_family = (STD_IP_IS_AFINDEX_V4(af_index))?
                                SAI_IP_ADDR_FAMILY_IPV4:SAI_IP_ADDR_FAMILY_IPV6;

    if (STD_IP_IS_AFINDEX_V4(af_index)) {
        p_ip_prefix->addr.ip4 = p_route_entry->prefix.u.v4_addr;
    } else {
        memcpy (p_ip_prefix->addr.ip6, p_route_entry->prefix.u.v6_addr, sizeof (sai_ip6_t));
    }

    std_ip_get_mask_from_prefix_len (af_index, p_route_entry->mask_len, &ip_mask);

    if (STD_IP_IS_AFINDEX_V4(af_index)) {
        p_ip_prefix->mask.ip4 = ip_mask.u.v4_addr;
    } else {
        memcpy (p_ip_prefix->mask.ip6, ip_mask.u.v6_addr, sizeof (sai_ip6_t));
    }
    return;
}

static sai_packet_action_t ndi_route_sai_action_get(ndi_route_action action)
{
    sai_packet_action_t sai_action;

    switch(action) {
        case NDI_ROUTE_PACKET_ACTION_FORWARD:
            sai_action = SAI_PACKET_ACTION_FORWARD;
            break;
        case NDI_ROUTE_PACKET_ACTION_TRAPCPU:
            sai_action = SAI_PACKET_ACTION_TRAP;
            break;
        case NDI_ROUTE_PACKET_ACTION_DROP:
            sai_action = SAI_PACKET_ACTION_DROP;
            break;
        case NDI_ROUTE_PACKET_ACTION_TRAPFORWARD:
            sai_action = SAI_PACKET_ACTION_LOG;
            break;
        default:
            sai_action = SAI_PACKET_ACTION_FORWARD;
            break;
    }
    return sai_action;
}

t_std_error ndi_route_add (ndi_route_t *p_route_entry)
{
    uint32_t                  attr_idx = 0;
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_unicast_route_entry_t sai_route;
    sai_attribute_t           sai_attr[NDI_MAX_ROUTE_ATTR];

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_route_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    ndi_route_params_copy(&sai_route, p_route_entry);

    sai_attr[attr_idx].value.s32 = ndi_route_sai_action_get(p_route_entry->action);
    sai_attr[attr_idx].id = SAI_ROUTE_ATTR_PACKET_ACTION;
    attr_idx++;

    if(p_route_entry->nh_handle){
        sai_attr[attr_idx].value.oid = p_route_entry->nh_handle;
        /*
         * Attribute type is same for both ECMP or non-ECMP case
         */
        sai_attr[attr_idx].id = SAI_ROUTE_ATTR_NEXT_HOP_ID;
        attr_idx++;
    }
    if ((sai_ret = ndi_route_api_get(ndi_db_ptr)->create_route(&sai_route, attr_idx, sai_attr))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_delete (ndi_route_t *p_route_entry)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_unicast_route_entry_t sai_route;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_route_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    ndi_route_params_copy(&sai_route, p_route_entry);

    if ((sai_ret = ndi_route_api_get(ndi_db_ptr)->remove_route(&sai_route))!= SAI_STATUS_SUCCESS){
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    return STD_ERR_OK;
}

t_std_error ndi_route_set_attribute (ndi_route_t *p_route_entry)
{
    sai_status_t              sai_ret = SAI_STATUS_FAILURE;
    sai_unicast_route_entry_t sai_route;
    sai_attribute_t           sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_route_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    ndi_route_params_copy(&sai_route, p_route_entry);

    switch(p_route_entry->flags) {
        case NDI_ROUTE_L3_PACKET_ACTION:
            sai_attr.value.s32 = ndi_route_sai_action_get(p_route_entry->action);
            sai_attr.id = SAI_ROUTE_ATTR_PACKET_ACTION;
            break;
        case NDI_ROUTE_L3_TRAP_PRIORITY:
            sai_attr.value.u8 = p_route_entry->priority;
            sai_attr.id = SAI_ROUTE_ATTR_TRAP_PRIORITY;
            break;
        case NDI_ROUTE_L3_NEXT_HOP_ID:
            sai_attr.value.oid = p_route_entry->nh_handle;
            sai_attr.id = SAI_ROUTE_ATTR_NEXT_HOP_ID;
            break;
        case NDI_ROUTE_L3_ECMP:
            sai_attr.value.oid = p_route_entry->nh_handle;
            sai_attr.id = SAI_ROUTE_ATTR_NEXT_HOP_ID;
            break;
        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE", "Invalid attribute");
            return STD_ERR(ROUTE, FAIL, 0);
    }

    if ((sai_ret = ndi_route_api_get(ndi_db_ptr)->set_route_attribute(&sai_route, &sai_attr))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_next_hop_add (ndi_neighbor_t *p_nbr_entry, next_hop_id_t *nh_handle)
{
    uint32_t          attr_idx = 0;
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t   sai_nh_id;
    sai_attribute_t   sai_attr[NDI_MAX_NEXT_HOP_ATTR];

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nbr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    sai_attr[attr_idx].value.s32 = SAI_NEXT_HOP_IP;
    sai_attr[attr_idx].id = SAI_NEXT_HOP_ATTR_TYPE;
    attr_idx++;

    ndi_sai_ip_address_copy(&sai_attr[attr_idx].value.ipaddr, &p_nbr_entry->ip_addr);
    sai_attr[attr_idx].id = SAI_NEXT_HOP_ATTR_IP;
    attr_idx++;

    sai_attr[attr_idx].value.oid = p_nbr_entry->rif_id;
    sai_attr[attr_idx].id = SAI_NEXT_HOP_ATTR_ROUTER_INTERFACE_ID;
    attr_idx++;

    if ((sai_ret = ndi_next_hop_api_get(ndi_db_ptr)->create_next_hop(&sai_nh_id, attr_idx, sai_attr))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    *nh_handle = sai_nh_id;

    return STD_ERR_OK;
}

t_std_error ndi_route_next_hop_delete (npu_id_t npu_id, next_hop_id_t nh_handle)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    if ((sai_ret = ndi_next_hop_api_get(ndi_db_ptr)->remove_next_hop(nh_handle))
                          != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    return STD_ERR_OK;
}

t_std_error ndi_route_neighbor_add (ndi_neighbor_t *p_nbr_entry)
{
    uint32_t             attr_idx = 0;
    sai_status_t         sai_ret = SAI_STATUS_FAILURE;
    sai_neighbor_entry_t sai_nbr_entry;
    sai_attribute_t      sai_attr[NDI_MAX_NEIGHBOR_ATTR];

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nbr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    sai_nbr_entry.rif_id = p_nbr_entry->rif_id;
    ndi_sai_ip_address_copy(&sai_nbr_entry.ip_address, &p_nbr_entry->ip_addr);

    memcpy (sai_attr[attr_idx].value.mac, p_nbr_entry->egress_data.neighbor_mac,
            HAL_MAC_ADDR_LEN);
    sai_attr[attr_idx].id = SAI_NEIGHBOR_ATTR_DST_MAC_ADDRESS;
    attr_idx++;

    sai_attr[attr_idx].value.s32 = ndi_route_sai_action_get(p_nbr_entry->action);
    sai_attr[attr_idx].id = SAI_NEIGHBOR_ATTR_PACKET_ACTION;
    attr_idx++;

    if ((sai_ret = ndi_neighbor_api_get(ndi_db_ptr)->create_neighbor_entry(&sai_nbr_entry,
                                                     attr_idx, sai_attr))!= SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_neighbor_delete (ndi_neighbor_t *p_nbr_entry)
{
    sai_status_t         sai_ret = SAI_STATUS_FAILURE;
    sai_neighbor_entry_t sai_nbr_entry;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nbr_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    sai_nbr_entry.rif_id = p_nbr_entry->rif_id;
    ndi_sai_ip_address_copy(&sai_nbr_entry.ip_address, &p_nbr_entry->ip_addr);

    if ((sai_ret = ndi_neighbor_api_get(ndi_db_ptr)->remove_neighbor_entry(&sai_nbr_entry))
                                        != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}


/*
 * NAS NDI Nexthop Group APIS for ECMP Functionality
 */
static inline  sai_next_hop_group_api_t *ndi_next_hop_group_api_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_next_hop_group_api_tbl);

}

t_std_error ndi_route_next_hop_group_create (ndi_nh_group_t *p_nh_group_entry,
                        next_hop_id_t *nh_group_handle)
{
    uint32_t          attr_idx = 0;
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t   sai_nh_group_id;
    sai_object_id_t   nexthops[NDI_MAX_NH_ENTRIES_PER_GROUP];
    sai_attribute_t   sai_attr[NDI_MAX_GROUP_NEXT_HOP_ATTR];
    uint32_t          nhop_count;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nh_group_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    sai_attr[attr_idx].value.s32 = SAI_NEXT_HOP_GROUP_ECMP;
    sai_attr[attr_idx].id = SAI_NEXT_HOP_GROUP_ATTR_TYPE;
    attr_idx++;


    nhop_count = p_nh_group_entry->nhop_count;
    /*
     * Add the nexthop id list to sai_next_hop_list_t
     */
    if (nhop_count == 0) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    /*
     * Copy nexthop-id to list
     */
    int i;
    for (i = 0; i <nhop_count; i++) {
        nexthops[i] = p_nh_group_entry->nh_list[i].id;
    }
    /* sai_next_hop_list_t */
    sai_attr[attr_idx].value.objlist.count = nhop_count;
    sai_attr[attr_idx].value.objlist.list = nexthops;
    sai_attr[attr_idx].id = SAI_NEXT_HOP_GROUP_ATTR_NEXT_HOP_LIST;
    attr_idx++;

    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
            create_next_hop_group(&sai_nh_group_id, attr_idx, sai_attr))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    *nh_group_handle = sai_nh_group_id;
    return STD_ERR_OK;
}

t_std_error ndi_route_next_hop_group_delete (npu_id_t npu_id,
                                    next_hop_id_t nh_handle)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t next_hop_group_id;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    next_hop_group_id = nh_handle;

    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
        remove_next_hop_group(next_hop_group_id)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    return STD_ERR_OK;
}

t_std_error ndi_route_set_next_hop_group_attribute (ndi_nh_group_t *p_nh_group_entry,
                                        next_hop_id_t nh_group_handle)
{
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t   sai_nh_group_id;
    sai_attribute_t   sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nh_group_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);


    switch(p_nh_group_entry->flags) {
        case NDI_ROUTE_NH_GROUP_ATTR_TYPE:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                                                    "Invalid attribute: Create-only");
            return STD_ERR(ROUTE, FAIL, sai_ret);
        case NDI_ROUTE_NH_GROUP_ATTR_NEXT_HOP_LIST:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                           "Invalid attribute: Create-only");
            return STD_ERR(ROUTE, FAIL, sai_ret);
        default:
            NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                                        "Invalid attribute");
            break;
    }
    sai_nh_group_id  = nh_group_handle;
    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
            set_next_hop_group_attribute(sai_nh_group_id, &sai_attr))
            != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    return STD_ERR_OK;
}

t_std_error ndi_route_get_next_hop_group_attribute (ndi_nh_group_t *p_nh_group_entry,
                                                    next_hop_id_t nh_group_handle)
{
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    uint32_t          attr_count= 1;
    sai_attribute_t   sai_attr[NDI_MAX_GROUP_NEXT_HOP_ATTR];
    sai_object_id_t   sai_nh_group_id;
    unsigned int      attr_idx = 0;
    sai_object_id_t  *next_hops;
    uint32_t          nhop_count;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nh_group_entry->npu_id);

    sai_nh_group_id  = nh_group_handle;
    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
            get_next_hop_group_attribute(sai_nh_group_id,attr_count, sai_attr))
            != SAI_STATUS_SUCCESS) {
        NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                      "get attribute failed");
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    for(attr_idx = 0; attr_idx < attr_count; attr_idx++) {
        switch(sai_attr[attr_idx].id) {
            case SAI_NEXT_HOP_GROUP_ATTR_NEXT_HOP_COUNT:
                p_nh_group_entry->nhop_count = sai_attr[attr_idx].value.u32;
                break;
            case SAI_NEXT_HOP_GROUP_ATTR_TYPE:
                if (sai_attr[attr_idx].value.u32 == SAI_NEXT_HOP_GROUP_ECMP) {
                    p_nh_group_entry->group_type = NDI_ROUTE_NH_GROUP_TYPE_ECMP;
                }
                /* @TODO WECMP case */
                break;
            case SAI_NEXT_HOP_GROUP_ATTR_NEXT_HOP_LIST:
                /*
                 * Get the nexthop id list from sai_next_hop_list_t
                 */
                nhop_count = sai_attr[attr_idx].value.objlist.count;
                if (nhop_count == 0) {
                    NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                                 "Invalid get nhlist attribute");
                    return STD_ERR(ROUTE, FAIL, sai_ret);
                }

                /* nexthops*/
                next_hops = sai_attr[attr_idx].value.objlist.list;
                /*
                 * Copy nexthop-id to list
                 */
                int i;
                for (i = 0; i <nhop_count; i++) {
                    p_nh_group_entry->nh_list[i].id = (*(next_hops+i));
                }

                break;
            default:
                NDI_LOG_TRACE(ev_log_t_NDI, "NDI-ROUTE-NHGROUP",
                             "Invalid get attribute");
                return STD_ERR(ROUTE, FAIL, 0);
        }
    }

    return STD_ERR_OK;
}
/*
 *  ndi_route_add_next_hop_to_group: t all new NH and new NH count
 */
t_std_error ndi_route_add_next_hop_to_group (ndi_nh_group_t *p_nh_group_entry,
                                             next_hop_id_t nh_group_handle)
{
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t   next_hop_group_id;
    sai_object_id_t   nexthops[NDI_MAX_NH_ENTRIES_PER_GROUP];
    uint32_t nhop_count;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nh_group_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);


    /* Set all new NH count */
    nhop_count = p_nh_group_entry->nhop_count;
    /*
     * Add the nexthop id list to SAI nexthps list
     */
    if (nhop_count == 0) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    /*
     * Copy nexthop-id to list
     */
    int i;
    for (i = 0; i <nhop_count; i++) {
        nexthops[i] = p_nh_group_entry->nh_list[i].id;

    }
    next_hop_group_id  = nh_group_handle;

    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
                    add_next_hop_to_group(next_hop_group_id, nhop_count,
                    nexthops)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_route_delete_next_hop_from_group (ndi_nh_group_t *p_nh_group_entry,
                            next_hop_id_t nh_group_handle)
{
    sai_status_t      sai_ret = SAI_STATUS_FAILURE;
    sai_object_id_t   next_hop_group_id;
    sai_object_id_t   nexthops[NDI_MAX_NH_ENTRIES_PER_GROUP];
    uint32_t nhop_count;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(p_nh_group_entry->npu_id);
    STD_ASSERT(ndi_db_ptr != NULL);

    nhop_count = p_nh_group_entry->nhop_count;
    /*
     * Add the nexthop id list to SAI nexthps list
     */
    if (nhop_count == 0) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }
    /*
     * Copy nexthop-id to list
     */
    int i;
    for (i = 0; i <nhop_count; i++) {
        nexthops[i] = p_nh_group_entry->nh_list[i].id;
    }

    next_hop_group_id  = nh_group_handle;
    if ((sai_ret = ndi_next_hop_group_api_get(ndi_db_ptr)->
                    remove_next_hop_from_group(next_hop_group_id, nhop_count,
                    nexthops)) != SAI_STATUS_SUCCESS) {
        return STD_ERR(ROUTE, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}
