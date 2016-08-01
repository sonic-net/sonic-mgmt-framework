
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
 * filename: nas_ndi_port_map.h
 */


#ifndef _NAS_NDI_PORT_MAP_H_
#define _NAS_NDI_PORT_MAP_H_

#include <stdio.h>
#include "std_error_codes.h"
#include "ds_common_types.h"
#include "dell-base-if-phy.h"
#include "sai.h"

#define NDI_MAX_HWPORT_PER_PORT  10

#ifdef __cplusplus
extern "C"{
#endif


size_t ndi_max_npu_get(void);

BASE_IF_PHY_BREAKOUT_MODE_t sai_break_to_ndi_break( int32_t sai_mode);

t_std_error ndi_sai_port_hwport_list_get(npu_id_t npu_id, sai_object_id_t sai_port, uint32_t *hwport_list, uint32_t *count);

t_std_error ndi_port_map_sai_port_add(npu_id_t npu, sai_object_id_t sai_port, npu_port_t *npu_port);

t_std_error ndi_port_map_sai_port_delete(npu_id_t npu, sai_object_id_t sai_port, npu_port_t *npu_port);

npu_id_t ndi_saiport_to_npu_id_get(sai_object_id_t sai_port);

t_std_error ndi_sai_port_map_create(void);

t_std_error ndi_npu_port_id_get(sai_object_id_t sai_port, npu_id_t *npu_id, npu_port_t *port_id);

void ndi_port_map_table_dump(void);

void ndi_saiport_map_table_dump(void);

#ifdef __cplusplus
}
#endif



#endif
