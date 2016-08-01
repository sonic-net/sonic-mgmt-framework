
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
 * nas_ndi_utils.h
 */

#ifndef _NAS_NDI_UTILS_H_
#define _NAS_NDI_UTILS_H_

#include "std_error_codes.h"
#include "ds_common_types.h"
#include "nas_ndi_common.h"
#include "nas_ndi_int.h"
#include "nas_ndi_port_map.h"
#include "saitypes.h"

#ifdef __cplusplus
extern "C"{
#endif

// NDI to SAI Object ID conversion routines
// for objects that are shared by multiple modules
#define ndi_utl_ndi2sai_mirror_id(x)     (sai_object_id_t) (x)
#define ndi_utl_ndi2sai_policer_id(x)    (sai_object_id_t) (x)
#define ndi_utl_ndi2sai_ip_nexthop_id(x) (sai_object_id_t) (x)
#define ndi_utl_ndi2sai_cpu_queue_id(x)  (sai_object_id_t) (x)

t_std_error ndi_db_global_tbl_alloc(size_t max_npu);

nas_ndi_db_t *ndi_db_ptr_get(npu_id_t npu_id);

npu_id_t ndi_npu_id_get(void);


t_std_error ndi_sai_vlan_id_get(npu_id_t npu_id, hal_vlan_id_t vlan_id, sai_vlan_id_t *sai_vlan);

ndi_switch_oper_status_t ndi_oper_status_translate(sai_switch_oper_status_t oper_status);

bool ndi_to_sai_if_stats(ndi_stat_id_t ndi_id, sai_port_stat_counter_t * sai_id);

bool ndi_to_sai_vlan_stats(ndi_stat_id_t ndi_id, sai_vlan_stat_counter_t * sai_id);

t_std_error ndi_switch_state_change_cb_register(npu_id_t npu_id,
                      sai_switch_state_change_notification_fn sw_state_change_cb);

t_std_error ndi_fdb_event_cb_register(npu_id_t npu_id,
                      sai_fdb_event_notification_fn fdb_event_cb);

t_std_error ndi_sai_oper_state_to_link_state_get(sai_port_oper_status_t sai_port_state,
                       ndi_port_oper_status_t *p_state);

t_std_error ndi_port_state_change_cb_register(npu_id_t npu_id,
                      sai_port_state_change_notification_fn port_state_change_cb);

t_std_error ndi_switch_shutdown_request_cb_register (npu_id_t npu_id,
                      sai_switch_shutdown_request_fn switch_shutdown_request_cb);

void ndi_packet_rx_cb(const void *buffer, sai_size_t buffer_size, uint32_t attr_count,
                                  const sai_attribute_t *attr_list);

bool ndi_port_to_sai_oid(ndi_port_t * ndi_port, sai_object_id_t *oid);

t_std_error ndi_sai_port_id_get(npu_id_t npu_id, npu_port_t ndi_port, sai_object_id_t *sai_port);

static inline t_std_error ndi_utl_mk_std_err (enum e_std_error_subsystems sub,
                                              sai_status_t st)
{
    return ((st == SAI_STATUS_TABLE_FULL) || (st == SAI_STATUS_INSUFFICIENT_RESOURCES)) ?
             STD_ERR_MK (sub, e_std_err_code_NORESOURCE, 0) :
             STD_ERR_MK (sub, e_std_err_code_FAIL, 0);
}

#ifdef __cplusplus
}
#endif

#endif  /*  _NAS_NDI_UTILS_H_ */


