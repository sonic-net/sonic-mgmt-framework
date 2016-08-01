
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
 * filename: nas_ndi_acl_utl.h
 */

#ifndef __NAS_NDI_ACL_UTL_H__
#define __NAS_NDI_ACL_UTL_H__

#include "saiacl.h"
#include <list>
#include <stdlib.h>

const sai_acl_api_t* ndi_acl_utl_api_get (const nas_ndi_db_t* ndi_db_ptr);

//////////////////////////////////////////////////////////
//Utilities to convert IDs from NDI to SAI and vice-versa
/////////////////////////////////////////////////////////
#define ndi_acl_utl_ndi2sai_table_id(x)   (sai_object_id_t) (x)
#define ndi_acl_utl_ndi2sai_entry_id(x)   (sai_object_id_t) (x)
#define ndi_acl_utl_ndi2sai_counter_id(x) (sai_object_id_t) (x)
#define ndi_acl_utl_sai2ndi_table_id(x)   (ndi_obj_id_t) (x)
#define ndi_acl_utl_sai2ndi_entry_id(x)   (ndi_obj_id_t) (x)
#define ndi_acl_utl_sai2ndi_counter_id(x) (ndi_obj_id_t) (x)

////////////////////////////////////////////////////////////////////////////////
// Utilities to map NAS-NDI values to SAI values and populate the SAI attribute
////////////////////////////////////////////////////////////////////////////////
t_std_error ndi_acl_utl_ndi2sai_filter_type (BASE_ACL_MATCH_TYPE_t ndi_filter_type,
                                             sai_attribute_t* attr_p);

t_std_error ndi_acl_utl_ndi2sai_action_type (BASE_ACL_ACTION_TYPE_t ndi_action_type,
                                             sai_attribute_t* attr_p);

t_std_error ndi_acl_utl_ndi2sai_tbl_filter_type (BASE_ACL_MATCH_TYPE_t ndi_filter_type,
                                                 sai_attribute_t* sai_attr_p);

t_std_error   ndi_acl_utl_fill_sai_filter (sai_attribute_t *sai_attr_p,
                                           const ndi_acl_entry_filter_t *ndi_filter_p,
                                           nas::mem_alloc_helper_t& malloc_tracker);

t_std_error   ndi_acl_utl_fill_sai_action (sai_attribute_t* sai_attr_p,
                                           const ndi_acl_entry_action_t* ndi_action_p,
                                           nas::mem_alloc_helper_t& malloc_tracker);

///////////////////////////////////////////////////////
// Utilities to Set/Get counter attributes to/from SAI
//////////////////////////////////////////////////////
t_std_error ndi_acl_utl_set_counter_attr (npu_id_t npu_id,
                                          ndi_obj_id_t ndi_counter_id,
                                          const sai_attribute_t* sai_counter_attr_p);

t_std_error ndi_acl_utl_get_counter_attr (npu_id_t npu_id,
                                          ndi_obj_id_t ndi_counter_id,
                                          sai_attribute_t* sai_counter_attr_p);

#endif
