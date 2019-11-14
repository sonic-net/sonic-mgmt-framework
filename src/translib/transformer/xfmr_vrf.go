package transformer

import (
        "errors"
        log "github.com/golang/glog"
        "github.com/openconfig/ygot/ygot"
        "strings"
        "translib/ocbinds"
)

type NwInstMapKey struct {
        NwInstName    string
        NwInstType    string
}

const (
        MGMT_VRF_ENABLE        = "mgmtVrfEnabled"
)

const (
        MGMT_VRF_NAME          = "mgmt-vrf-name"
)

var nwInstTypeMap = map[ocbinds.E_OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE] string {
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_UNSET: "",
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_DEFAULT_INSTANCE: "DEFAULT_INSTANCE",
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_L2L3: "L2L3",
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_L2P2P: "L2P2P",
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_L2VSI: "L2VSI",
        ocbinds.OpenconfigNetworkInstanceTypes_NETWORK_INSTANCE_TYPE_L3VRF: "L3VRF",
}

/* Top level network instance table name based on key name and type */
var NwInstTblNameMapWithNameAndType = map[NwInstMapKey]string {
        {NwInstName: "mgmt", NwInstType: "L3VRF"}: "MGMT_VRF_CONFIG",
        {NwInstName: "Vrf",  NwInstType: "L3VRF"}: "VRF",
        {NwInstName: "default", NwInstType: "L3VRF"}: "",
}

/* Top level network instance table name based on key name */
var NwInstTblNameMapWithName = map[string]string {
	"mgmt": "MGMT_VRF_CONFIG",
	"Vrf": "VRF",
	"default": "",
}

/*
 * Get internal network instance name based on the incoming network instance name
 * and use it for top level table map lookup
 */
func getInternalNwInstName (name string) (string, error) {
        var err error

        if name == "" {
                return "", errors.New("Network instance name is empty")
        } else if (strings.Compare(name, "mgmt") == 0) {
                return "mgmt", err
        } else if (strings.HasPrefix(name, "Vrf") == true) {
                return "Vrf", err
        } else if (strings.Compare(name, "default") == 0) {
                return "default", err
        } else {
                /* For other types */
                return "", errors.New("Network instance name uknonwn")
        }
}

/* Get table entry key based on the network instance name */
func getVrfTblKeyByName (name string) (string) {
        var vrf_key string

        if name == "" {
                /* Shouldn't even come here */
                log.Info("getVrfTblKeyByName:  network instance name is empty")
                return  vrf_key
        }

        if (strings.Compare(name, "mgmt") == 0) { 
                vrf_key = "vrf_global"
        } else {
                vrf_key = name
        }

        log.Info("getVrfTblKeyByName: vrf key is ", vrf_key)

        return vrf_key
}

/* Check if this is "MGMT_VRF_CONFIG" */
func isMgmtVrfDbTbl (inParams XfmrParams) (bool) {
        data := (*inParams.dbDataMap)[inParams.curDb]
        log.Info("isMgmtVrfDbTbl ", data, "inParams :", inParams)

        mgmtTbl := data["MGMT_VRF_CONFIG"]
        if (mgmtTbl != nil) {
                return true
        } else {
                return false
        }
}

/* Check if this is "VRF" table */
func isVrfDbTbl (inParams XfmrParams) (bool)  {
        data := (*inParams.dbDataMap)[inParams.curDb]
        log.Info("isVrfDbTbl: ", data, "inParams :", inParams)

        vrfTbl := data["VRF"]
        if (vrfTbl != nil) {
                return true
        } else {
                return false
        }
}

/* Check if "mgmtVrfEnabled" is set to true in the "MGMT_VRF_CONFIG" table */
func mgmtVrfEnabledInDb (inParams XfmrParams) (string) {
        data := (*inParams.dbDataMap)[inParams.curDb]
        log.Info("mgmtVrfEnabledInDb ", data, "inParams :", inParams)

        mgmtTbl := data["MGMT_VRF_CONFIG"]
        mgmtVrf := mgmtTbl[inParams.key]
        enabled_status := mgmtVrf.Field["mgmtVrfEnabled"]
        return enabled_status;
}

/* Get the top level network instance type. Note this is used for the create, update only */
func getNwInstType (nwInstObj *ocbinds.OpenconfigNetworkInstance_NetworkInstances, keyName string) (string, error) {
        var err error

        nwInstType := nwInstTypeMap[nwInstObj.NetworkInstance[keyName].Config.Type]
        if nwInstType == "" {
                return nwInstType, errors.New("network instance type not set")
        }

        return nwInstType, err
}

/* Check if this is mgmt vrf configuration. Note this is used for create, update only */
func isMgmtVrf(inParams XfmrParams) (bool, error) {
        var err error

        log.Info("isMgmtVrf ")
        nwInstObj := getNwInstRoot(inParams.ygRoot)
        if nwInstObj.NetworkInstance == nil {
                /* Should not even come here */
                return false, errors.New("No network instance in the path")
        }

        pathInfo := NewPathInfo(inParams.uri)

        /* get the name at the top network-instance table level, this is the key */
        keyName := pathInfo.Var("name")
        oc_nwInstType, ierr := getNwInstType(nwInstObj, keyName)
        if (ierr != nil) {
                return false, errors.New("Network instance type not set")
        }

        if ((strings.Compare(keyName, "mgmt") == 0) &&
            (strings.Compare(oc_nwInstType, "L3VRF") == 0)) {
                return true, err
        } else {
                return false, err
        }
}

func init() {
        XlateFuncBind("network_instance_table_name_xfmr", network_instance_table_name_xfmr)
        XlateFuncBind("YangToDb_network_instance_table_key_xfmr", YangToDb_network_instance_table_key_xfmr)
        XlateFuncBind("DbToYang_network_instance_table_key_xfmr", DbToYang_network_instance_table_key_xfmr)
        XlateFuncBind("YangToDb_network_instance_enabled_field_xfmr", YangToDb_network_instance_enabled_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_enabled_field_xfmr", DbToYang_network_instance_enabled_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_name_field_xfmr", YangToDb_network_instance_name_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_name_field_xfmr", DbToYang_network_instance_name_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_type_field_xfmr", YangToDb_network_instance_type_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_type_field_xfmr", DbToYang_network_instance_type_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_mtu_field_xfmr", YangToDb_network_instance_mtu_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_mtu_field_xfmr", DbToYang_network_instance_mtu_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_description_field_xfmr", YangToDb_network_instance_description_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_description_field_xfmr", DbToYang_network_instance_description_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_router_id_field_xfmr", YangToDb_network_instance_router_id_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_router_id_field_xfmr", DbToYang_network_instance_router_id_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_route_distinguisher_field_xfmr", YangToDb_network_instance_route_distinguisher_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_route_distinguisher_field_xfmr", DbToYang_network_instance_route_distinguisher_field_xfmr)
        XlateFuncBind("YangToDb_network_instance_enabled_addr_family_field_xfmr", YangToDb_network_instance_enabled_addr_family_field_xfmr)
        XlateFuncBind("DbToYang_network_instance_enabled_addr_family_field_xfmr", DbToYang_network_instance_enabled_addr_family_field_xfmr)
}

func getNwInstRoot(s *ygot.GoStruct) *ocbinds.OpenconfigNetworkInstance_NetworkInstances  {
        deviceObj := (*s).(*ocbinds.Device)
        return deviceObj.NetworkInstances
}

/* Table name in config DB correspoinding to the top level network instance name */
var network_instance_table_name_xfmr TableXfmrFunc = func (inParams XfmrParams)  ([]string, error) {
        var tblList []string
        var err error

        log.Info("network_instance_table_name_xfmr")

        nwInstObj := getNwInstRoot(inParams.ygRoot)
        if nwInstObj.NetworkInstance == nil {
                return tblList, errors.New("Network instance not set")
        }

        pathInfo := NewPathInfo(inParams.uri)

        /* get the name at the top network-instance table level, this is the key */
        keyName := pathInfo.Var("name")

        if keyName == "" {
                /* for GET with no keyName, return table name for mgmt VRF and data VRF */
                if (inParams.oper == GET) {
                        //tblList  = {"MGMT_VRF_CONFIG", "VRF"}
                        tblList = append(tblList , "MGMT_VRF_CONFIG")
                        tblList = append(tblList, "VRF")
                        log.Info("network_instance_table_name_xfmr: tblList ", tblList)
                        return tblList, err
                } else {
                        log.Info("network_instance_table_name_xfmr, for key name not present")
                        return tblList, errors.New("Empty network instance name")
                }
        }

        /* get internal network instance name in order to fetch the DB table name */
        intNwInstName, ierr := getInternalNwInstName(keyName)
        if intNwInstName == "" || ierr != nil {
                log.Info("network_instance_table_name_xfmr, invalid network instance name ", keyName)
                return tblList, errors.New("Invalide network instance name")
        }

        /*
         * for GET or DELETE, use the key "name" only to get the DB table name
         * for CREATE, UPDATE or REPLACE, use the key "name" and type to get the DB table name
         */
        if ((inParams.oper == CREATE) ||
            (inParams.oper == UPDATE) ||
            (inParams.oper == REPLACE))  {
                /* get the type for the network instance config */
                oc_nwInstType, ierr := getNwInstType(nwInstObj, keyName)
                if (ierr != nil ) {
                        log.Info("network_instance_table_name_xfmr, network instance type not correct")
                        return tblList, errors.New("network instance type incorrect")
                }

                log.Info("network_instance_table_name_xfmr, name ", keyName)
                log.Info("network_instance_table_name_xfmr, type ", oc_nwInstType)

                tblList = append(tblList, NwInstTblNameMapWithNameAndType[NwInstMapKey{intNwInstName, oc_nwInstType}])
        } else {
                tblList = append(tblList, NwInstTblNameMapWithName[intNwInstName])
        }

        log.Info("network_instance_table_name_xfmr,  DB table name ", tblList)

        return tblList, err
}

/* YangToDB Field transformer for top level network instance config "enabled" */
var YangToDb_network_instance_enabled_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_enabled_field_xfmr")

        nwInstObj := getNwInstRoot(inParams.ygRoot)
        if nwInstObj.NetworkInstance == nil {
                return res_map, errors.New("Network instance not set")
        }

        pathInfo := NewPathInfo(inParams.uri)

        if len(pathInfo.Vars) < 1 {
                /* network instance table has 1 key "name" */
                return res_map, errors.New("Invalid xpath, key attributes not found")
        }

        targetUriPath, err := getYangPathFromUri(pathInfo.Path)

        log.Info("YangToDb_network_instance_enabled_field_xfmr targetUri: ", targetUriPath)

        /* get the name at the top network-instance table level, this is the key */
        keyName := pathInfo.Var("name")
        if keyName != "mgmt" {
                log.Info("YangToDb_network_instance_enabled_field_xfmr, not mgmt vrf ", keyName)

                /* ToDo, put this until sonic yang default value is implemented */
                res_map["fallback"] = "false"
                return res_map, err 
        }

        enabled, _ := inParams.param.(*bool)

        var enStr string
        if *enabled == true {
                enStr = "true"
        } else {
                enStr = "false"
        }

        res_map[MGMT_VRF_ENABLE] = enStr
        log.Info("YangToDb_network_instance_enabled_field_xfmr: ", res_map)

        return res_map, err
}

/* DbToYang Field transformer for top level network instance config "enabled" */
var DbToYang_network_instance_enabled_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_enabled_field_xfmr: ")

        if (mgmtVrfEnabledInDb(inParams) == "true") {
                res_map["enabled"] = true
        } else if (mgmtVrfEnabledInDb(inParams) == "false") {
                res_map["enabled"] = false
        }
        return res_map, err
}

/* YangToDB key transformer for top level network instance */
var YangToDb_network_instance_table_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
        var vrfTbl_key  string
        var err error

        log.Info("YangToDb_network_instance_table_key_xfmr: ")

        pathInfo := NewPathInfo(inParams.uri)

        /* Get key for the respective table based on the network instance key "name" */
        vrfTbl_key = getVrfTblKeyByName(pathInfo.Var("name"))

        log.Info("YangToDb_network_instance_table_key_xfmr: ", vrfTbl_key)

        return vrfTbl_key, err
}

/* DbToYang key transformer for top level network instance */
var DbToYang_network_instance_table_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_table_key_xfmr: ", inParams.key)

        return  res_map, err
}

/* YangToDb Field transformer for name in the top level network instance config */
var YangToDb_network_instance_name_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_name_field_xfmr")

        /* the key name is not repeated as attr name in the DB */
        res_map["NULL"] = "NULL"

        return res_map, err
}

/* DbToYang Field transformer for name in the top level network instance config */
var DbToYang_network_instance_name_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_name_field_xfmr")

        if (isMgmtVrfDbTbl(inParams) == true) {
                res_map["name"] = "mgmt"
        } else if (isVrfDbTbl(inParams) == true){
                res_map["name"] = inParams.key
        }

        /* ToDo in else if cases */
        return  res_map, err
}

/* YangToDb Field transformer for type in the top level network instance config */
var YangToDb_network_instance_type_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_type_field_xfmr")

        return res_map, err
}

/* DbToYang Field transformer for type in the top level network instance config */
var DbToYang_network_instance_type_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_type_field_xfmr")

        if ((isMgmtVrfDbTbl(inParams) == true) || (isVrfDbTbl(inParams) == true)) {
                res_map["type"] = "L3VRF"
        }

        /* ToDo in else if cases */
        return  res_map, err
}

/* YangToDb Field transformer for enabled_address_family in the top level network instance config */
var YangToDb_network_instance_enabled_addr_family_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_enabled_addr_fam_field_xfmr")

        return res_map, err
}

/* DbToYang Field transformer for enabled_address_family in the top level network instance config */
var DbToYang_network_instance_enabled_addr_family_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_enabled_addr_fam_field_xfmr")

        return res_map, err
}

/* YangToDb Field transformer for mtu in the top level network instance config */
var YangToDb_network_instance_mtu_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_mtu_field_xfmr")

        return res_map, err
}

/* DbToYang Field transformer for mtu in the top level network instance config */
var DbToYang_network_instance_mtu_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_mtu_field_xfmr")

        return res_map, err
}

/* YangToDb Field transformer for description in the top level network instance config */
var YangToDb_network_instance_description_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_description_field_xfmr")

        return res_map, err
}

/* DbToYang Field transformer for description in the top level network instance config */
var DbToYang_network_instance_description_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_description_field_xfmr")

        return res_map, err
}

/* YangToDb Field transformer for router_id in the top level network instance config */
var YangToDb_network_instance_router_id_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_router_id_field_xfmr")

        return res_map, err
}

/* DbToYang Field transformer for router_id in the top level network instance config */
var DbToYang_network_instance_router_id_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_router_id_field_xfmr")

        return res_map, err
}

/* TBD for data vrf YangToDb Field transformer for route_distinguisher in the top level network instance config */
var YangToDb_network_instance_route_distinguisher_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
        res_map := make(map[string]string)
        var err error

        log.Info("YangToDb_network_instance_route_distinguisher_field_xfmr")

        return res_map, err
}

/* TBD for data vrf DbToYang Field transformer for route_distinguisher in the top level network instance config */
var DbToYang_network_instance_route_distinguisher_field_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
        res_map := make(map[string]interface{})
        var err error

        log.Info("DbToYang_network_instance_route_distinguisher_field_xfmr")

        return res_map, err
}


