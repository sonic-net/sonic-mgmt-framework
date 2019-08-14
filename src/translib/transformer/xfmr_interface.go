package transformer

import (
	"github.com/openconfig/ygot/ygot"
	"translib/db"
        log "github.com/golang/glog"
//	"translib/ocbinds"
)
/**
 * KeyXfmrYangToDb type is defined to use for conversion of Yang key to DB Key 
 * Transformer function definition.
 * Param: Database info, YgotRoot, Xpath
 * Return: Database keys to access db entry, error
 **/
type KeyXfmrYangToDb func (*db.DB, *ygot.GoStruct, string) (string, error)
/**
 * KeyXfmrDbToYang type is defined to use for conversion of DB key to Yang key
 * Transformer function definition.
 * Param: Database info, Database keys to access db entry
 * Return: multi dimensional map to hold the yang key attributes of complete xpath, error
 **/
type KeyXfmrDbToYang func (*db.DB, string) (map[string]map[string]string, error)

/**
 * FieldXfmrYangToDb type is defined to use for conversion of yang Field to DB field
 * Transformer function definition.
 * Param: Database info, YgotRoot, Xpath
 * Return: multi dimensional map to hold the DB data, error
 **/
type FieldXfmrYangToDb func (*db.DB, *ygot.GoStruct, string, interface {}) (map[string]string, error)
/**
 * FieldXfmrDbtoYang type is defined to use for conversion of DB field to Yang field
 * Transformer function definition.
 * Param: Database info, DB data in multidimensional map, output param YgotRoot
 * Return: error
 **/
type FieldXfmrDbtoYang func (*db.DB, map[string]map[string]db.Value, *ygot.GoStruct)  (error)

/**
 * SubTreeXfmrYangToDb type is defined to use for handling the yang subtree to DB
 * Transformer function definition.
 * Param: Database info, YgotRoot, Xpath
 * Return: multi dimensional map to hold the DB data, error
 **/
type SubTreeXfmrYangToDb func (*db.DB, *ygot.GoStruct, string) (map[string]map[string]db.Value, error)
/**
 * SubTreeXfmrDbToYang type is defined to use for handling the DB to Yang subtree
 * Transformer function definition.
 * Param : Database info, DB data in multidimensional map, output param YgotRoot
 * Return :  error
 **/
type SubTreeXfmrDbToYang func (*db.DB, map[string]map[string]db.Value, *ygot.GoStruct, string) (error)

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
