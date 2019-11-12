package transformer

import (
    "errors"
    "strings"
    log "github.com/golang/glog"
)

func init () {
	XlateFuncBind("YangToDb_bgp_gbl_afi_safi_key_xfmr", YangToDb_bgp_gbl_afi_safi_key_xfmr)
	XlateFuncBind("DbToYang_bgp_gbl_afi_safi_key_xfmr", DbToYang_bgp_gbl_afi_safi_key_xfmr) 
}

var YangToDb_bgp_gbl_afi_safi_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

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

var DbToYang_bgp_gbl_afi_safi_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: ", entry_key)

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

    rmap["afi-safi-name"] = afi

	log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: rmap:", rmap)
    return rmap, nil
}


