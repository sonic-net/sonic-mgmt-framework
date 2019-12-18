////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Dell, Inc.                                                 //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//  http://www.apache.org/licenses/LICENSE-2.0                                //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package transformer

import (
	"github.com/openconfig/ygot/ygot"
	"translib/db"
    log "github.com/golang/glog"
)

type RedisDbMap = map[db.DBNum]map[string]map[string]db.Value

type XfmrParams struct {
	d *db.DB
	dbs [db.MaxDB]*db.DB
	curDb db.DBNum
	ygRoot *ygot.GoStruct
	uri string
	requestUri string //original uri using which a curl/NBI request is made
	oper int
	key string
	dbDataMap *map[db.DBNum]map[string]map[string]db.Value
	subOpDataMap map[int]*RedisDbMap // used to add an in-flight data with a sub-op
	param interface{}
	txCache interface{}
	skipOrdTblChk *bool
}

/**
 * KeyXfmrYangToDb type is defined to use for conversion of Yang key to DB Key 
 * Transformer function definition.
 * Param: XfmrParams structure having Database info, YgotRoot, operation, Xpath
 * Return: Database keys to access db entry, error
 **/
type KeyXfmrYangToDb func (inParams XfmrParams) (string, error)
/**
 * KeyXfmrDbToYang type is defined to use for conversion of DB key to Yang key
 * Transformer function definition.
 * Param: XfmrParams structure having Database info, operation, Database keys to access db entry
 * Return: multi dimensional map to hold the yang key attributes of complete xpath, error
 **/
type KeyXfmrDbToYang func (inParams XfmrParams) (map[string]interface{}, error)

/**
 * FieldXfmrYangToDb type is defined to use for conversion of yang Field to DB field
 * Transformer function definition.
 * Param: Database info, YgotRoot, operation, Xpath
 * Return: multi dimensional map to hold the DB data, error
 **/
type FieldXfmrYangToDb func (inParams XfmrParams) (map[string]string, error)
/**
 * FieldXfmrDbtoYang type is defined to use for conversion of DB field to Yang field
 * Transformer function definition.
 * Param: XfmrParams structure having Database info, operation, DB data in multidimensional map, output param YgotRoot
 * Return: error
 **/
type FieldXfmrDbtoYang func (inParams XfmrParams)  (map[string]interface{}, error)

/**
 * SubTreeXfmrYangToDb type is defined to use for handling the yang subtree to DB
 * Transformer function definition.
 * Param: XfmrParams structure having Database info, YgotRoot, operation, Xpath
 * Return: multi dimensional map to hold the DB data, error
 **/
type SubTreeXfmrYangToDb func (inParams XfmrParams) (map[string]map[string]db.Value, error)
/**
 * SubTreeXfmrDbToYang type is defined to use for handling the DB to Yang subtree
 * Transformer function definition.
 * Param : XfmrParams structure having Database pointers, current db, operation, DB data in multidimensional map, output param YgotRoot, uri
 * Return :  error
 **/
type SubTreeXfmrDbToYang func (inParams XfmrParams) (error)
/**
 * ValidateCallpoint is used to validate a YANG node during data translation back to YANG as a response to GET
 * Param : XfmrParams structure having Database pointers, current db, operation, DB data in multidimensional map, output param YgotRoot, uri
 * Return :  bool
 **/
type ValidateCallpoint func (inParams XfmrParams) (bool)
/**
 * RpcCallpoint is used to invoke a callback for action
 * Param : []byte input payload, dbi indices
 * Return :  []byte output payload, error
 **/
type RpcCallpoint func (body []byte, dbs [db.MaxDB]*db.DB) ([]byte, error)
/**
 * PostXfmrFunc type is defined to use for handling any default handling operations required as part of the CREATE
 * Transformer function definition.
 * Param: XfmrParams structure having database pointers, current db, operation, DB data in multidimensional map, YgotRoot, uri
 * Return: Multi dimensional map to hold the DB data Map (tblName, key and Fields), error
 **/
type PostXfmrFunc func (inParams XfmrParams) (map[string]map[string]db.Value, error)


/**
 * TableXfmrFunc type is defined to use for table transformer function for dynamic derviation of redis table.
 * Param: XfmrParams structure having database pointers, current db, operation, DB data in multidimensional map, YgotRoot, uri
 * Return: List of table names, error
 **/
type TableXfmrFunc func (inParams XfmrParams) ([]string, error)


/**
 * Xfmr validation interface for validating the callback registration of app modules 
 * transformer methods.
 **/
type XfmrInterface interface {
    xfmrInterfaceValiidate()
}

func (KeyXfmrYangToDb) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for KeyXfmrYangToDb")
}
func (KeyXfmrDbToYang) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for KeyXfmrDbToYang")
}
func (FieldXfmrYangToDb) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for FieldXfmrYangToDb")
}
func (FieldXfmrDbtoYang) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for FieldXfmrDbtoYang")
}
func (SubTreeXfmrYangToDb) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for SubTreeXfmrYangToDb")
}
func (SubTreeXfmrDbToYang) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for SubTreeXfmrDbToYang")
}
func (TableXfmrFunc) xfmrInterfaceValiidate () {
    log.Info("xfmrInterfaceValiidate for TableXfmrFunc")
}
