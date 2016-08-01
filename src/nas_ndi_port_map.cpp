
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
 * filename: nas_ndi_port_map.cpp
 */

#include "dell-base-phy-interface.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_common.h"
#include "nas_ndi_event_logs.h"
#include "sai.h"

#include <stdio.h>
#include <stdlib.h>
#define __STDC_FORMAT_MACROS
#include <inttypes.h>
#include <vector>
#include <unordered_map>

#define NDI_MAX_NPU          1
#define NDI_CPU_PORT_ID      0

/*  This file contains ndi_port to sai_port mapping table definition and APIs to
 *  add/delete/get sai_port. Two dinemtional vector [npu_id][npu_port_id] is used
 *  for storing the mapping from ndi_port to sai_port. the table contains flags for
 *   Active and list of hwports associated with the npu_port. the list of hwports
 *   are further used by NAS to identify the front-panal ports associated with the
 *   ndi_port.
 *  In case of fanout mode configuration, old npu_port is set to in-active and new
 *  ndi_ports are added in the table along with the new set of hwports. NAS is notified
 *  for the port add/delete event.
 *
 *  [hwportl] : is the pyhsical port (eg. WC lane ID in BCM T2 chip) associated with
 *  a sai_port. one sai_port can be associated with multiple hwports.
 *  eg. in T2, one 40g npu_port is mapped to one sai_port and the sai_port is mapped
 *  to 4 hwports belonging to the same Warpcore.
 *
 * Another mapping is from sai_port to ndi_port store in the one dinemtional map. Table is
 * created in the beginning and then updated when ever port ADD/DELETE event is sent to
 * NDI from SAI. This map table is used for converting sai_port to ndi port.
 *
 * */
/*  Following is for mapping from ndi_port to saiport and hwport */
#define NDI_PORT_MAP_ACTIVE_MASK        0x00000001
#define NDI_PORT_MAP_CPU_PORT_MASK      0x00000002
typedef struct _ndi_npu_port_map_t {
    sai_object_id_t sai_port;
    uint32_t hwport_count;
    std::vector<uint32_t> hwport_list;
    uint32_t flags = (uint32_t) 0;
} ndi_port_map_t;

/*  two dimentional vector based on the [npu_id][npu_port_id] */
typedef std::vector <std::vector<ndi_port_map_t > > ndi_port_to_sai_port_map_tbl_t;

static ndi_port_to_sai_port_map_tbl_t g_ndi_port_map_tbl;

/*  following is for sai_port to ndi_port  */

typedef struct ndi_saiport_map_t {
    npu_id_t npu_id;
    npu_port_t npu_port;
} ndi_saiport_map_t;

typedef std::unordered_map<sai_object_id_t, ndi_saiport_map_t> saiport_map_t;

/*  Global saiport to ndi port map table */
saiport_map_t g_saiport_map;

extern "C" {
static bool ndi_saiport_map_add_entry(sai_object_id_t sai_port, ndi_saiport_map_t *entry)
{
     do {
         try {
             g_saiport_map[sai_port] = *entry;
         } catch(...) {
             break;
         }
         return true;
     } while(0);

     NDI_PORT_LOG_ERROR("SAI port entry Add failure %" PRIx64 " ",  sai_port);
     return false;
}

static bool ndi_saiport_map_delete_entry(sai_object_id_t sai_port)
{
    try {
        g_saiport_map.erase(sai_port);
    } catch(...) {
        return false;
    }
    return true;
}

void ndi_saiport_map_table_dump(void)
{
    auto it = g_saiport_map.begin();

    printf("\n SAI PORT                NPU ID    PORT \n");
    printf("---------------------------------------------------");
    for (; it != g_saiport_map.end(); it++) {
        printf("0x%" PRIx64 "         %5d    %5d \n",
                it->first, it->second.npu_id, it->second.npu_port);
    }
}

void ndi_port_map_table_dump(void)
{

    size_t max_npu = g_ndi_port_map_tbl.size();

    printf("\nNPU ID    NPU PORT     SAI PORT        HW PORTS \n");
    printf("--------------------------------------------------------------\n");
    for (uint npu_ix = 0; npu_ix < max_npu; npu_ix++) {
        size_t max_port = g_ndi_port_map_tbl[npu_ix].size();
        for (uint port_ix =0; port_ix < max_port; port_ix++) {
            ndi_port_map_t *entry = &g_ndi_port_map_tbl[npu_ix][port_ix];
            if (entry->flags & NDI_PORT_MAP_ACTIVE_MASK) {
                printf("\n%5d    %5d    0x%" PRIx64 "     ",
                        npu_ix, port_ix, entry->sai_port);
                for (uint hw_ix = 0; hw_ix < entry->hwport_count; hw_ix++) {
                    printf(" %d ",entry->hwport_list[hw_ix]);
                }
            }
        }
    }
}

t_std_error ndi_npu_port_id_get(sai_object_id_t sai_port, npu_id_t *npu_id, npu_port_t *port_id)
{
    auto it = g_saiport_map.find(sai_port);
    if (it == g_saiport_map.end()) {
        NDI_PORT_LOG_ERROR("SAI port entry does not exist %" PRIx64 " ",  sai_port);
        return (STD_ERR(NPU, FAIL, 0));
    }
    *npu_id = it->second.npu_id;
    *port_id = it->second.npu_port;
    return(STD_ERR_OK);
}

size_t ndi_max_npu_get(void)
{
    return(NDI_MAX_NPU);
}

static inline  sai_port_api_t *ndi_sai_port_api_tbl_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_port_api_tbl);
}

size_t ndi_max_npu_port_get(npu_id_t npu)
{
    return(g_ndi_port_map_tbl[npu].size());
}
/*  Get the CPU port id */
t_std_error ndi_cpu_port_get(npu_id_t npu_id, npu_port_t *cpu_port)
{
    if (cpu_port == NULL) {
        return(STD_ERR(NPU, PARAM, 0));
    }
    *cpu_port = NDI_CPU_PORT_ID;
    return(STD_ERR_OK);
}


/* function for getting max port from SAI */
static t_std_error ndi_max_sai_port_get(npu_id_t npu_id, size_t *max_port)
{
    sai_attribute_t sai_attr;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t  *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_attr.id = SAI_SWITCH_ATTR_PORT_NUMBER;
    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->get_switch_attribute(1, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(NPU, CFG, sai_ret);
    }

    *max_port = (size_t)sai_attr.value.u32;
    return(STD_ERR_OK);
}

/*  function for getting list of all sai ports  */

static t_std_error ndi_sai_port_list_get(npu_id_t npu_id, size_t *max_npu_port, sai_object_id_t *sai_port_list)
{

    sai_attribute_t sai_attr;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t  *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return (STD_ERR(NPU, PARAM, 0));
    }

    sai_attr.id = SAI_SWITCH_ATTR_PORT_LIST;
    sai_attr.value.objlist.count = (uint32_t) *max_npu_port;
    sai_attr.value.objlist.list = sai_port_list;
    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->get_switch_attribute(1, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_PORT_LOG_ERROR("SAI SWITCH API Failure API ID %d, Error %d ",  sai_attr.id, sai_ret);
        return (STD_ERR(NPU, CFG, sai_ret));
    }
    *max_npu_port = (size_t) sai_attr.value.objlist.count;
    return(STD_ERR_OK);
}

/*  public function  for fetching list of hwports for a given ndi port */
t_std_error ndi_sai_port_hwport_list_get(npu_id_t npu_id, sai_object_id_t sai_port, uint32_t *hwport_list, uint32_t *count)
{
    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_attr;

    if ((hwport_list == NULL) || (count == NULL)) {
        return STD_ERR(NPU, PARAM, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_attr.id = SAI_PORT_ATTR_HW_LANE_LIST;
    sai_attr.value.u32list.count = *count;
    sai_attr.value.u32list.list = hwport_list;
    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->get_port_attribute(sai_port, 1, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        NDI_PORT_LOG_ERROR("SAI port API Failure API ID %d, Error %d ",  sai_attr.id, sai_ret);
        return STD_ERR(NPU, CFG, sai_ret);
    }
    /*  this is actual number of hwports belonging to the sai port */
    *count = sai_attr.value.u32list.count;

    return ret_code;
}

/*  function for create and intialize the vector of port info based on the SAI PORT
 *      and HW ports fetched. HW POR tis used as a index into the vector  */

static t_std_error ndi_port_map_tbl_allocate(void) {
    npu_id_t npu         = 0;
    npu_id_t npu_max     = (npu_id_t)ndi_max_npu_get();
    size_t max_port   = 0;
    t_std_error ret_code = STD_ERR(NPU, NOMEM, 0);

    try {
       g_ndi_port_map_tbl.resize(npu_max);
    } catch (...) {
        return (ret_code);
    }

    for ( npu = 0; npu <  npu_max ; ++npu){
         if ((ndi_max_sai_port_get(npu, &max_port)) != STD_ERR_OK) {
             return(STD_ERR(NPU, CFG, 0));
         }
        try {
            g_ndi_port_map_tbl[npu].resize(max_port);
        } catch (...) {
            return ret_code;
        }
    }
    return STD_ERR_OK;
}

/*  Add sai port in to the port map table */
t_std_error ndi_port_map_sai_port_add(npu_id_t npu, sai_object_id_t sai_port, npu_port_t *npu_port)
{
    uint32_t hwport_list[NDI_MAX_HWPORT_PER_PORT];
    uint32_t hwport_count = NDI_MAX_HWPORT_PER_PORT;
    t_std_error rc = STD_ERR_OK;
    uint32_t first_hwport = 0;
    ndi_saiport_map_t sai_entry;

    if ((rc = ndi_sai_port_hwport_list_get(npu, sai_port, hwport_list, &hwport_count)) != STD_ERR_OK) {
        return(rc);
    }
    /*  use first HW port as index in the port map table */
    first_hwport = hwport_list[0];
    if (first_hwport > g_ndi_port_map_tbl[npu].size()-1) {
        try {
             g_ndi_port_map_tbl[npu].resize(first_hwport+1);
        } catch(...) {
            return(STD_ERR(NPU,NOMEM,0));
        }
    }
    /*  Check if the entry is already filled  */
    if (g_ndi_port_map_tbl[npu][first_hwport].flags & NDI_PORT_MAP_ACTIVE_MASK) {
        /* Entry already present  */
        return(STD_ERR(NPU,CFG,0));
    }

    /*  add the entry  */
    g_ndi_port_map_tbl[npu][first_hwport].sai_port = sai_port;
    g_ndi_port_map_tbl[npu][first_hwport].hwport_count = hwport_count;
    g_ndi_port_map_tbl[npu][first_hwport].flags |= NDI_PORT_MAP_ACTIVE_MASK;
    g_ndi_port_map_tbl[npu][first_hwport].hwport_list.resize(hwport_count);

    for (uint32_t idx =0; idx < hwport_count; idx++) {
       g_ndi_port_map_tbl[npu][first_hwport].hwport_list[idx] = hwport_list[idx];
    }

    /*  Now add an entry in the sai port map   */
    sai_entry.npu_id = npu;
    sai_entry.npu_port = first_hwport;
    if (ndi_saiport_map_add_entry(sai_port, &sai_entry) != true) {
        return STD_ERR(NPU, FAIL, 0);
    }

    NDI_LOG_INFO(0,"NAS-NDI-PORT"," Initializing ports hwport %X - sai port%" PRIx64 " ",first_hwport,sai_port);
    *npu_port = first_hwport;
    return(STD_ERR_OK);
}

t_std_error ndi_sai_cpu_port_add(npu_id_t npu_id)
{
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    sai_switch_api_t *sai_switch_api_tbl = ndi_sai_switch_api_tbl_get(ndi_db_ptr);
    sai_attribute_t sai_attr;
    sai_object_id_t sai_cpu_port;
    npu_port_t ndi_cpu_port = 0;
    ndi_saiport_map_t sai_entry;
    t_std_error ret_code = STD_ERR_OK;

    /*  Get ndi cpu port */
    if (( ndi_cpu_port_get(npu_id, &ndi_cpu_port)) != STD_ERR_OK) {
        ret_code = STD_ERR(NPU, PARAM, 0);
        return(ret_code);
    }

    /*  Fetch SAI CPU port Id  */
    sai_attr.id  = SAI_SWITCH_ATTR_CPU_PORT;
    if ((sai_ret = sai_switch_api_tbl->get_switch_attribute(1, &sai_attr)) != SAI_STATUS_SUCCESS) {
        NDI_INIT_LOG_ERROR(" SAI CPU PORT Attribute get API failed for NPU %d\n", npu_id);
        ret_code = STD_ERR(NPU, CFG, sai_ret);
        return(ret_code);
    }

    sai_cpu_port = sai_attr.value.oid;

    g_ndi_port_map_tbl[npu_id][ndi_cpu_port].sai_port = sai_cpu_port;
    /*  There is no HWport for CPU port. set it to 0 */
    g_ndi_port_map_tbl[npu_id][ndi_cpu_port].hwport_count = 1;
    g_ndi_port_map_tbl[npu_id][ndi_cpu_port].flags |= NDI_PORT_MAP_ACTIVE_MASK | NDI_PORT_MAP_CPU_PORT_MASK;
    g_ndi_port_map_tbl[npu_id][ndi_cpu_port].hwport_list.resize(1);
    g_ndi_port_map_tbl[npu_id][ndi_cpu_port].hwport_list[0] = 0; /*  hwport is 0 for cpu port */

    sai_entry.npu_id = npu_id;
    sai_entry.npu_port = ndi_cpu_port;
    if (ndi_saiport_map_add_entry(sai_cpu_port, &sai_entry) != true) {
        return STD_ERR(NPU, FAIL, 0);
    }
    return(STD_ERR_OK);
}

t_std_error ndi_port_map_sai_port_delete(npu_id_t npu, sai_object_id_t sai_port, npu_port_t *npu_port)
{
    t_std_error rc = STD_ERR_OK;
    uint32_t first_hwport = 0;

    auto it = g_saiport_map.find(sai_port);
    if (it==g_saiport_map.end()) {
        return rc;
    }
    if ((g_ndi_port_map_tbl.size()<= (size_t)it->second.npu_id)) {
        //error
        g_saiport_map.erase(sai_port);
        return STD_ERR(NPU,FAIL,0);
    }
    if ((size_t)it->second.npu_port >= g_ndi_port_map_tbl[it->second.npu_id].size()) {
        g_saiport_map.erase(sai_port);
        return STD_ERR(NPU,FAIL,0);
    }

    npu = it->second.npu_id;
    first_hwport = it->second.npu_port;

    /*  check if the the sai sai_port exist and flag is set to ACTIVE */
    if (( g_ndi_port_map_tbl[npu][first_hwport].sai_port != sai_port) ||
        ((g_ndi_port_map_tbl[npu][first_hwport].flags & NDI_PORT_MAP_ACTIVE_MASK) == false)) {
        /*  it means the entry does not exist for the same hwport */
        NDI_PORT_LOG_TRACE("sai_port does not exist npu %d sai_port 0x%" PRIx64 " ", npu, sai_port);
        return STD_ERR(NPU, FAIL, 0);
    }

    g_ndi_port_map_tbl[npu][first_hwport].hwport_list.resize(0);
    g_ndi_port_map_tbl[npu][first_hwport].flags &= ~NDI_PORT_MAP_ACTIVE_MASK;
    g_ndi_port_map_tbl[npu][first_hwport].hwport_count = 0;

    *npu_port = first_hwport;
    /*  Now delete from saiport map table  */
    if (ndi_saiport_map_delete_entry(sai_port) != true) {
        return STD_ERR(NPU, FAIL, 0);
    }
    return(STD_ERR_OK);
}

/*  Extract npu_id from the sai port object*/
/*  TODO confirm if this conversion is as per SAI spec */
npu_id_t ndi_saiport_to_npu_id_get(sai_object_id_t sai_port)
{
    npu_id_t npu_id = 0;
#define NDI_NPU_ID_BITMASK       0x00FF00000000
#define NDI_NPU_ID_BITPOS (32)
    npu_id = (npu_id_t) ((sai_port & NDI_NPU_ID_BITMASK) >> NDI_NPU_ID_BITPOS);
    return(npu_id);
}

static t_std_error ndi_per_npu_sai_port_map_update(npu_id_t npu_id)
{
    /*  store the sai port in the ndi port map table */
    sai_object_id_t *sai_port_list = NULL;
    t_std_error rc = STD_ERR_OK;
    size_t port_count = 0;
    uint_t p_idx =0;
    npu_port_t npu_port = 0;

    /*  Get the list of sai ports */

    port_count = g_ndi_port_map_tbl[npu_id].size();
    sai_port_list = (sai_object_id_t *)calloc(port_count, sizeof(sai_object_id_t));
    if (sai_port_list == NULL) {
        return(STD_ERR(NPU, NOMEM, 0));
    }

    do {
        /* cpu port */
        if ((rc = ndi_sai_cpu_port_add(npu_id)) != STD_ERR_OK) {
            NDI_PORT_LOG_ERROR("unable to add sai cpu port into the port map table for npu %d ",
                                npu_id);
            break;
        }

        /*  Get the list of sai ports  */
        if ((rc = ndi_sai_port_list_get(npu_id, &port_count, sai_port_list)) != STD_ERR_OK) {
            break;
        }

        /* run through the list of sai_ports  */
        for (p_idx = 0; p_idx < port_count; p_idx++) {
            /*  Now add the sai port and hw port in the port map */

            if ((rc = ndi_port_map_sai_port_add(npu_id, sai_port_list[p_idx], &npu_port) != STD_ERR_OK)) {
                NDI_PORT_LOG_ERROR("unable to add sai port index %d into the port map table for npu %d ",
                                    p_idx, npu_id);
                break; /*  from the for loop */
            }
        }

        if (p_idx != port_count) {
            /*  Not all sai ports are added in the port map  */
            NDI_PORT_LOG_ERROR("unable to create the saiport map table size %d \n", port_count);
        }

        /*  Imp: no return inside while  */
    }while (0);

    free(sai_port_list);
    return(rc);
}

t_std_error ndi_sai_port_map_create(void)
{

    t_std_error rc = STD_ERR_OK;


    /*  Allocate memory of the ndi port table */
    if ((rc = ndi_port_map_tbl_allocate()) != STD_ERR_OK) {
        return(rc);
    }
    for (uint_t npu = 0; npu < g_ndi_port_map_tbl.size(); npu++) {
        if ((rc = ndi_per_npu_sai_port_map_update(npu)) != STD_ERR_OK) {
            break;
        }
    }
    return(rc);
}

/*  public function for checking if the port is invalid */
bool ndi_port_is_valid(npu_id_t npu, npu_port_t ndi_port)
{
    if ((npu >= (npu_id_t) g_ndi_port_map_tbl.size()) ||
        (ndi_port >= g_ndi_port_map_tbl[npu].size())) {
        return(false);
    }
    if ((g_ndi_port_map_tbl[npu][ndi_port].flags & NDI_PORT_MAP_ACTIVE_MASK) == false) {
        return(false);
    }
    return(true);
}

t_std_error ndi_sai_port_id_get(npu_id_t npu_id, npu_port_t ndi_port, sai_object_id_t *sai_port)
{
    if ((ndi_port_is_valid(npu_id, ndi_port) == false) ||
           (sai_port == NULL)) {
        return(STD_ERR(NPU,PARAM,0));
    }
    *sai_port = g_ndi_port_map_tbl[npu_id][ndi_port].sai_port;
    return(STD_ERR_OK);
}

t_std_error ndi_hwport_list_get(npu_id_t npu, npu_port_t ndi_port, uint32_t *hwport)
{

    if (ndi_port_is_valid(npu, ndi_port) == false) {
        NDI_PORT_LOG_ERROR(" invalid npu %d or port id %d", npu, ndi_port);
        return(STD_ERR(NPU,PARAM,0));
    }
    *hwport = g_ndi_port_map_tbl[npu][ndi_port].hwport_list[0];

    return(STD_ERR_OK);
}

BASE_IF_PHY_BREAKOUT_MODE_t sai_break_to_ndi_break( int32_t sai_mode) {
    static const std::unordered_map<int32_t, BASE_IF_PHY_BREAKOUT_MODE_t> _m = {
            { SAI_PORT_BREAKOUT_MODE_1_LANE, BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_1X1 },
            { SAI_PORT_BREAKOUT_MODE_2_LANE, BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_2X1 },
            { SAI_PORT_BREAKOUT_MODE_4_LANE, BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_4X1 }
    };
    auto it = _m.find(sai_mode);
    if (it!=_m.end()) return it->second;
    return BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_1X1;
}

sai_port_breakout_mode_type_t sai_break_from_ndi_break( BASE_IF_PHY_BREAKOUT_MODE_t mode) {
    static const std::unordered_map<uint32_t,sai_port_breakout_mode_type_t> _m = {
            { BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_1X1, SAI_PORT_BREAKOUT_MODE_1_LANE },
            { BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_2X1, SAI_PORT_BREAKOUT_MODE_2_LANE },
            { BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_4X1, SAI_PORT_BREAKOUT_MODE_4_LANE }
    };
    auto it = _m.find(mode);
    if (it!=_m.end()) return it->second;
    return SAI_PORT_BREAKOUT_MODE_1_LANE;
}


/*  public function for setting breakout mode for a ndi_port */
t_std_error ndi_port_breakout_mode_set(npu_id_t npu_id, npu_port_t ndi_port,
        BASE_IF_PHY_BREAKOUT_MODE_t mode, npu_port_t *effected_ports, size_t len)
{
    t_std_error rc = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_attr_set;
    memset(&sai_attr_set, 0, sizeof(sai_attribute_t));

    if (mode != BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_1X1 &&
            mode!=BASE_IF_PHY_BREAKOUT_MODE_BREAKOUT_4X1) {
        return STD_ERR(NPU, PARAM, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_object_id_t sai_port;
    /*  Get the sai port  */
    if ((rc = ndi_sai_port_id_get(npu_id, ndi_port,&sai_port)) != STD_ERR_OK) {
        EV_LOG(ERR,NDI,0,"NDI-BREAKOUT","Invaid port breakout - master port %d:%d",npu_id,ndi_port);
        return(rc);
    }

    sai_object_id_t sai_effected_port[len];
    size_t ix = 0;
    for ( ; ix < len ; ++ix ) {
        if ((rc = ndi_sai_port_id_get(npu_id, effected_ports[ix],&sai_effected_port[ix])) != STD_ERR_OK) {
            EV_LOG(ERR,NDI,0,"NDI-BREAKOUT","Invaid port breakout - slave port %d:%d",npu_id,effected_ports[ix]);
            return(rc);
        }
    }

    sai_port_breakout_mode_type_t sai_mode = sai_break_from_ndi_break(mode);

    /*  TODO first check if the break out mode is same  */
    sai_attr_set.id = SAI_SWITCH_ATTR_PORT_BREAKOUT ;
    sai_attr_set.value.portbreakout.breakout_mode =  sai_mode;
    sai_attr_set.value.portbreakout.port_list.count = len ;
    sai_attr_set.value.portbreakout.port_list.list =  sai_effected_port;

    /*  call sai api for setting switch attribute */
    if ((sai_ret = ndi_sai_switch_api_tbl_get(ndi_db_ptr)->set_switch_attribute(&sai_attr_set))
                         != SAI_STATUS_SUCCESS) {
        rc =  STD_ERR(NPU, CFG, sai_ret);
    }

    return(rc);

}

/*  public funtion for registering port event callback  */

t_std_error ndi_port_event_cb_register(npu_id_t npu_id, ndi_port_event_update_fn func)
{
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        NDI_INIT_LOG_ERROR("Ignore SAI port event notification if SAI is not UP yet \n");
        return STD_ERR(NPU, PARAM, 0);
    }
    ndi_db_ptr->switch_notification->port_event_update_cb = func;
    return(STD_ERR_OK);
}

} //extern "C"

