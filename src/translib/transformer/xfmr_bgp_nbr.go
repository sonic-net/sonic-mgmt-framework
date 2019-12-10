package transformer

import (
    "errors"
    "strings"
    "translib/ocbinds"
    "translib/db"
    "strconv"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("YangToDb_bgp_nbr_tbl_key_xfmr", YangToDb_bgp_nbr_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_tbl_key_xfmr", DbToYang_bgp_nbr_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_address_fld_xfmr", YangToDb_bgp_nbr_address_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_address_fld_xfmr", DbToYang_bgp_nbr_address_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_peer_type_fld_xfmr", YangToDb_bgp_nbr_peer_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_peer_type_fld_xfmr", DbToYang_bgp_nbr_peer_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_af_nbr_tbl_key_xfmr", YangToDb_bgp_af_nbr_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_nbr_tbl_key_xfmr", DbToYang_bgp_af_nbr_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_afi_safi_name_fld_xfmr", YangToDb_bgp_nbr_afi_safi_name_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_afi_safi_name_fld_xfmr", DbToYang_bgp_nbr_afi_safi_name_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr", YangToDb_bgp_af_nbr_proto_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_af_nbr_proto_tbl_key_xfmr", DbToYang_bgp_af_nbr_proto_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_nbrs_nbr_state_xfmr", DbToYang_bgp_nbrs_nbr_state_xfmr)
    XlateFuncBind("DbToYang_bgp_nbrs_nbr_af_state_xfmr", DbToYang_bgp_nbrs_nbr_af_state_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_community_type_fld_xfmr", YangToDb_bgp_nbr_community_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_community_type_fld_xfmr", DbToYang_bgp_nbr_community_type_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_plist_direction_fld_xfmr", YangToDb_bgp_nbr_plist_direction_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_plist_direction_fld_xfmr", DbToYang_bgp_nbr_plist_direction_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_flist_direction_fld_xfmr", YangToDb_bgp_nbr_flist_direction_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_flist_direction_fld_xfmr", DbToYang_bgp_nbr_flist_direction_fld_xfmr)
    XlateFuncBind("YangToDb_bgp_nbr_orf_type_fld_xfmr", YangToDb_bgp_nbr_orf_type_fld_xfmr)
    XlateFuncBind("DbToYang_bgp_nbr_orf_type_fld_xfmr", DbToYang_bgp_nbr_orf_type_fld_xfmr)
}

var YangToDb_bgp_nbr_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_nbr_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, neighbor name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbrAddr   := pathInfo.Var("neighbor-address")

    if len(pathInfo.Vars) <  3 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
        log.Info("VRF Name is Missing")
        return vrfName, err
    }
    if strings.Contains(bgpId,"BGP") == false {
        err = errors.New("BGP ID is missing");
        log.Info("BGP ID is missing")
        return bgpId, err
    }
    if len(protoName) == 0 {
        err = errors.New("Protocol Name is missing");
        log.Info("Protocol Name is Missing")
        return protoName, err
    }
    if len(pNbrAddr) == 0 {
        err = errors.New("Neighbor address  is missing")
        log.Info("Neighbor address is Missing")
        return pNbrAddr, err
    }

    log.Info("URI VRF", vrfName)
    log.Info("URI Neighbor address", pNbrAddr)

    var pNbrKey string

    pNbrKey = vrfName + "|" + pNbrAddr

    log.Info("YangToDb_bgp_nbr_tbl_key_xfmr: pNbrKey:", pNbrKey)
    return pNbrKey, nil
}

var DbToYang_bgp_nbr_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_nbr_tbl_key: ", entry_key)

    nbrKey := strings.Split(entry_key, "|")
    nbrName:= nbrKey[1]

    rmap["neighbor-address"] = nbrName

    return rmap, nil
}

var YangToDb_bgp_nbr_peer_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    peer_type, _ := inParams.param.(ocbinds.E_OpenconfigBgp_PeerType)
    log.Info("YangToDb_bgp_nbr_peer_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " peer-type: ", peer_type)

    if (peer_type == ocbinds.OpenconfigBgp_PeerType_INTERNAL) {
        res_map["peer_type"] = "internal"
    }  else if (peer_type == ocbinds.OpenconfigBgp_PeerType_EXTERNAL) {
        res_map["peer_type"] = "external"
    } else {
        err = errors.New("Peer Type Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_nbr_peer_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_nbr_peer_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_NEIGHBOR"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_nbr_peer_type_fld_xfmr BGP neighbor not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    peer_type, ok := pGrpKey.Field["peer_type"]

    if ok {
        if (peer_type == "internal") {
            result["peer-type"] = "INTERNAL"
        } else if (peer_type == "external") {
            result["peer-type"] = "EXTERNAL"
        }
    } else {
        log.Info("peer_type field not found in DB")
    }
    return result, err
}


var YangToDb_bgp_nbr_address_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_nbr_address_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    entry_key := inParams.key
    nbrAddrKey := strings.Split(entry_key, "|")
    nbrAddr:= nbrAddrKey[1]

    result["neighbor-address"] = nbrAddr

    return result, err
}

var YangToDb_bgp_nbr_afi_safi_name_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    res_map["NULL"] = "NULL"
    return res_map, nil
}

var DbToYang_bgp_nbr_afi_safi_name_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    var nbrAfName string
    result := make(map[string]interface{})

    entry_key := inParams.key
    nbrAfKey := strings.Split(entry_key, "|")

    switch nbrAfKey[2] {
        case "ipv4_unicast":
            nbrAfName = "IPV4_UNICAST"
        case "ipv6_unicast":
            nbrAfName = "IPV6_UNICAST"
        case "l2vpn_evpn":
            nbrAfName = "L2VPN_EVPN"
       default:
            return result, nil
    }
    result["afi-safi-name"] = nbrAfName

    return result, err
}


var YangToDb_bgp_af_nbr_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_nbr_tbl_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <vrf name, protocol name, neighbor name> */

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbr   := pathInfo.Var("neighbor-address")
    afName     := pathInfo.Var("afi-safi-name")

    if len(pathInfo.Vars) <  4 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
        log.Info("VRF Name is Missing")
        return vrfName, err
    }
    if strings.Contains(bgpId,"BGP") == false {
        err = errors.New("BGP ID is missing");
        log.Info("BGP ID is missing")
        return bgpId, err
    }
    if len(protoName) == 0 {
        err = errors.New("Protocol Name is missing");
        log.Info("Protocol Name is Missing")
        return protoName, err
    }
    if len(pNbr) == 0 {
        err = errors.New("Neighbor is missing")
        log.Info("Neighbor is Missing")
        return pNbr, err
    }

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return afName, err
    }

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "ipv4_unicast"
    } else if strings.Contains(afName, "IPV6_UNICAST") {
        afName = "ipv6_unicast"
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "l2vpn_evpn"
    } else  {
	err = errors.New("Unsupported AFI SAFI")
	log.Info("Unsupported AFI SAFI ", afName);
	return afName, err
    }

    log.Info("URI VRF ", vrfName)
    log.Info("URI Nbr ", pNbr)
    log.Info("URI AFI SAFI ", afName)

    var nbrAfKey string

    nbrAfKey = vrfName + "|" + pNbr + "|" + afName

    log.Info("YangToDb_bgp_af_nbr_tbl_key_xfmr: afPgrpKey:", nbrAfKey)
    return nbrAfKey, nil
}

var DbToYang_bgp_af_nbr_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var afName string
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_nbr_tbl_key: ", entry_key)

    nbrAfKey := strings.Split(entry_key, "|")

    switch nbrAfKey[2] {
        case "ipv4_unicast":
            afName = "IPV4_UNICAST"
        case "ipv6_unicast":
            afName = "IPV6_UNICAST"
        case "l2vpn_evpn":
            afName = "L2VPN_EVPN"
       default:
            return rmap, nil
    }

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

var YangToDb_bgp_af_nbr_proto_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error
    var vrfName string

    log.Info("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    vrfName    =  pathInfo.Var("name")
    bgpId      := pathInfo.Var("identifier")
    protoName  := pathInfo.Var("name#2")
    pNbr   := pathInfo.Var("neighbor-address")
    afName     := pathInfo.Var("afi-safi-name")

    if len(pathInfo.Vars) <  4 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(vrfName) == 0 {
        err = errors.New("vrf name is missing");
        log.Info("VRF Name is Missing")
        return vrfName, err
    }
    if strings.Contains(bgpId,"BGP") == false {
        err = errors.New("BGP ID is missing");
        log.Info("BGP ID is missing")
        return bgpId, err
    }
    if len(protoName) == 0 {
        err = errors.New("Protocol Name is missing");
        log.Info("Protocol Name is Missing")
        return protoName, err
    }
    if len(pNbr) == 0 {
        err = errors.New("Neighbor missing")
        log.Info("Neighbo Missing")
        return pNbr, err
    }

    if len(afName) == 0 {
        err = errors.New("AFI SAFI is missing")
        log.Info("AFI SAFI is Missing")
        return afName, err
    }

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "ipv4_unicast"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") {
        afName = "ipv6_unicast"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "l2vpn_evpn"
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

    log.Info("URI VRF ", vrfName)
    log.Info("URI Nbr ", pNbr)
    log.Info("URI AFI SAFI ", afName)

    var nbrAfKey string

  nbrAfKey = vrfName + "|" + pNbr + "|" + afName

    log.Info("YangToDb_bgp_af_nbr_proto_tbl_key_xfmr: nbrAfKey:", nbrAfKey)
    return nbrAfKey, nil
}

var DbToYang_bgp_af_nbr_proto_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
   var afName string
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_af_nbr_proto_tbl_key_xfmr: ", entry_key)

    nbrAfKey := strings.Split(entry_key, "|")

    switch nbrAfKey[2] {
        case "ipv4_unicast":
            afName = "IPV4_UNICAST"
        case "ipv6_unicast":
            afName = "IPV6_UNICAST"
        case "l2vpn_evpn":
            afName = "L2VPN_EVPN"
       default:
            return rmap, nil
    }

    rmap["afi-safi-name"]   = afName

    return rmap, nil
}

type _xfmr_bgp_nbr_state_key struct {
    niName string
    nbrAddr string
}

func get_spec_nbr_cfg_tbl_entry (cfgDb *db.DB, nbr_key *_xfmr_bgp_nbr_state_key) (map[string]string, error) {
    var err error

    nbrCfgTblTs := &db.TableSpec{Name: "BGP_NEIGHBOR"}
    nbrEntryKey := db.Key{Comp: []string{nbr_key.niName, nbr_key.nbrAddr}}

    var entryValue db.Value
    if entryValue, err = cfgDb.GetEntry(nbrCfgTblTs, nbrEntryKey) ; err != nil {
        return nil, err
    }

    return entryValue.Field, err
}

func fill_nbr_state_cmn_info (nbr_key *_xfmr_bgp_nbr_state_key, frrNbrDataValue interface{}, cfgDb *db.DB,
                              nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    var err error
    nbrState := nbr_obj.State
    nbrState.NeighborAddress = &nbr_key.nbrAddr
    frrNbrDataJson := frrNbrDataValue.(map[string]interface{})

    if value, ok := frrNbrDataJson["bgpState"] ; ok {
        switch value {
            case "Idle":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_IDLE
            case "Connect":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_CONNECT
            case "Active":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_ACTIVE
            case "OpenSent":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_OPENSENT
            case "OpenConfirm":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_OPENCONFIRM
            case "Established":
                nbrState.SessionState = ocbinds.OpenconfigBgp_Bgp_Neighbors_Neighbor_State_SessionState_ESTABLISHED
        }
    }

    if value, ok := frrNbrDataJson["localAs"] ; ok {
        _localAs := uint32(value.(float64))
        nbrState.LocalAs = &_localAs
    }

    if value, ok := frrNbrDataJson["remoteAs"] ; ok {
        _peerAs := uint32(value.(float64))
        nbrState.PeerAs = &_peerAs
    }

    if value, ok := frrNbrDataJson["bgpTimerUpEstablishedEpoch"] ; ok {
        _lastEstablished := uint64(value.(float64))
        nbrState.LastEstablished = &_lastEstablished
    }

    if value, ok := frrNbrDataJson["connectionsEstablished"] ; ok {
        _establishedTransitions := uint64(value.(float64))
        nbrState.EstablishedTransitions = &_establishedTransitions
    }

    if statsMap, ok := frrNbrDataJson["messageStats"].(map[string]interface{}) ; ok {
        var _rcvd_msgs ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_State_Messages_Received
        var _sent_msgs ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_State_Messages_Sent
        var _msgs ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_State_Messages
        var _queues ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_State_Queues
        _msgs.Received = &_rcvd_msgs
        _msgs.Sent = &_sent_msgs
        nbrState.Messages = &_msgs
        nbrState.Queues = &_queues

        if value, ok := statsMap["updatesRecv"] ; ok {
            _updates_rcvd := uint64(value.(float64))
            _rcvd_msgs.UPDATE = &_updates_rcvd
        }
        if value, ok := statsMap["notificationsRecv"] ; ok {
            _notifs_rcvd := uint64(value.(float64))
            _rcvd_msgs.NOTIFICATION = &_notifs_rcvd
        }
        if value, ok := statsMap["updatesSent"] ; ok {
            _updates_sent := uint64(value.(float64))
            _sent_msgs.UPDATE = &_updates_sent
        }
        if value, ok := statsMap["notificationsSent"] ; ok {
            _notifs_sent := uint64(value.(float64))
            _sent_msgs.NOTIFICATION = &_notifs_sent
        }
        if value, ok := statsMap["depthOutq"] ; ok {
            _output := uint32(value.(float64))
            _queues.Output = &_output
        }
        if value, ok := statsMap["depthInq"] ; ok {
            _input := uint32(value.(float64))
            _queues.Input = &_input
        }
    }

    if capabMap, ok := frrNbrDataJson["neighborCapabilities"].(map[string]interface{}) ; ok {
        for capability,_ := range capabMap {
            switch capability {
                case "4byteAs":
                    nbrState.SupportedCapabilities = append(nbrState.SupportedCapabilities, ocbinds.OpenconfigBgpTypes_BGP_CAPABILITY_ASN32)
                case "addPath":
                    nbrState.SupportedCapabilities = append(nbrState.SupportedCapabilities, ocbinds.OpenconfigBgpTypes_BGP_CAPABILITY_ADD_PATHS)
                case "routeRefresh":
                    nbrState.SupportedCapabilities = append(nbrState.SupportedCapabilities, ocbinds.OpenconfigBgpTypes_BGP_CAPABILITY_ROUTE_REFRESH)
                case "multiprotocolExtensions":
                    nbrState.SupportedCapabilities = append(nbrState.SupportedCapabilities, ocbinds.OpenconfigBgpTypes_BGP_CAPABILITY_MPBGP)
                case "gracefulRestart":
                    nbrState.SupportedCapabilities = append(nbrState.SupportedCapabilities, ocbinds.OpenconfigBgpTypes_BGP_CAPABILITY_GRACEFUL_RESTART)
            }
        }
    }

    _dynamically_cfred := true

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, nbr_key) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["peer_group_name"] ; ok {
            nbrState.PeerGroup = &value
        }

        if value, ok := cfgDbEntry["admin_status"] ; ok {
            _enabled, _ := strconv.ParseBool(value)
            nbrState.Enabled = &_enabled
        }

        if value, ok := cfgDbEntry["name"] ; ok {
            nbrState.Description = &value
        }

        if value, ok := cfgDbEntry["auth_password"] ; ok {
            nbrState.AuthPassword = &value
        }

        if value, ok := cfgDbEntry["peer_type"] ; ok {
            switch value {
                case "internal":
                    nbrState.PeerType = ocbinds.OpenconfigBgp_PeerType_INTERNAL
                case "external":
                    nbrState.PeerType = ocbinds.OpenconfigBgp_PeerType_EXTERNAL
            }
        }

        if value, ok := cfgDbEntry["disable_ebgp_connected_route_check"] ; ok {
            _disableEbgpConnectedRouteCheck, _ := strconv.ParseBool(value)
            nbrState.DisableEbgpConnectedRouteCheck = &_disableEbgpConnectedRouteCheck
        }

        if value, ok := cfgDbEntry["enforce_first_as"] ; ok {
            _enforceFirstAs, _ := strconv.ParseBool(value)
            nbrState.EnforceFirstAs = &_enforceFirstAs
        }

        if value, ok := cfgDbEntry["solo_peer"] ; ok {
            _soloPeer, _ := strconv.ParseBool(value)
            nbrState.SoloPeer = &_soloPeer
        }

        if value, ok := cfgDbEntry["ttl_security_hops"] ; ok {
            if _ttlSecurityHops_u64, err := strconv.ParseUint(value, 10, 8) ; err == nil {
                _ttlSecurityHops_u8 := uint8(_ttlSecurityHops_u64)
                nbrState.TtlSecurityHops = &_ttlSecurityHops_u8
            }
        }

        if value, ok := cfgDbEntry["capability_ext_nexthop"] ; ok {
            _capabilityExtendedNexthop, _ := strconv.ParseBool(value)
            nbrState.CapabilityExtendedNexthop = &_capabilityExtendedNexthop
        }

        _dynamically_cfred = false
        nbrState.DynamicallyConfigured = &_dynamically_cfred
    } else {
        nbrState.DynamicallyConfigured = &_dynamically_cfred
    }

    return err
}

func fill_nbr_state_timers_info (nbr_key *_xfmr_bgp_nbr_state_key, frrNbrDataValue interface{}, cfgDb *db.DB,
                                 nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    var err error
    nbrTimersState := nbr_obj.Timers.State
    frrNbrDataJson := frrNbrDataValue.(map[string]interface{})

    if value, ok := frrNbrDataJson["bgpTimerHoldTimeMsecs"] ; ok {
        _neg_hold_time := (value.(float64))/1000
        nbrTimersState.NegotiatedHoldTime = &_neg_hold_time
    }

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, nbr_key) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["conn_retry"] ; ok {
            _connectRetry, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.ConnectRetry = &_connectRetry
        }
        if value, ok := cfgDbEntry["holdtime"] ; ok {
            _holdTime, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.HoldTime = &_holdTime
        }
        if value, ok := cfgDbEntry["keepalive"] ; ok {
            _keepaliveInterval, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.KeepaliveInterval = &_keepaliveInterval
        }
        if value, ok := cfgDbEntry["min_adv_interval"] ; ok {
            _minimumAdvertisementInterval, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.MinimumAdvertisementInterval = &_minimumAdvertisementInterval
        }
    }

    return err
}

func fill_nbr_state_transport_info (nbr_key *_xfmr_bgp_nbr_state_key, frrNbrDataValue interface{}, cfgDb *db.DB,
                                    nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    var err error

    nbrTransportState := nbr_obj.Transport.State
    frrNbrDataJson := frrNbrDataValue.(map[string]interface{})

    if value, ok := frrNbrDataJson["hostLocal"] ; ok {
        _localAddress := string(value.(string))
        nbrTransportState.LocalAddress = &_localAddress
    }
    if value, ok := frrNbrDataJson["portLocal"] ; ok {
        _localPort := uint16(value.(float64))
        nbrTransportState.LocalPort = &_localPort
    }
    if value, ok := frrNbrDataJson["hostForeign"] ; ok {
        _remoteAddress := string(value.(string))
        nbrTransportState.RemoteAddress = &_remoteAddress
    }
    if value, ok := frrNbrDataJson["portForeign"] ; ok {
        _remotePort := uint16(value.(float64))
        nbrTransportState.RemotePort = &_remotePort
    }

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, nbr_key) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["passive_mode"] ; ok {
            _passiveMode, _ := strconv.ParseBool(value)
            nbrTransportState.PassiveMode = &_passiveMode
        }
    }

    return err
}

func fill_nbr_state_info (get_req_uri_type E_bgp_nbr_state_get_req_uri_t, nbr_key *_xfmr_bgp_nbr_state_key, frrNbrDataValue interface{}, cfgDb *db.DB,
                          nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    switch get_req_uri_type {
        case E_bgp_nbr_state_get_req_uri_nbr_state:
            return fill_nbr_state_cmn_info (nbr_key, frrNbrDataValue, cfgDb, nbr_obj)
        case E_bgp_nbr_state_get_req_uri_nbr_timers_state:
            return fill_nbr_state_timers_info (nbr_key, frrNbrDataValue, cfgDb, nbr_obj)
        case E_bgp_nbr_state_get_req_uri_nbr_transport_state:
            return fill_nbr_state_transport_info (nbr_key, frrNbrDataValue, cfgDb, nbr_obj)
    }

    return errors.New("Opertational error")
}

func get_specific_nbr_state (get_req_uri_type E_bgp_nbr_state_get_req_uri_t,
                             nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor,
                             cfgDb *db.DB, nbr_key *_xfmr_bgp_nbr_state_key) error {
    var err error

    vtysh_cmd := "show ip bgp vrf " + nbr_key.niName + " neighbors " + nbr_key.nbrAddr + " json"
    nbrMapJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp neighbors state info for niName:%s nbrAddr:%s. Err: %s\n", nbr_key.niName, nbr_key.nbrAddr, err)
        return cmd_err
    }

    if frrNbrDataJson, ok := nbrMapJson[nbr_key.nbrAddr].(map[string]interface{}) ; ok {
        err = fill_nbr_state_info (get_req_uri_type, nbr_key, frrNbrDataJson, cfgDb, nbr_obj)
    }

    return err
}

func validate_nbr_state_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor, _xfmr_bgp_nbr_state_key, error) {
    var err error
    oper_err := errors.New("Opertational error")
    var nbr_key _xfmr_bgp_nbr_state_key
    var bgp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp

    bgp_obj, nbr_key.niName, err = getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", dbg_log , err);
        return nil, nbr_key, err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, _ := getYangPathFromUri(pathInfo.Path)
    nbr_key.nbrAddr = pathInfo.Var("neighbor-address")
    log.Infof("%s : path:%s; template:%s targetUriPath:%s niName:%s nbrAddr:%s",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, nbr_key.niName, nbr_key.nbrAddr)

    nbrs_obj := bgp_obj.Neighbors
    if nbrs_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors container missing", dbg_log)
        return nil, nbr_key, oper_err
    }

    nbr_obj, ok := nbrs_obj.Neighbor[nbr_key.nbrAddr]
    if !ok {
        log.Infof("%s Neighbor object missing, add new", dbg_log)
        nbr_obj,_ = nbrs_obj.NewNeighbor(nbr_key.nbrAddr)
    }
    ygot.BuildEmptyTree(nbr_obj)
    return nbr_obj, nbr_key, err
}

type E_bgp_nbr_state_get_req_uri_t string
const (
    E_bgp_nbr_state_get_req_uri_nbr_state E_bgp_nbr_state_get_req_uri_t = "GET_REQ_URI_BGP_NBR_STATE"
    E_bgp_nbr_state_get_req_uri_nbr_timers_state E_bgp_nbr_state_get_req_uri_t = "GET_REQ_URI_BGP_NBR_TIMERS_STATE"
    E_bgp_nbr_state_get_req_uri_nbr_transport_state E_bgp_nbr_state_get_req_uri_t = "GET_REQ_URI_BGP_NBR_TRANSPORT_STATE"
)

var DbToYang_bgp_nbrs_nbr_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    cmn_log := "GET: xfmr for BGP-nbrs state"
    get_req_uri_type := E_bgp_nbr_state_get_req_uri_nbr_state

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/timers/state":
            cmn_log = "GET: xfmr for BGP-nbrs timers state"
            get_req_uri_type = E_bgp_nbr_state_get_req_uri_nbr_timers_state
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/neighbors/neighbor/transport/state":
            cmn_log = "GET: xfmr for BGP-nbrs transport state"
            get_req_uri_type = E_bgp_nbr_state_get_req_uri_nbr_transport_state
    }

    nbr_obj, nbr_key, get_err := validate_nbr_state_get (inParams, cmn_log);
    if get_err != nil {
        log.Info("Neighbor state get subtree error: ", get_err)
        return get_err
    }

    err = get_specific_nbr_state (get_req_uri_type, nbr_obj, inParams.dbs[db.ConfigDB], &nbr_key)
    return err;
}

type _xfmr_bgp_nbr_af_state_key struct {
    niName string
    nbrAddr string
    afiSafiNameStr string
    afiSafiNameDbStr string
    afiSafiNameEnum ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE
}

func get_afi_safi_name_enum_dbstr_for_ocstr (afiSafiNameStr string) (ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, string, bool) {
    switch afiSafiNameStr {
        case "IPV4_UNICAST":
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, "ipv4_unicast", true
        case "IPV6_UNICAST":
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, "ipv6_unicast", true
        default:
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_UNSET, "", false
    }
}

func validate_nbr_af_state_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_AfiSafis_AfiSafi_State,
                                                                      _xfmr_bgp_nbr_af_state_key, error) {
    var err error
    var ok bool
    oper_err := errors.New("Opertational error")
    var nbr_af_key _xfmr_bgp_nbr_af_state_key
    var bgp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp

    bgp_obj, nbr_af_key.niName, err = getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", dbg_log , err);
        return nil, nbr_af_key, err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    nbr_af_key.nbrAddr = pathInfo.Var("neighbor-address")
    nbr_af_key.afiSafiNameStr = pathInfo.Var("afi-safi-name")
    nbr_af_key.afiSafiNameEnum, nbr_af_key.afiSafiNameDbStr, ok = get_afi_safi_name_enum_dbstr_for_ocstr (nbr_af_key.afiSafiNameStr)
    if !ok {
        log.Errorf("%s failed !! Error: AFI-SAFI ==> %s not supported", dbg_log, nbr_af_key.afiSafiNameStr)
        return nil, nbr_af_key, oper_err
    }

    log.Infof("%s : path:%s; template:%s targetUriPath:%s niName:%s nbrAddr:%s afiSafiNameStr:%s afiSafiNameEnum:%d afiSafiNameDbStr:%s",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, nbr_af_key.niName, nbr_af_key.nbrAddr, nbr_af_key.afiSafiNameStr, nbr_af_key.afiSafiNameEnum, nbr_af_key.afiSafiNameDbStr)

    nbrs_obj := bgp_obj.Neighbors
    if nbrs_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors container missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    nbr_obj, ok := nbrs_obj.Neighbor[nbr_af_key.nbrAddr]
    if !ok {
        log.Errorf("%s Neighbor object missing, add new", dbg_log)
        nbr_obj,_ = nbrs_obj.NewNeighbor(nbr_af_key.nbrAddr)
    }
    ygot.BuildEmptyTree(nbr_obj)

    afiSafis_obj := nbr_obj.AfiSafis
    if afiSafis_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors AfiSafis container missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }
    ygot.BuildEmptyTree(afiSafis_obj)

    afiSafi_obj, ok := afiSafis_obj.AfiSafi[nbr_af_key.afiSafiNameEnum]
    if !ok {
        log.Errorf("%s Neighbor AfiSafi object missing, allocate new", dbg_log)
        afiSafi_obj, _ = afiSafis_obj.NewAfiSafi(nbr_af_key.afiSafiNameEnum)
    }

    ygot.BuildEmptyTree(afiSafi_obj)

    afiSafiState_obj := afiSafi_obj.State
    if afiSafiState_obj == nil {
        log.Errorf("%s failed !! Error: Neighbor AfiSafi State object missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }
    ygot.BuildEmptyTree(afiSafiState_obj)

    return afiSafiState_obj, nbr_af_key, err
}

func get_spec_nbr_af_cfg_tbl_entry (cfgDb *db.DB, key *_xfmr_bgp_nbr_af_state_key) (map[string]string, error) {
    var err error

    nbrAfCfgTblTs := &db.TableSpec{Name: "BGP_NEIGHBOR_AF"}
    nbrAfEntryKey := db.Key{Comp: []string{key.niName, key.nbrAddr, key.afiSafiNameDbStr}}

    var entryValue db.Value
    if entryValue, err = cfgDb.GetEntry(nbrAfCfgTblTs, nbrAfEntryKey) ; err != nil {
        return nil, err
    }

    return entryValue.Field, err
}

var DbToYang_bgp_nbrs_nbr_af_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    cmn_log := "GET: xfmr for BGP-nbrs-nbr-af state"

    nbrs_af_state_obj, nbr_af_key, get_err := validate_nbr_af_state_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    var afiSafi_cmd string
    switch (nbr_af_key.afiSafiNameEnum) {
        case ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST:
            afiSafi_cmd = "ipv4 unicast"
        case ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST:
            afiSafi_cmd = "ipv6 unicast"
    }

    vtysh_cmd := "show ip bgp vrf " + nbr_af_key.niName + " " + afiSafi_cmd + " neighbors " + nbr_af_key.nbrAddr + " received-routes json"
    rcvdRoutesJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp neighbors received-routes state info for niName:%s nbrAddr:%s afi-safi-name:%s. Err: %s\n",
                   nbr_af_key.niName, nbr_af_key.nbrAddr, afiSafi_cmd, err)
        return cmd_err
    }

    vtysh_cmd = "show ip bgp vrf " + nbr_af_key.niName + " " + afiSafi_cmd + " neighbors " + nbr_af_key.nbrAddr + " advertised-routes json"
    advRoutesJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp neighbors advertised-routes state info for niName:%s nbrAddr:%s afi-safi-name:%s. Err: %s\n",
                   nbr_af_key.niName, nbr_af_key.nbrAddr, afiSafi_cmd, err)
        return cmd_err
    }

    nbrs_af_state_obj.AfiSafiName = nbr_af_key.afiSafiNameEnum

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_af_cfg_tbl_entry (inParams.dbs[db.ConfigDB], &nbr_af_key) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["admin_status"] ; ok {
            _enabled, _ := strconv.ParseBool(value)
            nbrs_af_state_obj.Enabled = &_enabled
        }

        if value, ok := cfgDbEntry["soft_reconfiguration_in"] ; ok {
            _softReconfigurationIn, _ := strconv.ParseBool(value)
            nbrs_af_state_obj.SoftReconfigurationIn = &_softReconfigurationIn
        }

        if value, ok := cfgDbEntry["unsuppress_map_name"] ; ok {
            nbrs_af_state_obj.UnsuppressMapName = &value
        }

        if value, ok := cfgDbEntry["weight"] ; ok {
            if _weight_u64, err := strconv.ParseUint(value, 10, 32) ; err == nil {
                _weight_u32 := uint32(_weight_u64)
                nbrs_af_state_obj.Weight = &_weight_u32
            }
        }

        if value, ok := cfgDbEntry["as_override"] ; ok {
            _asOverride, _ := strconv.ParseBool(value)
            nbrs_af_state_obj.AsOverride = &_asOverride
        }

        if value, ok := cfgDbEntry["send_community"] ; ok {
            switch value {
                case "standard":
                    nbrs_af_state_obj.SendCommunity = ocbinds.OpenconfigBgpExt_CommunityType_STANDARD
                case "extended":
                    nbrs_af_state_obj.SendCommunity = ocbinds.OpenconfigBgpExt_CommunityType_EXTENDED
                case "both":
                    nbrs_af_state_obj.SendCommunity = ocbinds.OpenconfigBgpExt_CommunityType_BOTH
                case "none":
                    nbrs_af_state_obj.SendCommunity = ocbinds.OpenconfigBgpExt_CommunityType_NONE
            }
        }

        if value, ok := cfgDbEntry["rrclient"] ; ok {
            _routeReflectorClient, _ := strconv.ParseBool(value)
            nbrs_af_state_obj.RouteReflectorClient = &_routeReflectorClient
        }
    }

    _active := false
    var _prefixes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor_AfiSafis_AfiSafi_State_Prefixes

    var _receivedPrePolicy, _rcvdFilteredByPolicy, _activeRcvdPrefixes uint32
    if value, ok := rcvdRoutesJson["totalPrefixCounter"] ; ok {
        _active = true
        _receivedPrePolicy = uint32(value.(float64))
        _prefixes.ReceivedPrePolicy = &_receivedPrePolicy
    }
    if value, ok := rcvdRoutesJson["filteredPrefixCounter"] ; ok {
        _rcvdFilteredByPolicy = uint32(value.(float64))
    }
    _activeRcvdPrefixes = _receivedPrePolicy - _rcvdFilteredByPolicy
    _prefixes.Received = &_activeRcvdPrefixes

    var _sentPrePolicy, _sentFilteredByPolicy, _activeSentPrefixes uint32
    if value, ok := advRoutesJson["totalPrefixCounter"] ; ok {
        _active = true
        _sentPrePolicy = uint32(value.(float64))
    }
    if value, ok := advRoutesJson["filteredPrefixCounter"] ; ok {
        _sentFilteredByPolicy = uint32(value.(float64))
    }
    _activeSentPrefixes = _sentPrePolicy - _sentFilteredByPolicy
    _prefixes.Sent = &_activeSentPrefixes

    nbrs_af_state_obj.Active = &_active
    nbrs_af_state_obj.Prefixes = &_prefixes

    return err;
}

var YangToDb_bgp_nbr_community_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    community_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_CommunityType)
    log.Info("YangToDb_bgp_nbr_community_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " community_type: ", community_type)

    if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_STANDARD) {
        res_map["send_community"] = "standard"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_EXTENDED) {
        res_map["send_community"] = "extended"
    }  else if (community_type == ocbinds.OpenconfigBgpExt_CommunityType_BOTH) {
        res_map["send_community"] = "both"
    } else {
        err = errors.New("send_community  Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_nbr_community_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_nbr_community_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_NEIGHBOR_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_nbr_community_type_fld_xfmr BGP Peer group not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    community_type, ok := pGrpKey.Field["send_community"]

    if ok {
        if (community_type == "standard") {
            result["send-community"] = "STANDARD"
        } else if (community_type == "extended") {
            result["send-community"] = "EXTENDED"
        } else if (community_type == "both") {
            result["send-community"] = "BOTH"
        }
    } else {
        log.Info("send_community not found in DB")
    }
    return result, err
}

var YangToDb_bgp_nbr_plist_direction_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    direction, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpDirection)
    log.Info("YangToDb_bgp_nbr_plist_direction_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " direction: ", direction)

    if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_INBOUND) {
        res_map["prefix_list_direction"] = "inbound"
    }  else if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_OUTBOUND) {
        res_map["prefix_list_direction"] = "outbound"
    } else {
        err = errors.New("direction Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_nbr_plist_direction_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_nbr_plist_direction_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_NEIGHBOR_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_nbr_plist_direction_fld_xfmr BGP neighbor not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    direction, ok := pGrpKey.Field["prefix_list_direction"]

    if ok {
        if (direction == "inbound") {
            result["direction"] = "INBOUND"
        } else if (direction == "outbound") {
            result["direction"] = "OUTBOUND"
        }
    } else {
        log.Info("prefix_list_direction field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_nbr_flist_direction_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    direction, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpDirection)
    log.Info("YangToDb_bgp_nbr_flist_direction_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " direction: ", direction)

    if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_INBOUND) {
        res_map["filter_list_direction"] = "inbound"
    }  else if (direction == ocbinds.OpenconfigBgpExt_BgpDirection_OUTBOUND) {
        res_map["filter_list_direction"] = "outbound"
    } else {
        err = errors.New("direction Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_nbr_flist_direction_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_nbr_flist_direction_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_NEIGHBOR_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_nbr_flist_direction_fld_xfmr BGP neighbor not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pGrpKey := pTbl[inParams.key]
    direction, ok := pGrpKey.Field["filter_list_direction"]

    if ok {
        if (direction == "inbound") {
            result["direction"] = "INBOUND"
        } else if (direction == "outbound") {
            result["direction"] = "OUTBOUND"
        }
    } else {
        log.Info("filter_list_direction field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_nbr_orf_type_fld_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    var err error
    if inParams.param == nil {
        err = errors.New("No Params");
        return res_map, err
    }
    orf_type, _ := inParams.param.(ocbinds.E_OpenconfigBgpExt_BgpOrfType)
    log.Info("YangToDb_bgp_nbr_orf_type_fld_xfmr: ", inParams.ygRoot, " Xpath: ", inParams.uri, " orf_type: ", orf_type)

    if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_SEND) {
        res_map["cap_orf"] = "send"
    }  else if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_RECEIVE) {
        res_map["cap_orf"] = "receive"
    }  else if (orf_type == ocbinds.OpenconfigBgpExt_BgpOrfType_BOTH) {
        res_map["cap_orf"] = "both"
    } else {
        err = errors.New("ORF type Missing");
        return res_map, err
    }

    return res_map, nil

}

var DbToYang_bgp_nbr_orf_type_fld_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {

    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_nbr_orf_type_fld_xfmr : ", data, "inParams : ", inParams)

    pTbl := data["BGP_NEIGHBOR_AF"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_nbr_orf_type_fld_xfmr BGP neighbor not found : ", inParams.key)
        return result, errors.New("BGP neighbor not found : " + inParams.key)
    }
    pNbrKey := pTbl[inParams.key]
    orf_type, ok := pNbrKey.Field["cap_orf"]

    if ok {
        if (orf_type == "send") {
            result["orf-type"] = "SEND"
        } else if (orf_type == "receive") {
            result["orf-type"] = "RECEIVE"
        } else if (orf_type == "both") {
            result["orf-type"] = "BOTH"
        }
    } else {
        log.Info("cap_orf_direction field not found in DB")
    }
    return result, err
}
