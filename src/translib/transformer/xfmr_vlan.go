package transformer

import (
        "errors"
        log "github.com/golang/glog"
        "strings"
        "strconv"
        "translib/ocbinds"
        "translib/db"
        "github.com/openconfig/ygot/ygot"        
)

func init() {
    XlateFuncBind("DbToYang_netinst_vlans_subtree_xfmr", DbToYang_netinst_vlans_subtree_xfmr)
}


func getUriAttributes(inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans, string, string, uint16, error) {

    var err error
    var vlanName string
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    vlanIdStr := pathInfo.Var("vlan-id")
    var vlanId uint16

    targetUriPath, _ := getYangPathFromUri(pathInfo.Path)

    if len(niName) == 0 {
        return nil, "", "", 0, errors.New("Network-instance name is missing")
    }

    deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    netInstsObj := deviceObj.NetworkInstances

    if netInstsObj.NetworkInstance == nil {
        return nil, "", "", 0, errors.New("Network-instances container missing")
    }

    netInstObj := netInstsObj.NetworkInstance[niName]
    if netInstObj == nil {
        netInstObj, _ = netInstsObj.NewNetworkInstance(niName)
        ygot.BuildEmptyTree(netInstObj)
    }

    netInstVlansObj := netInstObj.Vlans
    if netInstVlansObj == nil {
        ygot.BuildEmptyTree(netInstObj)
    }

    // Add prefix "VLAN"
    if len(vlanIdStr) > 0 {
        vlanName = "Vlan" + vlanIdStr
        var vlanId64 uint64
        if vlanId64,  err = strconv.ParseUint(vlanIdStr, 10, 16); err != nil {
            return nil,  "", "", 0, errors.New("Invalid Vlan id")
        }
        vlanId = uint16(vlanId64)
    }
    log.Infof(" niName %s vlanName %s targetUriPath %s", niName, vlanName, targetUriPath)

    return netInstObj.Vlans, niName, vlanName, vlanId, err
}

func dbToYangFillVlanMemberEntry(ocVlansVlan *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan, vlanName string, dbVal db.Value) (error){

    var dbVlanMembers []string
    if ocVlansVlan == nil {
        return errors.New("Operational Error")
    }
    var members ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan_Members
    ocVlansVlan.Members = &members

    ygot.BuildEmptyTree(&members)

    if dbVal.Has("members@") && len(dbVal.Field["members@"]) > 0 {
        dbVlanMembers = strings.SplitN(dbVal.Field["members@"], ",", -1 )
    } else {
        return nil
    }

    for i:= 0 ; i < len(dbVlanMembers); i++ {
        var vlanMember ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan_Members_Member
        var memberState ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan_Members_Member_State
        memberState.Interface = &dbVlanMembers[i]
        vlanMember.State = &memberState
        ocVlansVlan.Members.Member= append(ocVlansVlan.Members.Member, &vlanMember)
    }

    return nil
}

func dbToYangFillVlanEntry(ocVlansVlan *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan, vlanName string, vlanId uint16, dbVal db.Value) (error) {

    if ocVlansVlan == nil {
        return errors.New("Operational Error")
    }

    var state ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan_State

    ygot.BuildEmptyTree(&state)
    state.Name = &vlanName
    // convert to uint16
    state.VlanId = &vlanId
    state.Status = ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans_Vlan_Config_Status_UNSET

    ocVlansVlan.State = &state
    return dbToYangFillVlanMemberEntry(ocVlansVlan, vlanName, dbVal)
}

var DbToYang_netinst_vlans_subtree_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) error {
    var err error
    var vlansObj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Vlans
    var niName string
    var vlanName string
    var vlanId uint16
    log.Infof("DbToYang_netinst_vlans_subtree_xfmr: ")

    vlansObj, niName, vlanName, vlanId, err  = getUriAttributes(inParams)

    if (err != nil) {
        return err
    }

    if strings.HasPrefix(niName, "Vrf") {
        return nil
    }

    tblName := "VLAN"
    dbspec := &db.TableSpec { Name: tblName }

    // Vlan name given, get corresponding entry from Db.
    if len(vlanName) > 0 {
        // if network instance is vlan, return nil for other vlans.
        if strings.HasPrefix(niName, "Vlan") &&
            vlanName != niName {
            log.Infof("vlan_tbl_key_xfmr: vlanTbl_key %s, ntwk_inst %s ", vlanName, niName)
            return err
        }

        dbEntry, derr := inParams.d.GetEntry(dbspec, db.Key{Comp: []string{vlanName}})
        if derr != nil {
            log.Infof(" dbEntry get failure for Key %s", vlanName)
            return errors.New("Operational Error")
        }
        VlansVlanObj := vlansObj.Vlan[vlanId]
        err = dbToYangFillVlanEntry(VlansVlanObj, vlanName, vlanId, dbEntry)
    } else {
           var keys []db.Key
           if keys, err = inParams.d.GetKeys(&db.TableSpec{Name:tblName, CompCt:2} ); err != nil {
                return errors.New("Operational Error")
           }

           for _, key := range keys {
                dbEntry, dbErr := inParams.d.GetEntry(dbspec, key)
                if dbErr != nil {
                    log.Error("DB GetEntry failed for key : ", key)
                    continue
                }
                vlanName = key.Comp[0]
                vlanIdStr := dbEntry.Field["vlanid"]
                vlanId64, _ := strconv.ParseUint(vlanIdStr, 10, 16)
                vlanId = uint16(vlanId64)

                // if network instance is vlan, return nil for other vlans.
                if strings.HasPrefix(niName, "Vlan") &&
                    vlanName != niName {
                    log.Infof("vlan_tbl_key_xfmr: vlanTbl_key %s, ntwk_inst %s ", vlanName, niName)
                    continue
                }
                VlansVlanObj, _ := vlansObj.NewVlan(vlanId)
                if err = dbToYangFillVlanEntry(VlansVlanObj, vlanName, vlanId, dbEntry); err != nil {
                    log.Error("dbToYangFillVlanEntry failure for %s", vlanName)
                }
           }
    }
    return err
}