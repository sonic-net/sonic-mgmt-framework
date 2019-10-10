////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

const (
	YANG_MODULE    = "module"
	YANG_LIST      = "list"
	YANG_CONTAINER = "container"
	YANG_LEAF      = "leaf"
	YANG_LEAF_LIST = "leaflist"

	YANG_ANNOT_DB_NAME    = "db-name"
	YANG_ANNOT_TABLE_NAME = "table-name"
	YANG_ANNOT_FIELD_NAME = "field-name"
	YANG_ANNOT_KEY_DELIM  = "key-delimiter"
	YANG_ANNOT_TABLE_XFMR = "table-transformer"
	YANG_ANNOT_FIELD_XFMR = "field-transformer"
	YANG_ANNOT_KEY_XFMR   = "key-transformer"
	YANG_ANNOT_POST_XFMR  = "post-transformer"
	YANG_ANNOT_SUBTREE_XFMR  = "subtree-transformer"
	YANG_ANNOT_VALIDATE_FUNC = "get-validate"

	REDIS_DB_TYPE_APPLN   = "APPL_DB"
	REDIS_DB_TYPE_ASIC    = "ASIC_DB"
	REDIS_DB_TYPE_CONFIG  = "CONFIG_DB"
	REDIS_DB_TYPE_COUNTER = "COUNTERS_DB"
	REDIS_DB_TYPE_LOG_LVL = "LOGLEVEL_DB"
	REDIS_DB_TYPE_STATE   = "STATE_DB"
	REDIS_DB_TYPE_FLX_COUNTER = "FLEX_COUNTER_DB"

	XPATH_SEP_FWD_SLASH = "/"
	XFMR_EMPTY_STRING   = ""
	SONIC_TABLE_INDEX = 2

)
