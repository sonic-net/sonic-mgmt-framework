
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
 * nas_ndi_int.h
 */

#ifndef _NAS_NDI_INT_H
#define _NAS_NDI_INT_H

#include "std_error_codes.h"
#include "ds_common_types.h"
#include "nas_ndi_port.h"
#include "nas_ndi_common.h"
#include "nas_ndi_mac.h"
#include "sai.h"
#include "saiswitch.h"
#include "saistatus.h"
#include "saitypes.h"
#include "saimirror.h"
#include "saipolicer.h"
#include "saiwred.h"
#include "saiqosmaps.h"
#include "saiqueue.h"
#include "saischeduler.h"
#include "saischedulergroup.h"

#include "nas_ndi_mac.h"

#ifdef __cplusplus
extern "C"{
#endif



/**
 * @class SAI API Table
 * @brief Contains pointer to the SAI API tables for L2 and L3 feature-set.
 */
typedef struct _ndi_sai_api_tbl_t
{
    sai_switch_api_t            *n_sai_switch_api_tbl;
    sai_port_api_t              *n_sai_port_api_tbl;
    sai_fdb_api_t               *n_sai_fdb_api_tbl;
    sai_vlan_api_t              *n_sai_vlan_api_tbl ;
    sai_virtual_router_api_t    *n_sai_virtual_router_api_tbl;
    sai_route_api_t             *n_sai_route_api_tbl;
    sai_next_hop_api_t          *n_sai_next_hop_api_tbl;
    sai_next_hop_group_api_t    *n_sai_next_hop_group_api_tbl;
    sai_router_interface_api_t  *n_sai_route_interface_api_tbl;
    sai_neighbor_api_t          *n_sai_neighbor_api_tbl;
    sai_policer_api_t           *n_sai_policer_api_tbl;
    sai_wred_api_t              *n_sai_wred_api_tbl;
    sai_qos_map_api_t           *n_sai_qos_map_api_tbl;
    sai_queue_api_t             *n_sai_qos_queue_api_tbl;
    sai_scheduler_api_t         *n_sai_scheduler_api_tbl;
    sai_scheduler_group_api_t   *n_sai_scheduler_group_api_tbl;
    sai_acl_api_t               *n_sai_acl_api_tbl;
    sai_mirror_api_t            *n_sai_mirror_api_tbl;
    sai_stp_api_t               *n_sai_stp_api_tbl;
    sai_lag_api_t               *n_sai_lag_api_tbl;
    sai_samplepacket_api_t      *n_sai_samplepacket_api_tbl;
    sai_hostif_api_t            *n_sai_hostif_api_tbl;
    sai_buffer_api_t            *n_sai_buffer_api_tbl;
} ndi_sai_api_tbl_t;

typedef struct _ndi_switch_notification_t_
{

    /*  port state change callback */
    ndi_port_oper_status_change_fn     port_oper_status_change_cb;

    /*  port event update callback */
    ndi_port_event_update_fn           port_event_update_cb;

    /*  Rx packet callback */
    ndi_packet_rx_type             packet_rx_cb;

    /*  shutdown callback */
    ndi_switch_shutdown_request_fn  switch_shutdown_cb;

    /* mac event notification callback */
    ndi_mac_event_notification_fn mac_event_notify_cb;

} ndi_switch_notification_t;
/**
 * @class NAS NDI DB
 * @brief contains NPU specific information
 */
typedef struct _nas_ndi_db_t {
    sai_switch_profile_id_t npu_profile_id;
    ndi_switch_oper_status_t npu_oper_status;
    const char *npu_vendor_id;
    const char *npu_chip_id;
    const char *microcode;
    void *npu_key_value_tbl;
    service_method_table_t *ndi_services;
    ndi_sai_api_tbl_t ndi_sai_api_tbl; /*  pointer to the SAI API table */
    ndi_switch_notification_t *switch_notification;

} nas_ndi_db_t;

sai_switch_api_t *ndi_sai_switch_api_tbl_get(nas_ndi_db_t *ndi_db_ptr);
#ifdef __cplusplus
}
#endif



#endif
