package transformer

import (
    "errors"
    "translib/ocbinds"
	"strings"
	"encoding/json"
	"os/exec"
	"io/ioutil"
    "github.com/openconfig/ygot/ygot"
    log "github.com/golang/glog"
)

func init () {
	XlateFuncBind("YangToDb_bfd_shop_session_key_xfmr", YangToDb_bfd_shop_session_key_xfmr)
    XlateFuncBind("DbToYang_bfd_shop_session_key_xfmr", DbToYang_bfd_shop_session_key_xfmr)
    XlateFuncBind("YangToDb_bfd_smhop_session_key_xfmr", YangToDb_bfd_mhop_session_key_xfmr)
	XlateFuncBind("DbToYang_bfd_mhop_session_key_xfmr", DbToYang_bfd_mhop_session_key_xfmr)
    XlateFuncBind("DbToYang_bfd_state_xfmr", DbToYang_bfd_state_xfmr)
    XlateFuncBind("DbToYang_bfd_shop_state_xfmr", DbToYang_bfd_shop_state_xfmr)
    XlateFuncBind("DbToYang_bfd_mhop_state_xfmr", DbToYang_bfd_mhop_state_xfmr)
}

var YangToDb_bfd_shop_session_key_xfmr  = func(inParams XfmrParams) (string, error) {
    var err error

    log.Info("DbToYang_bfd_shop_session_key_xfmr ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <remote-address, vrf, interface, local-address> */

    remoteAddress    :=  pathInfo.Var("remote-address")
    vrfName         := pathInfo.Var("vrf")
    ifName          := pathInfo.Var("interface")
    localAddress    := pathInfo.Var("local-address")

    if len(pathInfo.Vars) <  4 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(remoteAddress) == 0 {
        err = errors.New("Remote-address is missing");
        log.Info("Remote-address is Missing")
        return remoteAddress, err
    }
    if len(vrfName) == 0 {
        err = errors.New("Vrf Name is missing");
        log.Info("Vrf Name is Missing")
        return vrfName, err
    }

    if len(ifName) == 0 {
        err = errors.New("Interface name is missing")
        log.Info("Interface name is Missing")
        return ifName, err
    }

    if len(localAddress) == 0 {
        err = errors.New("Local Address is missing")
        log.Info("Local Address is Missing")
        return localAddress, err
    }

    log.Info("URI REMOTE ADDRESS ", remoteAddress)
    log.Info("URI VRF NAME ", vrfName)
    log.Info("URI INTERFACE NAME ", ifName)
    log.Info("URI LOCAL ADDRESS ", localAddress)

    var bfdTableKey string

    bfdTableKey = remoteAddress + "|" + vrfName + "|" + ifName + "|" + localAddress

    log.Info("DbToYang_bfd_shop_session_key_xfmr : bfdTableKey:", bfdTableKey)
    return bfdTableKey, nil
}

var DbToYang_bfd_shop_session_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bfd_shop_session_key_xfmr: ", entry_key)

    bfdshopKey := strings.Split(entry_key, "|")

    rmap["remote-address"] = bfdshopKey[1]
    rmap["vrf"]            = bfdshopKey[2]
    rmap["interface"]      = bfdshopKey[3]
    rmap["local-address"]  = bfdshopKey[4]

    log.Info("Rmap", rmap)

    return rmap, nil
}

var YangToDb_bfd_mhop_session_key_xfmr = func(inParams XfmrParams) (string, error) {
    var err error

    log.Info("DbToYang_bfd_mhop_session_key_xfmr  ***", inParams.uri)
    pathInfo := NewPathInfo(inParams.uri)

    /* Key should contain, <remote-address, vrf, interface, local-address> */

    remoteAddress    :=  pathInfo.Var("remote-address")
    vrfName         := pathInfo.Var("vrf")
    localAddress    := pathInfo.Var("local-address")

    if len(pathInfo.Vars) <  4 {
        err = errors.New("Invalid Key length");
        log.Info("Invalid Key length", len(pathInfo.Vars))
        return vrfName, err
    }

    if len(remoteAddress) == 0 {
        err = errors.New("Remote-address is missing");
        log.Info("Remote-address is Missing")
        return remoteAddress, err
    }

    if len(vrfName) == 0 {
        err = errors.New("Vrf Name is missing");
        log.Info("Vrf Name is Missing")
        return vrfName, err
    }

    if len(localAddress) == 0 {
        err = errors.New("Local Address is missing")
        log.Info("Local Address is Missing")
        return localAddress, err
    }

    log.Info("URI REMOTE ADDRESS ", remoteAddress)
    log.Info("URI VRF NAME ", vrfName)
    log.Info("URI LOCAL ADDRESS ", localAddress)

    var bfdTableKey string

    bfdTableKey = remoteAddress + "|" + vrfName + "|" + localAddress

    log.Info("DbToYang_bfd_mhop_session_key_xfmr : bfdTableKey:", bfdTableKey)
    return bfdTableKey, nil
}

var DbToYang_bfd_mhop_session_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bfd_mhop_session_key_xfmr: ", entry_key)

    bfdmhopKey := strings.Split(entry_key, "|")

    rmap["remote-address"] = bfdmhopKey[1]
    rmap["vrf"]            = bfdmhopKey[2]
    rmap["local-address"]  = bfdmhopKey[3]

    log.Info("Rmap", rmap)

    return rmap, nil
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

func get_bfd_specific_shop_peer (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, inParams XfmrParams) error {
	var err error
	var vtysh_cmd string
	var bfdshop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState
	var bfdshop_key ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key
	bfdMapJson := make(map[string]interface{})	
	bfdCounterMapJson := make(map[string]interface{})

	pathInfo := NewPathInfo(inParams.uri)

    log.Info(inParams.uri)

	bfdshop_key.RemoteAddress = pathInfo.Var("neighbor-address")
	bfdshop_key.Vrf = pathInfo.Var("vrf")
	bfdshop_key.Interface = pathInfo.Var("interface")
	bfdshop_key.LocalAddress = pathInfo.Var("local-address")

    log.Info(bfdshop_key)

	bfdshop_obj = bfd_obj.SingleHopState[bfdshop_key]
	if bfdshop_obj == nil {
		get_bfd_peers(bfd_obj)
		return err
		//return errors.New("BFD shop State container missing")
	}

	if (bfdshop_key.LocalAddress == "null") {
		vtysh_cmd = "show bfd peer " + bfdshop_key.RemoteAddress + " vrf " + bfdshop_key.Vrf + " interface " + bfdshop_key.Interface
	}else {
		vtysh_cmd = "show bfd peer " + bfdshop_key.RemoteAddress + " vrf " + bfdshop_key.Vrf + " interface " + bfdshop_key.Interface + " local-address " + bfdshop_key.LocalAddress
	}

    output, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bfd peers:, err")
        return cmd_err
    }

    bfdMapJson["output"] = output

	log.Info(bfdMapJson)

	if (bfdshop_key.LocalAddress == "null") {
		vtysh_cmd = "show bfd peer " + bfdshop_key.RemoteAddress + " vrf " + bfdshop_key.Vrf + " interface " + bfdshop_key.Interface + " counters"
	}else {
		vtysh_cmd = "show bfd peer " + bfdshop_key.RemoteAddress + " vrf " + bfdshop_key.Vrf + " interface " + bfdshop_key.Interface + " local-address " + bfdshop_key.LocalAddress + " counters"
	}
	
    output, cmd_err = exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bfd peers counter:, err")

        return cmd_err
    }

    bfdCounterMapJson["output"] = output

    log.Info(bfdCounterMapJson)

	session, _ := bfdMapJson["output"].(map[string]interface{})
	counter, _ := bfdCounterMapJson["output"].(map[string]interface{})

	//session_data, ok := session.(map[string][]reflect.Type); 
	//counter_data, ok := counter.(map[string]interface{}); 

    fill_bfd_shop_data (bfd_obj, bfdshop_obj, session, counter) ;

    return err
}

var DbToYang_bfd_shop_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

    var err error
    cmn_log := "GET: xfmr for BFD shop peers state"

    bfd_obj, get_err := validate_bfd_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    err = get_bfd_specific_shop_peer (bfd_obj, inParams)

    return err;
}

func get_bfd_specific_mhop_peer (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, inParams XfmrParams) error {
	var err error
	var bfdmhop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState
	var bfdmhop_key ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key
        bfdMapJson := make(map[string]interface{})
        bfdCounterMapJson := make(map[string]interface{})
	

	pathInfo := NewPathInfo(inParams.uri)

    log.Info(pathInfo)

	bfdmhop_key.RemoteAddress = pathInfo.Var("neighbor-address")
	bfdmhop_key.Vrf = pathInfo.Var("vrf")
	bfdmhop_key.LocalAddress = pathInfo.Var("local-address")

	bfdmhop_obj = bfd_obj.MultiHopState[bfdmhop_key]
	if bfdmhop_obj == nil {
        return errors.New("BFD mhop state container missing")
	}

    log.Info(bfdmhop_key)
	//bfdmhop_key =  bfdmhop_obj[key]

	vtysh_cmd := "show bfd peer " + bfdmhop_key.RemoteAddress + " vrf " + bfdmhop_key.Vrf + " multihop " + " local-address " + bfdmhop_key.LocalAddress

    output, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
		log.Errorf("Failed to fetch bfd mhop peer:, err")
        return cmd_err
    }

    bfdMapJson["output"] = output

	log.Info(bfdMapJson)

	vtysh_cmd = "show bfd peer " + bfdmhop_key.RemoteAddress + " vrf " + bfdmhop_key.Vrf + " multihop " + " local-address " + bfdmhop_key.LocalAddress + " counters"
	
    output, cmd_err = exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch mhop bfd peers counter:, err")

        return cmd_err
    }

    bfdCounterMapJson["output"] = output

    log.Info(bfdCounterMapJson)

	session, _ := bfdMapJson["output"].(map[string]interface{})
	counter, _ := bfdCounterMapJson["output"].(map[string]interface{})

	//session_data, ok := session.(map[string]interface{}); 
	//counter_data, ok := counter.(map[string]interface{}); 

    fill_bfd_mhop_data (bfd_obj, bfdmhop_obj, session, counter) ;

    return err
}

var DbToYang_bfd_mhop_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

    var err error
    cmn_log := "GET: xfmr for BFD mhop peers state"

    bfd_obj, get_err := validate_bfd_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    err = get_bfd_specific_mhop_peer (bfd_obj, inParams)

    return err;
}

func exec_vtysh_bfd_cmd (vtysh_cmd string) (map[string]interface{}, error) {
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

    var outputJson interface{}
    var output1Json map[string]interface{}
    b, err := ioutil.ReadAll(out_stream)
	if err != nil { 
		log.Fatal(err) 
	}

	//fmt.Printf("%s", b) 
    
    err = json.Unmarshal(b, &outputJson)
    if err != nil {
        log.Errorf("Not able to decode vtysh json output: %s\n", err)
        return nil, oper_err
    }


	//log.Infof(outputJson)

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

    return output1Json, err
}


func get_bfd_peers (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState) error {
	var err error
	output := make(map[string]interface{})
	bfdMapJson := make(map[string]interface{})
	bfdCounterMapJson := make(map[string]interface{})

    //vtysh_cmd := "show bfd peers " + "json"
	vtysh_cmd := "show peer 1.1.1.1 local-address 1.1.1.2 vrf default interface Ethernet0 " + "json"
    output, cmd_err := exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
        log.Errorf("Failed to fetch bfd peers:, err")

        return cmd_err
    }

    log.Info(output)

	//output = [{"multihop":false}]	

    //output = map[string]interface{}{
                //"local-address": "1.1.1.1",
        //}
	
    bfdMapJson["output"] = output

	log.Info(bfdMapJson)
	
	vtysh_cmd = "show bfd peers counters " + " json"
    output, cmd_err = exec_vtysh_cmd (vtysh_cmd)
    if cmd_err != nil {
		log.Errorf("Failed to fetch bfd peers counters:, err")

        return cmd_err
    }

	log.Info(output)

    bfdCounterMapJson["output"] = output

    log.Info(bfdCounterMapJson)

    var bfdmhop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState
    //if bfdmhop_obj = bfd_obj.MultiHop; bfdmhop_obj == nil {
        //var _bfdmhop ocbinds.OpenconfigBfd_Bfd_Sessions_MultiHop
        //bfd_obj.MultiHop = &_bfdmhop
        //bfdmhop_obj = bfd_obj.MultiHop
    //}

    var bfdshop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState
    //if bfdshop_obj = bfd_obj.SingleHop ; bfdshop_obj == nil {
        //var _bfdshop ocbinds.OpenconfigBfd_Bfd_Sessions_SingleHop
        //bfd_obj.SingleHop = &_bfdshop
        //bfdshop_obj = bfd_obj.SingleHop
    //}

	sessions, _ := bfdMapJson["output"].(map[string]interface{})
	counters, _ := bfdCounterMapJson["output"].(map[string]interface{})

    for session, _ := range sessions {
	//sessioncount := len(sessions)
	//countercount := len(counters)

	//for i, j := 0, 0; i < sessioncount && j < countercount; i, j = i+1, j+1 {
		session_data, ok := sessions[session].(map[string]interface{}) ; if !ok {continue}
        counter_data, ok := counters[session].(map[string]interface{}) ; if !ok {continue}
        if value, ok := session_data["multihop"].(string) ; ok {
            if value == "false" {
                if ok := fill_bfd_shop_data (bfd_obj, bfdshop_obj, session_data, counter_data) ; !ok {continue}
            }else {
                if ok := fill_bfd_mhop_data (bfd_obj, bfdmhop_obj, session_data, counter_data) ; !ok {continue}
            }
        }
	}

	//bfd_obj.SingleHop = &bfdshop_obj
	//bfd_obj.MultiHop = &bfdmhop_obj

    return err
}

var DbToYang_bfd_state_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {

    var err error
    cmn_log := "GET: xfmr for BFD peers state"

    bfd_obj, get_err := validate_bfd_get (inParams, cmn_log);
    if get_err != nil {
        return get_err
    }

    err = get_bfd_peers (bfd_obj)

    return err;
}

func fill_bfd_shop_data (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, bfdshop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState, session_data map[string]interface{}, counter_data map[string]interface{}) bool {

	var bfdshopkey ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Key
	var bfdasyncstats *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Async
	var bfdechocstats *ocbinds.OpenconfigBfd_Bfd_BfdState_SingleHopState_Echo

    if value, ok := session_data["status"].(string) ; ok {
        bfdshopkey.RemoteAddress = value
    }

    if value, ok := session_data["interface"].(string) ; ok {
        bfdshopkey.Interface = value
    }

    if value, ok := session_data["vrf"].(string) ; ok {
        bfdshopkey.Vrf = value
    }

    if value, ok := session_data["peer"].(string) ; ok {
        bfdshopkey.LocalAddress = value
    }

    bfdshop_obj, err := bfd_obj.NewSingleHopState(bfdshopkey.RemoteAddress, bfdshopkey.Interface, bfdshopkey.Vrf, bfdshopkey.LocalAddress)
    if err != nil {return false}
    ygot.BuildEmptyTree(bfdshop_obj)

    if value, ok := session_data["status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdshop_obj.SessionState = value
    }
/*
    if value, ok := session_data["status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdshop_obj.RemoteSessionState = value
    }
*/
    if value, ok := session_data["downtime"].(uint64) ; ok {
        bfdshop_obj.LastFailureTime = &value
	}	

    if value, ok := session_data["id"].(string) ; ok {
        bfdshop_obj.LocalDiscriminator = &value
    }

    if value, ok := session_data["remote-id"].(string) ; ok {
        bfdshop_obj.RemoteDiscriminator = &value
    }

    if value, ok := session_data["diagnostic"].(ocbinds.E_OpenconfigBfd_BfdDiagnosticCode) ; ok {
        bfdshop_obj.LocalDiagnosticCode = value
    }

    if value, ok := session_data["remote-diagnostic"].(ocbinds.E_OpenconfigBfd_BfdDiagnosticCode) ; ok {
        bfdshop_obj.RemoteDiagnosticCode = value
    }

    if value, ok := session_data["remote-receive-interval"].(uint32) ; ok {
        bfdshop_obj.RemoteMinimumReceiveInterval = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.DemandModeRequested = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.RemoteAuthenticationEnabled = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
        bfdshop_obj.RemoteControlPlaneIndependent = &value
    }

    if value, ok := session_data[""].(ocbinds.E_OpenconfigBfdExt_BfdSessionType) ; ok {
        bfdshop_obj.SessionType = value
    }

    if value, ok := session_data[""].(uint32) ; ok {
        bfdshop_obj.RemoteMultiplier = &value
    }

    if value, ok := session_data[""].(uint32) ; ok {
        bfdshop_obj.LocalMultiplier = &value
    }

    if value, ok := session_data["transmit-interval"].(uint32) ; ok {
        bfdshop_obj.NegotiatedTransmissionInterval = &value
    }

    if value, ok := session_data["receive-interval"].(uint32) ; ok {
        bfdshop_obj.NegotiatedReceiveInterval = &value
    }

    if value, ok := session_data["remote-transmit-interval"].(uint32) ; ok {
        bfdshop_obj.RemoteDesiredTransmissionInterval = &value
    }

    if value, ok := session_data["remote-echo-interval"].(uint32) ; ok {
        bfdshop_obj.RemoteEchoReceiveInterval = &value
    }

    if value, ok := session_data[""].(uint64) ; ok {
        bfdshop_obj.LastUpTime = &value
	}

	bfdasyncstats = bfdshop_obj.Async
	bfdechocstats = bfdshop_obj.Echo
/*
	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketReceived = &value
	}

	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketTransmitted = &value
	}
*/
	if value, ok := counter_data["control-packet-input"].(uint64) ; ok {
        bfdasyncstats.ReceivedPackets = &value
	}

	if value, ok := counter_data["control-packet-output"].(uint64) ; ok {
        bfdasyncstats.TransmittedPackets = &value
	}

	if value, ok := counter_data["stats.session_up"].(uint64) ; ok {
        bfdasyncstats.UpTransitions = &value
	}

	if value, ok := counter_data["session-down"].(uint64) ; ok {
        bfdshop_obj.FailureTransitions = &value
	}

	if value, ok := counter_data[""].(bool) ; ok {
        bfdechocstats.Active = &value
	}
/*
	if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketReceived = &value
	}

	if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.LastPacketTransmitted = &value
	}
*/
	if value, ok := counter_data["bs->stats.rx_echo_pkt"].(uint64) ; ok {
        bfdechocstats.ReceivedPackets = &value
	}

	if value, ok := counter_data["bs->stats.tx_echo_pkt"].(uint64) ; ok {
        bfdechocstats.TransmittedPackets = &value
	}

	if value, ok := counter_data[""].(uint64) ; ok {
        bfdechocstats.UpTransitions = &value
	}

    return true;
}


func fill_bfd_mhop_data (bfd_obj *ocbinds.OpenconfigBfd_Bfd_BfdState, bfdmhop_obj *ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState, session_data map[string]interface{}, counter_data map[string]interface{}) bool {

    var bfdmhopkey ocbinds.OpenconfigBfd_Bfd_BfdState_MultiHopState_Key

    if value, ok := session_data["status"].(string) ; ok {
        bfdmhopkey.RemoteAddress = value
    }

    if value, ok := session_data["vrf"].(string) ; ok {
        bfdmhopkey.Vrf = value
    }

    if value, ok := session_data["peer"].(string) ; ok {
        bfdmhopkey.LocalAddress = value
    }

    bfdmhop_obj, err := bfd_obj.NewMultiHopState(bfdmhopkey.RemoteAddress, bfdmhopkey.Vrf, bfdmhopkey.LocalAddress)
    if err != nil {return false}
    ygot.BuildEmptyTree(bfdmhop_obj)


    if value, ok := session_data["status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdmhop_obj.SessionState = value
    }
/*
    if value, ok := session_data["status"].(ocbinds.E_OpenconfigBfd_BfdSessionState) ; ok {
        bfdmhop_obj.RemoteSessionState = value
    }
*/
    if value, ok := session_data["downtime"].(uint64) ; ok {
        bfdmhop_obj.LastFailureTime = &value
	}	

    if value, ok := session_data["id"].(string) ; ok {
        bfdmhop_obj.LocalDiscriminator = &value
    }

    if value, ok := session_data["remote-id"].(string) ; ok {
        bfdmhop_obj.RemoteDiscriminator = &value
    }

    if value, ok := session_data["diagnostic"].(ocbinds.E_OpenconfigBfd_BfdDiagnosticCode) ; ok {
        bfdmhop_obj.LocalDiagnosticCode = value
    }

    if value, ok := session_data["remote-diagnostic"].(ocbinds.E_OpenconfigBfd_BfdDiagnosticCode) ; ok {
        bfdmhop_obj.RemoteDiagnosticCode = value
    }

    if value, ok := session_data["remote-receive-interval"].(uint32) ; ok {
        bfdmhop_obj.RemoteMinimumReceiveInterval = &value
    }

    if value, ok := session_data[""].(bool) ; ok {
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
    }

    if value, ok := session_data[""].(uint32) ; ok {
        bfdmhop_obj.RemoteMultiplier = &value
    }

    if value, ok := session_data[""].(uint32) ; ok {
        bfdmhop_obj.LocalMultiplier = &value
    }

    if value, ok := session_data["transmit-interval"].(uint32) ; ok {
        bfdmhop_obj.NegotiatedTransmissionInterval = &value
    }

    if value, ok := session_data["receive-interval"].(uint32) ; ok {
        bfdmhop_obj.NegotiatedReceiveInterval = &value
    }

    if value, ok := session_data["remote-transmit-interval"].(uint32) ; ok {
        bfdmhop_obj.RemoteDesiredTransmissionInterval = &value
    }

    if value, ok := session_data["remote-echo-interval"].(uint32) ; ok {
        bfdmhop_obj.RemoteEchoReceiveInterval = &value
    }

    if value, ok := session_data[""].(uint64) ; ok {
        bfdmhop_obj.LastUpTime = &value
	}

	bfdasyncstats := bfdmhop_obj.Async
/*
	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketReceived = &value
	}

	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.LastPacketTransmitted = &value
	}
*/
	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.ReceivedPackets = &value
	}

	if value, ok := counter_data[""].(uint64) ; ok {
        bfdasyncstats.TransmittedPackets = &value
	}

	if value, ok := counter_data["bs->stats.session_up"].(uint64) ; ok {
        bfdasyncstats.UpTransitions = &value
	}

    return true;
}
