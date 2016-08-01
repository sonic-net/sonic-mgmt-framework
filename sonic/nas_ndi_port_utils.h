
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
 * nas_ndi_port_utils.h
 */

#ifndef _NAS_NDI_PORT_UTILS_H_
#define _NAS_NDI_PORT_UTILS_H_

#include "dell-base-phy-interface.h"
#include "nas_ndi_port.h"
#include "saiport.h"

#ifdef __cplusplus
extern "C"{
#endif

sai_port_fdb_learning_mode_t ndi_port_get_sai_mac_learn_mode
                             (BASE_PORT_MAC_LEARN_MODE_t ndi_fdb_learn_mode);


BASE_PORT_MAC_LEARN_MODE_t ndi_port_get_mac_learn_mode
                             (sai_port_fdb_learning_mode_t sai_fdb_learn_mode);

sai_port_internal_loopback_mode_t ndi_port_get_sai_loopback_mode(BASE_CMN_LOOPBACK_TYPE_t lpbk_mode);
BASE_CMN_LOOPBACK_TYPE_t ndi_port_get_ndi_loopback_mode(sai_port_internal_loopback_mode_t lpbk_mode);
bool ndi_port_get_sai_speed(BASE_IF_SPEED_t speed, uint32_t *sai_speed);
bool ndi_port_get_ndi_speed(uint32_t sai_speed, BASE_IF_SPEED_t *ndi_speed);

#ifdef __cplusplus
}
#endif

#endif  /*  _NAS_NDI_PORT_UTILS_H_ */
