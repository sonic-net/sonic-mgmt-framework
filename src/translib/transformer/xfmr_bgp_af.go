package transformer

import (
    "errors"
    "strings"
    log "github.com/golang/glog"
)

func init () {
	XlateFuncBind("YangToDb_bgp_afi_safi_use_multi_path_key_xfmr", YangToDb_bgp_afi_safi_use_multi_path_key_xfmr)
	XlateFuncBind("DbToYang_bgp_afi_safi_use_multi_path_key_xfmr", DbToYang_bgp_afi_safi_use_multi_path_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_table_conn_key_xfmr", YangToDb_bgp_table_conn_key_xfmr)
	XlateFuncBind("DbToYang_bgp_table_conn_key_xfmr", DbToYang_bgp_table_conn_key_xfmr)
}

var YangToDb_bgp_afi_safi_use_multi_path_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
	afiName := pathInfo.Var("afi-safi-name")
	afi := ""

	if strings.Contains(afiName, "IPV4_UNICAST") {
		afi = "ipv4"
	} else if strings.Contains(afiName, "IPV6_UNICAST") {
		afi = "ipv6"
	} else if strings.Contains(afiName, "L2VPN") {
		afi = "l2vpn"
	} else {
		log.Info("Unsupported AFI type " + afiName)
        return afi, errors.New("Unsupported AFI type " + afiName)
	}

	key := niName + "|" + afi
	
	log.Info("AFI key: ", key)

    return key, nil
}

var DbToYang_bgp_afi_safi_use_multi_path_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_afi_safi_use_multi_path_key_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")
	afi := ""

	switch mpathKey[1] {
	case "ipv4":
		afi = "IPV4_UNICAST"
	case "ipv6":
		afi = "IPV6_UNICAST"
	case "l2vpn":
		afi = "L2VPN_EVPN"
	}

    rmap["name"] = mpathKey[0]
    rmap["name#2"] = "BGP"
    rmap["afi-safi-name"] = afi

	log.Info("DbToYang_bgp_afi_safi_use_multi_path_key_xfmr: rmap:", rmap)
    return rmap, nil
}

var YangToDb_bgp_table_conn_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
	afName := pathInfo.Var("address-family")
	family := ""

	if strings.Contains(afName, "IPV4") {
		family = "ipv4"
	} else if strings.Contains(afName, "IPV6") {
		family = "ipv6"
	} else {
		log.Info("Unsupported address-family " + afName)
        return family, errors.New("Unsupported address-family " + afName)
	}

	key := niName + "|" + family
	
	log.Info("TableConnection key: ", key)

    return key, nil
}

var DbToYang_bgp_table_conn_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
	log.Info("DbToYang_bgp_table_conn_key_xfmr: key:", entry_key)

    connKey := strings.Split(entry_key, "|")
	conn := ""

	switch connKey[1] {
	case "ipv4":
		conn = "IPV4"
	case "ipv6":
		conn = "IPV6"
	}

    rmap["name"] = connKey[0]
    rmap["name#2"] = "BGP"
    rmap["address-family"] = conn

	log.Info("DbToYang_bgp_table_conn_key_xfmr: rmap:", rmap)
    return rmap, nil
}
