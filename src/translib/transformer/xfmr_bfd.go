package transformer

import (
    "errors"
    "translib/ocbinds"
    "strings"
    "encoding/json"
    "strconv"
    "os/exec"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

func init () {
    XlateFuncBind("DbToYang_bfd_state_xfmr", DbToYang_bfd_state_xfmr)
}


func validate_bfd_get (inParams XfmrParams, dbg_log string) (*ocbinds.OpenconfigBfd_Bfd_BfdState, error) {
    var err error
    var bfd_obj *ocbinds.OpenconfigBfd_Bfd

    deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    bfd_obj = deviceObj.Bfd

    if bfd_obj.BfdState == nil {
        return nil, errors.New("BFD State container missing")
    }

    return bfd_obj.BfdState, err
}

func exec_vtysh_cmd_array (vtysh_cmd string) ([]interface{}, error) {
    var err error
    oper_err := errors.New("Operational error")

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

    var outputJson []interface{}
    err = json.NewDecoder(out_stream).Decode(&outputJson)
    if err != nil {
        log.Errorf("Not able to decode vtysh json output as array of objects: %s\n", err)
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

func get_bfd_shop_peers (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, inParams XfmrParams) (error) {
    var bfdshop_key ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key
    var vtysh_cmd string
    var err error

    bfdMapJson := make(map[string]interface{})
    bfdCounterMapJson := make(map[string]interface{})

    pathInfo := NewPathInfo(inParams.uri)

    bfdshop_key.RemoteAddress = pathInfo.Var("remote-address")
    bfdshop_key.Vrf = pathInfo.Var("vrf")
    bfdshop_key.Interface = pathInfo.Var("interface")
    bfdshop_key.LocalAddress = pathInfo.Var("local-address")

    if (bfdshop_key.LocalAddress == "null") {
        vtysh_cmd = "show bfd vrf " + bfdshop_key.Vrf + " peer " + bfdshop_key.RemoteAddress + " interface " + bfdshop_key.Interface + " json"
    } else {
        vtysh_cmd = "show bfd vrf " + bfdshop_key.Vrf + " peer " + bfdshop_key.RemoteAddress + " interface " + bfdshop_key.Interface + " local-address " + bfdshop_key.LocalAddress + " json"
    }

    output_peer, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch shop bfd peers:, err")
        return cmd_err;
    }

    if (bfdshop_key.LocalAddress == "null") {
        vtysh_cmd = "show bfd vrf " + bfdshop_key.Vrf + " peer " + bfdshop_key.RemoteAddress + " interface " + bfdshop_key.Interface + " counters" + " json"
    } else {
        vtysh_cmd = "show bfd vrf " + bfdshop_key.Vrf + " peer " + bfdshop_key.RemoteAddress + " interface " + bfdshop_key.Interface + " local-address " + bfdshop_key.LocalAddress + " counters" + " json"
    }

    output_counter, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch shop bfd peers counters array:, err")
        return cmd_err;
    }

    log.Info(output_peer)
    bfdMapJson["output"] = output_peer

    log.Info(output_counter)
    bfdCounterMapJson["output"] = output_counter

    if sessions, ok := bfdMapJson["output"].(map[string]interface{}) ; ok {
        log.Info(sessions)
        if counters, ok := bfdCounterMapJson["output"].(map[string]interface{}) ; ok {
            log.Info(counters)
            fill_bfd_shop_data (bfd_obj, sessions, counters, &bfdshop_key)
        }
    }

    return err;
}

func get_bfd_mhop_peers (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, inParams XfmrParams) (error) {
    var bfdmhop_key ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key
    var err error
    var vtysh_cmd string

    bfdMapJson := make(map[string]interface{})
    bfdCounterMapJson := make(map[string]interface{})

    pathInfo := NewPathInfo(inParams.uri)

    bfdmhop_key.RemoteAddress = pathInfo.Var("remote-address")
    bfdmhop_key.Interface = pathInfo.Var("interface")
    bfdmhop_key.Vrf = pathInfo.Var("vrf")
    bfdmhop_key.LocalAddress = pathInfo.Var("local-address")

    if (bfdmhop_key.LocalAddress == "null") {
        vtysh_cmd = "show bfd vrf " + bfdmhop_key.Vrf + " peer " + bfdmhop_key.RemoteAddress + " multihop " + " local-address " + bfdmhop_key.LocalAddress + " json"
    } else {
        vtysh_cmd = "show bfd vrf " + bfdmhop_key.Vrf + " peer " + bfdmhop_key.RemoteAddress + " multihop " + " local-address " + bfdmhop_key.LocalAddress + " interface " + bfdmhop_key.Interface + " json"
    }

    output_peer, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch shop bfd peers array:, err")
        return cmd_err;
    }

    if (bfdmhop_key.LocalAddress == "null") {
        vtysh_cmd = "show bfd vrf " + bfdmhop_key.Vrf + " peer " + bfdmhop_key.RemoteAddress + " multihop " + " local-address " + bfdmhop_key.LocalAddress + " counters" + " json"
    } else {
        vtysh_cmd = "show bfd vrf " + bfdmhop_key.Vrf + " peer " + bfdmhop_key.RemoteAddress + " multihop " + " local-address " + bfdmhop_key.LocalAddress + " interface " + bfdmhop_key.Interface + " counters" + " json"
    }

    output_counter, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch shop bfd peers counters array:, err")
        return cmd_err;
    }

    log.Info(output_peer)
    bfdMapJson["output"] = output_peer

    log.Info(output_counter)
    bfdCounterMapJson["output"] = output_counter

    if sessions, ok := bfdMapJson["output"].(map[string]interface{}) ; ok {
        log.Info(sessions)
        if counters, ok := bfdCounterMapJson["output"].(map[string]interface{}) ; ok {
            log.Info(counters)
            fill_bfd_mhop_data (bfd_obj, sessions, counters, &bfdmhop_key)
        }
    }

    return err;
}

func get_bfd_peers (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, inParams XfmrParams) error {
    var err error
    var cmd_err error
    var output_peer []interface{}
    var output_counter []interface{}

    bfdMapJson := make(map[string]interface{})
    bfdCounterMapJson := make(map[string]interface{})

    pathInfo := NewPathInfo(inParams.uri)

    targetUriPath, err := getYangPathFromUri(pathInfo.Path)
    log.Info(targetUriPath)
    if strings.HasPrefix(targetUriPath, "/openconfig-bfd:bfd/openconfig-bfd-ext:bfd-state/single-hop-state") {
        cmd_err = get_bfd_shop_peers (bfd_obj, inParams)
        return cmd_err
    } else if strings.HasPrefix(targetUriPath, "/openconfig-bfd:bfd/openconfig-bfd-ext:bfd-state/multi-hop-state") {
        cmd_err = get_bfd_mhop_peers (bfd_obj, inParams)
        return cmd_err
    } else {
        vtysh_cmd := "show bfd peers json"
        output_peer, cmd_err = exec_vtysh_cmd_array (vtysh_cmd)
        if cmd_err != nil {
            log.Errorf("Failed to fetch bfd peers array:, err")
            return cmd_err
        }

        vtysh_cmd = "show bfd peers counters json"
        output_counter, cmd_err = exec_vtysh_cmd_array (vtysh_cmd)
        if cmd_err != nil {
            log.Errorf("Failed to fetch bfd peers counters array:, err")
            return cmd_err
        }
    }

    log.Info(output_peer)
    bfdMapJson["output"] = output_peer

    log.Info(output_counter)
    bfdCounterMapJson["output"] = output_counter

    sessions, _ := bfdMapJson["output"].([]interface{})
    counters, _ := bfdCounterMapJson["output"].([]interface{})

    for i, session := range sessions {
        session_data, _ := session.(map[string]interface{})
        counter_data, _ := counters[i].(map[string]interface{})
        log.Info(session_data)
        log.Info(counter_data)
        if value, ok := session_data["multihop"].(bool) ; ok {
            if value == false {
                if ok := fill_bfd_shop_data (bfd_obj, session_data, counter_data, nil) ; !ok {return err}
            }else {
                if ok := fill_bfd_mhop_data (bfd_obj, session_data, counter_data, nil) ; !ok {return err}
            }
        }
    }

    log.Info(bfd_obj)

    return err
}

var DbToYang_bfd_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

    var err error
    cmn_log := "GET: xfmr for BFD peers state"

    bfd_obj, get_err := validate_bfd_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    err = get_bfd_peers (bfd_obj, inParams)

    return err;
}

func fill_bfd_shop_data (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, session_data map[string]interface{}, counter_data map[string]interface{}, bfdshop_Input_key *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key) bool {
    var err error
    var bfdshop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState
    var bfdshopkey ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key
    var bfdshop_tempkey ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key
    var bfdasyncstats *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Async
    var bfdechocstats *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Echo

    log.Info("fill_bfd_shop_data")

    if (nil != bfdshop_Input_key) {
        bfdshop_tempkey = *bfdshop_Input_key
	log.Info("fill_bfd_shop_data1")
        bfdshop_obj = bfd_obj.SingleHopState[bfdshop_tempkey]
        if (nil == bfdshop_obj) {
            log.Info("Peer with input key not found")
            return false; 
        }
    } else {
        if value, ok := session_data["peer"].(string) ; ok {
            bfdshopkey.RemoteAddress = value
        }

        if value, ok := session_data["interface"].(string) ; ok {
           bfdshopkey.Interface = value
        }

        if value, ok := session_data["vrf"].(string) ; ok {
            bfdshopkey.Vrf = value
        }

        if value, ok := session_data["local"].(string) ; ok {
            bfdshopkey.LocalAddress = value
        } else {
            bfdshopkey.LocalAddress = "null"
        }

        bfdshop_obj, err = bfd_obj.NewSingleHopState(bfdshopkey.RemoteAddress, bfdshopkey.Interface, bfdshopkey.Vrf, bfdshopkey.LocalAddress)
        if (err != nil) {
            return false;
        }
    }
 
    ygot.BuildEmptyTree(bfdshop_obj)

    if value, ok := session_data["status"].(string) ; ok {
        if value == "down" {
            bfdshop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_DOWN
        } else if value == "up" {
            bfdshop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_UP
        } else if value == "shutdown" {
            bfdshop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_ADMIN_DOWN
        } else if value == "init" {
            bfdshop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_INIT
        } else {
            bfdshop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_UNSET
        }

    }

    /*if value, ok := session_data["remote-status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdshop_obj.RemoteSessionState = value
    }*/

    if bfdshop_obj.SessionState != ocbinds.OpenconfigBfd_BfdSessionState_UP { 
        if value, ok := session_data["downtime"].(float64) ; ok {
            value64 := uint64(value)
            bfdshop_obj.LastFailureTime = &value64
        }   
    }

    if value, ok := session_data["id"].(float64) ; ok {
        s := strconv.FormatFloat(value, 'f', -1, 64)
        bfdshop_obj.LocalDiscriminator = &s
    }

    if value, ok := session_data["remote-id"].(float64) ; ok {
        s := strconv.FormatFloat(value, 'f', -1, 64)
        bfdshop_obj.RemoteDiscriminator = &s
    }

    if value, ok := session_data["diagnostic"].(string) ; ok {
        if value == "ok" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "control detection time expired" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_DETECTION_TIMEOUT
        } else if value == "echo function failed" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ECHO_FAILED
        } else if value == "neighbor signaled session down" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "forwarding plane reset" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_FORWARDING_RESET
        } else if value == "path down" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_PATH_DOWN
        } else if value == "concatenated path down" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_CONCATENATED_PATH_DOWN
        } else if value == "administratively down" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ADMIN_DOWN
        } else if value == "reverse concatenated path down" {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_REVERSE_CONCATENATED_PATH_DOWN
        } else {
            bfdshop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_NO_DIAGNOSTIC
        }

    }

    if value, ok := session_data["remote-diagnostic"].(string) ; ok {
        if value == "ok" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "control detection time expired" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_DETECTION_TIMEOUT
        } else if value == "echo function failed" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ECHO_FAILED
        } else if value == "neighbor signaled session down" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "forwarding plane reset" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_FORWARDING_RESET
        } else if value == "path down" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_PATH_DOWN
        } else if value == "concatenated path down" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_CONCATENATED_PATH_DOWN
        } else if value == "administratively down" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ADMIN_DOWN
        } else if value == "reverse concatenated path down" {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_REVERSE_CONCATENATED_PATH_DOWN
        } else {
            bfdshop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_NO_DIAGNOSTIC
        }
    }

    if value, ok := session_data["remote-receive-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.RemoteMinimumReceiveInterval = &value32
    }

    /*if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.DemandModeRequested = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.RemoteAuthenticationEnabled = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.RemoteControlPlaneIndependent = &value
    }*/

    if value, ok := session_data["peer_type"].(string) ; ok {
        if value == "configured" {
            bfdshop_obj.SessionType = ocbinds.OpenconfigBfdExt_BfdSessionType_CONFIGURED
        } else {
            bfdshop_obj.SessionType = ocbinds.OpenconfigBfdExt_BfdSessionType_DYNAMIC
        }
    }

    if value, ok := session_data["remote-detect-multiplier"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.RemoteMultiplier = &value32
    }

    if value, ok := session_data["detect-multiplier"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.LocalMultiplier = &value32
    }

    if value, ok := session_data["transmit-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.NegotiatedTransmissionInterval = &value32
    }

    if value, ok := session_data["receive-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.NegotiatedReceiveInterval = &value32
    }

    if value, ok := session_data["remote-transmit-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.RemoteDesiredTransmissionInterval = &value32
    }

    if value, ok := session_data["remote-echo-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.RemoteEchoReceiveInterval = &value32
    }

    if value, ok := session_data["echo-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdshop_obj.MinimumEchoInterval = &value32
    }

    /*if value, ok := session_data[""].(uint64) ; ok {
        bfdshop_obj.LastUpTime = &value
    }*/

    if bfdshop_obj.SessionState == ocbinds.OpenconfigBfd_BfdSessionState_UP {
        if value, ok := session_data["uptime"].(float64) ; ok {
            value64 := uint64(value)
            bfdshop_obj.LastUpTime = &value64
        }
    }

    bfdasyncstats = bfdshop_obj.Async
    bfdechocstats = bfdshop_obj.Echo

    /*if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketReceived = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketTransmitted = &value
    }*/

    if value, ok := counter_data["control-packet-input"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.ReceivedPackets = &value64
    }

    if value, ok := counter_data["control-packet-output"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.TransmittedPackets = &value64
    }

    if value, ok := counter_data["session-up"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.UpTransitions = &value64
    }

    if value, ok := counter_data["session-down"].(float64) ; ok {
        value64 := uint64(value)
        bfdshop_obj.FailureTransitions = &value64
    }

    /*if value, ok := counter_data[""].(bool) ; ok {
        bfdechocstats.Active = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketReceived = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketTransmitted = &value
    }*/

    if value, ok := counter_data["echo-packet-input"].(float64) ; ok {
        value64 := uint64(value)
        bfdechocstats.ReceivedPackets = &value64
    }

    if value, ok := counter_data["echo-packet-output"].(float64) ; ok {
        value64 := uint64(value)
        bfdechocstats.TransmittedPackets = &value64
    }

    /*if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.UpTransitions = &value
    }*/

    return true;
}



func fill_bfd_mhop_data (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, session_data map[string]interface{}, counter_data map[string]interface{}, bfdmhop_Input_key *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key) bool {
    var err error
    var bfdmhop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState
    var bfdmhop_tempkey ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key
    var bfdmhopkey ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key
    var bfdasyncstats *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Async

    log.Info("fill_bfd_mhop_data")

    if (nil != bfdmhop_Input_key) {
        bfdmhop_tempkey = *bfdmhop_Input_key
        bfdmhop_obj = bfd_obj.MultiHopState[bfdmhop_tempkey]
        if (nil == bfdmhop_obj) {
            log.Info("Peer with input key not found")
            return false;
        }
    } else {
        if value, ok := session_data["peer"].(string) ; ok {
            bfdmhopkey.RemoteAddress = value
        }

        if value, ok := session_data["interface"].(string) ; ok {
            bfdmhopkey.Interface = value
        } else {
            bfdmhopkey.Interface = "null"
        }

        if value, ok := session_data["vrf"].(string) ; ok {
            bfdmhopkey.Vrf = value
        }

        if value, ok := session_data["local"].(string) ; ok {
            bfdmhopkey.LocalAddress = value
        }

        bfdmhop_obj, err = bfd_obj.NewMultiHopState(bfdmhopkey.RemoteAddress, bfdmhopkey.Interface, bfdmhopkey.Vrf, bfdmhopkey.LocalAddress)
        if err != nil {return false}
    }

    ygot.BuildEmptyTree(bfdmhop_obj)

    if value, ok := session_data["status"].(string) ; ok {
        if value == "down" {
            bfdmhop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_DOWN
        } else if value == "up" {
            bfdmhop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_UP
        } else if value == "shutdown" {
            bfdmhop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_ADMIN_DOWN
        } else if value == "init" {
            bfdmhop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_INIT
        } else {
            bfdmhop_obj.SessionState = ocbinds.OpenconfigBfd_BfdSessionState_UNSET
        }
    }

    /*if value, ok := session_data["remote-status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdmhop_obj.RemoteSessionState = value
    }*/

    if bfdmhop_obj.SessionState != ocbinds.OpenconfigBfd_BfdSessionState_UP {
        if value, ok := session_data["downtime"].(float64) ; ok {
            value64 := uint64(value)
            bfdmhop_obj.LastFailureTime = &value64
        }   
    }
    if value, ok := session_data["id"].(float64) ; ok {
        s := strconv.FormatFloat(value, 'f', -1, 64)
        bfdmhop_obj.LocalDiscriminator = &s
    }

    if value, ok := session_data["remote-id"].(float64) ; ok {
        s := strconv.FormatFloat(value, 'f', -1, 64)
        bfdmhop_obj.RemoteDiscriminator = &s
    }

    if value, ok := session_data["diagnostic"].(string) ; ok {
        if value == "ok" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "control detection time expired" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_DETECTION_TIMEOUT
        } else if value == "echo function failed" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ECHO_FAILED
        } else if value == "neighbor signaled session down" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "forwarding plane reset" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_FORWARDING_RESET
        } else if value == "path down" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_PATH_DOWN
        } else if value == "concatenated path down" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_CONCATENATED_PATH_DOWN
        } else if value == "administratively down" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ADMIN_DOWN
        } else if value == "reverse concatenated path down" {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_REVERSE_CONCATENATED_PATH_DOWN
        } else {
            bfdmhop_obj.LocalDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_NO_DIAGNOSTIC
        }

    }

    if value, ok := session_data["remote-diagnostic"].(string) ; ok {
        if value == "ok" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "control detection time expired" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_DETECTION_TIMEOUT
        } else if value == "echo function failed" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ECHO_FAILED
        } else if value == "neighbor signaled session down" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_UNSET
        } else if value == "forwarding plane reset" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_FORWARDING_RESET
        } else if value == "path down" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_PATH_DOWN
        } else if value == "concatenated path down" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_CONCATENATED_PATH_DOWN
        } else if value == "administratively down" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_ADMIN_DOWN
        } else if value == "reverse concatenated path down" {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_REVERSE_CONCATENATED_PATH_DOWN
        } else {
            bfdmhop_obj.RemoteDiagnosticCode = ocbinds.OpenconfigBfd_BfdDiagnosticCode_NO_DIAGNOSTIC
        }
    }

    if value, ok := session_data["remote-receive-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.RemoteMinimumReceiveInterval = &value32
    }

    /*if value, ok := session_data[""].(bool) ; ok {
        bfdmhop_obj.DemandModeRequested = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdmhop_obj.RemoteAuthenticationEnabled = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdmhop_obj.RemoteControlPlaneIndependent = &value
    }

    if value, ok := session_data[""].(ocbinds.E_OpenconfigBfdExt_BfdSessionType) ; ok {
        bfdmhop_obj.SessionType = value
    }*/

    if value, ok := session_data["remote-detect-multiplier"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.RemoteMultiplier = &value32
    }

    if value, ok := session_data["detect-multiplier"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.LocalMultiplier = &value32
    }

    if value, ok := session_data["peer_type"].(string) ; ok {
        if value == "configured" {
            bfdmhop_obj.SessionType = ocbinds.OpenconfigBfdExt_BfdSessionType_CONFIGURED
        } else {
            bfdmhop_obj.SessionType = ocbinds.OpenconfigBfdExt_BfdSessionType_DYNAMIC
        }
    }

    if value, ok := session_data["transmit-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.NegotiatedTransmissionInterval = &value32
    }

    if value, ok := session_data["receive-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.NegotiatedReceiveInterval = &value32
    }

    if value, ok := session_data["remote-transmit-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.RemoteDesiredTransmissionInterval = &value32
    }

    if value, ok := session_data["remote-echo-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.RemoteEchoReceiveInterval = &value32
    }

    /*if value, ok := session_data["echo-interval"].(float64) ; ok {
        value32 := uint32(value)
        bfdmhop_obj.MinimumEchoInterval = &value32
    }*/

    if bfdmhop_obj.SessionState == ocbinds.OpenconfigBfd_BfdSessionState_UP {
        if value, ok := session_data["uptime"].(float64) ; ok {
            value64 := uint64(value)
            bfdmhop_obj.LastUpTime = &value64
        }
    }

    bfdasyncstats = bfdmhop_obj.Async
    //bfdechocstats = bfdmhop_obj.Echo

    /*if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketReceived = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketTransmitted = &value
    }*/

    if value, ok := counter_data["control-packet-input"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.ReceivedPackets = &value64
    }

    if value, ok := counter_data["control-packet-output"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.TransmittedPackets = &value64
    }

    if value, ok := counter_data["session-up"].(float64) ; ok {
        value64 := uint64(value)
        bfdasyncstats.UpTransitions = &value64
    }

    if value, ok := counter_data["session-down"].(float64) ; ok {
        value64 := uint64(value)
        bfdmhop_obj.FailureTransitions = &value64
    }

    /*if value, ok := counter_data[""].(bool) ; ok {
        bfdechocstats.Active = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketReceived = &value
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketTransmitted = &value
    }

    if value, ok := counter_data["echo-packet-input"].(float64) ; ok {
        value64 := uint64(value)
        bfdechocstats.ReceivedPackets = &value64
    }

    if value, ok := counter_data["echo-packet-output"].(float64) ; ok {
        value64 := uint64(value)
        bfdechocstats.TransmittedPackets = &value64
    }

    if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.UpTransitions = &value
    }*/

    return true;
}
