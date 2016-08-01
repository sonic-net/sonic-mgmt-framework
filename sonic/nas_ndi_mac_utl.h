
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
 * nas_ndi_mac_utl.h
 */

#ifndef _NAS_NDI_MAC_UTL_H_
#define _NAS_NDI_MAC_UTL_H_

#include "std_error_codes.h"
#include "ds_common_types.h"
#include "nas_ndi_common.h"
#include "nas_ndi_int.h"
#include "nas_ndi_port_map.h"
#include "saitypes.h"

#ifdef __cplusplus
extern "C"{
#endif

sai_packet_action_t ndi_mac_sai_packet_action_get(BASE_MAC_PACKET_ACTION_t action);

BASE_MAC_PACKET_ACTION_t ndi_mac_packet_action_get(sai_packet_action_t action);

ndi_mac_event_type_t ndi_mac_event_type_get(sai_fdb_event_t sai_event_type);

#ifdef __cplusplus
}
#endif

#endif  /*  _NAS_NDI_MAC_UTL_H_ */
