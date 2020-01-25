package transformer

import (
    "errors"
    "strconv"
    "strings"
    "encoding/json"
    "translib/ocbinds"
    "translib/tlerr"
    "translib/db"
    "io"
    "bytes"
    "net"
    "encoding/binary"
    log "github.com/golang/glog"
)
const sock_addr = "/etc/sonic/frr/bgpd_client_sock"

func getBgpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp, string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
    var err error

    if len(pathInfo.Vars) <  3 {
        return nil, "", errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return nil, "", errors.New("vrf name is missing")
    }
    if strings.Contains(bgpId,"BGP") == false {
        return nil, "", errors.New("BGP ID is missing")
    }
    if len(protoName) == 0 {
        return nil, "", errors.New("Protocol Name is missing")
    }

	deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    netInstsObj := deviceObj.NetworkInstances

    if netInstsObj.NetworkInstance == nil {
        return nil, "", errors.New("Network-instances container missing")
    }

    netInstObj := netInstsObj.NetworkInstance[niName]
    if netInstObj == nil {
        return nil, "", errors.New("Network-instance obj missing")
    }

    if netInstObj.Protocols == nil || len(netInstObj.Protocols.Protocol) == 0 {
        return nil, "", errors.New("Network-instance protocols-container missing or protocol-list empty")
    }

    var protoKey ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Key
    protoKey.Identifier = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_BGP
    protoKey.Name = protoName
    protoInstObj := netInstObj.Protocols.Protocol[protoKey]
    if protoInstObj == nil {
        return nil, "", errors.New("Network-instance BGP-Protocol obj missing")
    }
    return protoInstObj.Bgp, niName, err
}
func exec_vtysh_cmd (vtysh_cmd string) (map[string]interface{}, error) {
    var err error
    oper_err := errors.New("Operational error")
    
    log.Infof("Going to connect UDS socket call to reach BGP container  ==> \"%s\"", vtysh_cmd)
    conn, err := net.DialUnix("unix", nil, &net.UnixAddr{sock_addr, "unix"})
    if err != nil {
        log.Infof("Failed to connect proxy server: %s\n", err)
        return nil, oper_err
    }
    defer conn.Close()
    bs := make([]byte, 4)
    binary.BigEndian.PutUint32(bs, uint32(len(vtysh_cmd)))
    _, err = conn.Write(bs)
    if err != nil {
        log.Infof("Failed to write command length to server: %s\n", err)
        return nil, oper_err
    }
    _, err = conn.Write([]byte(vtysh_cmd))
    if err != nil {
        log.Infof("Failed to write command length to server: %s\n", err)
        return nil, oper_err
    }
    log.Infof("Reading data from server\n")
    var buffer bytes.Buffer
    data := make([]byte, 10240)
    for {
        count, err := conn.Read(data)
        if err == io.EOF {
            log.Infof("End reading\n")
            break
        }
        if err != nil {
            log.Infof("Failed to read from server: %s\n", err)
            return nil, oper_err
        }
        buffer.WriteString(string(data[:count]))
    }
    json_str := buffer.String()
    log.Infof("Successfully got data from BGP container thru UDS socket ==> \"%s\"", vtysh_cmd)

    log.Infof("Data decoding started ==> \"%s\"", vtysh_cmd)
    var outputJson map[string]interface{}
    err = json.NewDecoder(strings.NewReader(json_str)).Decode(&outputJson)
    if err != nil {
        log.Infof("Not able to decode vtysh json output: %s\n", err)
        return nil, oper_err
    }

    if outputJson == nil {
        log.Infof("VTYSH output empty\n")
        return nil, oper_err
    }
    log.Infof("Data decoding completed ==> \"%s\"", vtysh_cmd)

    return outputJson, err
}

func init () {
    XlateFuncBind("YangToDb_bgp_gbl_tbl_key_xfmr", YangToDb_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_tbl_key_xfmr", DbToYang_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_local_asn_fld_xfmr", YangToDb_bgp_local_asn_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_local_asn_fld_xfmr", DbToYang_bgp_local_asn_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_gbl_afi_safi_field_xfmr", YangToDb_bgp_gbl_afi_safi_field_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_afi_safi_field_xfmr", DbToYang_bgp_gbl_afi_safi_field_xfmr)
	XlateFuncBind("YangToDb_bgp_dyn_neigh_listen_key_xfmr", YangToDb_bgp_dyn_neigh_listen_key_xfmr)
	XlateFuncBind("DbToYang_bgp_dyn_neigh_listen_key_xfmr", DbToYang_bgp_dyn_neigh_listen_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_gbl_afi_safi_key_xfmr", YangToDb_bgp_gbl_afi_safi_key_xfmr)
	XlateFuncBind("DbToYang_bgp_gbl_afi_safi_key_xfmr", DbToYang_bgp_gbl_afi_safi_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_gbl_afi_safi_addr_key_xfmr", YangToDb_bgp_gbl_afi_safi_addr_key_xfmr)
	XlateFuncBind("DbToYang_bgp_gbl_afi_safi_addr_key_xfmr", DbToYang_bgp_gbl_afi_safi_addr_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_dyn_neigh_listen_field_xfmr", YangToDb_bgp_dyn_neigh_listen_field_xfmr)
	XlateFuncBind("DbToYang_bgp_dyn_neigh_listen_field_xfmr", DbToYang_bgp_dyn_neigh_listen_field_xfmr) 
	XlateFuncBind("YangToDb_bgp_gbl_afi_safi_addr_field_xfmr", YangToDb_bgp_gbl_afi_safi_addr_field_xfmr)
	XlateFuncBind("DbToYang_bgp_gbl_afi_safi_addr_field_xfmr", DbToYang_bgp_gbl_afi_safi_addr_field_xfmr) 
    XlateFuncBind("YangToDb_bgp_global_subtree_xfmr", YangToDb_bgp_global_subtree_xfmr)
    XlateFuncBind("rpc_clear_bgp", rpc_clear_bgp)
}

func bgp_global_get_local_asn(d *db.DB , niName string, tblName string) (string, error) {
    var err error

    dbspec := &db.TableSpec { Name: tblName }

    log.Info("bgp_global_get_local_asn", db.Key{Comp: []string{niName}})
    dbEntry, err := d.GetEntry(dbspec, db.Key{Comp: []string{niName}})
    if err != nil {
        return "", err
    }
    asn, ok := dbEntry.Field["local_asn"]
    if ok {
        log.Info("Current ASN ", asn)
    } else {
        log.Info("No ASN assigned")
    }
    return asn, nil;
}


var YangToDb_bgp_local_asn_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error
    if inParams.param == nil {
        rmap["local_asn"] = ""
        return rmap, err
    }

    if inParams.oper == DELETE {
        rmap["local_asn"] = ""
        return rmap, nil
    }

    log.Info("YangToDb_bgp_local_asn_fld_xfmr")
    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")

    asn, _ := inParams.param.(*uint32)

    curr_asn, err_val := bgp_global_get_local_asn (inParams.d, niName, "BGP_GLOBALS")
    if err_val == nil {
       local_asn64, err_conv := strconv.ParseUint(curr_asn, 10, 32)
       local_asn := uint32(local_asn64)
       if err_conv == nil && local_asn != *asn {
           log.Info("YangToDb_bgp_local_asn_fld_xfmr Local ASN is already present", local_asn, *asn)
           return rmap, tlerr.InvalidArgs("BGP is already running with AS number %s", 
                                          strconv.FormatUint(local_asn64, 10))
       }
    }
    rmap["local_asn"] = strconv.FormatUint(uint64(*asn), 10)
    return rmap, err
}

var DbToYang_bgp_local_asn_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_local_asn_fld_xfmr: ")

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        return result, err
    }
    pGblKey := pTbl[inParams.key]
    curr_asn, ok := pGblKey.Field["local_asn"]
    if ok {
       local_asn64, _:= strconv.ParseUint(curr_asn, 10, 32)
       local_asn := uint32(local_asn64)
       result["as"] = local_asn
    } else {
        log.Info("Local ASN field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_gbl_afi_safi_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error

    log.Info("YangToDb_bgp_gbl_afi_safi_field_xfmr")
    rmap["NULL"] = "NULL"
    
    return rmap, err
}

var DbToYang_bgp_gbl_afi_safi_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_field_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")
	afi := ""

	switch mpathKey[1] {
	case "ipv4_unicast":
		afi = "IPV4_UNICAST"
	case "ipv6_unicast":
		afi = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afi = "L2VPN_EVPN"
    default:
        return rmap, nil
	}

    rmap["afi-safi-name"] = afi

    return rmap, err
}

var YangToDb_bgp_dyn_neigh_listen_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error

    log.Info("YangToDb_bgp_dyn_neigh_listen_field_xfmr")
    rmap["NULL"] = "NULL"
    
    return rmap, err
}

var YangToDb_bgp_gbl_afi_safi_addr_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error

    log.Info("YangToDb_bgp_gbl_afi_safi_addr_field_xfmr")
    rmap["NULL"] = "NULL"
    
    return rmap, err
}


var DbToYang_bgp_dyn_neigh_listen_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    
    entry_key := inParams.key
    log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: ", entry_key)

    dynKey := strings.Split(entry_key, "|")

    rmap["prefix"] = dynKey[1]

    return rmap, err
}

var DbToYang_bgp_gbl_afi_safi_addr_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_addr_field_xfmr: ", entry_key)

    dynKey := strings.Split(entry_key, "|")

    rmap["prefix"] = dynKey[2]

    return rmap, err
}

var YangToDb_bgp_gbl_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")

    if len(pathInfo.Vars) <  3 {
        return "", errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return "", errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return "", errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return "", errors.New("Protocol Name is missing")
    }

    log.Info("URI VRF ", niName)

    return niName, err
}

var DbToYang_bgp_gbl_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_tbl_key: ", entry_key)

    rmap["name"] = entry_key
    return rmap, err
}

var YangToDb_bgp_dyn_neigh_listen_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	log.Info("YangToDb_bgp_dyn_neigh_listen_key_xfmr key: ", inParams.uri)

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
	prefix := pathInfo.Var("prefix")

    if len(pathInfo.Vars) < 4 {
        return "", errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return "", errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return "", errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return "", errors.New("Protocol Name is missing")
    }

	key := niName + "|" + prefix
	
	log.Info("YangToDb_bgp_dyn_neigh_listen_key_xfmr key: ", key)

    return key, nil
}

var DbToYang_bgp_dyn_neigh_listen_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: ", entry_key)

    dynKey := strings.Split(entry_key, "|")

    rmap["prefix"] = dynKey[1]

	log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: rmap:", rmap)
    return rmap, nil
}

var YangToDb_bgp_gbl_afi_safi_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
	afName := pathInfo.Var("afi-safi-name")
	afi := ""
    var err error

    if len(pathInfo.Vars) < 4 {
        return afi, errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return afi, errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return afi, errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return afi, errors.New("Protocol Name is missing")
    }

	if strings.Contains(afName, "IPV4_UNICAST") {
		afi = "ipv4_unicast"
	} else if strings.Contains(afName, "IPV6_UNICAST") {
		afi = "ipv6_unicast"
	} else if strings.Contains(afName, "L2VPN_EVPN") {
		afi = "l2vpn_evpn"
	} else {
		log.Info("Unsupported AFI type " + afName)
        return afi, errors.New("Unsupported AFI type " + afName)
	}

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "IPV4_UNICAST"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		    log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") {
        afName = "IPV6_UNICAST"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		    log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "L2VPN_EVPN"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "ipv4-unicast") {
		    err = errors.New("L2VPN_EVPN supported only on l2vpn-evpn container")
		    log.Info("L2VPN_EVPN supported only on l2vpn-evpn container: ", afName);
		    return afName, err
        }
    } else  {
	    err = errors.New("Unsupported AFI SAFI")
	    log.Info("Unsupported AFI SAFI ", afName);
	    return afName, err
    }

	key := niName + "|" + afi
	
	log.Info("AFI key: ", key)

    return key, nil
}

var DbToYang_bgp_gbl_afi_safi_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")
	afi := ""

	switch mpathKey[1] {
	case "ipv4_unicast":
		afi = "IPV4_UNICAST"
	case "ipv6_unicast":
		afi = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afi = "L2VPN_EVPN"
    default:
        return rmap, nil
	}

    rmap["afi-safi-name"] = afi

	log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: rmap:", rmap)
    return rmap, nil
}

var YangToDb_bgp_global_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
	log.Info("YangToDb_bgp_global_subtree_xfmr:", inParams.oper)
    if inParams.oper == DELETE {
        return nil, errors.New("Invalid request")
    }
    return nil, err
}

var YangToDb_bgp_gbl_afi_safi_addr_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
	afName := pathInfo.Var("afi-safi-name")
	prefix := pathInfo.Var("prefix")
	afi := ""
    var err error

    if len(pathInfo.Vars) < 5 {
        return afi, errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return afi, errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return afi, errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return afi, errors.New("Protocol Name is missing")
    }

	if strings.Contains(afName, "IPV4_UNICAST") {
		afi = "ipv4_unicast"
	} else if strings.Contains(afName, "IPV6_UNICAST") {
		afi = "ipv6_unicast"
	} else if strings.Contains(afName, "L2VPN_EVPN") {
		afi = "l2vpn_evpn"
	} else {
		log.Info("Unsupported AFI type " + afName)
        return afi, errors.New("Unsupported AFI type " + afName)
	}

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "IPV4_UNICAST"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		    log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") {
        afName = "IPV6_UNICAST"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		    log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "L2VPN_EVPN"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "ipv4-unicast") {
		    err = errors.New("L2VPN_EVPN supported only on l2vpn-evpn container")
		    log.Info("L2VPN_EVPN supported only on l2vpn-evpn container: ", afName);
		    return afName, err
        }
    } else  {
	    err = errors.New("Unsupported AFI SAFI")
	    log.Info("Unsupported AFI SAFI ", afName);
	    return afName, err
    }

	key := niName + "|" + afi + "|" + prefix
	
	log.Info("YangToDb_bgp_gbl_afi_safi_addr_key_xfmr AFI key: ", key)

    return key, nil
}

var DbToYang_bgp_gbl_afi_safi_addr_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_addr_key_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")

    rmap["prefix"] = mpathKey[2]

	log.Info("DbToYang_bgp_gbl_afi_safi_addr_key_xfmr: rmap:", rmap)
    return rmap, nil
}


var rpc_clear_bgp RpcCallpoint = func(body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error) {
    log.Info("In rpc_clear_bgp")
    var err error
    var status string
    var clear_all string
    var af_str, vrf_name, all, soft, in, out,neigh_address, prefix, peer_group, asn, intf, external string
    var cmd string
    var mapData map[string]interface{}
    err = json.Unmarshal(body, &mapData)
    if err != nil {
        log.Info("Failed to unmarshall given input data")
        return nil, err
    }

    var result struct {
        Output struct {
              Status string `json:"response"`
        } `json:"sonic-bgp-clear:output"`
    }

    log.Info("In rpc_clear_bgp", mapData)

    input, _ := mapData["sonic-bgp-clear:input"]
    mapData = input.(map[string]interface{})

    log.Info("In rpc_clear_bgp", mapData)

    if value, ok := mapData["clear-all"].(bool) ; ok {
        log.Info("In clearall", value)
        if value == true {
           clear_all = "* "
        }
    }

    log.Info("In clearall", clear_all)
    if value, ok := mapData["vrf-name"].(string) ; ok {
        log.Info("In vrf", value)
        if value != "" {
            vrf_name = "vrf " + value + " "
        }
    }

    if value, ok := mapData["family"].(string) ; ok {
        if value == "IPv4" {
            af_str = "ipv4 "
        } else if value == "IPv6" {
            af_str = "ipv6 "
        }
    }

    if value, ok := mapData["all"].(bool) ; ok {
        if value == true {
           all = "* "
        }
    }

    if value, ok := mapData["external"].(bool) ; ok {
        if value == true {
           external = "external "
        }
    }

    if value, ok := mapData["address"].(string) ; ok {
        if value != "" {
            neigh_address = value + " "
        }
    }

    if value, ok := mapData["interface"].(string) ; ok {
        if value != "" {
            intf = value + " "
        }
    }

    if value, ok := mapData["asn"].(float64) ; ok {
        _asn := uint64(value)
        asn = strconv.FormatUint(_asn, 10)
        asn = asn + " "
    }

    if value, ok := mapData["prefix"].(string) ; ok {
        if value != "" {
            prefix = "prefix " + value + " "
        }
    }

    if value, ok := mapData["peer-group"].(string) ; ok {
        if value != "" {
            peer_group = "peer_group " + value + " "
        }
    }

    if value, ok := mapData["in"].(bool) ; ok {
        if value == true {
           in = "in "
        }
    }

    if value, ok := mapData["out"].(bool) ; ok {
        if value == true {
            out = "out "
        }
    }

    if value, ok := mapData["soft"].(bool) ; ok {
        if value == true {
            soft = "soft "
        }
    }

    log.Info("In rpc_clear_bgp", clear_all, vrf_name, af_str, all, neigh_address, intf, asn, prefix, peer_group, in, out, soft)

    if clear_all != "" {
        cmd = "clear ip bgp " + clear_all
    } else {
        cmd = "clear ip bgp "
        if vrf_name != "" {
            cmd = "clear ip bgp " + vrf_name
        }

        if af_str != "" {
            cmd = cmd + af_str
        }

        if neigh_address != "" {
            cmd = cmd + neigh_address
        }

        if intf != "" {
            cmd = cmd + intf
        }

        if prefix != "" {
            cmd = cmd + prefix
        }

        if peer_group != "" {
            cmd = cmd + peer_group
        }

        if external != "" {
            cmd = cmd + external
        }

        if asn != "" {
            cmd = cmd + asn
        }

        if all != "" {
            cmd = cmd + all
        }

        if soft != "" {
            cmd = cmd + soft
        }

        if in != "" {
            cmd = cmd + in
        }

        if out != "" {
            cmd = cmd + out
        }

    }
    cmd = strings.TrimSuffix(cmd, " ")
    exec_vtysh_cmd (cmd)
    status = "Success"
    result.Output.Status = status
    return json.Marshal(&result)
}
