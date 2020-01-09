package transformer

import (
	"errors"
	_"fmt"
	"reflect"
	_ "strconv"
	"translib/ocbinds"
	log "github.com/golang/glog"
	    "github.com/openconfig/ygot/ygot"
    )


func init () {
	XlateFuncBind("DbToYang_ipv4_route_get_xfmr", DbToYang_ipv4_route_get_xfmr)
	XlateFuncBind("DbToYang_ipv6_route_get_xfmr", DbToYang_ipv6_route_get_xfmr)
}


func getIpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts, string, string, error) {
	pathInfo := NewPathInfo(inParams.uri)
	niName := pathInfo.Var("name")
	prefix := pathInfo.Var("prefix")
	var err error

	targetUriPath, _ := getYangPathFromUri(pathInfo.Path)

	if len(niName) == 0 {
		return nil, "", "", errors.New("vrf name is missing")
	}

	deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
	netInstsObj := deviceObj.NetworkInstances

	if netInstsObj.NetworkInstance == nil {
		return nil, "", "", errors.New("Network-instances container missing")
	}

	netInstObj := netInstsObj.NetworkInstance[niName]
	if netInstObj == nil {
		return nil, "", "", errors.New("Network-instances obj missing")
	}

	netInstAftsObj := netInstObj.Afts

	if netInstAftsObj == nil {
		return nil, "", "", errors.New("Network-instaces aft obj missing")
	}
	log.Infof(" niName %s targetUriPath %s", niName, targetUriPath)

	return netInstAftsObj, niName, prefix, err
}


func parse_protocol_type (jsonProtocolType string, originType *ocbinds.E_OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE) {

    switch jsonProtocolType {
        case "static":
            *originType = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_STATIC
        case "connected":
            *originType =  ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_DIRECTLY_CONNECTED
        case "bgp":
            *originType = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_BGP
        case "ospf":
        	*originType = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_OSPF
        case "ospf3":
        	*originType = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_OSPF3
        default:
        	*originType=  ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_UNSET
   	} 	
}

func fill_ipv4_nhop_entry(nexthopsArr []interface{},
						  ipv4NextHops *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv4Unicast_Ipv4Entry_NextHops) (error) {
	var err error
	var index uint64

	for _, nextHops := range nexthopsArr {

	switch  t := nextHops.(type) {

		case map[string]interface{}:
			var nextHop *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv4Unicast_Ipv4Entry_NextHops_NextHop
			nextHopsMap := nextHops.(map[string]interface{})
			isactive, ok := nextHopsMap["active"]

			if ok == false || isactive == false {
				log.Infof("Nexthop is not active, skip")
				break
			}
			index += 1
			nextHop, err = ipv4NextHops.NewNextHop(uint64(index))
			if err != nil {
				return errors.New("Operational Error")
			}
			ygot.BuildEmptyTree(nextHop)
			var state ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv4Unicast_Ipv4Entry_NextHops_NextHop_State
			for nextHopKey, nextHopVal := range nextHopsMap {
				if nextHopKey == "interfaceName" {
					intfName := nextHopVal.(string)
					nextHop.InterfaceRef.State.Interface = &intfName
   				} else if nextHopKey == "ip" {
   					ip := nextHopVal.(string)
					state.IpAddress = &ip
   				} else if nextHopKey == "directlyConnected" {
   					isDirectlyConnected := nextHopVal.(bool)
   					state.DirectlyConnected = &isDirectlyConnected
  				}
			}
			nextHop.State = &state
   		default:
   			log.Infof("Unhandled nextHops type [%s]", t)
   		}
   	}
   	return err
}


func fill_ipv4_entry (prfxValArr []interface{},
					  prfxKey string,
					  aftsObjIpv4 *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv4Unicast) (error) {
	var err error
 	var ipv4Entry *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv4Unicast_Ipv4Entry
	for _, prfxValArrVal := range prfxValArr {
		log.Infof("prfxValMap_type[%s]", reflect.TypeOf(prfxValArrVal))
		switch t := prfxValArrVal.(type) {
   	 			
		case map[string]interface{}:

			prfxValArrValMap := prfxValArrVal.(map[string]interface{})
			if _, ok := prfxValArrValMap["selected"]; ok == false {
				log.Infof("Route is not selected, skip %s", prfxKey)
			    break
			}
			var ok bool
			if ipv4Entry, ok = aftsObjIpv4.Ipv4Entry[prfxKey] ; !ok {
				ipv4Entry, err = aftsObjIpv4.NewIpv4Entry(prfxKey)
				if err != nil {
					return errors.New("Operational Error")
				}
			}
			ygot.BuildEmptyTree(ipv4Entry)

			for prfxValKey, prfxValVal := range prfxValArrValMap {

				if prfxValKey == "protocol" {
   	 			    parse_protocol_type(prfxValVal.(string), &ipv4Entry.OriginProtocol)
                } else if prfxValKey == "distance" {
   	 				distance := (uint32)(prfxValVal.(float64))
   	 				ipv4Entry.Distance = &distance
   	 			} else if prfxValKey == "metric" {
   	 				metric := (uint32)(prfxValVal.(float64))
   	 				ipv4Entry.Metric = &metric
   	 			}  else if prfxValKey == "uptime" {
   	 				uptime := prfxValVal.(string)
   	 				ipv4Entry.Uptime = &uptime
					log.Infof("uptime: [%s]", ipv4Entry.Uptime)   	 				
   	 			} else if prfxValKey == "nexthops" {
  	 				err = fill_ipv4_nhop_entry(prfxValVal.([]interface{}), ipv4Entry.NextHops)
  	 				if err != nil {
  	 					return err
  	 				}
  	 			}
   	 		}
   	 		
   	 	default:
   	 		log.Infof("Unhandled prfxValArrVal : type [%s]", t)
   	 	}
   	}
   	return err
}


func fill_ipv6_nhop_entry(nexthopsArr []interface{},
						  ipv6NextHops *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv6Unicast_Ipv6Entry_NextHops) (error) {

	var err error
	var index uint64			  	
	for _, nextHops := range nexthopsArr {

	switch  t := nextHops.(type) {

		case map[string]interface{}:
			var nextHop *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv6Unicast_Ipv6Entry_NextHops_NextHop
			
			nextHopsMap := nextHops.(map[string]interface{})
			isactive, ok := nextHopsMap["active"]

			if ok == false || isactive == false {
				log.Infof("Nexthop is not active, skip")
			    break	
			}
			index += 1
			nextHop, err = ipv6NextHops.NewNextHop(uint64(index))
			if err != nil {
				return errors.New("Operational Error")
			}
			ygot.BuildEmptyTree(nextHop)
			var state ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv6Unicast_Ipv6Entry_NextHops_NextHop_State
			for nextHopKey, nextHopVal := range nextHopsMap {
				if nextHopKey == "interfaceName" {
					intfName := nextHopVal.(string)
					nextHop.InterfaceRef.State.Interface = &intfName
   				} else if nextHopKey == "ip" {
   					ip := nextHopVal.(string)
					state.IpAddress = &ip
   				} else if nextHopKey == "directlyConnected" {
   					isDirectlyConnected := nextHopVal.(bool)
   					state.DirectlyConnected = &isDirectlyConnected
  				}
			}
			nextHop.State = &state
   		default:
   			log.Infof("Unhandled nextHops type [%s]", t)
   		}
   	}
   	return err
}

func fill_ipv6_entry (prfxValArr []interface{},
					  prfxKey string,
					  aftsObjIpv6 *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv6Unicast) (error) {

 	var err error
 	var ipv6Entry *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts_Ipv6Unicast_Ipv6Entry
	for _, prfxValArrVal := range prfxValArr {
		log.Infof("prfxValMap_type[%s]", reflect.TypeOf(prfxValArrVal))
		switch t := prfxValArrVal.(type) {
   	 			
		case map[string]interface{}:
			// skip non-selected routes.

			prfxValArrValMap := prfxValArrVal.(map[string]interface{})
			if _, ok := prfxValArrValMap["selected"]; ok == false {
				log.Infof("Route is not selected, skip %s", prfxKey)
			    break
			}
			var ok bool
			if ipv6Entry, ok = aftsObjIpv6.Ipv6Entry[prfxKey] ; !ok {
				ipv6Entry, err = aftsObjIpv6.NewIpv6Entry(prfxKey)
				if err != nil {
					return errors.New("Operational Error")
				}
			}

			ygot.BuildEmptyTree(ipv6Entry)

			for prfxValKey, prfxValVal := range prfxValArrValMap {

				if prfxValKey == "protocol" {
   	 			    parse_protocol_type(prfxValVal.(string), &ipv6Entry.OriginProtocol)
   	 			} else if prfxValKey == "distance" {
   	 				distance := (uint32)(prfxValVal.(float64))
   	 				ipv6Entry.Distance = &distance
   	 			} else if prfxValKey == "metric" {
   	 				metric := (uint32)(prfxValVal.(float64))
   	 				ipv6Entry.Metric = &metric
   	 			} else if prfxValKey == "uptime" {
   	 				uptime := prfxValVal.(string)
   	 				ipv6Entry.Uptime = &uptime
   	 				log.Infof("uptime: [%s]", ipv6Entry.Uptime)
   	 			}  else if prfxValKey == "nexthops" {
  	 				err = fill_ipv6_nhop_entry(prfxValVal.([]interface{}), ipv6Entry.NextHops)
  	 				if err != nil {
  	 					return err
  	 				}
  	 			}
   	 		}
   	 		
   	 	default:
   	 		log.Infof("Unhandled prfxValArrVal : type [%s]", t)
   	 	}
   	}
   	return err
}

var DbToYang_ipv4_route_get_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) error {

	var err error
	var aftsObj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts
	var niName string
	var prefix string

    aftsObj, niName, prefix, err = getIpRoot(inParams)
    _ = niName

    if (err != nil) {
		return err
	}

	aftsObjIpv4 := aftsObj.Ipv4Unicast
	if aftsObjIpv4  == nil {
		return errors.New("Network-instance IPv4 unicast object missing")
	}

	var outputJson map[string]interface{}
	cmd := "show ip route vrf " + niName
	if len(prefix) > 0 {
		cmd += " "
		cmd += prefix
	}
	cmd += " json"
 	log.Infof("vty cmd [%s]", cmd)
 
   	if outputJson, err = exec_vtysh_cmd(cmd); err == nil {

   		for prfxKey, prfxVal := range outputJson {

   			err = fill_ipv4_entry(prfxVal.([]interface{}), prfxKey, aftsObjIpv4)

   			if (err != nil) {
   				return err
   			}
   		}
   	}
   	return err
}

var DbToYang_ipv6_route_get_xfmr SubTreeXfmrDbToYang = func (inParams XfmrParams) error {

	var err error
	var aftsObj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Afts
	var niName string
	var prefix string

    aftsObj, niName, prefix, err = getIpRoot(inParams)
    _ = niName

    if (err != nil) {
		return err
	}

	aftsObjIpv6 := aftsObj.Ipv6Unicast
	if aftsObjIpv6  == nil {
		return errors.New("Network-instance IPv6 unicast object missing")
	}

	var outputJson map[string]interface{}
	cmd := "show ipv6 route vrf " + niName
	if len(prefix) > 0 {
		cmd += " "
		cmd += prefix
	}
	cmd += " json"
 	log.Infof("vty cmd [%s]", cmd)

   	if outputJson, err = exec_vtysh_cmd(cmd); err == nil {

   		for prfxKey, prfxVal := range outputJson {

   			err = fill_ipv6_entry(prfxVal.([]interface{}), prfxKey, aftsObjIpv6)

   			if (err != nil) {
   				return err
   			}
   		}
   	}
   	return err
}
