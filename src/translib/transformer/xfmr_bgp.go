package transformer

import (
    "errors"
    "encoding/json"
    "translib/ocbinds"
    "translib/db"
    "strconv"
    "os/exec"
    log "github.com/golang/glog"
    ygot "github.com/openconfig/ygot/ygot"
)

func getBgpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp, string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
    var err error

    if len(niName) == 0 {
        return nil, "", errors.New("Network-instance-name missing")
    }

    if bgpId != "BGP" {
        return nil, "", errors.New("Protocol-id is not BGP!! Incoming Protocol-id:" + bgpId)
    }

    if len(protoName) == 0 {
        return nil, "", errors.New("Network-instance Protocol-name missing")
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

func init () {
    XlateFuncBind("YangToDb_bgp_gbl_tbl_key_xfmr", YangToDb_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_tbl_key_xfmr", DbToYang_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_always_compare_med_enable_xfmr", YangToDb_bgp_always_compare_med_enable_xfmr)
    XlateFuncBind("DbToYang_bgp_always_compare_med_enable_xfmr", DbToYang_bgp_always_compare_med_enable_xfmr)
    XlateFuncBind("YangToDb_bgp_allow_multiple_as_xfmr", YangToDb_bgp_allow_multiple_as_xfmr)
    XlateFuncBind("DbToYang_bgp_allow_multiple_as_xfmr", DbToYang_bgp_allow_multiple_as_xfmr)
    XlateFuncBind("YangToDb_bgp_graceful_restart_status_xfmr", YangToDb_bgp_graceful_restart_status_xfmr)
    XlateFuncBind("DbToYang_bgp_graceful_restart_status_xfmr", DbToYang_bgp_graceful_restart_status_xfmr)
    XlateFuncBind("DbToYang_bgp_nbrs_nbr_state_xfmr", DbToYang_bgp_nbrs_nbr_state_xfmr)
    XlateFuncBind("DbToYang_bgp_nbrs_nbr_af_state_xfmr", DbToYang_bgp_nbrs_nbr_af_state_xfmr)
    XlateFuncBind("YangToDb_bgp_ignore_as_path_length_xfmr", YangToDb_bgp_ignore_as_path_length_xfmr)
    XlateFuncBind("DbToYang_bgp_ignore_as_path_length_xfmr", DbToYang_bgp_ignore_as_path_length_xfmr)
    XlateFuncBind("YangToDb_bgp_external_compare_router_id_xfmr", YangToDb_bgp_external_compare_router_id_xfmr)
    XlateFuncBind("DbToYang_bgp_external_compare_router_id_xfmr", DbToYang_bgp_external_compare_router_id_xfmr)
    XlateFuncBind("DbToYang_bgp_routes_get_xfmr", DbToYang_bgp_routes_get_xfmr)
}

var YangToDb_bgp_gbl_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    protoName := pathInfo.Var("name#2")

    if protoName != "bgp" {
        return niName, errors.New("Invalid protocol name : " + protoName)
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

var YangToDb_bgp_always_compare_med_enable_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_always_compare_med_enable_xfmr: ")
    if inParams.param == nil {
	    res_map["always_compare_med"] = ""
	    return res_map, nil
	}

    enabled, _ := inParams.param.(*bool)
    var enStr string
    if *enabled == true {
        enStr = "true"
    } else {
        enStr = "false"
    }
    res_map["always_compare_med"] = enStr

    return res_map, nil
}

var DbToYang_bgp_always_compare_med_enable_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_always_compare_med_enable_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_always_compare_med_enable_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    always_compare_med_enable, ok := niInst.Field["always_compare_med"]
    if ok {
        if always_compare_med_enable == "true" {
            result["always-compare-med"] = true
        } else {
            result["always-compare-med"] = false
        }
    } else {
        log.Info("always_compare_med field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_allow_multiple_as_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_allow_multiple_as_xfmr: ")
    if inParams.param == nil {
	    res_map["load_balance_mp_relax"] = ""
	    return res_map, nil
	}

    allow_multiple_as, _ := inParams.param.(*bool)
    var allowMultipleAsStr string
    if *allow_multiple_as == true {
        allowMultipleAsStr = "true"
    } else {
        allowMultipleAsStr = "false"
    }
    res_map["load_balance_mp_relax"] = allowMultipleAsStr

    return res_map, nil
}

var DbToYang_bgp_allow_multiple_as_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_allow_multiple_as_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_allow_multiple_as_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    load_balance_mp_relax_val, ok := niInst.Field["load_balance_mp_relax"]
    if ok {
        if load_balance_mp_relax_val == "true" {
            result["load_balance_mp_relax"] = true
        } else {
            result["load_balance_mp_relax"] = false
        }
    } else {
        log.Info("load_balance_mp_relax field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_graceful_restart_status_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_graceful_restart_status_xfmr: ")
    if inParams.param == nil {
	    res_map["graceful_restart_enable"] = ""
	    return res_map, nil
	}

    gr_status, _ := inParams.param.(*bool)
    var gr_statusStr string
    if *gr_status == true {
        gr_statusStr = "true"
    } else {
        gr_statusStr = "false"
    }
    res_map["graceful_restart_enable"] = gr_statusStr

    return res_map, nil
}

var DbToYang_bgp_graceful_restart_status_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_graceful_restart_status_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_graceful_restart_status_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    gr_enable_val, ok := niInst.Field["graceful_restart_enable"]
    if ok {
        if gr_enable_val == "true" {
            result["graceful_restart_enable"] = true
        } else {
            result["graceful_restart_enable"] = false
        }
    } else {
        log.Info("graceful_restart_enable field not found in DB")
    }
    return result, err
}

func exec_vtysh_cmd (vtysh_cmd string) (map[string]interface{}, error) {
    var err error
    oper_err := errors.New("Opertational error")

    log.Infof("Going to execute vtysh cmd ==> \"%s\"", vtysh_cmd)

    cmd := exec.Command("/usr/bin/docker", "exec", "bgp", "vtysh", "-c", vtysh_cmd)
    out_stream, err := cmd.StdoutPipe()
    if err != nil {
        log.Errorf("Can't get stdout pipe: %s\n", err)
        return nil, oper_err
    }

    err = cmd.Start()
    if err != nil {
        log.Errorf("cmd.Start() failed with %s\n", err)
        return nil, oper_err
    }

    var outputJson map[string]interface{}
    err = json.NewDecoder(out_stream).Decode(&outputJson)
    if err != nil {
        log.Errorf("Not able to decode vtysh json output: %s\n", err)
        return nil, oper_err
    }

    err = cmd.Wait()
    if err != nil {
        log.Errorf("Command execution completion failed with %s\n", err)
        return nil, oper_err
    }

    log.Infof("Successfully executed vtysh-cmd ==> \"%s\"", vtysh_cmd)

    if outputJson == nil {
        log.Errorf("VTYSH output empty !!!")
        return nil, oper_err
    }

    return outputJson, err
}

func get_spec_nbr_cfg_tbl_entry (cfgDb *db.DB, niName string, nbrAddr string) (map[string]string, error) {
    var err error

    nbrCfgTblTs := &db.TableSpec{Name: "BGP_NEIGHBOR"}
    nbrEntryKey := db.Key{Comp: []string{niName,nbrAddr}}

    var entryValue db.Value
    if entryValue, err = cfgDb.GetEntry(nbrCfgTblTs, nbrEntryKey) ; err != nil {
        return nil, err
    }

    return entryValue.Field, err
}

func fill_nbr_state_cmn_info (niName string, nbrAddr string, frrNbrDataValue interface{}, cfgDb *db.DB,
                              nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    var err error
    nbrState := nbr_obj.State
    nbrState.NeighborAddress = &nbrAddr
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
            default:
                log.Infof("bgp_session_state for nbrAddr:%s ==> %s", nbrAddr, value)
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

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, niName, nbrAddr) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["peer_group_name"] ; ok {
            nbrState.PeerGroup = &value
        }

        if value, ok := cfgDbEntry["enabled"] ; ok {
            _enabled, _ := strconv.ParseBool(value)
            nbrState.Enabled = &_enabled
        }

        if value, ok := cfgDbEntry["description"] ; ok {
            nbrState.Description = &value
        }

        if value, ok := cfgDbEntry["auth_password"] ; ok {
            nbrState.AuthPassword = &value
        }
        
        _dynamically_cfred = false
        nbrState.DynamicallyConfigured = &_dynamically_cfred
    } else {
        nbrState.DynamicallyConfigured = &_dynamically_cfred
    }

    return err
}

func fill_nbr_state_timers_info (niName string, nbrAddr string, frrNbrDataValue interface{}, cfgDb *db.DB,
                                 nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    var err error
    nbrTimersState := nbr_obj.Timers.State
    frrNbrDataJson := frrNbrDataValue.(map[string]interface{})

    if value, ok := frrNbrDataJson["bgpTimerHoldTimeMsecs"] ; ok {
        _neg_hold_time := (value.(float64))/1000
        nbrTimersState.NegotiatedHoldTime = &_neg_hold_time
    }

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, niName, nbrAddr) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["conn_retry"] ; ok {
            _connectRetry, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.ConnectRetry = &_connectRetry
        }
        if value, ok := cfgDbEntry["hold_time"] ; ok {
            _holdTime, _ := strconv.ParseFloat(value, 64)
            nbrTimersState.HoldTime = &_holdTime
        }
        if value, ok := cfgDbEntry["keepalive_interval"] ; ok {
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

func fill_nbr_state_transport_info (niName string, nbrAddr string, frrNbrDataValue interface{}, cfgDb *db.DB,
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

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_cfg_tbl_entry (cfgDb, niName, nbrAddr) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["passive_mode"] ; ok {
            _passiveMode, _ := strconv.ParseBool(value)
            nbrTransportState.PassiveMode = &_passiveMode
        }
    }

    return err
}

func fill_nbr_state_info (get_req_uri_type E_bgp_nbr_state_get_req_uri_t, niName string, nbrAddr string, frrNbrDataValue interface{}, cfgDb *db.DB,
                          nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor) error {
    switch get_req_uri_type {
        case E_bgp_nbr_state_get_req_uri_nbr_state:
            return fill_nbr_state_cmn_info (niName, nbrAddr, frrNbrDataValue, cfgDb, nbr_obj)
        case E_bgp_nbr_state_get_req_uri_nbr_timers_state:
            return fill_nbr_state_timers_info (niName, nbrAddr, frrNbrDataValue, cfgDb, nbr_obj)
        case E_bgp_nbr_state_get_req_uri_nbr_transport_state:
            return fill_nbr_state_transport_info (niName, nbrAddr, frrNbrDataValue, cfgDb, nbr_obj)
    }

    return errors.New("Opertational error")
}

func get_specific_nbr_state (get_req_uri_type E_bgp_nbr_state_get_req_uri_t,
                             nbrs_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors,
                             cfgDb *db.DB, niName string, nbrAddr string) error {
    var err error

    vtysh_cmd := "show ip bgp vrf " + niName + " neighbors " + nbrAddr + " json"
    nbrMapJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bgp neighbors state info for niName:%s nbrAddr:%s. Err: %s\n", niName, nbrAddr, err)
        return cmd_err
    }

    var ok bool
    var nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors_Neighbor
    if len(nbrs_obj.Neighbor) == 0 {
        nbr_obj, _ = nbrs_obj.NewNeighbor (nbrAddr)
    } else {
        if nbr_obj, ok = nbrs_obj.Neighbor[nbrAddr] ; !ok {
            nbr_obj, _ = nbrs_obj.NewNeighbor (nbrAddr)
        }
    }
    ygot.BuildEmptyTree(nbr_obj)

    if frrNbrDataJson, ok := nbrMapJson[nbrAddr].(map[string]interface{}) ; ok {
        err = fill_nbr_state_info (get_req_uri_type, niName, nbrAddr, frrNbrDataJson, cfgDb, nbr_obj)
    }

    return err
}

func get_all_nbr_state (get_req_uri_type E_bgp_nbr_state_get_req_uri_t,
                        nbrs_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors,
                        cfgDb *db.DB, niName string) error {
    var err error

    vtysh_cmd := "show ip bgp vrf " + niName + " neighbors " + " json"
    nbrsMapJson, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch all bgp neighbors state info for niName:%s. Err: %s\n", niName, err)
        return cmd_err
    }

    for nbrAddr, frrNbrDataJson := range nbrsMapJson {
        nbr_obj, _ := nbrs_obj.NewNeighbor (nbrAddr)
        ygot.BuildEmptyTree(nbr_obj)
        err = fill_nbr_state_info (get_req_uri_type, niName, nbrAddr, frrNbrDataJson, cfgDb, nbr_obj)
    }

    return err
}

var YangToDb_bgp_ignore_as_path_length_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_ignore_as_path_length_xfmr: ")
    if inParams.param == nil {
	    res_map["ignore_as_path_length"] = ""
	    return res_map, nil
	}

    ignore_as_path_length, _ := inParams.param.(*bool)
    var ignoreAsPathLen string
    if *ignore_as_path_length == true {
        ignoreAsPathLen = "true"
    } else {
        ignoreAsPathLen = "false"
    }
    res_map["ignore_as_path_length"] = ignoreAsPathLen

    return res_map, nil
}

var DbToYang_bgp_ignore_as_path_length_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_ignore_as_path_length_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_ignore_as_path_length_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    ignore_as_path_length_val, ok := niInst.Field["ignore_as_path_length"]
    if ok {
        if ignore_as_path_length_val == "true" {
            result["ignore_as_path_length"] = true
        } else {
            result["ignore_as_path_length"] = false
        }
    } else {
        log.Info("ignore_as_path_length field not found in DB")
    }
    return result, err
}

var YangToDb_bgp_external_compare_router_id_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    res_map := make(map[string]string)

    log.Info("YangToDb_bgp_external_compare_router_id_xfmr: ")
    if inParams.param == nil {
	    res_map["external_compare_router_id"] = ""
	    return res_map, nil
	}

    external_compare_router_id, _ := inParams.param.(*bool)
    var externalCompareRouterIdStr string
    if *external_compare_router_id == true {
        externalCompareRouterIdStr = "true"
    } else {
        externalCompareRouterIdStr = "false"
    }
    res_map["external_compare_router_id"] = externalCompareRouterIdStr

    return res_map, nil
}

var DbToYang_bgp_external_compare_router_id_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    var err error
    result := make(map[string]interface{})

    data := (*inParams.dbDataMap)[inParams.curDb]
    log.Info("DbToYang_bgp_external_compare_router_id_xfmr", data, "inParams : ", inParams)

    pTbl := data["BGP_GLOBALS"]
    if _, ok := pTbl[inParams.key]; !ok {
        log.Info("DbToYang_bgp_external_compare_router_id_xfmr BGP globals not found : ", inParams.key)
        return result, errors.New("BGP globals not found : " + inParams.key)
    }
    niInst := pTbl[inParams.key]
    external_compare_router_id_val, ok := niInst.Field["external_compare_router_id"]
    if ok {
        if external_compare_router_id_val == "true" {
            result["external_compare_router_id"] = true
        } else {
            result["external_compare_router_id"] = false
        }
    } else {
        log.Info("external_compare_router_id field not found in DB")
    }
    return result, err
}

func validate_nbr_state_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Neighbors, string, string, error) {
    var err error
    oper_err := errors.New("Opertational error")

    bgp_obj, niName, err := getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", dbg_log , err);
        return nil, "", "", err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    nbrAddr := pathInfo.Var("neighbor-address")
    log.Infof("%s : path:%s; template:%s targetUriPath:%s niName:%s nbrAddr:%s",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, niName, nbrAddr)

    nbrs_obj := bgp_obj.Neighbors
    if nbrs_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors container missing", dbg_log)
        return nil, "", "", oper_err
    }

    return nbrs_obj, niName, nbrAddr, err
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

    nbrs_obj, niName, nbrAddr, get_err := validate_nbr_state_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    if len(nbrAddr) != 0 {
        err = get_specific_nbr_state (get_req_uri_type, nbrs_obj, inParams.dbs[db.ConfigDB], niName, nbrAddr);
    } else {
        err = get_all_nbr_state (get_req_uri_type, nbrs_obj, inParams.dbs[db.ConfigDB], niName);
    }

    return err;
}

type _xfmr_bgp_nbr_af_state_key struct {
    niName string
    nbrAddr string
    afiSafiNameStr string
    afiSafiNameEnum ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE
}

func get_afi_safi_name_enum_for_str (afiSafiNameStr string) (ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, bool) {
    switch afiSafiNameStr {
        case "IPV4_UNICAST":
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, true
        case "IPV6_UNICAST":
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, true
        default:
            return ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_UNSET, false
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
    nbr_af_key.afiSafiNameEnum, ok = get_afi_safi_name_enum_for_str (nbr_af_key.afiSafiNameStr)
    if !ok {
        log.Errorf("%s failed !! Error: AFI-SAFI ==> %s not supported", dbg_log, nbr_af_key.afiSafiNameStr)
        return nil, nbr_af_key, oper_err
    }

    log.Infof("%s : path:%s; template:%s targetUriPath:%s niName:%s nbrAddr:%s afiSafiNameStr:%s afiSafiNameEnum:%d",
              dbg_log, pathInfo.Path, pathInfo.Template, targetUriPath, nbr_af_key.niName, nbr_af_key.nbrAddr, nbr_af_key.afiSafiNameStr, nbr_af_key.afiSafiNameEnum)

    nbrs_obj := bgp_obj.Neighbors
    if nbrs_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors container missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    nbr_obj, ok := nbrs_obj.Neighbor[nbr_af_key.nbrAddr]
    if !ok {
        log.Errorf("%s failed !! Error: Neighbor object missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    afiSafis_obj := nbr_obj.AfiSafis
    if afiSafis_obj == nil {
        log.Errorf("%s failed !! Error: Neighbors AfiSafis container missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    afiSafi_obj, ok := afiSafis_obj.AfiSafi[nbr_af_key.afiSafiNameEnum]
    if !ok {
        log.Errorf("%s failed !! Error: Neighbor AfiSafi object missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    afiSafiState_obj := afiSafi_obj.State
    if afiSafiState_obj == nil {
        log.Errorf("%s failed !! Error: Neighbor AfiSafi State object missing", dbg_log)
        return nil, nbr_af_key, oper_err
    }

    return afiSafiState_obj, nbr_af_key, err
}

func get_spec_nbr_af_cfg_tbl_entry (cfgDb *db.DB, key _xfmr_bgp_nbr_af_state_key) (map[string]string, error) {
    var err error

    nbrAfCfgTblTs := &db.TableSpec{Name: "BGP_NEIGHBOR_AF"}
    nbrAfEntryKey := db.Key{Comp: []string{key.niName, key.nbrAddr, key.afiSafiNameStr}}

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

    if cfgDbEntry, cfgdb_get_err := get_spec_nbr_af_cfg_tbl_entry (inParams.dbs[db.ConfigDB], nbr_af_key) ; cfgdb_get_err == nil {
        if value, ok := cfgDbEntry["enabled"] ; ok {
            _enabled, _ := strconv.ParseBool(value)
            nbrs_af_state_obj.Enabled = &_enabled
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

var DbToYang_bgp_routes_get_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    /*

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib/routes":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib/routes/route":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes/route":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes/route":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-pre":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-pre/routes":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-pre/routes/route":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post/routes":
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post/routes/route":
    }

    /* Address Family
     * 1) IPv4
     *       - Local RIB - get all & specific prefix get
     *       - Neighbors - get all for adj-rib-in-pre, adj-rib-in-post, adj-rib-out-pre & adj-rib-out-post
     *                     and specific route get in above containers.
     * 2) IPv6
     *       - Local RIB - get all & specific prefix get
     *       - Neighbors - get all for adj-rib-in-pre, adj-rib-in-post, adj-rib-out-pre & adj-rib-out-post
     *                     and specific route get in above containers.
     **/
    return err;
}


