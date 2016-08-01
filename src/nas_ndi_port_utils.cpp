
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
 * filename: nas_ndi_port_utils.cpp
 */

#include "dell-base-phy-interface.h"
#include "nas_ndi_port_utils.h"
#include<unordered_map>
#include <algorithm>

static std::unordered_map<BASE_PORT_MAC_LEARN_MODE_t, sai_port_fdb_learning_mode_t,std::hash<int>>
ndi_to_sai_fdb_learn_mode =
{
    {BASE_PORT_MAC_LEARN_MODE_DROP, SAI_PORT_LEARN_MODE_DROP},
    {BASE_PORT_MAC_LEARN_MODE_DISABLE, SAI_PORT_LEARN_MODE_DISABLE},
    {BASE_PORT_MAC_LEARN_MODE_HW, SAI_PORT_LEARN_MODE_HW},
    {BASE_PORT_MAC_LEARN_MODE_CPU_TRAP, SAI_PORT_LEARN_MODE_CPU_TRAP},
    {BASE_PORT_MAC_LEARN_MODE_CPU_LOG, SAI_PORT_LEARN_MODE_CPU_LOG},
};


static std::unordered_map<sai_port_fdb_learning_mode_t, BASE_PORT_MAC_LEARN_MODE_t,std::hash<int>>
sai_to_ndi_fdb_learn_mode =
{
    {SAI_PORT_LEARN_MODE_DROP, BASE_PORT_MAC_LEARN_MODE_DROP},
    {SAI_PORT_LEARN_MODE_DISABLE, BASE_PORT_MAC_LEARN_MODE_DISABLE},
    {SAI_PORT_LEARN_MODE_HW, BASE_PORT_MAC_LEARN_MODE_HW},
    {SAI_PORT_LEARN_MODE_CPU_TRAP, BASE_PORT_MAC_LEARN_MODE_CPU_TRAP},
    {SAI_PORT_LEARN_MODE_CPU_LOG, BASE_PORT_MAC_LEARN_MODE_CPU_LOG},
};

sai_port_fdb_learning_mode_t ndi_port_get_sai_mac_learn_mode
                             (BASE_PORT_MAC_LEARN_MODE_t ndi_fdb_learn_mode){

    sai_port_fdb_learning_mode_t mode = SAI_PORT_LEARN_MODE_HW;

    auto it = ndi_to_sai_fdb_learn_mode.find(ndi_fdb_learn_mode);

    if(it != ndi_to_sai_fdb_learn_mode.end()){
        mode = it->second;
    }

    return mode;
}


BASE_PORT_MAC_LEARN_MODE_t ndi_port_get_mac_learn_mode
                             (sai_port_fdb_learning_mode_t sai_fdb_learn_mode){

    BASE_PORT_MAC_LEARN_MODE_t mode = BASE_PORT_MAC_LEARN_MODE_HW;

    auto it = sai_to_ndi_fdb_learn_mode.find(sai_fdb_learn_mode);

    if(it != sai_to_ndi_fdb_learn_mode.end()){
        mode = it->second;
    }

    return mode;
}
static std::unordered_map<BASE_CMN_LOOPBACK_TYPE_t, sai_port_internal_loopback_mode_t,std::hash<int>>
ndi2sai_int_loopback_mode =
{
    {BASE_CMN_LOOPBACK_TYPE_NONE, SAI_PORT_INTERNAL_LOOPBACK_NONE},
    {BASE_CMN_LOOPBACK_TYPE_PHY, SAI_PORT_INTERNAL_LOOPBACK_PHY},
    {BASE_CMN_LOOPBACK_TYPE_MAC, SAI_PORT_INTERNAL_LOOPBACK_MAC},
};

static std::unordered_map<sai_port_internal_loopback_mode_t,BASE_CMN_LOOPBACK_TYPE_t,std::hash<int>>
sai2ndi_int_loopback_mode =
{
    {SAI_PORT_INTERNAL_LOOPBACK_NONE, BASE_CMN_LOOPBACK_TYPE_NONE},
    {SAI_PORT_INTERNAL_LOOPBACK_PHY, BASE_CMN_LOOPBACK_TYPE_PHY},
    {SAI_PORT_INTERNAL_LOOPBACK_MAC, BASE_CMN_LOOPBACK_TYPE_MAC},
};

sai_port_internal_loopback_mode_t ndi_port_get_sai_loopback_mode
                             (BASE_CMN_LOOPBACK_TYPE_t lpbk_mode){

     sai_port_internal_loopback_mode_t mode = SAI_PORT_INTERNAL_LOOPBACK_NONE;

    auto it = ndi2sai_int_loopback_mode.find(lpbk_mode);

    if(it != ndi2sai_int_loopback_mode.end()){
        mode = it->second;
    }
    return mode;
}

BASE_CMN_LOOPBACK_TYPE_t ndi_port_get_ndi_loopback_mode(sai_port_internal_loopback_mode_t lpbk_mode){

     BASE_CMN_LOOPBACK_TYPE_t mode = BASE_CMN_LOOPBACK_TYPE_NONE;

    auto it = sai2ndi_int_loopback_mode.find(lpbk_mode);

    if(it != sai2ndi_int_loopback_mode.end()){
        mode = it->second;
    }
    return mode;
}

static std::unordered_map<BASE_IF_SPEED_t, uint32_t,std::hash<int>>
ndi2sai_speed =
{
    {BASE_IF_SPEED_10MBPS,       10},
    {BASE_IF_SPEED_100MBPS,     100},
    {BASE_IF_SPEED_1GIGE,      1000},
    {BASE_IF_SPEED_10GIGE,    10000},
    {BASE_IF_SPEED_25GIGE,    25000},
    {BASE_IF_SPEED_40GIGE,    40000},
    {BASE_IF_SPEED_100GIGE,  100000},
};

bool ndi_port_get_sai_speed(BASE_IF_SPEED_t speed, uint32_t *sai_speed){

    auto it = ndi2sai_speed.find(speed);
    if(it == ndi2sai_speed.end()) return false;
    *sai_speed = it->second;
    return true;
}
bool ndi_port_get_ndi_speed(uint32_t sai_speed, BASE_IF_SPEED_t *ndi_speed) {
    auto it = std::find_if(ndi2sai_speed.begin(), ndi2sai_speed.end(),
            [&sai_speed](decltype(*ndi2sai_speed.begin()) &speed_map){return speed_map.second == sai_speed;});
    if (it == ndi2sai_speed.end()) return false;
    *ndi_speed = it->first;
    return true;
}

