
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
 * filename: nas_ndi_sflow.cpp
 */

#include "nas_ndi_sflow.h"
#include "dell-base-sflow.h"
#include "event_log.h"
#include "saitypes.h"
#include "saistatus.h"
#include "nas_ndi_init.h"
#include "nas_ndi_utils.h"
#include "saisamplepacket.h"
#include "saiport.h"


#include <inttypes.h>

#define NDI_SFLOW_LOG(type,LVL,msg, ...) \
        EV_LOG(type, NAS_L2, LVL,"NDI-SFLOW", msg, ##__VA_ARGS__)


static inline  sai_samplepacket_api_t * ndi_sflow_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_samplepacket_api_tbl);
}

static inline  sai_port_api_t * ndi_port_api_get(nas_ndi_db_t *ndi_db_ptr) {
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_port_api_tbl);
}


t_std_error ndi_sflow_create_session(ndi_sflow_entry_t *sflow_entry){

    sai_status_t sai_ret;
    sai_object_id_t  sai_sflow_id;
    sai_attribute_t sflow_rate_attr;
    const unsigned int attr_list_size = 1;

    sflow_rate_attr.id = SAI_SAMPLEPACKET_ATTR_SAMPLE_RATE;
    sflow_rate_attr.value.u32 = sflow_entry->sampling_rate;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(sflow_entry->npu_id);

    if(ndi_db_ptr == NULL){
        NDI_SFLOW_LOG(ERR,0,"invalid NPU Id %d",sflow_entry->npu_id);
        return STD_ERR(SFLOW,PARAM,0);
    }

    if ((sai_ret = ndi_sflow_api_get(ndi_db_ptr)->create_samplepacket_session(&sai_sflow_id,
                                     attr_list_size,&sflow_rate_attr))!= SAI_STATUS_SUCCESS) {
        NDI_SFLOW_LOG(ERR,0,"Failed to create new sflow session in the NPU %d, error code %d",
                  sflow_entry->npu_id,sai_ret);
        return STD_ERR(SFLOW, FAIL, sai_ret);
    }

    NDI_SFLOW_LOG(INFO,3,"Created new sflow session %" PRIx64 " ",sai_sflow_id);
    sflow_entry->ndi_sflow_id = sai_sflow_id;

    return STD_ERR_OK;
}


t_std_error ndi_sflow_delete_session(ndi_sflow_entry_t *sflow_entry){

    sai_status_t sai_ret;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(sflow_entry->npu_id);

    if(ndi_db_ptr == NULL){
        NDI_SFLOW_LOG(ERR,0,"invalid NPU Id %d",sflow_entry->npu_id);
        return STD_ERR(SFLOW,PARAM,0);
    }

    if ((sai_ret = ndi_sflow_api_get(ndi_db_ptr)->remove_samplepacket_session((sai_object_id_t )
                                               sflow_entry->ndi_sflow_id))!= SAI_STATUS_SUCCESS) {
        NDI_SFLOW_LOG(ERR,0,"Error Deleting a sflow session % " PRIx64 " in the NPU",
                      sflow_entry->ndi_sflow_id);
        return STD_ERR(SFLOW, FAIL, sai_ret);
    }

    NDI_SFLOW_LOG(INFO,3,"Deleted sflow session % " PRIx64 " in the NPU",sflow_entry->ndi_sflow_id);
    return STD_ERR_OK;
}


t_std_error ndi_sflow_update_session(ndi_sflow_entry_t *sflow_entry,BASE_SFLOW_ENTRY_t attr_id){

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_sflow_attr;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(sflow_entry->npu_id);

    if(ndi_db_ptr == NULL){
        NDI_SFLOW_LOG(ERR,0,"invalid NPU Id %d",sflow_entry->npu_id);
        return STD_ERR(SFLOW,PARAM,0);
    }

    switch(attr_id){
        case BASE_SFLOW_ENTRY_SAMPLING_RATE:
            sai_sflow_attr.id = SAI_SAMPLEPACKET_ATTR_SAMPLE_RATE;
            sai_sflow_attr.value.u32 = sflow_entry->sampling_rate;
            break;

        default:
            NDI_SFLOW_LOG(ERR,0,"Unsupported sflow attribute id %d passed",attr_id);
            return STD_ERR(SFLOW,PARAM,0);
    }

    if ((sai_ret = ndi_sflow_api_get(ndi_db_ptr)->set_samplepacket_attribute((sai_object_id_t)
                             sflow_entry->ndi_sflow_id,&sai_sflow_attr))!= SAI_STATUS_SUCCESS) {
        NDI_SFLOW_LOG(ERR,0,"Failed to updated sampling rate for the sflow session %" PRIx64 ""    ,
                      sflow_entry->ndi_sflow_id);
        return STD_ERR(SFLOW, FAIL, sai_ret);
    }

    NDI_SFLOW_LOG(INFO,3,"Updated Attribute %d for sflow session % " PRIx64 " ",attr_id,
                  sflow_entry->ndi_sflow_id);
    return STD_ERR_OK;
}


t_std_error ndi_sflow_update_direction(ndi_sflow_entry_t *sflow_entry,
                        BASE_CMN_TRAFFIC_PATH_t direction,bool enable){

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_sflow_attr;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(sflow_entry->npu_id);

    if(ndi_db_ptr == NULL){
        NDI_SFLOW_LOG(ERR,0,"invalid NPU Id %d",sflow_entry->npu_id);
        return STD_ERR(SFLOW,PARAM,0);
    }

    if(enable){
        sai_sflow_attr.value.oid = (sai_object_id_t)sflow_entry->ndi_sflow_id;
    }else {
        sai_sflow_attr.value.oid = SAI_NULL_OBJECT_ID;
    }

    switch(direction){
        case BASE_CMN_TRAFFIC_PATH_INGRESS:
            sai_sflow_attr.id = SAI_PORT_ATTR_INGRESS_SAMPLEPACKET_ENABLE;
            break;

        case BASE_CMN_TRAFFIC_PATH_EGRESS:
            sai_sflow_attr.id = SAI_PORT_ATTR_EGRESS_SAMPLEPACKET_ENABLE;
            break;

        default:
            NDI_SFLOW_LOG(ERR,0,"Error invalid direction is passed %d",
                      sflow_entry->sflow_direction);
            return STD_ERR(SFLOW,PARAM,0);
    }

    sai_object_id_t port_obj_id;
    if(ndi_sai_port_id_get(sflow_entry->npu_id,sflow_entry->port_id,&port_obj_id)!= STD_ERR_OK){
        NDI_SFLOW_LOG(ERR,0,"Failed to get oid for npu %d and port %d",sflow_entry->npu_id,
                  sflow_entry->port_id);
        return STD_ERR(SFLOW,FAIL,0);
    }

    if ((sai_ret = ndi_port_api_get(ndi_db_ptr)->set_port_attribute(port_obj_id,&sai_sflow_attr))
                                         != SAI_STATUS_SUCCESS){
        NDI_SFLOW_LOG(ERR,0,"Failed to update sampling direction %d with val %d for the sflow "
                      "session %" PRIx64 " ",direction,enable,sflow_entry->ndi_sflow_id);
        return STD_ERR(SFLOW, FAIL, sai_ret);
    }

    NDI_SFLOW_LOG(ERR,0,"Updated sampling direction %d with val %d for the sflow "
          "session %" PRIx64 " ",direction,enable,sflow_entry->ndi_sflow_id);
    return STD_ERR_OK;

}


t_std_error ndi_sflow_get_session(ndi_sflow_entry_t *sflow_entry,npu_id_t npu_id, ndi_sflow_id_t id){

    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    sai_attribute_t sai_sflow_attr;
    const unsigned int attr_count = 1;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if(ndi_db_ptr == NULL){
        NDI_SFLOW_LOG(ERR,0,"invalid NPU Id %d",sflow_entry->npu_id);
        return STD_ERR(SFLOW,PARAM,0);
    }

    sai_sflow_attr.id = SAI_SAMPLEPACKET_ATTR_SAMPLE_RATE;

    if ((sai_ret = ndi_sflow_api_get(ndi_db_ptr)->get_samplepacket_attribute((sai_object_id_t)
                   sflow_entry->ndi_sflow_id,attr_count,&sai_sflow_attr))!= SAI_STATUS_SUCCESS) {
        NDI_SFLOW_LOG(ERR,0,"Failed to get the sflow attributes from the NPU for session %d",id);
        return STD_ERR(SFLOW, FAIL, sai_ret);
    }

    sflow_entry->sampling_rate = sai_sflow_attr.value.u32;

    NDI_SFLOW_LOG(INFO,3,"Session Information for %" PRIx64 " retrived from SAI",sflow_entry->ndi_sflow_id);
    return STD_ERR_OK;
}
