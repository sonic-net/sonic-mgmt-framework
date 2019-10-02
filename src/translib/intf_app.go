package translib

import (
	"errors"
	"fmt"
	log "github.com/golang/glog"
	"github.com/openconfig/ygot/ygot"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"translib/db"
	"translib/ocbinds"
	"translib/tlerr"
	"unsafe"
)

type reqType int

const (
	opCreate reqType = iota + 1
	opDelete
	opUpdate
)

type dbEntry struct {
	op    reqType
	entry db.Value
}

const (
	PORT              = "PORT"
	PORT_INDEX        = "index"
	PORT_MTU          = "mtu"
	PORT_ADMIN_STATUS = "admin_status"
	PORT_SPEED        = "speed"
	PORT_DESC         = "description"
	PORT_OPER_STATUS  = "oper_status"
)

type Table int

const (
	IF_TABLE_MAP Table = iota
	PORT_STAT_MAP
)

type IntfApp struct {
	path       *PathInfo
	reqData    []byte
	ygotRoot   *ygot.GoStruct
	ygotTarget *interface{}

	respJSON  interface{}
	allIpKeys []db.Key

	appDB      *db.DB
	countersDB *db.DB

	ifTableMap   map[string]dbEntry
	ifIPTableMap map[string]map[string]dbEntry
	portOidMap   dbEntry
	portStatMap  map[string]dbEntry

	portTs             *db.TableSpec
	portTblTs          *db.TableSpec
	intfIPTs           *db.TableSpec
	intfIPTblTs        *db.TableSpec
	intfCountrTblTs    *db.TableSpec
	portOidCountrTblTs *db.TableSpec
}

func init() {
	log.Info("Init called for INTF module")
	err := register("/openconfig-interfaces:interfaces",
		&appInfo{appType: reflect.TypeOf(IntfApp{}),
			ygotRootType: reflect.TypeOf(ocbinds.OpenconfigInterfaces_Interfaces{}),
			isNative:     false})
	if err != nil {
		log.Fatal("Register INTF app module with App Interface failed with error=", err)
	}

	err = addModel(&ModelData{Name: "openconfig-interfaces",
		Org: "OpenConfig working group",
		Ver: "1.0.2"})
	if err != nil {
		log.Fatal("Adding model data to appinterface failed with error=", err)
	}
}

func (app *IntfApp) initialize(data appData) {
	log.Info("initialize:if:path =", data.path)

	app.path = NewPathInfo(data.path)
	app.reqData = data.payload
	app.ygotRoot = data.ygotRoot
	app.ygotTarget = data.ygotTarget

	app.portTs = &db.TableSpec{Name: "PORT"}
	app.portTblTs = &db.TableSpec{Name: "PORT_TABLE"}
	app.intfIPTs = &db.TableSpec{Name: "INTERFACE"}
	app.intfIPTblTs = &db.TableSpec{Name: "INTF_TABLE", CompCt: 2}
	app.intfCountrTblTs = &db.TableSpec{Name: "COUNTERS"}
	app.portOidCountrTblTs = &db.TableSpec{Name: "COUNTERS_PORT_NAME_MAP"}

	app.ifTableMap = make(map[string]dbEntry)
	app.ifIPTableMap = make(map[string]map[string]dbEntry)
	app.portStatMap = make(map[string]dbEntry)
}

func (app *IntfApp) getAppRootObject() *ocbinds.OpenconfigInterfaces_Interfaces {
	deviceObj := (*app.ygotRoot).(*ocbinds.Device)
	return deviceObj.Interfaces
}

func (app *IntfApp) translateCreate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateCreate:intf:path =", app.path)

	err = errors.New("Not implemented")
	return keys, err
}

func (app *IntfApp) translateUpdate(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys

	log.Info("translateUpdate:intf:path =", app.path)

	keys, err = app.translateCommon(d, opUpdate)

	if err != nil {
		log.Info("Something wrong:=", err)
	}

	return keys, err
}

func (app *IntfApp) translateReplace(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	log.Info("translateReplace:intf:path =", app.path)
	err = errors.New("Not implemented")
	return keys, err
}

func (app *IntfApp) translateDelete(d *db.DB) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	pathInfo := app.path

	log.Infof("Received Delete for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	intfObj := app.getAppRootObject()

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("uripath:=", targetUriPath)
	log.Info("err:=", err)

	if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
		log.Info("len:=", len(intfObj.Interface))
		for ifKey, _ := range intfObj.Interface {
			log.Info("Name:=", ifKey)
			intf := intfObj.Interface[ifKey]

			if intf.Subinterfaces == nil {
				continue
			}
			subIf := intf.Subinterfaces.Subinterface[0]
			if subIf != nil {
				if subIf.Ipv4 != nil && subIf.Ipv4.Addresses != nil {
					for ip, _ := range subIf.Ipv4.Addresses.Address {
						addr := subIf.Ipv4.Addresses.Address[ip]
						if addr != nil {
							ipAddr := addr.Ip
							log.Info("IPv4 address = ", *ipAddr)
							if !validIPv4(*ipAddr) {
								errStr := "Invalid IPv4 address " + *ipAddr
								ipValidErr := tlerr.InvalidArgsError{Format: errStr}
								return keys, ipValidErr
							}
							err = app.validateIp(d, ifKey, *ipAddr)
							if err != nil {
								errStr := "Invalid IPv4 address " + *ipAddr
								ipValidErr := tlerr.InvalidArgsError{Format: errStr}
								return keys, ipValidErr
							}
						}
					}
				}
				if subIf.Ipv6 != nil && subIf.Ipv6.Addresses != nil {
					for ip, _ := range subIf.Ipv6.Addresses.Address {
						addr := subIf.Ipv6.Addresses.Address[ip]
						if addr != nil {
							ipAddr := addr.Ip
							log.Info("IPv6 address = ", *ipAddr)
							if !validIPv6(*ipAddr) {
								errStr := "Invalid IPv6 address " + *ipAddr
								ipValidErr := tlerr.InvalidArgsError{Format: errStr}
								return keys, ipValidErr
							}
							err = app.validateIp(d, ifKey, *ipAddr)
							if err != nil {
								errStr := "Invalid IPv6 address:" + *ipAddr
								ipValidErr := tlerr.InvalidArgsError{Format: errStr}
								return keys, ipValidErr
							}
						}
					}
				}
			} else {
				err = errors.New("Only subinterface index 0 is supported")
				return keys, err
			}
		}
	} else {
		err = errors.New("Not implemented")
	}
	return keys, err
}

func (app *IntfApp) translateGet(dbs [db.MaxDB]*db.DB) error {
	var err error
	log.Info("translateGet:intf:path =", app.path)
	return err
}

func (app *IntfApp) translateAction(dbs [db.MaxDB]*db.DB) error {
    err := errors.New("Not supported")
    return err
}

func (app *IntfApp) translateSubscribe(dbs [db.MaxDB]*db.DB, path string) (*notificationOpts, *notificationInfo, error) {
	app.appDB = dbs[db.ApplDB]
	pathInfo := NewPathInfo(path)
	notifInfo := notificationInfo{dbno: db.ApplDB}
	notSupported := tlerr.NotSupportedError{Format: "Subscribe not supported", Path: path}

	if isSubtreeRequest(pathInfo.Template, "/openconfig-interfaces:interfaces") {
		if pathInfo.HasSuffix("/interface{}") ||
			pathInfo.HasSuffix("/config") ||
			pathInfo.HasSuffix("/state") {
			log.Errorf("Subscribe not supported for %s!", pathInfo.Template)
			return nil, nil, notSupported
		}
		ifKey := pathInfo.Var("name")
		if len(ifKey) == 0 {
			return nil, nil, errors.New("ifKey given is empty!")
		}
		log.Info("Interface name = ", ifKey)
		err := app.validateInterface(app.appDB, ifKey, db.Key{Comp: []string{ifKey}})
		if err != nil {
			return nil, nil, err
		}
		if pathInfo.HasSuffix("/state/oper-status") {
			notifInfo.table = db.TableSpec{Name: "PORT_TABLE"}
			notifInfo.key = asKey(ifKey)
			notifInfo.needCache = true
			return &notificationOpts{pType: OnChange}, &notifInfo, nil
		}
	}
	return nil, nil, notSupported
}

func (app *IntfApp) processCreate(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCreate:intf:path =", app.path)
	log.Info("ProcessCreate: Target Type is " + reflect.TypeOf(*app.ygotTarget).Elem().Name())

	err = errors.New("Not implemented")
	return resp, err
}

func (app *IntfApp) processUpdate(d *db.DB) (SetResponse, error) {

	log.Infof("Calling processCommon()")

	resp, err := app.processCommon(d)
	return resp, err
}

func (app *IntfApp) processReplace(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processReplace:intf:path =", app.path)
	err = errors.New("Not implemented")
	return resp, err
}

func (app *IntfApp) processDelete(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse
	log.Info("processDelete:intf:path =", app.path)

	if len(app.ifIPTableMap) == 0 {
		return resp, err
	}
	for ifKey, entrylist := range app.ifIPTableMap {
		for ip, _ := range entrylist {
			err = d.DeleteEntry(app.intfIPTs, db.Key{Comp: []string{ifKey, ip}})
			log.Infof("Deleted IP : %s for Interface : %s", ip, ifKey)
		}
	}
	return resp, err
}

/* Note : Registration already happened, followed by filling the internal DS and filling the JSON */
func (app *IntfApp) processGet(dbs [db.MaxDB]*db.DB) (GetResponse, error) {

	var err error
	var payload []byte
	pathInfo := app.path

	log.Infof("Received GET for path %s; template: %s vars=%v", pathInfo.Path, pathInfo.Template, pathInfo.Vars)
	app.appDB = dbs[db.ApplDB]
	app.countersDB = dbs[db.CountersDB]

	intfObj := app.getAppRootObject()

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("URI Path = ", targetUriPath)

	if isSubtreeRequest(targetUriPath, "/openconfig-interfaces:interfaces/interface") {
		/* Request for a specific interface */
		if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
			/* Interface name is the key */
			for ifKey, _ := range intfObj.Interface {
				log.Info("Interface Name = ", ifKey)
				ifInfo := intfObj.Interface[ifKey]
				/* Filling Interface Info to internal DS */
				err = app.convertDBIntfInfoToInternal(app.appDB, ifKey, db.Key{Comp: []string{ifKey}})
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}

				/*Check if the request is for a specific attribute in Interfaces state container*/
				oc_val := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State{}
				ok, e := app.getSpecificAttr(targetUriPath, ifKey, oc_val)
				if ok {
					if e != nil {
						return GetResponse{Payload: payload, ErrSrc: AppErr}, e
					}

					payload, err = dumpIetfJson(oc_val, false)
					if err == nil {
						return GetResponse{Payload: payload}, err
					} else {
						return GetResponse{Payload: payload, ErrSrc: AppErr}, err
					}
				}

				/* Filling the counter Info to internal DS */
				err = app.getPortOidMapForCounters(app.countersDB)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}
				err = app.convertDBIntfCounterInfoToInternal(app.countersDB, ifKey)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}

				/*Check if the request is for a specific attribute in Interfaces state COUNTERS container*/
				counter_val := &ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters{}
				ok, e = app.getSpecificCounterAttr(targetUriPath, ifKey, counter_val)
				if ok {
					if e != nil {
						return GetResponse{Payload: payload, ErrSrc: AppErr}, e
					}

					payload, err = dumpIetfJson(counter_val, false)
					if err == nil {
						return GetResponse{Payload: payload}, err
					} else {
						return GetResponse{Payload: payload, ErrSrc: AppErr}, err
					}
				}

				/* Filling Interface IP info to internal DS */
				err = app.convertDBIntfIPInfoToInternal(app.appDB, ifKey)
				if err != nil {
					return GetResponse{Payload: payload, ErrSrc: AppErr}, err
				}

				/* Filling the tree with the info we have in Internal DS */
				ygot.BuildEmptyTree(ifInfo)
				if *app.ygotTarget == ifInfo.State {
					ygot.BuildEmptyTree(ifInfo.State)
				}
				app.convertInternalToOCIntfInfo(&ifKey, ifInfo)
				if *app.ygotTarget == ifInfo {
					payload, err = dumpIetfJson(intfObj, false)
				} else {
					dummyifInfo := &ocbinds.OpenconfigInterfaces_Interfaces_Interface{}
					if *app.ygotTarget == ifInfo.Config {
						dummyifInfo.Config = ifInfo.Config
						payload, err = dumpIetfJson(dummyifInfo, false)
					} else if *app.ygotTarget == ifInfo.State {
						dummyifInfo.State = ifInfo.State
						payload, err = dumpIetfJson(dummyifInfo, false)
					} else {
						log.Info("Not supported get type!")
						err = errors.New("Requested get-type not supported!")
					}
				}
			}
		}
		return GetResponse{Payload: payload}, err
	}

	/* Get all Interfaces */
	if isSubtreeRequest(targetUriPath, "/openconfig-interfaces:interfaces") {
		log.Info("Get all Interfaces request!")
		/* Filling Interface Info to internal DS */
		err = app.convertDBIntfInfoToInternal(app.appDB, "", db.Key{})
		if err != nil {
			return GetResponse{Payload: payload, ErrSrc: AppErr}, err
		}
		/* Filling Interface IP info to internal DS */
		err = app.convertDBIntfIPInfoToInternal(app.appDB, "")
		if err != nil {
			return GetResponse{Payload: payload, ErrSrc: AppErr}, err
		}
		/* Filling the counter Info to internal DS */
		err = app.getPortOidMapForCounters(app.countersDB)
		if err != nil {
			return GetResponse{Payload: payload, ErrSrc: AppErr}, err
		}
		err = app.convertDBIntfCounterInfoToInternal(app.countersDB, "")
		if err != nil {
			return GetResponse{Payload: payload, ErrSrc: AppErr}, err
		}
		ygot.BuildEmptyTree(intfObj)
		for ifKey, _ := range app.ifTableMap {
			log.Info("If Key = ", ifKey)
			ifInfo, err := intfObj.NewInterface(ifKey)
			if err != nil {
				log.Errorf("Creation of interface subtree for %s failed!", ifKey)
				return GetResponse{Payload: payload, ErrSrc: AppErr}, err
			}
			ygot.BuildEmptyTree(ifInfo)
			app.convertInternalToOCIntfInfo(&ifKey, ifInfo)
		}
		if *app.ygotTarget == intfObj {
			payload, err = dumpIetfJson((*app.ygotRoot).(*ocbinds.Device), true)
		} else {
			log.Error("Wrong request!")
		}
	}
	return GetResponse{Payload: payload}, err
}

func (app *IntfApp) processAction(dbs [db.MaxDB]*db.DB) (ActionResponse, error) {
    var resp ActionResponse
    err := errors.New("Not implemented")

    return resp, err
}

/* Checking IP adderss is v4 */
func validIPv4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}

/* Checking IP address is v6 */
func validIPv6(ip6Address string) bool {
	ip6Address = strings.Trim(ip6Address, " ")
	re, _ := regexp.Compile(`(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))`)
	if re.MatchString(ip6Address) {
		return true
	}
	return false
}

func (app *IntfApp) doGetAllIpKeys(d *db.DB, dbSpec *db.TableSpec) ([]db.Key, error) {

	var keys []db.Key

	intfTable, err := d.GetTable(dbSpec)
	if err != nil {
		return keys, err
	}

	keys, err = intfTable.GetKeys()
	log.Infof("Found %d INTF table keys", len(keys))
	return keys, err
}

func (app *IntfApp) getSpecificAttr(targetUriPath string, ifKey string, oc_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State) (bool, error) {
	switch targetUriPath {
	case "/openconfig-interfaces:interfaces/interface/state/oper-status":
		val, e := app.getIntfAttr(ifKey, PORT_OPER_STATUS, IF_TABLE_MAP)
		if len(val) > 0 {
			switch val {
			case "up":
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UP
			case "down":
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_DOWN
			default:
				oc_val.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UNSET
			}
			return true, nil
		} else {
			return true, e
		}
	case "/openconfig-interfaces:interfaces/interface/state/admin-status":
		val, e := app.getIntfAttr(ifKey, PORT_ADMIN_STATUS, IF_TABLE_MAP)
		if len(val) > 0 {
			switch val {
			case "up":
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UP
			case "down":
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_DOWN
			default:
				oc_val.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UNSET
			}
			return true, nil
		} else {
			return true, e
		}
	case "/openconfig-interfaces:interfaces/interface/state/mtu":
		val, e := app.getIntfAttr(ifKey, PORT_MTU, IF_TABLE_MAP)
		if len(val) > 0 {
			v, e := strconv.ParseUint(val, 10, 16)
			if e == nil {
				oc_val.Mtu = (*uint16)(unsafe.Pointer(&v))
				return true, nil
			}
		}
		return true, e
	case "/openconfig-interfaces:interfaces/interface/state/ifindex":
		val, e := app.getIntfAttr(ifKey, PORT_INDEX, IF_TABLE_MAP)
		if len(val) > 0 {
			v, e := strconv.ParseUint(val, 10, 32)
			if e == nil {
				oc_val.Ifindex = (*uint32)(unsafe.Pointer(&v))
				return true, nil
			}
		}
		return true, e
	case "/openconfig-interfaces:interfaces/interface/state/description":
		val, e := app.getIntfAttr(ifKey, PORT_DESC, IF_TABLE_MAP)
		if e == nil {
			oc_val.Description = &val
			return true, nil
		}
		return true, e

	default:
		log.Infof(targetUriPath + " - Not an interface state attribute")
	}
	return false, nil
}

func (app *IntfApp) getSpecificCounterAttr(targetUriPath string, ifKey string, counter_val *ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_Counters) (bool, error) {

	var e error

	switch targetUriPath {
	case "/openconfig-interfaces:interfaces/interface/state/counters/in-octets":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_OCTETS", &counter_val.InOctets)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-unicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &counter_val.InUnicastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-broadcast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_BROADCAST_PKTS", &counter_val.InBroadcastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-multicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_MULTICAST_PKTS", &counter_val.InMulticastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-errors":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_ERRORS", &counter_val.InErrors)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-discards":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_DISCARDS", &counter_val.InDiscards)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/in-pkts":
		var inNonUCastPkt, inUCastPkt *uint64
		var in_pkts uint64

		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS", &inNonUCastPkt)
		if e == nil {
			e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_IN_UCAST_PKTS", &inUCastPkt)
			if e != nil {
				return true, e
			}
			in_pkts = *inUCastPkt + *inNonUCastPkt
			counter_val.InPkts = &in_pkts
			return true, e
		} else {
			return true, e
		}

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-octets":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_OCTETS", &counter_val.OutOctets)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-unicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &counter_val.OutUnicastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-broadcast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS", &counter_val.OutBroadcastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-multicast-pkts":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS", &counter_val.OutMulticastPkts)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-errors":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_ERRORS", &counter_val.OutErrors)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-discards":
		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_DISCARDS", &counter_val.OutDiscards)
		return true, e

	case "/openconfig-interfaces:interfaces/interface/state/counters/out-pkts":
		var outNonUCastPkt, outUCastPkt *uint64
		var out_pkts uint64

		e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS", &outNonUCastPkt)
		if e == nil {
			e = app.getCounters(ifKey, "SAI_PORT_STAT_IF_OUT_UCAST_PKTS", &outUCastPkt)
			if e != nil {
				return true, e
			}
			out_pkts = *outUCastPkt + *outNonUCastPkt
			counter_val.OutPkts = &out_pkts
			return true, e
		} else {
			return true, e
		}

	default:
		log.Infof(targetUriPath + " - Not an interface state counter attribute")
	}
	return false, nil
}

func (app *IntfApp) getCounters(ifKey string, attr string, counter_val **uint64) error {
	val, e := app.getIntfAttr(ifKey, attr, PORT_STAT_MAP)
	if len(val) > 0 {
		v, e := strconv.ParseUint(val, 10, 64)
		if e == nil {
			*counter_val = &v
			return nil
		}
	}
	return e
}

func (app *IntfApp) getIntfAttr(ifName string, attr string, table Table) (string, error) {

	var ok bool = false
	var entry dbEntry

	if table == IF_TABLE_MAP {
		entry, ok = app.ifTableMap[ifName]
	} else if table == PORT_STAT_MAP {
		entry, ok = app.portStatMap[ifName]
	} else {
		return "", errors.New("Unsupported table")
	}

	if ok {
		ifData := entry.entry

		if val, ok := ifData.Field[attr]; ok {
			return val, nil
		}
	}
	return "", errors.New("Attr " + attr + "doesn't exist in IF table Map!")
}

/***********  Translation Helper fn to convert DB Interface info to Internal DS   ***********/
func (app *IntfApp) getPortOidMapForCounters(dbCl *db.DB) error {
	var err error
	ifCountInfo, err := dbCl.GetMapAll(app.portOidCountrTblTs)
	if err != nil {
		log.Error("Port-OID (Counters) get for all the interfaces failed!")
		return err
	}
	if ifCountInfo.IsPopulated() {
		app.portOidMap.entry = ifCountInfo
	} else {
		return errors.New("Get for OID info from all the interfaces from Counters DB failed!")
	}
	return err
}

func (app *IntfApp) convertDBIntfCounterInfoToInternal(dbCl *db.DB, ifKey string) error {
	var err error

	if len(ifKey) > 0 {
		oid := app.portOidMap.entry.Field[ifKey]
		log.Infof("OID : %s received for Interface : %s", oid, ifKey)

		/* Get the statistics for the port */
		var ifStatKey db.Key
		ifStatKey.Comp = []string{oid}

		ifStatInfo, err := dbCl.GetEntry(app.intfCountrTblTs, ifStatKey)
		if err != nil {
			log.Infof("Fetching port-stat for port : %s failed!", ifKey)
			return err
		}
		app.portStatMap[ifKey] = dbEntry{entry: ifStatInfo}
	} else {
		for ifKey, _ := range app.ifTableMap {
			app.convertDBIntfCounterInfoToInternal(dbCl, ifKey)
		}
	}
	return err
}

func (app *IntfApp) validateInterface(dbCl *db.DB, ifName string, ifKey db.Key) error {
	var err error
	if len(ifName) == 0 {
		return errors.New("Empty Interface name")
	}
	app.portTblTs = &db.TableSpec{Name: "PORT_TABLE"}
	_, err = dbCl.GetEntry(app.portTblTs, ifKey)
	if err != nil {
		log.Errorf("Error found on fetching Interface info from App DB for If Name : %s", ifName)
		errStr := "Invalid Interface:" + ifName
		err = tlerr.InvalidArgsError{Format: errStr}
		return err
	}
	return err
}

func (app *IntfApp) convertDBIntfInfoToInternal(dbCl *db.DB, ifName string, ifKey db.Key) error {

	var err error
	/* Fetching DB data for a specific Interface */
	if len(ifName) > 0 {
		log.Info("Updating Interface info from APP-DB to Internal DS for Interface name : ", ifName)
		ifInfo, err := dbCl.GetEntry(app.portTblTs, ifKey)
		if err != nil {
			log.Errorf("Error found on fetching Interface info from App DB for If Name : %s", ifName)
			errStr := "Invalid Interface:" + ifName
			err = tlerr.InvalidArgsError{Format: errStr}
			return err
		}
		if ifInfo.IsPopulated() {
			log.Info("Interface Info populated for ifName : ", ifName)
			app.ifTableMap[ifName] = dbEntry{entry: ifInfo}
		} else {
			return errors.New("Populating Interface info for " + ifName + "failed")
		}
	} else {
		log.Info("App-DB get for all the interfaces")
		tbl, err := dbCl.GetTable(app.portTblTs)
		if err != nil {
			log.Error("App-DB get for list of interfaces failed!")
			return err
		}
		keys, _ := tbl.GetKeys()
		for _, key := range keys {
			app.convertDBIntfInfoToInternal(dbCl, key.Get(0), db.Key{Comp: []string{key.Get(0)}})
		}
	}
	return err
}

/***********  Translation Helper fn to convert DB Interface IP info to Internal DS   ***********/
func (app *IntfApp) convertDBIntfIPInfoToInternal(dbCl *db.DB, ifName string) error {

	var err error
	log.Info("Updating Interface IP Info from APP-DB to Internal DS for Interface Name : ", ifName)
	app.allIpKeys, _ = app.doGetAllIpKeys(dbCl, app.intfIPTblTs)

	for _, key := range app.allIpKeys {
		if len(key.Comp) <= 1 {
			continue
		}
		ipInfo, err := dbCl.GetEntry(app.intfIPTblTs, key)
		if err != nil {
			log.Errorf("Error found on fetching Interface IP info from App DB for Interface Name : %s", ifName)
			return err
		}
		if len(app.ifIPTableMap[key.Get(0)]) == 0 {
			app.ifIPTableMap[key.Get(0)] = make(map[string]dbEntry)
			app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
		} else {
			app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
		}
	}
	return err
}

func (app *IntfApp) convertInternalToOCIntfInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {
	app.convertInternalToOCIntfAttrInfo(ifName, ifInfo)
	app.convertInternalToOCIntfIPAttrInfo(ifName, ifInfo)
	app.convertInternalToOCPortStatInfo(ifName, ifInfo)
}

func (app *IntfApp) convertInternalToOCIntfAttrInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {

	/* Handling the Interface attributes */
	if entry, ok := app.ifTableMap[*ifName]; ok {
		ifData := entry.entry

		name := *ifName
		ifInfo.Config.Name = &name
		ifInfo.State.Name = &name

		for ifAttr := range ifData.Field {
			switch ifAttr {
			case PORT_ADMIN_STATUS:
				adminStatus := ifData.Get(ifAttr)
				ifInfo.State.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_DOWN
				if adminStatus == "up" {
					ifInfo.State.AdminStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_AdminStatus_UP
				}
			case PORT_OPER_STATUS:
				operStatus := ifData.Get(ifAttr)
				ifInfo.State.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_DOWN
				if operStatus == "up" {
					ifInfo.State.OperStatus = ocbinds.OpenconfigInterfaces_Interfaces_Interface_State_OperStatus_UP
				}
			case PORT_DESC:
				descVal := ifData.Get(ifAttr)
				descr := new(string)
				*descr = descVal
				ifInfo.Config.Description = descr
				ifInfo.State.Description = descr
			case PORT_MTU:
				mtuStr := ifData.Get(ifAttr)
				mtuVal, err := strconv.Atoi(mtuStr)
				mtu := new(uint16)
				*mtu = uint16(mtuVal)
				if err == nil {
					ifInfo.Config.Mtu = mtu
					ifInfo.State.Mtu = mtu
				}
			case PORT_SPEED:
				speed := ifData.Get(ifAttr)
				var speedEnum ocbinds.E_OpenconfigIfEthernet_ETHERNET_SPEED

				switch speed {
				case "2500":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_2500MB
				case "1000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_1GB
				case "5000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_5GB
				case "10000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_10GB
				case "25000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_25GB
				case "40000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_40GB
				case "50000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_50GB
				case "100000":
					speedEnum = ocbinds.OpenconfigIfEthernet_ETHERNET_SPEED_SPEED_100GB
				default:
					log.Infof("Not supported speed: %s!", speed)
				}
				ifInfo.Ethernet.Config.PortSpeed = speedEnum
				ifInfo.Ethernet.State.PortSpeed = speedEnum
			case PORT_INDEX:
				ifIdxStr := ifData.Get(ifAttr)
				ifIdxNum, err := strconv.Atoi(ifIdxStr)
				if err == nil {
					ifIdx := new(uint32)
					*ifIdx = uint32(ifIdxNum)
					ifInfo.State.Ifindex = ifIdx
				}
			default:
				log.Info("Not a valid attribute!")
			}
		}
	}

}

func (app *IntfApp) convertInternalToOCIntfIPAttrInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {

	/* Handling the Interface IP attributes */
	subIntf, err := ifInfo.Subinterfaces.NewSubinterface(0)
	if err != nil {
		log.Error("Creation of subinterface subtree failed!")
		return
	}
	ygot.BuildEmptyTree(subIntf)
	if ipMap, ok := app.ifIPTableMap[*ifName]; ok {
		for ipKey, _ := range ipMap {
			log.Info("IP address = ", ipKey)
			ipB, ipNetB, _ := net.ParseCIDR(ipKey)

			v4Flag := false
			v6Flag := false

			var v4Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv4_Addresses_Address
			var v6Address *ocbinds.OpenconfigInterfaces_Interfaces_Interface_Subinterfaces_Subinterface_Ipv6_Addresses_Address

			if validIPv4(ipB.String()) {
				v4Address, err = subIntf.Ipv4.Addresses.NewAddress(ipB.String())
				v4Flag = true
			} else if validIPv6(ipB.String()) {
				v6Address, err = subIntf.Ipv6.Addresses.NewAddress(ipB.String())
				v6Flag = true
			} else {
				log.Error("Invalid IP address " + ipB.String())
				continue
			}

			if err != nil {
				log.Error("Creation of address subtree failed!")
				return
			}
			if v4Flag {
				ygot.BuildEmptyTree(v4Address)

				ipStr := new(string)
				*ipStr = ipB.String()
				v4Address.Ip = ipStr
				v4Address.Config.Ip = ipStr
				v4Address.State.Ip = ipStr

				ipNetBNum, _ := ipNetB.Mask.Size()
				prfxLen := new(uint8)
				*prfxLen = uint8(ipNetBNum)
				v4Address.Config.PrefixLength = prfxLen
				v4Address.State.PrefixLength = prfxLen
			}
			if v6Flag {
				ygot.BuildEmptyTree(v6Address)

				ipStr := new(string)
				*ipStr = ipB.String()
				v6Address.Ip = ipStr
				v6Address.Config.Ip = ipStr
				v6Address.State.Ip = ipStr

				ipNetBNum, _ := ipNetB.Mask.Size()
				prfxLen := new(uint8)
				*prfxLen = uint8(ipNetBNum)
				v6Address.Config.PrefixLength = prfxLen
				v6Address.State.PrefixLength = prfxLen
			}
		}
	}
}

func (app *IntfApp) convertInternalToOCPortStatInfo(ifName *string, ifInfo *ocbinds.OpenconfigInterfaces_Interfaces_Interface) {
	if len(app.portStatMap) == 0 {
		log.Errorf("Port stat info not present for interface : %s", *ifName)
		return
	}
	if portStatInfo, ok := app.portStatMap[*ifName]; ok {

		inOctet := new(uint64)
		inOctetVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_OCTETS"])
		*inOctet = uint64(inOctetVal)
		ifInfo.State.Counters.InOctets = inOctet

		inUCastPkt := new(uint64)
		inUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_UCAST_PKTS"])
		*inUCastPkt = uint64(inUCastPktVal)
		ifInfo.State.Counters.InUnicastPkts = inUCastPkt

		inNonUCastPkt := new(uint64)
		inNonUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_NON_UCAST_PKTS"])
		*inNonUCastPkt = uint64(inNonUCastPktVal)

		inPkt := new(uint64)
		*inPkt = *inUCastPkt + *inNonUCastPkt
		ifInfo.State.Counters.InPkts = inPkt

		inBCastPkt := new(uint64)
		inBCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_BROADCAST_PKTS"])
		*inBCastPkt = uint64(inBCastPktVal)
		ifInfo.State.Counters.InBroadcastPkts = inBCastPkt

		inMCastPkt := new(uint64)
		inMCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_MULTICAST_PKTS"])
		*inMCastPkt = uint64(inMCastPktVal)
		ifInfo.State.Counters.InMulticastPkts = inMCastPkt

		inErrPkt := new(uint64)
		inErrPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_ERRORS"])
		*inErrPkt = uint64(inErrPktVal)
		ifInfo.State.Counters.InErrors = inErrPkt

		inDiscPkt := new(uint64)
		inDiscPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_IN_DISCARDS"])
		*inDiscPkt = uint64(inDiscPktVal)
		ifInfo.State.Counters.InDiscards = inDiscPkt

		outOctet := new(uint64)
		outOctetVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_OCTETS"])
		*outOctet = uint64(outOctetVal)
		ifInfo.State.Counters.OutOctets = outOctet

		outUCastPkt := new(uint64)
		outUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_UCAST_PKTS"])
		*outUCastPkt = uint64(outUCastPktVal)
		ifInfo.State.Counters.OutUnicastPkts = outUCastPkt

		outNonUCastPkt := new(uint64)
		outNonUCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_NON_UCAST_PKTS"])
		*outNonUCastPkt = uint64(outNonUCastPktVal)

		outPkt := new(uint64)
		*outPkt = *outUCastPkt + *outNonUCastPkt
		ifInfo.State.Counters.OutPkts = outPkt

		outBCastPkt := new(uint64)
		outBCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_BROADCAST_PKTS"])
		*outBCastPkt = uint64(outBCastPktVal)
		ifInfo.State.Counters.OutBroadcastPkts = outBCastPkt

		outMCastPkt := new(uint64)
		outMCastPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_MULTICAST_PKTS"])
		*outMCastPkt = uint64(outMCastPktVal)
		ifInfo.State.Counters.OutMulticastPkts = outMCastPkt

		outErrPkt := new(uint64)
		outErrPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_ERRORS"])
		*outErrPkt = uint64(outErrPktVal)
		ifInfo.State.Counters.OutErrors = outErrPkt

		outDiscPkt := new(uint64)
		outDiscPktVal, _ := strconv.Atoi(portStatInfo.entry.Field["SAI_PORT_STAT_IF_OUT_DISCARDS"])
		*outDiscPkt = uint64(outDiscPktVal)
		ifInfo.State.Counters.OutDiscards = outDiscPkt
	}
}

func (app *IntfApp) translateCommon(d *db.DB, inpOp reqType) ([]db.WatchKeys, error) {
	var err error
	var keys []db.WatchKeys
	pathInfo := app.path

	log.Infof("Received UPDATE for path %s; vars=%v", pathInfo.Template, pathInfo.Vars)

	app.allIpKeys, _ = app.doGetAllIpKeys(d, app.intfIPTs)

	intfObj := app.getAppRootObject()

	targetUriPath, err := getYangPathFromUri(app.path.Path)
	log.Info("uripath:=", targetUriPath)
	log.Info("err:=", err)

	if intfObj.Interface != nil && len(intfObj.Interface) > 0 {
		log.Info("len:=", len(intfObj.Interface))
		for ifKey, _ := range intfObj.Interface {
			log.Info("Name:=", ifKey)
			intf := intfObj.Interface[ifKey]
			curr, err := d.GetEntry(app.portTs, db.Key{Comp: []string{ifKey}})
			if err != nil {
				errStr := "Invalid Interface:" + ifKey
				ifValidErr := tlerr.InvalidArgsError{Format: errStr}
				return keys, ifValidErr
			}
			if !curr.IsPopulated() {
				log.Error("Interface ", ifKey, " doesn't exist in DB")
				return keys, errors.New("Interface: " + ifKey + " doesn't exist in DB")
			}
			if intf.Config != nil {
				if intf.Config.Description != nil {
					log.Info("Description = ", *intf.Config.Description)
					curr.Field["description"] = *intf.Config.Description
				} else if intf.Config.Mtu != nil {
					log.Info("mtu:= ", *intf.Config.Mtu)
					curr.Field["mtu"] = strconv.Itoa(int(*intf.Config.Mtu))
				} else if intf.Config.Enabled != nil {
					log.Info("enabled = ", *intf.Config.Enabled)
					if *intf.Config.Enabled == true {
						curr.Field["admin_status"] = "up"
					} else {
						curr.Field["admin_status"] = "down"
					}
				}
				log.Info("Writing to db for ", ifKey)
				var entry dbEntry
				entry.op = opUpdate
				entry.entry = curr

				app.ifTableMap[ifKey] = entry
			}
			if intf.Subinterfaces == nil {
				continue
			}
			subIf := intf.Subinterfaces.Subinterface[0]
			if subIf != nil {
				if subIf.Ipv4 != nil && subIf.Ipv4.Addresses != nil {
					for ip, _ := range subIf.Ipv4.Addresses.Address {
						addr := subIf.Ipv4.Addresses.Address[ip]
						if addr.Config != nil {
							log.Info("Ip:=", *addr.Config.Ip)
							log.Info("prefix:=", *addr.Config.PrefixLength)
							if !validIPv4(*addr.Config.Ip) {
								errStr := "Invalid IPv4 address " + *addr.Config.Ip
								err = tlerr.InvalidArgsError{Format: errStr}
								return keys, err
							}
							err = app.translateIpv4(d, ifKey, *addr.Config.Ip, int(*addr.Config.PrefixLength))
							if err != nil {
								return keys, err
							}
						}
					}
				}
				if subIf.Ipv6 != nil && subIf.Ipv6.Addresses != nil {
					for ip, _ := range subIf.Ipv6.Addresses.Address {
						addr := subIf.Ipv6.Addresses.Address[ip]
						if addr.Config != nil {
							log.Info("Ip:=", *addr.Config.Ip)
							log.Info("prefix:=", *addr.Config.PrefixLength)
							if !validIPv6(*addr.Config.Ip) {
								errStr := "Invalid IPv6 address " + *addr.Config.Ip
								err = tlerr.InvalidArgsError{Format: errStr}
								return keys, err
							}
							err = app.translateIpv4(d, ifKey, *addr.Config.Ip, int(*addr.Config.PrefixLength))
							if err != nil {
								return keys, err
							}
						}
					}
				}
			} else {
				err = errors.New("Only subinterface index 0 is supported")
				return keys, err
			}
		}
	} else {
		err = errors.New("Not implemented")
	}

	return keys, err
}

/* Validates whether the IP exists in the DB */
func (app *IntfApp) validateIp(dbCl *db.DB, ifName string, ip string) error {
	app.allIpKeys, _ = app.doGetAllIpKeys(dbCl, app.intfIPTs)

	for _, key := range app.allIpKeys {
		if len(key.Comp) < 2 {
			continue
		}
		if key.Get(0) != ifName {
			continue
		}
		ipAddr, _, _ := net.ParseCIDR(key.Get(1))
		ipStr := ipAddr.String()
		if ipStr == ip {
			log.Infof("IP address %s exists, updating the DS for deletion!", ipStr)
			ipInfo, err := dbCl.GetEntry(app.intfIPTs, key)
			if err != nil {
				log.Error("Error found on fetching Interface IP info from App DB for Interface Name : ", ifName)
				return err
			}
			if len(app.ifIPTableMap[key.Get(0)]) == 0 {
				app.ifIPTableMap[key.Get(0)] = make(map[string]dbEntry)
				app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
			} else {
				app.ifIPTableMap[key.Get(0)][key.Get(1)] = dbEntry{entry: ipInfo}
			}
			return nil
		}
	}
	return errors.New(fmt.Sprintf("IP address : %s doesn't exist!", ip))
}

func (app *IntfApp) translateIpv4(d *db.DB, intf string, ip string, prefix int) error {
	var err error
	var ifsKey db.Key

	ifsKey.Comp = []string{intf}

	ipPref := ip + "/" + strconv.Itoa(prefix)
	ifsKey.Comp = []string{intf, ipPref}

	log.Info("ifsKey:=", ifsKey)

	log.Info("Checking for IP overlap ....")
	ipA, ipNetA, _ := net.ParseCIDR(ipPref)

	for _, key := range app.allIpKeys {
		if len(key.Comp) < 2 {
			continue
		}
		ipB, ipNetB, _ := net.ParseCIDR(key.Get(1))

		if ipNetA.Contains(ipB) || ipNetB.Contains(ipA) {
			log.Info("IP ", ipPref, "overlaps with ", key.Get(1), " of ", key.Get(0))

			if intf != key.Get(0) {
				//IP overlap across different interface, reject
				log.Error("IP ", ipPref, " overlaps with ", key.Get(1), " of ", key.Get(0))

				errStr := "IP " + ipPref + " overlaps with IP " + key.Get(1) + " of Interface " + key.Get(0)
				err = tlerr.InvalidArgsError{Format: errStr}
				return err
			} else {
				//IP overlap on same interface, replace
				var entry dbEntry
				entry.op = opDelete

				log.Info("Entry ", key.Get(1), " on ", intf, " needs to be deleted")
				if app.ifIPTableMap[intf] == nil {
					app.ifIPTableMap[intf] = make(map[string]dbEntry)
				}
				app.ifIPTableMap[intf][key.Get(1)] = entry
			}
		}
	}

	//At this point, we need to add the entry to db
	{
		var entry dbEntry
		entry.op = opCreate

		m := make(map[string]string)
		m["NULL"] = "NULL"
		value := db.Value{Field: m}
		entry.entry = value
		if app.ifIPTableMap[intf] == nil {
			app.ifIPTableMap[intf] = make(map[string]dbEntry)
		}
		app.ifIPTableMap[intf][ipPref] = entry
	}
	return err
}

func (app *IntfApp) processCommon(d *db.DB) (SetResponse, error) {
	var err error
	var resp SetResponse

	log.Info("processCommon:intf:path =", app.path)
	log.Info("ProcessCommon: Target Type is " + reflect.TypeOf(*app.ygotTarget).Elem().Name())

	for key, entry := range app.ifTableMap {
		if entry.op == opUpdate {
			log.Info("Updating entry for ", key)
			err = d.SetEntry(app.portTs, db.Key{Comp: []string{key}}, entry.entry)
		}
	}

	for key, entry1 := range app.ifIPTableMap {
		ifEntry, err := d.GetEntry(app.intfIPTs, db.Key{Comp: []string{key}})
		if err != nil || !ifEntry.IsPopulated() {
			log.Infof("Interface Entry not present for Key:%s for IP config!", key)
			m := make(map[string]string)
			m["NULL"] = "NULL"
			err = d.CreateEntry(app.intfIPTs, db.Key{Comp: []string{key}}, db.Value{Field: m})
			if err != nil {
				return resp, err
			}
			log.Infof("Created Interface entry with Interface name : %s alone!", key)
		}
		for ip, entry := range entry1 {
			if entry.op == opCreate {
				log.Info("Creating entry for ", key, ":", ip)
				err = d.CreateEntry(app.intfIPTs, db.Key{Comp: []string{key, ip}}, entry.entry)
			} else if entry.op == opDelete {
				log.Info("Deleting entry for ", key, ":", ip)
				err = d.DeleteEntry(app.intfIPTs, db.Key{Comp: []string{key, ip}})
			}
		}
	}
	return resp, err
}
