/*
Copyright 2019 Broadcom. All rights reserved.
The term “Broadcom” refers to Broadcom Inc. and/or its subsidiaries.
*/

package translib

import (
	"errors"
	"reflect"
	"strings"

	log "github.com/golang/glog"
	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"

	"translib/ocbinds"
	"translib/tlerr"
)

const (
	GET = 1 + iota
	CREATE
	REPLACE
	UPDATE
	DELETE
)

var ygSchema *ytypes.Schema

func init() {
	initSchema()
}

func initSchema() {
	log.Flush()
	var err error
	if ygSchema, err = ocbinds.Schema(); err != nil {
		panic("Error in getting the schema: " + err.Error())
	}
}

type requestBinder struct {
	uri             *string
	payload         *[]byte
	opcode          int
	appRootNodeType *reflect.Type
	pathTmp         *gnmi.Path
}

func getRequestBinder(uri *string, payload *[]byte, opcode int, appRootNodeType *reflect.Type) *requestBinder {
	return &requestBinder{uri, payload, opcode, appRootNodeType, nil}
}

func (binder *requestBinder) unMarshallPayload(workObj *interface{}) error {
	targetObj, ok := (*workObj).(ygot.GoStruct)
	if ok == false {
		err := errors.New("Error in casting the target object")
		log.Error(err)
		return tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
	}

	if len(*binder.payload) == 0 {
		err := errors.New("Request payload is empty")
		log.Error(err)
		return tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
	}

	err := ocbinds.Unmarshal(*binder.payload, targetObj)
	if err != nil {
		log.Error(err)
		return tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
	}

	return nil
}

func (binder *requestBinder) validateRequest(deviceObj *ocbinds.Device) error {
	if binder.pathTmp == nil || len(binder.pathTmp.Elem) == 0 {
		if binder.opcode == UPDATE || binder.opcode == REPLACE {
			log.Info("validateRequest: path is base node")
			devObjTmp, ok := (reflect.ValueOf(*deviceObj).Interface()).(ygot.ValidatedGoStruct)
			if ok == true {
				err := devObjTmp.Validate(&ytypes.LeafrefOptions{IgnoreMissingData: true})
				if err != nil {
					return err
				}
			} else {
				return errors.New("Invalid base Object in the binding: Not able to cast to type ValidatedGoStruct")
			}
			return nil
		} else {
			return errors.New("Path is empty")
		}
	}

	path, err := ygot.StringToPath(binder.pathTmp.Elem[0].Name, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		return err
	} else {
		baseTreeNode, err := ytypes.GetNode(ygSchema.RootSchema(), deviceObj, path)
		if err != nil {
			return err
		} else if len(baseTreeNode) == 0 {
			return errors.New("Invalid base URI node")
		} else {
			basePathObj, ok := (baseTreeNode[0].Data).(ygot.ValidatedGoStruct)
			if ok == true {
				err := basePathObj.Validate(&ytypes.LeafrefOptions{IgnoreMissingData: true})
				if err != nil {
					return err
				}
			} else {
				return errors.New("Invalid Object in the binding: Not able to cast to type ValidatedGoStruct")
			}
		}
	}

	return nil
}

func (binder *requestBinder) unMarshall() (*ygot.GoStruct, *interface{}, error) {
	var deviceObj ocbinds.Device = ocbinds.Device{}

	workObj, err := binder.unMarshallUri(&deviceObj)
	if err != nil {
		log.Error("Error in creating the target object : ", err)
		return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 404, ErrorStr: err}
	}

	rootIntf := reflect.ValueOf(&deviceObj).Interface()
	ygotObj := rootIntf.(ygot.GoStruct)
	var ygotRootObj *ygot.GoStruct = &ygotObj

	if binder.opcode == GET || binder.opcode == DELETE {
		return ygotRootObj, workObj, nil
	}

	switch binder.opcode {
	case CREATE:
		if reflect.ValueOf(*workObj).Kind() == reflect.Map {
			return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: errors.New("URI doesn't have keys in the request")}
		} else {
			err = binder.unMarshallPayload(workObj)
			if err != nil {
				return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
			}
		}

	case UPDATE, REPLACE:
		var tmpTargetNode *interface{}
		if binder.pathTmp != nil {
			treeNodeList, err2 := ytypes.GetNode(ygSchema.RootSchema(), &deviceObj, binder.pathTmp)
			if err2 != nil {
				return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err2}
			}

			if len(treeNodeList) == 0 {
				return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: errors.New("Invalid URI")}
			}

			tmpTargetNode = &(treeNodeList[0].Data)
		} else {
			tmpTargetNode = workObj
		}

		err = binder.unMarshallPayload(tmpTargetNode)
		if err != nil {
			return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
		}

	default:
		if binder.opcode != GET && binder.opcode != DELETE {
			return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: errors.New("Unknown HTTP METHOD in the request")}
		}
	}

	if err = binder.validateRequest(&deviceObj); err != nil {
		return nil, nil, tlerr.TranslibSyntaxValidationError{StatusCode: 400, ErrorStr: err}
	}

	return ygotRootObj, workObj, nil
}

func (binder *requestBinder) getUriPath() (*gnmi.Path, error) {
	var path *gnmi.Path
	var err error

	path, err = ygot.StringToPath(*binder.uri, ygot.StructuredPath, ygot.StringSlicePath)
	if err != nil {
		log.Error("Error in uri to path conversion: ", err)
		return nil, err
	}

	return path, nil
}

func (binder *requestBinder) unMarshallUri(deviceObj *ocbinds.Device) (*interface{}, error) {
	if len(*binder.uri) == 0 {
		errMsg := errors.New("Error: URI is empty")
		log.Error(errMsg)
		return nil, errMsg
	}

	path, err := binder.getUriPath()
	if err != nil {
		return nil, err
	} else {
		binder.pathTmp = path
	}

	for _, p := range path.Elem {
		pathSlice := strings.Split(p.Name, ":")
		p.Name = pathSlice[len(pathSlice)-1]
	}

	ygNode, ygEntry, errYg := ytypes.GetOrCreateNode(ygSchema.RootSchema(), deviceObj, path)

	if errYg != nil {
		log.Error("Error in creating the target object: ", errYg)
		return nil, errYg
	}

	switch binder.opcode {
	case UPDATE, REPLACE:
		if ygEntry.IsList() == false || reflect.ValueOf(ygNode).Kind() == reflect.Map {
			var pathList []*gnmi.PathElem = path.Elem

			gpath := &gnmi.Path{}

			for i := 0; i < (len(pathList) - 1); i++ {
				log.Info("pathList[i] ", pathList[i])
				gpath.Elem = append(gpath.Elem, pathList[i])
			}

			log.Info("modified path is: ", gpath)

			binder.pathTmp = gpath
		} else {
			log.Info("ygot type of the node is Map")
		}
	}

	if (binder.opcode == GET || binder.opcode == DELETE) && (ygEntry.IsLeaf() == false && ygEntry.IsLeafList() == false) {
		if err = binder.validateRequest(deviceObj); err != nil {
			return nil, err
		}
	}

	return &ygNode, nil
}
