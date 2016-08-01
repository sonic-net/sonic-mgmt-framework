
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
 * filename: nas_ndi_stg.cpp
 */

#include "dell-base-stg.h"
#include "event_log.h"
#include "std_error_codes.h"

#include "nas_ndi_stg.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"

#include "saitypes.h"
#include "saiport.h"
#include "saistp.h"
#include "saivlan.h"
#include "saiswitch.h"

#include <unordered_map>
#include <functional>
#include <stdint.h>
#include <inttypes.h>

#define NDI_STG_LOG(type,LVL,msg, ...) \
        EV_LOG( type, NAS_L2, LVL,"NDI-STG", msg, ##__VA_ARGS__)

static std::unordered_map<BASE_STG_INTERFACE_STATE_t, sai_port_stp_port_state_t,std::hash<int>>
ndi_to_sai_stp_state_map = {
        {BASE_STG_INTERFACE_STATE_DISABLED,SAI_PORT_STP_STATE_BLOCKING},
        {BASE_STG_INTERFACE_STATE_LEARNING,SAI_PORT_STP_STATE_LEARNING},
        {BASE_STG_INTERFACE_STATE_FORWARDING,SAI_PORT_STP_STATE_FORWARDING},
        {BASE_STG_INTERFACE_STATE_BLOCKING,SAI_PORT_STP_STATE_BLOCKING},
        {BASE_STG_INTERFACE_STATE_LISTENING,SAI_PORT_STP_STATE_BLOCKING}
};

static std::unordered_map<sai_port_stp_port_state_t, BASE_STG_INTERFACE_STATE_t, std::hash<int>>
sai_to_ndi_stp_state_map = {
    {SAI_PORT_STP_STATE_BLOCKING,BASE_STG_INTERFACE_STATE_BLOCKING},
    {SAI_PORT_STP_STATE_LEARNING,BASE_STG_INTERFACE_STATE_LEARNING},
    {SAI_PORT_STP_STATE_FORWARDING,BASE_STG_INTERFACE_STATE_FORWARDING}
};

static inline  sai_stp_api_t * ndi_stp_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_stp_api_tbl);
}

static inline  sai_switch_api_t * ndi_switch_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_switch_api_tbl);
}

static inline  sai_vlan_api_t * ndi_vlan_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_vlan_api_tbl);
}


t_std_error ndi_stg_add(npu_id_t npu_id, ndi_stg_id_t * stg_id){
    sai_status_t sai_ret;
    sai_object_id_t  sai_stp_id;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }

    const unsigned int attr_count = 0;

    if ((sai_ret = ndi_stp_api_get(ndi_db_ptr)->create_stp(&sai_stp_id,attr_count,NULL))
                                                                    != SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"NDI STG Creation Failed with return code %d",sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"New STG Id %" PRIu64 " created",sai_stp_id);
    *stg_id = sai_stp_id;
    return STD_ERR_OK;
}


t_std_error ndi_stg_delete(npu_id_t npu_id, ndi_stg_id_t stg_id){
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }

    if ((sai_ret = ndi_stp_api_get(ndi_db_ptr)->remove_stp(stg_id))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"STG Id %" PRIu64 " deletion failed with return code %d",stg_id,sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"STG Id %" PRIu64 " deleted",stg_id);

    return STD_ERR_OK;
}


t_std_error ndi_stg_update_vlan(npu_id_t npu_id, ndi_stg_id_t  stg_id, hal_vlan_id_t vlan_id){
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }
    sai_status_t sai_ret;
    sai_attribute_t vlan_attr;
    vlan_attr.id = SAI_VLAN_ATTR_STP_INSTANCE ;
    vlan_attr.value.oid= stg_id;

    if ((sai_ret = ndi_vlan_api_get(ndi_db_ptr)->set_vlan_attribute(vlan_id
                                                ,&vlan_attr))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"Associating VLAN ID %d to STG ID %" PRIu64 " failed with return code %d"
                    ,vlan_id,stg_id,sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"Associated VLAN ID %d to STG ID %" PRIu64 " ",vlan_id,stg_id);
    return STD_ERR_OK;
}


t_std_error ndi_stg_set_stp_port_state(npu_id_t npu_id, ndi_stg_id_t stg_id, npu_port_t port_id,
                                                    BASE_STG_INTERFACE_STATE_t port_stp_state){

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }

    sai_status_t sai_ret ;
    auto it = ndi_to_sai_stp_state_map.find(port_stp_state);
    if(it == ndi_to_sai_stp_state_map.end()) {
        NDI_STG_LOG(ERR,0,"NO SAI STP State found for %d",port_stp_state);
        return STD_ERR(STG,PARAM,0);
    }

    sai_object_id_t obj_id;
    if(ndi_sai_port_id_get( npu_id,port_id,&obj_id)!= STD_ERR_OK){
        NDI_STG_LOG(ERR,0,"Failed to get oid for npu %d and port %d",
                              npu_id,port_id);
        return STD_ERR(STG,FAIL,0);
    }

    if ((sai_ret = ndi_stp_api_get(ndi_db_ptr)->set_stp_port_state(stg_id,
                            obj_id ,it->second))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"Failed to Set stp state %d to port %d in stg id %d with return code %d",
                                                    it->second,port_id,stg_id,sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"Set stp state %d to port %d in stg id %" PRIu64 "",
                                                 it->second,port_id,stg_id);
    return STD_ERR_OK;
}


t_std_error ndi_stg_get_stp_port_state(npu_id_t npu_id, ndi_stg_id_t stg_id, npu_port_t port_id,
                                                BASE_STG_INTERFACE_STATE_t *port_stp_state){

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }

    sai_status_t sai_ret;
    sai_port_stp_port_state_t  sai_stp_state;

    sai_object_id_t obj_id;
    if(ndi_sai_port_id_get(npu_id,port_id,&obj_id)!= STD_ERR_OK){
        NDI_STG_LOG(ERR,0,"Failed to get oid for npu %d and port %d",
                          npu_id,port_id);
        return STD_ERR(STG,FAIL,0);
    }

    if ((sai_ret = ndi_stp_api_get(ndi_db_ptr)->get_stp_port_state(stg_id,
                                obj_id ,&sai_stp_state))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"Failed to get the STP Port State for STG id %" PRIu64 ""
                        "and Port id %d with return code %d",stg_id,port_id,sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"Got the STP Port State for STG id %" PRIu64 " "
                                        "and Port id %d",stg_id,port_id);

    auto it = sai_to_ndi_stp_state_map.find(sai_stp_state);
    if(it == sai_to_ndi_stp_state_map.end()){
        NDI_STG_LOG(ERR,0,"NO SAI STP State found for %d",sai_stp_state);
        return STD_ERR(STG,PARAM,0);
    }
    *port_stp_state = it->second;
    return STD_ERR_OK;
}


t_std_error ndi_stg_get_default_id(npu_id_t npu_id, ndi_stg_id_t *stg_id, hal_vlan_id_t *vlan_id){

    if( (stg_id == NULL) || (vlan_id == NULL) ){
        NDI_STG_LOG(ERR,0,"Null Pointers passed to get default STG instance info");
        return STD_ERR(STG,PARAM,0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_STG_LOG(ERR,0,"Invalid NPU id %d",npu_id);
        return STD_ERR(STG,PARAM,0);
    }

    sai_attribute_t default_stp;
    const unsigned int attr_count = 1;
    sai_status_t sai_ret;
    default_stp.id = SAI_SWITCH_ATTR_DEFAULT_STP_INST_ID;


    if ((sai_ret = ndi_switch_api_get(ndi_db_ptr)->get_switch_attribute
                        (attr_count,&default_stp))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"Failed to get the Default STP Id with return code %d",sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    *stg_id = default_stp.value.oid;

    NDI_STG_LOG(INFO,3,"Got the default STG instance id %" PRIu64 " ",*stg_id);

    sai_attribute_t default_vlan;
    default_vlan.id = SAI_STP_ATTR_VLAN_LIST;
    default_vlan.value.vlanlist.list = vlan_id;
    default_vlan.value.vlanlist.count = 1;


    if ((sai_ret = ndi_stp_api_get(ndi_db_ptr)->get_stp_attribute
                        (*stg_id,attr_count,&default_vlan))!= SAI_STATUS_SUCCESS) {
        NDI_STG_LOG(ERR,0,"Failed to get the VLAN associated with default STG"
                          " with return code %d",sai_ret);
        return STD_ERR(STG, FAIL, sai_ret);
    }

    NDI_STG_LOG(INFO,3,"Got the VLAN for default STG instance id %d ",*vlan_id);

    return STD_ERR_OK;

}
