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
 * filename: nas_ndi_port.c
 */

#include "std_error_codes.h"
#include "std_assert.h"
#include "ds_common_types.h"
#include "dell-base-platform-common.h"

#include "nas_ndi_event_logs.h"
#include "nas_ndi_int.h"
#include "nas_ndi_utils.h"
#include "nas_ndi_port.h"
#include "nas_ndi_port_utils.h"
#include "sai.h"
#include "saiport.h"
#include "saistatus.h"
#include "saitypes.h"

#include <stdio.h>
#include <stdlib.h>


/*  NDI Port specific APIs  */

static inline  sai_port_api_t *ndi_sai_port_api_tbl_get(nas_ndi_db_t *ndi_db_ptr)
{
    return(ndi_db_ptr->ndi_sai_api_tbl.n_sai_port_api_tbl);
}

typedef enum  {
    SAI_SG_ACT_SET,
    SAI_SG_ACT_GET
} SAI_SET_OR_GET_ACTION_t;

t_std_error _sai_port_attr_set_or_get(npu_id_t npu, port_t port, SAI_SET_OR_GET_ACTION_t set,
        sai_attribute_t *attr, size_t count) {
    STD_ASSERT(attr != NULL);

    if (count==0 || ((set==SAI_SG_ACT_SET) && (count >1))) {
        return STD_ERR(NPU, PARAM, 0);
    }

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_object_id_t sai_port;
    t_std_error ret_code = STD_ERR_OK;

    if ((ret_code = ndi_sai_port_id_get(npu, port, &sai_port) != STD_ERR_OK)) {
        return STD_ERR(NPU, PARAM, 0);
    }

    sai_status_t sai_ret = SAI_STATUS_SUCCESS;
    if (set==SAI_SG_ACT_SET) {
        sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->set_port_attribute(sai_port,attr);
    } else {
        sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->get_port_attribute(sai_port, count, attr);
    }

    return sai_ret == SAI_STATUS_SUCCESS ?
            STD_ERR_OK :
            STD_ERR(NPU, CFG, sai_ret);
}

t_std_error ndi_port_oper_state_notify_register(ndi_port_oper_status_change_fn reg_fn)
{
    t_std_error ret_code = STD_ERR_OK;
    npu_id_t npu_id = ndi_npu_id_get();
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }
    STD_ASSERT(reg_fn != NULL);

    ndi_db_ptr->switch_notification->port_oper_status_change_cb = reg_fn;

    return ret_code;
}

/*  public function for getting list of breakout mode supported. */
t_std_error ndi_port_supported_breakout_mode_get(npu_id_t npu_id, npu_port_t ndi_port,
        int *mode_count, BASE_IF_PHY_BREAKOUT_MODE_t *mode_list) {

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_SUPPORTED_BREAKOUT_MODE;
    sai_attr.value.s32list.count = SAI_PORT_BREAKOUT_MODE_MAX;
    int32_t modes[SAI_PORT_BREAKOUT_MODE_MAX];
    sai_attr.value.s32list.list = modes;
    if (_sai_port_attr_set_or_get(npu_id,ndi_port,SAI_SG_ACT_GET,&sai_attr,1)!=STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }
    size_t ix = 0;
    size_t mx = (*mode_count > sai_attr.value.s32list.count) ?
            sai_attr.value.s32list.count : *mode_count;
    for ( ; ix < mx ; ++ix ) {
        mode_list[ix] = sai_break_to_ndi_break(sai_attr.value.s32list.list[ix]);
    }
    *mode_count = mx;
    return(STD_ERR_OK);
}

t_std_error ndi_port_admin_state_set(npu_id_t npu_id, npu_port_t port_id,
        bool admin_state) {

    sai_attribute_t sai_attr;
    sai_attr.value.booldata = admin_state;
    sai_attr.id = SAI_PORT_ATTR_ADMIN_STATE;

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_admin_state_get(npu_id_t npu_id, npu_port_t port_id, IF_INTERFACES_STATE_INTERFACE_ADMIN_STATUS_t *admin_state)
{
    STD_ASSERT(admin_state != NULL);

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_ADMIN_STATE;
    t_std_error rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        *admin_state  = (sai_attr.value.booldata == true) ? IF_INTERFACES_STATE_INTERFACE_ADMIN_STATUS_UP : IF_INTERFACES_STATE_INTERFACE_ADMIN_STATUS_DOWN;
    }
    return rc;
}

t_std_error ndi_port_link_state_get(npu_id_t npu_id, npu_port_t port_id,
        ndi_intf_link_state_t *link_state)
{

    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    sai_port_oper_status_t sai_port_status;
    sai_object_id_t sai_port;
    sai_attribute_t sai_attr;

    STD_ASSERT(link_state != NULL);

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((ret_code = ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK)) {
        return ret_code;
    }

    sai_attr.id = SAI_PORT_ATTR_OPER_STATUS;
    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->get_port_attribute(sai_port, 1, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(NPU, CFG, sai_ret);
    }
    sai_port_status = (sai_port_oper_status_t) sai_attr.value.s32;
    ret_code = ndi_sai_oper_state_to_link_state_get(sai_port_status, &(link_state->oper_status));


    return ret_code;
}

t_std_error ndi_sai_oper_state_to_link_state_get(sai_port_oper_status_t sai_port_state,
                       ndi_port_oper_status_t *p_state)
{

    switch(sai_port_state) {
    case SAI_PORT_OPER_STATUS_UNKNOWN:
        *p_state = ndi_port_OPER_FAIL;
            NDI_INIT_LOG_TRACE(" port state is unknown\n");
        break;
    case SAI_PORT_OPER_STATUS_UP:
        *p_state = ndi_port_OPER_UP;
        break;
    case SAI_PORT_OPER_STATUS_DOWN:
        *p_state = ndi_port_OPER_DOWN;
        break;
    case SAI_PORT_OPER_STATUS_TESTING:
        *p_state = ndi_port_OPER_TESTING;
        NDI_INIT_LOG_TRACE(" port is under test\n");
        break;
    case SAI_PORT_OPER_STATUS_NOT_PRESENT:
        *p_state = ndi_port_OPER_FAIL;
        break;
    default:
        NDI_INIT_LOG_TRACE(" unknown port state is return\n");
        return (STD_ERR(NPU,FAIL,0));
    }

    return STD_ERR_OK;
}

t_std_error ndi_port_breakout_mode_get(npu_id_t npu, npu_port_t port,
        BASE_IF_PHY_BREAKOUT_MODE_t *mode) {

    sai_attribute_t sai_attr;
    memset(&sai_attr,0,sizeof(sai_attr));
    sai_attr.id = SAI_PORT_ATTR_CURRENT_BREAKOUT_MODE;

    t_std_error rc ;
    if ((rc=_sai_port_attr_set_or_get(npu,port,SAI_SG_ACT_GET,&sai_attr,1))==STD_ERR_OK) {
        *mode = sai_break_to_ndi_break(sai_attr.value.s32);
        return STD_ERR_OK;
    }
    return rc;
}

/*  public function for getting list of supported speed. */
t_std_error ndi_port_supported_speed_get(npu_id_t npu_id, npu_port_t ndi_port,
        size_t *speed_count, BASE_IF_SPEED_t *speed_list) {

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_SUPPORTED_SPEED;
    sai_attr.value.s32list.count = NDI_PORT_SUPPORTED_SPEED_MAX;
    int32_t speeds[NDI_PORT_SUPPORTED_SPEED_MAX];
    sai_attr.value.s32list.list = speeds;
    if (_sai_port_attr_set_or_get(npu_id,ndi_port,SAI_SG_ACT_GET,&sai_attr,1)!=STD_ERR_OK) {
        return STD_ERR(NPU, PARAM, 0);
    }
    size_t ix = 0;
    size_t mx = (*speed_count > sai_attr.value.s32list.count) ?
            sai_attr.value.s32list.count : *speed_count;
    BASE_IF_SPEED_t *speed = speed_list;
    for ( ; ix < mx ; ++ix ) {
        if (!ndi_port_get_ndi_speed(sai_attr.value.s32list.list[ix], speed)) {
            NDI_PORT_LOG_ERROR("unsupported Speed  returned from SAI%d", sai_attr.value.s32list.list[ix]);
            return STD_ERR(NPU, NEXIST, 0);
        }
        speed++;
    }
    *speed_count = mx;
    return(STD_ERR_OK);
}
t_std_error ndi_port_speed_set(npu_id_t npu_id, npu_port_t port_id, BASE_IF_SPEED_t speed) {
    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_SPEED;
    if (speed == BASE_IF_SPEED_AUTO)  {
        /*  speed==AUTO is not supported at BASE level
         *  TODO just return ok for the time being until it is supported at application layer
         */
        NDI_PORT_LOG_ERROR("Speed AUTO is not supported at BASE level");
        return STD_ERR_OK;
    }
    if (!ndi_port_get_sai_speed(speed, (uint32_t *)&sai_attr.value.u32)) {
        NDI_PORT_LOG_ERROR("unsupported Speed %d", (uint32_t)speed);
        return STD_ERR(NPU, PARAM, 0);
    }
    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_mtu_get(npu_id_t npu_id, npu_port_t port_id, uint_t *mtu) {
    STD_ASSERT(mtu!=NULL);

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_MTU;

    t_std_error rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        *mtu = sai_attr.value.u32;
    }
    return rc;
}

t_std_error ndi_port_mtu_set(npu_id_t npu_id, npu_port_t port_id, uint_t mtu) {
    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_MTU;
    sai_attr.value.u32 = mtu;

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_loopback_get(npu_id_t npu_id, npu_port_t port_id, BASE_CMN_LOOPBACK_TYPE_t *loopback) {
    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_INTERNAL_LOOPBACK;

    t_std_error rc =  _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        *loopback = ndi_port_get_ndi_loopback_mode(sai_attr.value.s32);
    }
    return rc;
}

t_std_error ndi_port_loopback_set(npu_id_t npu_id, npu_port_t port_id, BASE_CMN_LOOPBACK_TYPE_t loopback) {
    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_INTERNAL_LOOPBACK;
    sai_attr.value.s32 = ndi_port_get_sai_loopback_mode(loopback);

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

static inline sai_port_media_type_t ndi_sai_port_media_type_translate (PLATFORM_MEDIA_TYPE_t hal_media_type)
{
    sai_port_media_type_t sal_media_type;
    switch (hal_media_type) {
        case PLATFORM_MEDIA_TYPE_AR_POPTICS_NOTPRESENT:
           sal_media_type = SAI_PORT_MEDIA_TYPE_NOT_PRESENT;
            break;
        case PLATFORM_MEDIA_TYPE_AR_POPTICS_UNKNOWN:
           sal_media_type = SAI_PORT_MEDIA_TYPE_UNKNONWN;
            break;
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_USR:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_SR:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_LR:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_ER:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_ZR:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_LRM:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_DWDM:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_DWDM_40KM:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_DWDM_80KM:
            sal_media_type = SAI_PORT_MEDIA_TYPE_SFP_FIBER;
            break;
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_T:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CUHALFM:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU1M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU2M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU3M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU5M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU7M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CU10M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_ACU7M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_ACU10M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_ACU15M:
        case PLATFORM_MEDIA_TYPE_AR_SFPPLUS_10GBASE_CX4:
            sal_media_type = SAI_PORT_MEDIA_TYPE_SFP_COPPER;
            break;
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_SR4:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_SR4_EXT:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_LR4:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_PSM4_1490NM:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_PSM4_1490NM_1M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_PSM4_1490NM_3M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_PSM4_1490NM_5M:
        case PLATFORM_MEDIA_TYPE_QSFP_40GBASE_SM4:
            sal_media_type = SAI_PORT_MEDIA_TYPE_QSFP_FIBER;
            break;
        case PLATFORM_MEDIA_TYPE_AR_4X1_1000BASE_T:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_1M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_HAL_M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_2M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_3M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_5M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_7M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_10M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4_50M:
        case PLATFORM_MEDIA_TYPE_AR_QSFP_40GBASE_CR4:
        case PLATFORM_MEDIA_TYPE_AR_4X10_10GBASE_CR1_HAL_M:
        case PLATFORM_MEDIA_TYPE_AR_4X10_10GBASE_CR1_1M:
        case PLATFORM_MEDIA_TYPE_AR_4X10_10GBASE_CR1_3M:
        case PLATFORM_MEDIA_TYPE_AR_4X10_10GBASE_CR1_5M:
        case PLATFORM_MEDIA_TYPE_AR_4X10_10GBASE_CR1_7M:
            sal_media_type = SAI_PORT_MEDIA_TYPE_QSFP_COPPER;
            break;
        default:
            NDI_PORT_LOG_ERROR("media type is not recognized %d \n", hal_media_type);
            sal_media_type = SAI_PORT_MEDIA_TYPE_UNKNONWN;
    }
    return (sal_media_type);
}

t_std_error ndi_port_media_type_set(npu_id_t npu_id, npu_port_t port_id, PLATFORM_MEDIA_TYPE_t media)
{
    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    sai_attribute_t sai_attr;
    sai_object_id_t sai_port;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((ret_code = ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK)) {
        return ret_code;
    }

    sai_attr.value.s32 = ndi_sai_port_media_type_translate(media);
    sai_attr.id = SAI_PORT_ATTR_MEDIA_TYPE;

    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->set_port_attribute(sai_port, &sai_attr))
                         != SAI_STATUS_SUCCESS) {
        return STD_ERR(NPU, CFG, sai_ret);
    }

    return ret_code;
}


t_std_error ndi_port_speed_get(npu_id_t npu_id, npu_port_t port_id, BASE_IF_SPEED_t *speed) {
    STD_ASSERT(speed!=NULL);

    /*  in case if link is not UP then return speed = 0 Mbps */
    ndi_intf_link_state_t link_state;
    t_std_error rc = ndi_port_link_state_get(npu_id, port_id, &link_state);
    if ((rc != STD_ERR_OK) || /* unable to read link state */
        ((rc == STD_ERR_OK) && (link_state.oper_status != ndi_port_OPER_UP)))  /*  Link is not UP */
    {
        *speed = BASE_IF_SPEED_0MBPS;
        return rc;
    }

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_SPEED;

    rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        if (!ndi_port_get_ndi_speed((uint32_t)sai_attr.value.u32, speed)) return STD_ERR(NPU, PARAM, 0);
    }

    return rc;
}

t_std_error ndi_port_stats_get(npu_id_t npu_id, npu_port_t port_id,
                               ndi_stat_id_t *ndi_stat_ids,
                               uint64_t* stats_val, size_t len)
{
    sai_object_id_t sai_port;
    const unsigned int list_len = len;
    sai_port_stat_counter_t sai_port_stats_ids[list_len];

    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        NDI_PORT_LOG_ERROR("Invalid NPU Id %d", npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((ret_code = ndi_sai_port_id_get(npu_id, port_id, &sai_port)) != STD_ERR_OK) {
        return ret_code;
    }
    size_t ix = 0;
    for ( ; ix < len ; ++ix){
        if(!ndi_to_sai_if_stats(ndi_stat_ids[ix],&sai_port_stats_ids[ix])){
            return STD_ERR(NPU,PARAM,0);
        }
    }

    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->get_port_stats(sai_port,
                   sai_port_stats_ids, len, stats_val))
                   != SAI_STATUS_SUCCESS) {
        NDI_PORT_LOG_TRACE("Port stats Get failed for npu %d, port %d, ret %d \n",
                            npu_id, port_id, sai_ret);
        return STD_ERR(NPU, FAIL, sai_ret);
    }

    return ret_code;
}

t_std_error ndi_port_set_untagged_port_attrib(npu_id_t npu_id,
                                              npu_port_t port_id,
                                              BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_t mode) {
    sai_attribute_t cur_modes[2];
    cur_modes[0].id = SAI_PORT_ATTR_DROP_TAGGED;
    cur_modes[1].id = SAI_PORT_ATTR_DROP_UNTAGGED;

    sai_attribute_t targ_modes[2];
    t_std_error rc= _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,cur_modes,2);
    if (rc!=STD_ERR_OK) {
        return rc;
    }
    memcpy(targ_modes,cur_modes,sizeof(cur_modes));
    targ_modes[0].value.booldata = !(mode == BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_HYBRID ||
            mode == BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_TAGGED );

    targ_modes[1].value.booldata = !(mode == BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_HYBRID ||
            mode == BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_UNTAGGED );

    rc= _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&targ_modes[0],1);
    if (rc!=STD_ERR_OK) {
        return rc;
    }

    rc= _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&targ_modes[1],1);
    if (rc!=STD_ERR_OK) {
        _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&cur_modes[0],1);
        return rc;
    }
    return rc;
}

t_std_error ndi_port_get_untagged_port_attrib(npu_id_t npu_id,
                                              npu_port_t port_id,
                                              BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_t *mode) {
    sai_attribute_t cur_modes[2];
    cur_modes[0].id = SAI_PORT_ATTR_DROP_TAGGED;
    cur_modes[1].id = SAI_PORT_ATTR_DROP_UNTAGGED;

    t_std_error rc= _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,cur_modes,2);
    if (rc!=STD_ERR_OK) {
        return rc;
    }
    if (cur_modes[0].value.booldata == false && cur_modes[1].value.booldata == false) {
        *mode = BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_HYBRID;
    }
    //if drop tagged and not untagged
    if (cur_modes[0].value.booldata == true && cur_modes[1].value.booldata == false) {
        *mode = BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_UNTAGGED;
    }

    if (cur_modes[0].value.booldata == false && cur_modes[1].value.booldata == true) {
        *mode = BASE_IF_PHY_IF_INTERFACES_INTERFACE_TAGGING_MODE_TAGGED;
    }

    return STD_ERR_OK;
}

t_std_error ndi_set_port_vid(npu_id_t npu_id, npu_port_t port_id,
                             hal_vlan_id_t vlan_id)
{
    t_std_error ret_code = STD_ERR_OK;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;

    sai_object_id_t sai_port;
    sai_attribute_t sai_attr;

    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);
    if (ndi_db_ptr == NULL) {
        return STD_ERR(NPU, PARAM, 0);
    }


    if ((ret_code = ndi_sai_port_id_get(npu_id, port_id, &sai_port) != STD_ERR_OK)) {
        return ret_code;
    }

    sai_attr.value.u16 = (sai_vlan_id_t)vlan_id;
    sai_attr.id = SAI_PORT_ATTR_PORT_VLAN_ID;

    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->set_port_attribute(sai_port, &sai_attr))
                                  != SAI_STATUS_SUCCESS) {
        return STD_ERR(INTERFACE, CFG, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_port_mac_learn_mode_set(npu_id_t npu_id, npu_port_t port_id,
                                        BASE_IF_PHY_MAC_LEARN_MODE_t mode){
    sai_attribute_t sai_attr;
    sai_attr.value.u32 = (sai_port_fdb_learning_mode_t )ndi_port_get_sai_mac_learn_mode(mode);
    sai_attr.id = SAI_PORT_ATTR_FDB_LEARNING;

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_mac_learn_mode_get(npu_id_t npu_id, npu_port_t port_id,
                                        BASE_IF_PHY_MAC_LEARN_MODE_t * mode){
    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_FDB_LEARNING;

    t_std_error rc;
    if((rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1)) != STD_ERR_OK){
        return rc;
    }

    *mode = (BASE_IF_PHY_MAC_LEARN_MODE_t)ndi_port_get_mac_learn_mode(sai_attr.value.u32);
    return STD_ERR_OK;
}


t_std_error ndi_port_clear_all_stat(npu_id_t npu_id, npu_port_t port_id){
    sai_object_id_t sai_port;
    sai_status_t sai_ret = SAI_STATUS_FAILURE;
    t_std_error rc;
    nas_ndi_db_t *ndi_db_ptr = ndi_db_ptr_get(npu_id);

    if (ndi_db_ptr == NULL) {
        NDI_PORT_LOG_ERROR("Invalid NPU Id %d", npu_id);
        return STD_ERR(NPU, PARAM, 0);
    }

    if ((rc = ndi_sai_port_id_get(npu_id, port_id, &sai_port)) != STD_ERR_OK) {
        NDI_PORT_LOG_TRACE("Failed to convert  npu %d and port %d to sai port",
                            npu_id, port_id);
        return rc;
    }

    if ((sai_ret = ndi_sai_port_api_tbl_get(ndi_db_ptr)->clear_port_all_stats(sai_port))
                   != SAI_STATUS_SUCCESS) {
        NDI_PORT_LOG_TRACE("Port stats clear failed for npu %d, port %d, ret %d ",
                            npu_id, port_id, sai_ret);
        return STD_ERR(NPU, FAIL, sai_ret);
    }

    return STD_ERR_OK;
}

t_std_error ndi_port_set_ingress_filtering(npu_id_t npu_id, npu_port_t port_id, bool ing_filter) {
    sai_attribute_t sai_attr;
    sai_attr.value.booldata = ing_filter;
    sai_attr.id = SAI_PORT_ATTR_INGRESS_FILTERING;

    return _sai_port_attr_set_or_get(npu_id, port_id, SAI_SG_ACT_SET, &sai_attr, 1);
}

t_std_error ndi_port_auto_neg_set(npu_id_t npu_id, npu_port_t port_id,
        bool enable) {

    sai_attribute_t sai_attr;

    sai_attr.value.booldata = enable;
    sai_attr.id = SAI_PORT_ATTR_AUTO_NEG_MODE;

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_auto_neg_get(npu_id_t npu_id, npu_port_t port_id, bool *enable)
{
    STD_ASSERT(enable != NULL);

    sai_attribute_t sai_attr;
    sai_attr.id = SAI_PORT_ATTR_AUTO_NEG_MODE;
    t_std_error rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        *enable  = sai_attr.value.booldata;
    }
    return rc;
}
t_std_error ndi_port_duplex_set(npu_id_t npu_id, npu_port_t port_id,
        BASE_CMN_DUPLEX_TYPE_t duplex) {

    sai_attribute_t sai_attr;
   /* TODO in case of AUTO set it to FULL: TBD */
    sai_attr.value.booldata = (duplex == BASE_CMN_DUPLEX_TYPE_HALF) ? false : true;
    sai_attr.id = SAI_PORT_ATTR_FULL_DUPLEX_MODE;

    return _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_SET,&sai_attr,1);
}

t_std_error ndi_port_duplex_get(npu_id_t npu_id, npu_port_t port_id,  BASE_CMN_DUPLEX_TYPE_t *duplex)
{
    STD_ASSERT(duplex != NULL);

    sai_attribute_t sai_attr;

    sai_attr.id = SAI_PORT_ATTR_FULL_DUPLEX_MODE;
    t_std_error rc = _sai_port_attr_set_or_get(npu_id,port_id,SAI_SG_ACT_GET,&sai_attr,1);
    if (rc==STD_ERR_OK) {
        *duplex  = (sai_attr.value.booldata == true) ?
                          BASE_CMN_DUPLEX_TYPE_FULL : BASE_CMN_DUPLEX_TYPE_HALF ;
    }
    return rc;
}
