package transformer

import (
    "errors"
    "fmt"
    "strconv"
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
    pathId int
    pathIdKey string
    nbrAddr string
}

func print_rib_keys (rib_key *_xfmr_bgp_rib_key) string {
    return fmt.Sprintf("niName:%s ; afiSafiName:%s ; prefix:%s ; origin:%s ; pathId:%d ; pathIdKey:%s; nbrAddr:%s",
                       rib_key.niName, rib_key.afiSafiName, rib_key.prefix, rib_key.origin, rib_key.pathId, rib_key.pathIdKey, rib_key.nbrAddr)
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
        prefixData, ok := routes[prefix].(map[string]interface{})
        if !ok {continue}

        paths, ok := prefixData["paths"].(map[string]interface {})
        if !ok {continue }

        for pathId, _ := range paths {
            pathData, ok := paths[pathId].(map[string]interface{})
            if !ok {continue}

            peer, ok := pathData["peer"].(map[string]interface{})
            if !ok {continue}

            peerId, ok := peer["peerId"].(string)
            if !ok {continue}

            _route_prefix := prefix
            _route_origin := &ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_State_Origin_Union_String{peerId}
            _route_pathId_u64, _ := strconv.ParseUint(pathId, 10, 32)
            _route_pathId := uint32(_route_pathId_u64)
            ipv4LocRibRoute_obj, err := ipv4LocRibRoutes_obj.NewRoute (_route_prefix, _route_origin, _route_pathId)
            if err != nil {continue}

            var _state ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_LocRib_Routes_Route_State
            ipv4LocRibRoute_obj.State = &_state
            ipv4LocRibRouteState := ipv4LocRibRoute_obj.State

            ipv4LocRibRouteState.Prefix = &_route_prefix
            ipv4LocRibRouteState.Origin = _route_origin
            ipv4LocRibRouteState.PathId = &_route_pathId

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
        }
    }

    return err
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

    return err
}

func hdl_get_bgp_local_rib (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                            rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool

    log.Infof("%s ==> Local-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    bgpRibOutputJson, cmd_err := fake_rib_exec_vtysh_cmd ("")
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
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

func hdl_get_bgp_nbrs_adj_rib_in_pre (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                      rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    return err
}

func hdl_get_bgp_ipv4_nbrs_adj_rib_in_post (ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi,
                                rib_key *_xfmr_bgp_rib_key, bgpRibOutputJson map[string]interface{}, dbg_log *string) (error) {
    var err error
    var ok bool

    log.Infof("hdl_get_bgp_ipv4_nbrs_adj_rib_in_post: nbrAddr %s ", rib_key.nbrAddr)
    ipv4Ucast_obj := ribAfiSafi_obj.Ipv4Unicast
    ygot.BuildEmptyTree(ipv4Ucast_obj)

    ipv4NbrsRib_obj := ipv4Ucast_obj.Neighbors
    ygot.BuildEmptyTree(ipv4NbrsRib_obj)

    var ipv4NbrsRibNbr_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor

    if ipv4NbrsRibNbr_obj, ok = ipv4NbrsRib_obj.Neighbor[rib_key.nbrAddr]; !ok {
        ipv4NbrsRibNbr_obj, _ = ipv4NbrsRib_obj.NewNeighbor (rib_key.nbrAddr)
        ygot.BuildEmptyTree(ipv4NbrsRibNbr_obj)
    }

    if ipv4NbrsRibNbr_obj == nil {
        return err
    }
    ygot.BuildEmptyTree(ipv4NbrsRibNbr_obj)

    ipv4NbrsRibNbr_obj.State.NeighborAddress = &rib_key.nbrAddr

    nbrAdjRibInPost_obj := ipv4NbrsRibNbr_obj.AdjRibInPost
    ygot.BuildEmptyTree(nbrAdjRibInPost_obj)
    nbrAdjRibInPostRoutes_obj := nbrAdjRibInPost_obj.Routes

    log.Info("hdl_get_bgp_ipv4_nbrs_adj_rib_in_post: Get Routes ", bgpRibOutputJson)
    routes, ok := bgpRibOutputJson["routes"].(map[string]interface{})
    if !ok {return err}

    for prefix, _ := range routes {
        prefixData, ok := routes[prefix].(map[string]interface{})
        if !ok {continue}

        paths, ok := prefixData["paths"].([]interface {})
        if !ok {continue }

        for _, path := range paths {
            pathData, ok := path.(map[string]interface{})
            if !ok {continue}

            _route_prefix := prefix
            pathId, ok := pathData["pathId"].(float64)
            if !ok {continue}
            _route_pathId := uint(pathId)
            _route_pathId_u32 := uint32(_route_pathId)

            log.Infof("hdl_get_bgp_ipv4_nbrs_adj_rib_in_post: ****** prefix %s pathId %d ******",
                      _route_prefix, _route_pathId)

            if (rib_key.prefix != "" && (_route_prefix != rib_key.prefix)) {continue}
            if (rib_key.pathIdKey != "" && (int(_route_pathId) != rib_key.pathId)) {continue}

            var nbrRoute_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route

            key := ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_Key{
                Prefix: _route_prefix,
                PathId: uint32(_route_pathId),
            }

            if nbrRoute_obj, ok = nbrAdjRibInPostRoutes_obj.Route[key] ; !ok {
                nbrRoute_obj,  err = nbrAdjRibInPostRoutes_obj.NewRoute (_route_prefix, _route_pathId_u32)
               if err != nil {return err}
               ygot.BuildEmptyTree(nbrRoute_obj)
            }

            /* State Attributes */
            nbrRoute_obj.State.Prefix = &_route_prefix
            nbrRoute_obj.State.PathId = &_route_pathId_u32

            if value, ok := pathData["valid"].(bool) ; ok {
                nbrRoute_obj.State.ValidRoute = &value
            }

            bestPath, ok := pathData["bestpath"].(map[string]interface{})
            if ok {
                if value, ok := bestPath["overall"].(bool) ; ok {
                    nbrRoute_obj.State.BestPath = &value
                }
            }

            lastUpdate, ok := pathData["lastUpdate"].(map[string]interface{})
            if ok {
                if value, ok := lastUpdate["epoch"] ; ok {
                    _lastUpdateEpoch := uint64(value.(float64))
                    nbrRoute_obj.State.LastModified = &_lastUpdateEpoch
                }
            }
            /* Attr Sets */
           if aggAddr, ok := pathData["aggregatorId"].(string) ; ok {
                nbrRoute_obj.AttrSets.Aggregator.State.Address = &aggAddr
            }
            if value, ok := pathData["aggregatorAs"].(float64) ; ok {
                newValue := uint32(value)
                nbrRoute_obj.AttrSets.Aggregator.State.As = &newValue
            }
            if value, ok := pathData["aggregatorAs4"].(float64) ; ok {
                newValue := uint32(value)
                nbrRoute_obj.AttrSets.Aggregator.State.As4 = &newValue
            }
            if value, ok := pathData["atomicAggregate"].(bool) ; ok {
                nbrRoute_obj.AttrSets.AtomicAggregate = &value
            }
            if value, ok := pathData["localPref"].(float64) ; ok {
                newValue := uint32(value)
                nbrRoute_obj.AttrSets.LocalPref = &newValue
            }
            if value, ok := pathData["med"].(float64) ; ok {
                newValue := uint32(value)
                nbrRoute_obj.AttrSets.Med = &newValue
            }
            if value, ok := pathData["originatorId"].(string) ; ok {
                nbrRoute_obj.AttrSets.OriginatorId = &value
            }
            if value, ok := pathData["origin"].(string) ; ok {
                if value == "incomplete" {
                   nbrRoute_obj.AttrSets.Origin = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_INCOMPLETE
                }
                if value == "IGP" {
                   nbrRoute_obj.AttrSets.Origin = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_IGP
                }
                if value == "EGP" {
                   nbrRoute_obj.AttrSets.Origin = ocbinds.OpenconfigRibBgp_BgpOriginAttrType_EGP
                }
            }

            nexthops, ok := pathData["nexthops"].([]interface{})
            if ok {
              for _, nexthop := range nexthops {
                    data, ok := nexthop.(map[string]interface{})
                    if ok {
                        if value, ok := data["ip"].(string) ; ok {
                            nbrRoute_obj.AttrSets.NextHop = &value
                        }
                    }
                }
            }
            /* Cluster list */
            clusters, ok := pathData["cluster"].(map[string]interface{})
            if ok {
                lists, ok := clusters["list"].([]interface{})
                if ok {
                    for _, list := range lists {
                        nbrRoute_obj.AttrSets.ClusterList = append(nbrRoute_obj.AttrSets.ClusterList, list.(string))
                    }
                }
            }
            /* Community list */
            community, ok := pathData["community"].(map[string]interface{})
            if ok {
                lists, ok := community["list"].([]interface{})
                if ok {
                    for _, list := range lists {
                        temp, _ := nbrRoute_obj.AttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_Community_Union(list.(string))
                        nbrRoute_obj.AttrSets.Community = append(nbrRoute_obj.AttrSets.Community, temp)
                    }
                }
            }
            /* Ext Community list */

            extendedCommunity, ok := pathData["extendedCommunity"].(map[string]interface{})
            if ok {
                lists, ok := extendedCommunity["list"].([]interface{})
                if ok {
                    for _, list := range lists {
                        temp, _ := nbrRoute_obj.AttrSets.To_OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_ExtCommunity_Union(list.(string))
                        nbrRoute_obj.AttrSets.ExtCommunity = append(nbrRoute_obj.AttrSets.ExtCommunity, temp)
                    }
                }
            }

            /* asPath */
            ygot.BuildEmptyTree(nbrRoute_obj.AttrSets.AsPath)
            asPathData, ok := pathData["aspath"].(map[string]interface{})
            if ok {
                segments, ok := asPathData["segments"].([]interface{})
                if ok {
                    for _, segment := range segments {
                        data, ok := segment.(map[string]interface{})
                        if ok {
                            as := new(ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi_Ipv4Unicast_Neighbors_Neighbor_AdjRibInPost_Routes_Route_AttrSets_AsPath_AsSegment)
                            nbrRoute_obj.AttrSets.AsPath.AsSegment = append(nbrRoute_obj.AttrSets.AsPath.AsSegment, as)
                            ygot.BuildEmptyTree(as)
                            if value, ok := data["type"].(string) ; ok {
                                if value == "as-set" {
                                   as.State.Type = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_SET
                                }

                                if value == "as-sequence" {
                                   as.State.Type = ocbinds.OpenconfigRibBgp_AsPathSegmentType_AS_SEQ
                                }
                            }

                            lists, ok := data["list"].([]interface{})
                            if ok {
                                for _ , list := range lists {
                                    as.State.Member = append(as.State.Member, uint32(list.(float64)))
                                }
                            }
                        }
                    }
                }
            }
        }/* paths */
    } /* prefix */

    return err
}

func hdl_get_bgp_ipv6_nbrs_adj_rib_in_post (ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi,
                                rib_key *_xfmr_bgp_rib_key, bgpRibOutputJson map[string]interface{}, dbg_log *string) (error) {
    var err error

    log.Infof("hdl_get_bgp_ipv6_nbrs_adj_rib_in_post: nbrAddr ", rib_key.nbrAddr)

    return err
}
func hdl_get_bgp_nbrs_adj_rib_in_post (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                       rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
    oper_err := errors.New("Opertational error")
    var ok bool

    log.Infof("%s ==> NBRS-RIB invoke with keys {%s} afiSafiType:%d", *dbg_log, print_rib_keys(rib_key), afiSafiType)

    bgpRibOutputJson, cmd_err := fake_rib_nbrs_in_post_exec_vtysh_cmd ("")
    if (cmd_err != nil) {
        log.Errorf ("%s failed !! Error:%s", *dbg_log, cmd_err);
        return oper_err
    }

    log.Infof("NBRS-RIB ==> Got FRR response ---------------")

    if vrfName, ok := bgpRibOutputJson["vrfName"] ; (!ok || vrfName != rib_key.niName) {
        log.Errorf ("%s failed !! GET-req niName:%s not same as JSON-VRFname:%s", *dbg_log, rib_key.niName, vrfName)
        return oper_err
    }

    ygot.BuildEmptyTree(bgpRib_obj)

    ribAfiSafis_obj := bgpRib_obj.AfiSafis

    var ribAfiSafi_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib_AfiSafis_AfiSafi
    if ribAfiSafi_obj, ok = ribAfiSafis_obj.AfiSafi[afiSafiType] ; !ok {
        ribAfiSafi_obj, _ = ribAfiSafis_obj.NewAfiSafi(afiSafiType)
    }
    ygot.BuildEmptyTree(ribAfiSafi_obj)

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST {
        log.Info("NBRS-RIB ==> Get IPv4 UNicast Nbr data  ---------------", ribAfiSafi_obj)
        err = hdl_get_bgp_ipv4_nbrs_adj_rib_in_post (ribAfiSafi_obj, rib_key, bgpRibOutputJson, dbg_log)
    }

    if afiSafiType == ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV6_UNICAST {
        err = hdl_get_bgp_ipv6_nbrs_adj_rib_in_post (ribAfiSafi_obj, rib_key, bgpRibOutputJson, dbg_log)
    }
    return err
}

func hdl_get_bgp_nbrs_adj_rib_out_post (bgpRib_obj *ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp_Rib,
                                        rib_key *_xfmr_bgp_rib_key, afiSafiType ocbinds.E_OpenconfigBgpTypes_AFI_SAFI_TYPE, dbg_log *string) (error) {
    var err error
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
    rib_key.pathId, err = strconv.Atoi(pathInfo.Var("path-id"))
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

    log.Info("IPV6 LOCAL RIB -------------------")
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
                if err != nil {
                    log.Errorf("%s IPV6_UNICAST failed !! Error: BGP RIB container missing", cmn_log)
                return oper_err}
            }
    }

    log.Info("IPV4 NBRS RIB -------------------")
    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors/neighbor/adj-rib-in-pre/routes/route":
            if (rib_key.afiSafiName == "") || (rib_key.afiSafiName == "IPV4_UNICAST") {
                err = hdl_get_bgp_nbrs_adj_rib_in_pre (bgpRib_obj, &rib_key, ocbinds.OpenconfigBgpTypes_AFI_SAFI_TYPE_IPV4_UNICAST, &dbg_log)
                if err != nil {
                    log.Errorf("%s NBGR IPV4_UNICAST failed !! Error: BGP RIB container missing", cmn_log)
                return oper_err}
            }
    }

    switch targetUriPath {
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors": fallthrough
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
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors": fallthrough
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
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors": fallthrough
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
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv4-unicast/neighbors": fallthrough
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
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast": fallthrough
        case "/openconfig-network-instance:network-instances/network-instance/protocols/protocol/bgp/rib/afi-safis/afi-safi/ipv6-unicast/neighbors": fallthrough
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
