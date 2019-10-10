package transformer

const (
	YANG_MODULE    = "module"
	YANG_LIST      = "list"
	YANG_CONTAINER = "container"
	YANG_LEAF      = "leaf"
	YANG_LEAF_LIST = "leaflist"

	YANG_ANNOT_DB_NAME    = "redis-db-name"
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
