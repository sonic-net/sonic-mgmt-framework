package transformer

import (
    "errors"
    "fmt"
    "strconv"
    "strings"
    "translib/ocbinds"
    log "github.com/golang/glog"
    "github.com/openconfig/ygot/ygot"
)

func init () {
    XlateFuncBind("DbToYang_bgp_routes_get_xfmr", DbToYang_bgp_routes_get_xfmr)
}

type _xfmr_bgp_rib_key struct {
    niName string
    afiSafiName string
    prefix string
    origin string
    pathId uint32
    pathIdKey string
    nbrAddr string
}

func print_rib_keys (rib_key *_xfmr_bgp_rib_key) string {
    return fmt.Sprintf("niName:%s ; afiSafiName:%s ; prefix:%s ; origin:%s ; pathId:%d ; pathIdKey:%s; nbrAddr:%s",
                       rib_key.niName, rib_key.afiSafiName, rib_key.prefix, rib_key.origin, rib_key.pathId, rib_key.pathIdKey, rib_key.nbrAddr)
}

func parse_aspath_segment_data (asSegmentData map[string]interface{}, aspathSegmentType *ocbinds.E_OpenconfigRibBgp_AsPathSegmentType, aspathSegmentMember *[]uint32) bool {

    Type, ok := asSegmentData["type"].(string) ; if !ok {return false}
    switch Type {
        case "as-sequence":
            *aspathSegmentType = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_SEQ
        case "as-set":
            *aspathSegmentType = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_SET
        case "as-confed-sequence":
            *aspathSegmentType = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_CONFED_SEQUENCE
        case "as-confed-set":
            *aspathSegmentType = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_CONFED_SET
        default:
            return false
    }

    if _members, ok := asSegmentData["list"].([]interface {}) ; ok {
        for _, value := range _members {
            _memberValue := uint32(value.(float64))
            *aspathSegmentMember = append (*aspathSegmentMember, _memberValue)
        }
    }

    return true
}

func parse_origin_type_data (jsonOriginType string, originType *ocbinds.E_OpenconfigRibBgp_BgpOriginAttrType) {
    switch jsonOriginType {
        case "incomplete":
            *originType = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_INCOMPLETE
        case "IGP":
            *originType = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_IGP
        case "EGP":
            *originType = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_EGP
    }
}

func fill_ipv4_spec_pfx_path_loc_rib_data (ipv4LocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes,
                                           prefix string, pathId uint32, pathData map[string]interface{}) bool {
    var err error

    peer, ok := pathData["peer"].(map[string]interface{})
    if !ok {return false}

    peerId, ok := peer["peerId"].(string)
    if !ok {return false}

    _route_origin := &ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_State_Origin_Union_String{peerId}

	key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_Key{
		Prefix: prefix,
		Origin: _route_origin,
		PathId: pathId,
	}

    var ipv4LocRibRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route
    if ipv4LocRibRoute_obj, ok = ipv4LocRibRoutes_obj.Route[key] ; !ok {
        ipv4LocRibRoute_obj, err = ipv4LocRibRoutes_obj.NewRoute (prefix, _route_origin, pathId) ; if err != nil {return false}
    }
    ygot.BuildEmptyTree(ipv4LocRibRoute_obj)

    ipv4LocRibRouteState := ipv4LocRibRoute_obj.State

    ipv4LocRibRouteState.Prefix = &prefix
    ipv4LocRibRouteState.Origin = _route_origin
    ipv4LocRibRouteState.PathId = &pathId

    if value, ok := pathData["valid"].(bool) ; ok {
        ipv4LocRibRouteState.ValidRoute = &value
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            ipv4LocRibRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    ipv4LocRibRouteAttrSets := ipv4LocRibRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        ipv4LocRibRouteAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localpref"] ; ok {
        _localPref := uint32(value.(float64))
        ipv4LocRibRouteAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        ipv4LocRibRouteAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        ipv4LocRibRouteAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &ipv4LocRibRouteAttrSets.Origin)
    }

    ipv4LocRibRouteAggState := ipv4LocRibRouteAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        ipv4LocRibRouteAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        ipv4LocRibRouteAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   ipv4LocRibRouteAttrSets.AsPath.AsSegment = append (ipv4LocRibRouteAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if nexthops, ok := pathData["nexthops"].([]interface {}) ; ok {
        for _, _nexthopData := range nexthops {
            nexthopData := _nexthopData.(map[string]interface {})
            if ip, ok := nexthopData["ip"].(string) ; ok {
                ipv4LocRibRouteAttrSets.NextHop = &ip
                break
            }
        }
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                ipv4LocRibRouteAttrSets.ClusterList = append (ipv4LocRibRouteAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := ipv4LocRibRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    ipv4LocRibRouteAttrSets.Community = append (ipv4LocRibRouteAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := ipv4LocRibRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    ipv4LocRibRouteAttrSets.ExtCommunity = append (ipv4LocRibRouteAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func hdl_get_bgp_ipv4_local_rib (ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi,
                                 rib_key *_xfmr_bgp_rib_key, bgpRibOutputJson map[string]interface{}, dbg_log *string) (error) {
    var err error
    var ok bool

    var ipv4Ucast_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast
    if ipv4Ucast_obj = ribAfiSafi_obj.Ipv4Unicast ; ipv4Ucast_obj == nil {
        var _ipv4Ucast ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast
        ribAfiSafi_obj.Ipv4Unicast = &_ipv4Ucast
        ipv4Ucast_obj = ribAfiSafi_obj.Ipv4Unicast
    }

    var ipv4LocRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib
    if ipv4LocRib_obj = ipv4Ucast_obj.LocRib ; ipv4LocRib_obj == nil {
        var _ipv4LocRib ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib
        ipv4Ucast_obj.LocRib = &_ipv4LocRib
        ipv4LocRib_obj = ipv4Ucast_obj.LocRib
    }

    var ipv4LocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes
    if ipv4LocRibRoutes_obj = ipv4LocRib_obj.Routes ; ipv4LocRibRoutes_obj == nil {
        var _ipv4LocRibRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes
        ipv4LocRib_obj.Routes = &_ipv4LocRibRoutes
        ipv4LocRibRoutes_obj = ipv4LocRib_obj.Routes
    }

    routes, ok := bgpRibOutputJson["routes"].(map[string]interface{})
    if !ok {return err}

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        paths, ok := prefixData["paths"].([]interface {}) ; if !ok {continue}
        for _, _pathData := range paths {
            pathData := _pathData.(map[string]interface {})
            if value, ok := pathData["pathId"] ; ok {
                pathId := uint32(value.(float64))
                if ok := fill_ipv4_spec_pfx_path_loc_rib_data (ipv4LocRibRoutes_obj, prefix, pathId, pathData) ; !ok {continue}
            }
        }
    }

    return err
}

func fill_ipv6_spec_pfx_path_loc_rib_data (ipv6LocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes,
                                           prefix string, pathId uint32, pathData map[string]interface{}) bool {
    var err error
    peer, ok := pathData["peer"].(map[string]interface{})
    if !ok {return false}

    peerId, ok := peer["peerId"].(string)
    if !ok {return false}

    _route_origin := &ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route_State_Origin_Union_String{peerId}

	key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route_Key{
		Prefix: prefix,
		Origin: _route_origin,
		PathId: pathId,
	}

    var ipv6LocRibRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route
    if ipv6LocRibRoute_obj, ok = ipv6LocRibRoutes_obj.Route[key] ; !ok {
        ipv6LocRibRoute_obj, err = ipv6LocRibRoutes_obj.NewRoute (prefix, _route_origin, pathId) ; if err != nil {return false}
    }
    ygot.BuildEmptyTree(ipv6LocRibRoute_obj)

    ipv6LocRibRouteState := ipv6LocRibRoute_obj.State

    ipv6LocRibRouteState.Prefix = &prefix
    ipv6LocRibRouteState.Origin = _route_origin
    ipv6LocRibRouteState.PathId = &pathId

    if value, ok := pathData["valid"].(bool) ; ok {
        ipv6LocRibRouteState.ValidRoute = &value
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            ipv6LocRibRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    ipv6LocRibRouteAttrSets := ipv6LocRibRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        ipv6LocRibRouteAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localpref"] ; ok {
        _localPref := uint32(value.(float64))
        ipv6LocRibRouteAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        ipv6LocRibRouteAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        ipv6LocRibRouteAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &ipv6LocRibRouteAttrSets.Origin)
    }

    ipv6LocRibRouteAggState := ipv6LocRibRouteAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        ipv6LocRibRouteAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        ipv6LocRibRouteAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   ipv6LocRibRouteAttrSets.AsPath.AsSegment = append (ipv6LocRibRouteAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if nexthops, ok := pathData["nexthops"].([]interface {}) ; ok {
        for _, _nexthopData := range nexthops {
            nexthopData := _nexthopData.(map[string]interface {})
            if _scope, ok := nexthopData["used"] ; ok && _scope == true {
                if ip, ok := nexthopData["ip"].(string) ; ok {
                    ipv6LocRibRouteAttrSets.NextHop = &ip
                    break
                }
            }
        }
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                ipv6LocRibRouteAttrSets.ClusterList = append (ipv6LocRibRouteAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := ipv6LocRibRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    ipv6LocRibRouteAttrSets.Community = append (ipv6LocRibRouteAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := ipv6LocRibRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    ipv6LocRibRouteAttrSets.ExtCommunity = append (ipv6LocRibRouteAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func hdl_get_bgp_ipv6_local_rib (ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi,
                                 rib_key *_xfmr_bgp_rib_key, bgpRibOutputJson map[string]interface{}, dbg_log *string) (error) {
    var err error

    var ipv6Ucast_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast
    if ipv6Ucast_obj = ribAfiSafi_obj.Ipv6Unicast ; ipv6Ucast_obj == nil {
        var _ipv6Ucast ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast
        ribAfiSafi_obj.Ipv6Unicast = &_ipv6Ucast
        ipv6Ucast_obj = ribAfiSafi_obj.Ipv6Unicast
    }

    var ipv6LocRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib
    if ipv6LocRib_obj = ipv6Ucast_obj.LocRib ; ipv6LocRib_obj == nil {
        var _ipv6LocRib ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib
        ipv6Ucast_obj.LocRib = &_ipv6LocRib
        ipv6LocRib_obj = ipv6Ucast_obj.LocRib
    }

    var ipv6LocRibRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes
    if ipv6LocRibRoutes_obj = ipv6LocRib_obj.Routes ; ipv6LocRibRoutes_obj == nil {
        var _ipv6LocRibRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_LocRib_Routes
        ipv6LocRib_obj.Routes = &_ipv6LocRibRoutes
        ipv6LocRibRoutes_obj = ipv6LocRib_obj.Routes
    }

    routes, ok := bgpRibOutputJson["routes"].(map[string]interface{})
    if !ok {return err}

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        paths, ok := prefixData["paths"].([]interface {}) ; if !ok {continue}
        for _, _pathData := range paths {
            pathData := _pathData.(map[string]interface {})
            if value, ok := pathData["pathId"] ; ok {
                pathId := uint32(value.(float64))
                if ok := fill_ipv6_spec_pfx_path_loc_rib_data (ipv6LocRibRoutes_obj, prefix, pathId, pathData) ; !ok {continue}
            }
        }
    }

    return err
}

func hdl_get_bgp_local_rib (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                            rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool

    log.Infof("%s ==> Local-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    cmd := "show ip bgp vrf" + " " + rib_key.niName + " " + "verbose json"
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        cmd = "show bgp vrf" + " " + rib_key.niName + " " + "verbose json"
    }

    bgpRibOutputJson, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    if outError, ok := bgpRibOutputJson["warning"] ; ok {
        log.Errorf ("%s failed !!, %s", outError)
        return oper_err
    }

    if vrfName, ok := bgpRibOutputJson["vrfName"] ; (!ok || vrfName != rib_key.niName) {
        log.Errorf ("%s failed !! GET-req niName:%s not same as JSON-VRFname:%s", *dbg_log, rib_key.niName, vrfName)
        return oper_err
    }

    var ribAfiSafis_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
    if ribAfiSafis_obj = bgpRib_obj.AfiSafis ; ribAfiSafis_obj == nil {
        var _ribAfiSafis ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
        bgpRib_obj.AfiSafis = &_ribAfiSafis
        ribAfiSafis_obj = bgpRib_obj.AfiSafis
    }

    var ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi
    if ribAfiSafi_obj, ok = ribAfiSafis_obj.AfiSafi[afiSafiType] ; !ok {
        ribAfiSafi_obj, _ = ribAfiSafis_obj.NewAfiSafi (afiSafiType)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        err = hdl_get_bgp_ipv4_local_rib (ribAfiSafi_obj, rib_key, bgpRibOutputJson, dbg_log)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        err = hdl_get_bgp_ipv6_local_rib (ribAfiSafi_obj, rib_key, bgpRibOutputJson, dbg_log)
    }

    return err
}

func fill_ipv4_spec_pfx_nbr_in_pre_rib_data (ipv4InPreRoute_obj *ocbinds.
                                              OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route,
                                              prefix string, pathId uint32, pathData map[string]interface{}) bool {
    nbrRouteState := ipv4InPreRoute_obj.State
    nbrRouteState.Prefix = &prefix
    nbrRouteState.PathId = &pathId

    /* State Attributes */
    if value, ok := pathData["valid"].(bool) ; ok {
        nbrRouteState.ValidRoute = &value
    }

    symbol, ok := pathData["appliedStatusSymbols"].(map[string]interface{})
    if ok {
        if value, ok := symbol["*"] ; ok {
            _value := value.(bool)
            nbrRouteState.ValidRoute = &_value
        }
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            nbrRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    /* Attr Sets */
    routeAttrSets := ipv4InPreRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        routeAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localPref"] ; ok {
        _localPref := uint32(value.(float64))
        routeAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        routeAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        routeAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &routeAttrSets.Origin)
    }

    /* Attr Sets Aggregator */
    routeAggState := routeAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        routeAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        routeAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   routeAttrSets.AsPath.AsSegment = append (routeAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if value, ok := pathData["nextHop"] ; ok {
        _value := value.(string)
        routeAttrSets.NextHop = &_value
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                routeAttrSets.ClusterList = append (routeAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    routeAttrSets.Community = append (routeAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    routeAttrSets.ExtCommunity = append (routeAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}


func fill_bgp_ipv4_nbr_adj_rib_in_pre (ipv4Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor,
                                       rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error
   ipv4NbrAdjRibInPre_obj := ipv4Nbr_obj.AdjRibInPre

    var ipv4InPreRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes
    if ipv4InPreRoutes_obj = ipv4NbrAdjRibInPre_obj.Routes ; ipv4InPreRoutes_obj == nil {
        var _ipv4InPreRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes
        ipv4NbrAdjRibInPre_obj.Routes = &_ipv4InPreRoutes
        ipv4InPreRoutes_obj = ipv4NbrAdjRibInPre_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        value, ok := prefixData["pathId"] ; if !ok {continue}
        pathId := uint32(value.(float64))

        if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
        if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

        key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_Key{
            Prefix: prefix,
            PathId: pathId,
        }

        var ipv4InPreRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route
        if ipv4InPreRoute_obj, ok = ipv4InPreRoutes_obj.Route[key] ; !ok {
            ipv4InPreRoute_obj, err = ipv4InPreRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
        }
        ygot.BuildEmptyTree(ipv4InPreRoute_obj)
        if ok := fill_ipv4_spec_pfx_nbr_in_pre_rib_data (ipv4InPreRoute_obj, prefix, pathId, prefixData) ; !ok {continue}
    }

    return err
}

func fill_ipv6_spec_pfx_nbr_in_pre_rib_data (ipv6InPreRoute_obj *ocbinds.
                                              OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route,
                                              prefix string, pathId uint32, pathData map[string]interface{}) bool {
    nbrRouteState := ipv6InPreRoute_obj.State
    nbrRouteState.Prefix = &prefix
    nbrRouteState.PathId = &pathId

    /* State Attributes */
    if value, ok := pathData["valid"].(bool) ; ok {
        nbrRouteState.ValidRoute = &value
    }

    symbol, ok := pathData["appliedStatusSymbols"].(map[string]interface{})
    if ok {
        if value, ok := symbol["*"] ; ok {
            _value := value.(bool)
            nbrRouteState.ValidRoute = &_value
        }
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            nbrRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    /* Attr Sets */
    routeAttrSets := ipv6InPreRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        routeAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localPref"] ; ok {
        _localPref := uint32(value.(float64))
        routeAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        routeAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        routeAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &routeAttrSets.Origin)
    }

    /* Attr Sets Aggregator */
    routeAggState := routeAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        routeAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        routeAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   routeAttrSets.AsPath.AsSegment = append (routeAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if value, ok := pathData["nextHopGlobal"] ; ok {
        _value := value.(string)
        routeAttrSets.NextHop = &_value
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                routeAttrSets.ClusterList = append (routeAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    routeAttrSets.Community = append (routeAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    routeAttrSets.ExtCommunity = append (routeAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func fill_bgp_ipv6_nbr_adj_rib_in_pre (ipv6Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor,
                                       rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error
    ipv6NbrAdjRibInPre_obj := ipv6Nbr_obj.AdjRibInPre

    var ipv6InPreRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes
    if ipv6InPreRoutes_obj = ipv6NbrAdjRibInPre_obj.Routes ; ipv6InPreRoutes_obj == nil {
        var _ipv6InPreRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes
        ipv6NbrAdjRibInPre_obj.Routes = &_ipv6InPreRoutes
        ipv6InPreRoutes_obj = ipv6NbrAdjRibInPre_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        value, ok := prefixData["pathId"] ; if !ok {continue}
        pathId := uint32(value.(float64))

        if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
        if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

        key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route_Key{
            Prefix: prefix,
            PathId: pathId,
        }

        var ipv6InPreRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPre_Routes_Route
        if ipv6InPreRoute_obj, ok = ipv6InPreRoutes_obj.Route[key] ; !ok {
            ipv6InPreRoute_obj, err = ipv6InPreRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
        }
        ygot.BuildEmptyTree(ipv6InPreRoute_obj)
        if ok := fill_ipv6_spec_pfx_nbr_in_pre_rib_data (ipv6InPreRoute_obj, prefix, pathId, prefixData) ; !ok {continue}
    }
    return err
}

func fill_ipv4_spec_pfx_nbr_in_post_rib_data (ipv4InPostRoute_obj *ocbinds.
                                              OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route,
                                              prefix string, pathId uint32, pathData map[string]interface{}) bool {
    nbrRouteState := ipv4InPostRoute_obj.State
    nbrRouteState.Prefix = &prefix
    nbrRouteState.PathId = &pathId

    /* State Attributes */
    if value, ok := pathData["valid"].(bool) ; ok {
        nbrRouteState.ValidRoute = &value
    }

    bestPath, ok := pathData["bestpath"].(map[string]interface{})
    if ok {
        if value, ok := bestPath["overall"].(bool) ; ok {
            nbrRouteState.BestPath = &value
        }
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            nbrRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    /* Attr Sets */
    routeAttrSets := ipv4InPostRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        routeAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localpref"] ; ok {
        _localPref := uint32(value.(float64))
        routeAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        routeAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        routeAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &routeAttrSets.Origin)
    }

    /* Attr Sets Aggregator */
    routeAggState := routeAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        routeAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        routeAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   routeAttrSets.AsPath.AsSegment = append (routeAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    nexthops, ok := pathData["nexthops"].([]interface{})
    if ok {
        for _, nexthop := range nexthops {
            data, ok := nexthop.(map[string]interface{})
            if ok {
                if value, ok := data["ip"].(string) ; ok {
                    routeAttrSets.NextHop = &value
                }
            }
        }
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                routeAttrSets.ClusterList = append (routeAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    routeAttrSets.Community = append (routeAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    routeAttrSets.ExtCommunity = append (routeAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func fill_bgp_ipv4_nbr_adj_rib_in_post (ipv4Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor,
                                        rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error

    ipv4NbrAdjRibInPost_obj := ipv4Nbr_obj.AdjRibInPost

    var ipv4InPostRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes
    if ipv4InPostRoutes_obj = ipv4NbrAdjRibInPost_obj.Routes ; ipv4InPostRoutes_obj == nil {
        var _ipv4InPostRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes
        ipv4NbrAdjRibInPost_obj.Routes = &_ipv4InPostRoutes
        ipv4InPostRoutes_obj = ipv4NbrAdjRibInPost_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}

        paths, ok := prefixData["paths"].([]interface {})
        if !ok {continue }

        for _, path := range paths {
            pathData, ok := path.(map[string]interface{})
            if !ok {continue}

            id, ok := pathData["pathId"].(float64)
            if !ok {continue}
            pathId := uint32(id)

            if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
            if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

            key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_Key{
                Prefix: prefix,
                PathId: pathId,
            }

            var ipv4InPostRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route
            if ipv4InPostRoute_obj, ok = ipv4InPostRoutes_obj.Route[key] ; !ok {
                ipv4InPostRoute_obj, err = ipv4InPostRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
            }
            ygot.BuildEmptyTree(ipv4InPostRoute_obj)
            if ok := fill_ipv4_spec_pfx_nbr_in_post_rib_data (ipv4InPostRoute_obj, prefix, pathId, pathData) ; !ok {continue}
        }
    }

    return err
}

func fill_ipv6_spec_pfx_nbr_in_post_rib_data (ipv6InPostRoute_obj *ocbinds.
                                              OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route,
                                              prefix string, pathId uint32, pathData map[string]interface{}) bool {
    nbrRouteState := ipv6InPostRoute_obj.State
    nbrRouteState.Prefix = &prefix
    nbrRouteState.PathId = &pathId

    /* State Attributes */
    if value, ok := pathData["valid"].(bool) ; ok {
        nbrRouteState.ValidRoute = &value
    }

    bestPath, ok := pathData["bestpath"].(map[string]interface{})
    if ok {
        if value, ok := bestPath["overall"].(bool) ; ok {
            nbrRouteState.BestPath = &value
        }
    }

    lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            nbrRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    /* Attr Sets */
    routeAttrSets := ipv6InPostRoute_obj.AttrSets

    if value, ok := pathData["atomicAggregate"].(bool) ; ok {
        routeAttrSets.AtomicAggregate = &value
    }

    if value, ok := pathData["localpref"] ; ok {
        _localPref := uint32(value.(float64))
        routeAttrSets.LocalPref = &_localPref
    }

    if value, ok := pathData["med"] ; ok {
        _med := uint32(value.(float64))
        routeAttrSets.Med = &_med
    }

    if value, ok := pathData["originatorId"].(string) ; ok {
        routeAttrSets.OriginatorId = &value
    }

    if value, ok := pathData["origin"].(string) ; ok {
        parse_origin_type_data (value, &routeAttrSets.Origin)
    }

    /* Attr Sets Aggregator */
    routeAggState := routeAttrSets.Aggregator.State

    if value, ok := pathData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        routeAggState.As = &_as
    }

    if value, ok := pathData["aggregatorId"].(string) ; ok {
        routeAggState.Address = &value
    }

    if value, ok := pathData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   routeAttrSets.AsPath.AsSegment = append (routeAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    nexthops, ok := pathData["nexthops"].([]interface{})
    if ok {
        for _, nexthop := range nexthops {
            data, ok := nexthop.(map[string]interface{})
            if ok {
                if scope, ok := data["scope"].(string) ; ok {
                    if scope == "global" {
                        if value, ok := data["ip"].(string) ; ok {
                            routeAttrSets.NextHop = &value
                        }
                    }
                }
            }
        }
    }

    if value, ok := pathData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                routeAttrSets.ClusterList = append (routeAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := pathData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    routeAttrSets.Community = append (routeAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := pathData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := routeAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    routeAttrSets.ExtCommunity = append (routeAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}


func fill_bgp_ipv6_nbr_adj_rib_in_post (ipv6Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor,
                                        rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error

    ipv6NbrAdjRibInPost_obj := ipv6Nbr_obj.AdjRibInPost

    var ipv6InPostRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes
    if ipv6InPostRoutes_obj = ipv6NbrAdjRibInPost_obj.Routes ; ipv6InPostRoutes_obj == nil {
        var _ipv6InPostRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes
        ipv6NbrAdjRibInPost_obj.Routes = &_ipv6InPostRoutes
        ipv6InPostRoutes_obj = ipv6NbrAdjRibInPost_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}

        paths, ok := prefixData["paths"].([]interface {})
        if !ok {continue }

        for _, path := range paths {
            pathData, ok := path.(map[string]interface{})
            if !ok {continue}

            id, ok := pathData["pathId"].(float64)
            if !ok {continue}
            pathId := uint32(id)

            if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
            if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

            key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_Key{
                Prefix: prefix,
                PathId: pathId,
            }

            var ipv6InPostRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route
            if ipv6InPostRoute_obj, ok = ipv6InPostRoutes_obj.Route[key] ; !ok {
                ipv6InPostRoute_obj, err = ipv6InPostRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
            }
            ygot.BuildEmptyTree(ipv6InPostRoute_obj)
            if ok := fill_ipv6_spec_pfx_nbr_in_post_rib_data (ipv6InPostRoute_obj, prefix, pathId, pathData) ; !ok {continue}
        }
    }
    return err
}

func fill_ipv4_spec_pfx_nbr_out_post_rib_data (ipv4OutPostRoute_obj *ocbinds.
                                               OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route,
                                               prefix string, pathId uint32, prefixData map[string]interface{}) bool {
    ipv4NbrOutPostRouteState := ipv4OutPostRoute_obj.State
    ipv4NbrOutPostRouteState.Prefix = &prefix
    ipv4NbrOutPostRouteState.PathId = &pathId

    lastUpdate, ok := prefixData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            ipv4NbrOutPostRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    ipv4OutPostRouteAttrSets := ipv4OutPostRoute_obj.AttrSets

    if value, ok := prefixData["atomicAggregate"].(bool) ; ok {
        ipv4OutPostRouteAttrSets.AtomicAggregate = &value
    }

    if value, ok := prefixData["localPref"] ; ok {
        _localPref := uint32(value.(float64))
        ipv4OutPostRouteAttrSets.LocalPref = &_localPref
    }

    if value, ok := prefixData["med"] ; ok {
        _med := uint32(value.(float64))
        ipv4OutPostRouteAttrSets.Med = &_med
    }

    if value, ok := prefixData["originatorId"].(string) ; ok {
        ipv4OutPostRouteAttrSets.OriginatorId = &value
    }

    if value, ok := prefixData["origin"].(string) ; ok {
        parse_origin_type_data (value, &ipv4OutPostRouteAttrSets.Origin)
    }

    ipv4OutPostRouteAggState := ipv4OutPostRouteAttrSets.Aggregator.State

    if value, ok := prefixData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        ipv4OutPostRouteAggState.As = &_as
    }

    if value, ok := prefixData["aggregatorId"].(string) ; ok {
        ipv4OutPostRouteAggState.Address = &value
    }

    if value, ok := prefixData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   ipv4OutPostRouteAttrSets.AsPath.AsSegment = append (ipv4OutPostRouteAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if value, ok := prefixData["nextHop"] ; ok {
        _value := value.(string)
        ipv4OutPostRouteAttrSets.NextHop = &_value
    }

    if value, ok := prefixData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                ipv4OutPostRouteAttrSets.ClusterList = append (ipv4OutPostRouteAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := prefixData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := ipv4OutPostRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    ipv4OutPostRouteAttrSets.Community = append (ipv4OutPostRouteAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := prefixData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := ipv4OutPostRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    ipv4OutPostRouteAttrSets.ExtCommunity = append (ipv4OutPostRouteAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func fill_bgp_ipv4_nbr_adj_rib_out_post (ipv4Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor,
                                         rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error

    ipv4NbrAdjRibOutPost_obj := ipv4Nbr_obj.AdjRibOutPost

    var ipv4OutPostRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes
    if ipv4OutPostRoutes_obj = ipv4NbrAdjRibOutPost_obj.Routes ; ipv4OutPostRoutes_obj == nil {
        var _ipv4OutPostRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes
        ipv4NbrAdjRibOutPost_obj.Routes = &_ipv4OutPostRoutes
        ipv4OutPostRoutes_obj = ipv4NbrAdjRibOutPost_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        value, ok := prefixData["pathId"] ; if !ok {continue}
        pathId := uint32(value.(float64))

        if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
        if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

        key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_Key{
            Prefix: prefix,
            PathId: pathId,
        }

        var ipv4OutPostRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route
        if ipv4OutPostRoute_obj, ok = ipv4OutPostRoutes_obj.Route[key] ; !ok {
            ipv4OutPostRoute_obj, err = ipv4OutPostRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
        }
        ygot.BuildEmptyTree(ipv4OutPostRoute_obj)
        if ok := fill_ipv4_spec_pfx_nbr_out_post_rib_data (ipv4OutPostRoute_obj, prefix, pathId, prefixData) ; !ok {continue}
    }

    return err
}

func fill_ipv6_spec_pfx_nbr_out_post_rib_data (ipv6OutPostRoute_obj *ocbinds.
                                               OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route,
                                               prefix string, pathId uint32, prefixData map[string]interface{}) bool {
    ipv6NbrOutPostRouteState := ipv6OutPostRoute_obj.State
    ipv6NbrOutPostRouteState.Prefix = &prefix
    ipv6NbrOutPostRouteState.PathId = &pathId

    lastUpdate, ok := prefixData["lastUpdate"].(map[string]interface{})
    if ok {
        if value, ok := lastUpdate["epoch"] ; ok {
            _lastUpdateEpoch := uint64(value.(float64))
            ipv6NbrOutPostRouteState.LastModified = &_lastUpdateEpoch
        }
    }

    ipv6OutPostRouteAttrSets := ipv6OutPostRoute_obj.AttrSets

    if value, ok := prefixData["atomicAggregate"].(bool) ; ok {
        ipv6OutPostRouteAttrSets.AtomicAggregate = &value
    }

    if value, ok := prefixData["localPref"] ; ok {
        _localPref := uint32(value.(float64))
        ipv6OutPostRouteAttrSets.LocalPref = &_localPref
    }

    if value, ok := prefixData["med"] ; ok {
        _med := uint32(value.(float64))
        ipv6OutPostRouteAttrSets.Med = &_med
    }

    if value, ok := prefixData["originatorId"].(string) ; ok {
        ipv6OutPostRouteAttrSets.OriginatorId = &value
    }

    if value, ok := prefixData["origin"].(string) ; ok {
        parse_origin_type_data (value, &ipv6OutPostRouteAttrSets.Origin)
    }

    ipv6OutPostRouteAggState := ipv6OutPostRouteAttrSets.Aggregator.State

    if value, ok := prefixData["aggregatorAs"] ; ok {
        _as := uint32(value.(float64))
        ipv6OutPostRouteAggState.As = &_as
    }

    if value, ok := prefixData["aggregatorId"].(string) ; ok {
        ipv6OutPostRouteAggState.Address = &value
    }

    if value, ok := prefixData["aspath"].(map[string]interface{}) ; ok {
        if asPathSegments, ok := value["segments"].([]interface {}) ; ok {
            for _, asPathSegmentsData := range asPathSegments {
                var _segment ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_AsPath_AsSegment
                ygot.BuildEmptyTree (&_segment)
                if ok = parse_aspath_segment_data (asPathSegmentsData.(map[string]interface {}), &_segment.State.Type, &_segment.State.Member) ; ok {
                   ipv6OutPostRouteAttrSets.AsPath.AsSegment = append (ipv6OutPostRouteAttrSets.AsPath.AsSegment, &_segment)
                }
            }
        }
    }

    if value, ok := prefixData["nextHopGlobal"] ; ok {
        _value := value.(string)
        ipv6OutPostRouteAttrSets.NextHop = &_value
    }

    if value, ok := prefixData["clusterList"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                ipv6OutPostRouteAttrSets.ClusterList = append (ipv6OutPostRouteAttrSets.ClusterList, _listData.(string))
            }
        }
    }

    if value, ok := prefixData["community"].(map[string]interface{}) ; ok {
        if _list, ok := value["list"].([]interface{}) ; ok {
            for _, _listData := range _list {
                if _community_union, err := ipv6OutPostRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_Community_Union (_listData.(string)) ; err == nil {
                    ipv6OutPostRouteAttrSets.Community = append (ipv6OutPostRouteAttrSets.Community, _community_union)
                }
            }
        }
    }

    if value, ok := prefixData["extendedCommunity"].(map[string]interface{}) ; ok {
        if _value, ok := value["string"] ; ok {
            _community_slice := strings.Split (_value.(string), " ")
            for _, _data := range _community_slice {
                if _ext_community_union, err := ipv6OutPostRouteAttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_AttrSets_ExtCommunity_Union (_data) ; err == nil {
                    ipv6OutPostRouteAttrSets.ExtCommunity = append (ipv6OutPostRouteAttrSets.ExtCommunity, _ext_community_union)
                }
            }
        }
    }

    return true
}

func fill_bgp_ipv6_nbr_adj_rib_out_post (ipv6Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor,
                                         rib_key *_xfmr_bgp_rib_key, routes map[string]interface{}, dbg_log *string) (error) {
    var err error

    ipv6NbrAdjRibOutPost_obj := ipv6Nbr_obj.AdjRibOutPost

    var ipv6OutPostRoutes_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes
    if ipv6OutPostRoutes_obj = ipv6NbrAdjRibOutPost_obj.Routes ; ipv6OutPostRoutes_obj == nil {
        var _ipv6OutPostRoutes ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes
        ipv6NbrAdjRibOutPost_obj.Routes = &_ipv6OutPostRoutes
        ipv6OutPostRoutes_obj = ipv6NbrAdjRibOutPost_obj.Routes
    }

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{}) ; if !ok {continue}
        value, ok := prefixData["pathId"] ; if !ok {continue}
        pathId := uint32(value.(float64))

        if (rib_key.prefix != "" && (prefix != rib_key.prefix)) {continue}
        if (rib_key.pathIdKey != "" && (pathId != rib_key.pathId)) {continue}

        key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route_Key{
            Prefix: prefix,
            PathId: pathId,
        }

        var ipv6OutPostRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor_AdjRibOutPost_Routes_Route
        if ipv6OutPostRoute_obj, ok = ipv6OutPostRoutes_obj.Route[key] ; !ok {
            ipv6OutPostRoute_obj, err = ipv6OutPostRoutes_obj.NewRoute (prefix, pathId) ; if err != nil {continue}
        }
        ygot.BuildEmptyTree(ipv6OutPostRoute_obj)
        if ok := fill_ipv6_spec_pfx_nbr_out_post_rib_data (ipv6OutPostRoute_obj, prefix, pathId, prefixData) ; !ok {continue}
    }

    return err
}

func hdl_get_bgp_nbrs_adj_rib_in_pre (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                      rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool
    log.Infof("%s ==> NBRS-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    cmd := "show ip bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "received-routes verbose json"
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        cmd = "show bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "received-routes verbose json"
    }

    bgpRibOutputJson, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    if outError, ok := bgpRibOutputJson["warning"] ; ok {
        log.Errorf ("%s failed !!, %s", outError)
        return oper_err
    }

    log.Infof("NBRS-RIB ==> Got FRR response ")

    if vrfName, ok := bgpRibOutputJson["vrfName"] ; (!ok || vrfName != rib_key.niName) {
        log.Errorf ("%s failed !! GET-req niName:%s not same as JSON-VRFname:%s", *dbg_log, rib_key.niName, vrfName)
        return oper_err
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        var ipv4NbrsRibNbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor

        if ipv4NbrsRibNbr_obj, ok = bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv4Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv4NbrsRibNbr_obj)

        ipv4NbrsRibNbr_obj.State.NeighborAddress = &rib_key.nbrAddr
        routesData, ok := bgpRibOutputJson["receivedRoutes"].(map[string]interface{})
        if !ok {return err}
        err = fill_bgp_ipv4_nbr_adj_rib_in_pre (ipv4NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        var ipv6NbrsRibNbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor

        if ipv6NbrsRibNbr_obj, ok = bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv6Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv6NbrsRibNbr_obj)
        ipv6NbrsRibNbr_obj.State.NeighborAddress = &rib_key.nbrAddr
        routesData, ok := bgpRibOutputJson["receivedRoutes"].(map[string]interface{})
        if !ok {return err}

        err = fill_bgp_ipv6_nbr_adj_rib_in_pre (ipv6NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }

    return err
}

func hdl_get_bgp_nbrs_adj_rib_in_post (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                       rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool
    log.Infof("%s ==> NBRS-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    cmd := "show ip bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "routes verbose json"
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        cmd = "show bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "routes verbose json"
    }

    bgpRibOutputJson, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    if outError, ok := bgpRibOutputJson["warning"] ; ok {
        log.Errorf ("%s failed !!, %s", outError)
        return oper_err
    }

    log.Infof("NBRS-RIB ==> Got FRR response ")

    if vrfName, ok := bgpRibOutputJson["vrfName"] ; (!ok || vrfName != rib_key.niName) {
        log.Errorf ("%s failed !! GET-req niName:%s not same as JSON-VRFname:%s", *dbg_log, rib_key.niName, vrfName)
        return oper_err
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        var ipv4NbrsRibNbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor

        if ipv4NbrsRibNbr_obj, ok = bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv4Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv4NbrsRibNbr_obj)

        ipv4NbrsRibNbr_obj.State.NeighborAddress = &rib_key.nbrAddr
        routesData, ok := bgpRibOutputJson["routes"].(map[string]interface{})
        if !ok {return err}

        err = fill_bgp_ipv4_nbr_adj_rib_in_post (ipv4NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        var ipv6NbrsRibNbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor

        if ipv6NbrsRibNbr_obj, ok = bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv6Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv6NbrsRibNbr_obj)

        routesData, ok := bgpRibOutputJson["routes"].(map[string]interface{})
        if !ok {return err}

        err = fill_bgp_ipv6_nbr_adj_rib_in_post (ipv6NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }
    return err
}

func hdl_get_bgp_nbrs_adj_rib_out_post (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                        rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")

    log.Infof("%s ==> NBRS-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    if rib_key.nbrAddr == "" {
        log.Errorf ("%s failed !! Nbr-Addr not present !!", *dbg_log);
        return oper_err
    }

    cmd := "show ip bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "advertised-routes verbose json"
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        cmd = "show bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + rib_key.nbrAddr + " " + "advertised-routes verbose json"
    }

    bgpRibOutputJson, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    log.Infof("NBRS-RIB ==> Got FRR response ")
    if outError, ok := bgpRibOutputJson["warning"] ; ok {
        log.Errorf ("%s failed !!, %s", outError)
        return oper_err
    }

    if vrfName, ok := bgpRibOutputJson["vrfName"] ; (!ok || vrfName != rib_key.niName) {
        log.Errorf ("%s failed !! GET-req niName:%s not same as JSON-VRFname:%s", *dbg_log, rib_key.niName, vrfName)
        return oper_err
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        ipv4NbrsRibNbr_obj, ok := bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv4Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; if !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv4NbrsRibNbr_obj)

        ipv4NbrsRibNbr_obj.State.NeighborAddress = &rib_key.nbrAddr
        routesData, ok := bgpRibOutputJson["advertisedRoutes"].(map[string]interface{})
        if !ok {return err}

        err = fill_bgp_ipv4_nbr_adj_rib_out_post (ipv4NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        ipv6NbrsRibNbr_obj, ok := bgpRib_obj.AfiSafis.AfiSafi[afiSafiType].Ipv6Unicast.Neighbors.Neighbor[rib_key.nbrAddr]; if !ok {
            log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
            return err
        }
        ygot.BuildEmptyTree(ipv6NbrsRibNbr_obj)

        routesData, ok := bgpRibOutputJson["advertisedRoutes"].(map[string]interface{})
        if !ok {return err}

        err = fill_bgp_ipv6_nbr_adj_rib_out_post (ipv6NbrsRibNbr_obj, rib_key, routesData, dbg_log)
    }

    return err

}

func hdl_get_all_bgp_nbrs_adj_rib (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                   rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool

    log.Infof("%s ==> GET-ALL: Nbrs-Adj-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    cmd := "show ip bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + "routes verbose json"
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        cmd = "show bgp vrf" + " " + rib_key.niName + " " + "neighbors" + " " + "routes verbose json"
    }

    bgpRibOutputJson, cmd_err := exec_vtysh_cmd (cmd)
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    if outError, ok := bgpRibOutputJson["warning"] ; ok {
        log.Errorf ("%s failed !!, %s", outError)
        return oper_err
    }

    var ribAfiSafis_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
    if ribAfiSafis_obj = bgpRib_obj.AfiSafis ; ribAfiSafis_obj == nil {
        var _ribAfiSafis ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis
        bgpRib_obj.AfiSafis = &_ribAfiSafis
        ribAfiSafis_obj = bgpRib_obj.AfiSafis
    }

    var ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi
    if ribAfiSafi_obj, ok = ribAfiSafis_obj.AfiSafi[afiSafiType] ; !ok {
        ribAfiSafi_obj, _ = ribAfiSafis_obj.NewAfiSafi (afiSafiType)
    }

    var ipv4Ucast_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast
    var ipv4Nbrs_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        if ipv4Ucast_obj = ribAfiSafi_obj.Ipv4Unicast ; ipv4Ucast_obj == nil {
            var _ipv4Ucast ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast
            ribAfiSafi_obj.Ipv4Unicast = &_ipv4Ucast
            ipv4Ucast_obj = ribAfiSafi_obj.Ipv4Unicast
        }

        if ipv4Nbrs_obj = ipv4Ucast_obj.Neighbors ; ipv4Nbrs_obj == nil {
            var _ipv4Nbrs_obj ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors
            ipv4Ucast_obj.Neighbors = &_ipv4Nbrs_obj
            ipv4Nbrs_obj = ipv4Ucast_obj.Neighbors
        }
    }

    var ipv6Ucast_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast
    var ipv6Nbrs_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors
    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        if ipv6Ucast_obj = ribAfiSafi_obj.Ipv6Unicast ; ipv6Ucast_obj == nil {
            var _ipv6Ucast ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast
            ribAfiSafi_obj.Ipv6Unicast = &_ipv6Ucast
            ipv6Ucast_obj = ribAfiSafi_obj.Ipv6Unicast
        }

        if ipv6Nbrs_obj = ipv6Ucast_obj.Neighbors ; ipv6Nbrs_obj == nil {
            var _ipv6Nbrs_obj ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors
            ipv6Ucast_obj.Neighbors = &_ipv6Nbrs_obj
            ipv6Nbrs_obj = ipv6Ucast_obj.Neighbors
        }
    }

    for nbrAddr, _ := range bgpRibOutputJson {
        nbrData, ok := bgpRibOutputJson[nbrAddr].([]interface{}) ; if !ok {continue}

        rib_key.nbrAddr = nbrAddr
        log.Infof("%s ==> Local-RIB invoke with keys {%s} afiSafiType:%d *****************", *dbg_log, print_rib_keys(rib_key), afiSafiType)

        var ipv4Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor
        if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
            if ipv4Nbr_obj, err = ipv4Nbrs_obj.NewNeighbor (nbrAddr) ; err != nil {continue}
            ygot.BuildEmptyTree(ipv4Nbr_obj)
            _nbrAddr := nbrAddr
            ipv4Nbr_obj.State.NeighborAddress = &_nbrAddr
        }

        var ipv6Nbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv6Unicast_Neighbors_Neighbor
        if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
            if ipv6Nbr_obj, err = ipv6Nbrs_obj.NewNeighbor (nbrAddr) ; err != nil {continue}
            ygot.BuildEmptyTree(ipv6Nbr_obj)
            _nbrAddr := nbrAddr
            ipv6Nbr_obj.State.NeighborAddress = &_nbrAddr
        }

        for _, _routesData := range nbrData {
            routesData := _routesData.(map[string]interface {})

            if inPreRoutesData, ok := routesData["receivedRoutes"].(map[string]interface{}) ; ok {
                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
                    err = fill_bgp_ipv4_nbr_adj_rib_in_pre (ipv4Nbr_obj, rib_key, inPreRoutesData, dbg_log)
                }

                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
                    err = fill_bgp_ipv6_nbr_adj_rib_in_pre (ipv6Nbr_obj, rib_key, inPreRoutesData, dbg_log)
                }
            }

            if inPostRoutesData, ok := routesData["routes"].(map[string]interface{}) ; ok {
                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
                    err = fill_bgp_ipv4_nbr_adj_rib_in_post (ipv4Nbr_obj, rib_key, inPostRoutesData, dbg_log)
                }

                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
                    err = fill_bgp_ipv6_nbr_adj_rib_in_post (ipv6Nbr_obj, rib_key, inPostRoutesData, dbg_log)
                }
            }

            if outPostRoutesData, ok := routesData["advertisedRoutes"].(map[string]interface{}) ; ok {
                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
                    err = fill_bgp_ipv4_nbr_adj_rib_out_post (ipv4Nbr_obj, rib_key, outPostRoutesData, dbg_log)
                }

                if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
                    err = fill_bgp_ipv6_nbr_adj_rib_out_post (ipv6Nbr_obj, rib_key, outPostRoutesData, dbg_log)
                }
            }
        }
    }

    return err
}

var DbToYang_bgp_routes_get_xfmr SubTreeXfmrDbToYang = func(inParams XfmrParams) error {
    var err error
    oper_err := errors.New("Opertational error")
    cmn_log := "GET: xfmr for BGP-RIB"

    var bgp_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp
    var rib_key _xfmr_bgp_rib_key

    bgp_obj, rib_key.niName, err = getBgpRoot (inParams)
    if err != nil {
        log.Errorf ("%s failed !! Error:%s", cmn_log, err);
        return oper_err
    }

    bgpRib_obj := bgp_obj.Rib
    if bgpRib_obj == nil {
        log.Errorf("%s failed !! Error: BGP RIB container missing", cmn_log)
        return oper_err
    }

    pathInfo := NewPathInfo(inParams.uri)
    targetUriPath, err := getYangPathFromUri(pathInfo.Path)

    rib_key.afiSafiName = pathInfo.Var("afi-safi-name")
    rib_key.prefix = pathInfo.Var("prefix")
    rib_key.origin = pathInfo.Var("origin")
    rib_key.pathIdKey = pathInfo.Var("path-id")
    _pathId, err := strconv.Atoi(pathInfo.Var("path-id"))
    rib_key.pathId = uint32(_pathId)
    rib_key.nbrAddr = pathInfo.Var("neighbor-address")

    dbg_log := cmn_log + " Path: " + targetUriPath

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/loc-rib/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_bgp_local_rib (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {
                    log.Errorf("%s IPV4_UNICAST failed !! Error: BGP RIB container missing", cmn_log)
                    return oper_err
                }
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/loc-rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/loc-rib/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/loc-rib/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV6_UNICAST") {
                err = hdl_get_bgp_local_rib (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_all_bgp_nbrs_adj_rib (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV6_UNICAST") {
                err = hdl_get_all_bgp_nbrs_adj_rib (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_in_pre (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-pre": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-pre/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-pre/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV6_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_in_pre (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-post/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_in_post (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-post": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-post/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-in-post/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV6_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_in_post (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-out-post/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_out_post (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-out-post": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-out-post/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors/neighbor/adj-rib-out-post/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV6_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_out_post (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST, &dbg_log)
                if err != nil {return oper_err}
            }
    }

    return err;
}
