package transformer

import (
    "encoding/json"
    "os"
    "fmt"
    "errors"
    "io/ioutil"
    "cvl"
	log "github.com/golang/glog"
)

type xTablelist struct {
    TblInfo []xTblInfo `json:"tablelist"`
}

type xTblInfo struct {
    Name    string `json:"tablename"`
    Parent  string `json:"parent"`
}

type tblDirGph struct {
	node map[string]*gphNode
}

type gphNode struct {
    tableName string
	childTbl  []*gphNode
	visited   bool
}

func tblInfoAdd(tnMap map[string]*gphNode, tname, pname string) error {
	if _, ok := tnMap[tname]; !ok {
		node := new (gphNode)
		node.tableName = tname
		node.visited   = false
		tnMap[tname]   = node
	}
	node := tnMap[tname]
	if _, ok := tnMap[pname]; !ok {
		pnode := new (gphNode)
		pnode.tableName = pname
		pnode.visited   = false
		tnMap[pname]    = pnode
	}
	tnMap[pname].childTbl = append(tnMap[pname].childTbl, node)
	return nil
}

func childtblListGet (tnode *gphNode, ordTblList map[string][]string) (error, []string){
	var ctlist []string
	if len(tnode.childTbl) <= 0 {
		return nil, ctlist
	}

	if _, ok := ordTblList[tnode.tableName]; ok {
		return nil, ordTblList[tnode.tableName]
	}

	for _, ctnode := range tnode.childTbl {
		if ctnode.visited == false {
			ctnode.visited = true

			err, curTblList := childtblListGet(ctnode, ordTblList)
			if err != nil {
				ctlist = make([]string, 0)
				return err, ctlist
			}

			ordTblList[ctnode.tableName] = curTblList
			ctlist = append(ctlist, ctnode.tableName)
			ctlist = append(ctlist, curTblList...)
		} else {
			ctlist = append(ctlist, ctnode.tableName)
			ctlist = append(ctlist, ordTblList[ctnode.tableName]...)
		}
	}

	return nil, ctlist
}

func ordTblListCreate(ordTblList map[string][]string, tnMap map[string]*gphNode) {
	var tnodelist []*gphNode

	for _, tnode  := range tnMap {
		tnodelist = append(tnodelist, tnode)
	}

	for _, tnode := range tnodelist {
		if tnode != nil && tnode.visited == false {
			tnode.visited = true
			_, tlist := childtblListGet(tnode, ordTblList)
			ordTblList[tnode.tableName] = tlist
		}
	}
	return
}

 //sort transformer result table list based on dependenciesi(using CVL API) tables to be used for CRUD operations     
 //func sortPerTblDeps(tblList []string) ([]string, error) {
 func sortPerTblDeps(ordTblListMap map[string][]string) error {
	 var err error

	 errStr := fmt.Errorf("%v", "Failed to create cvl session")
	 cvSess, status := cvl.ValidationSessOpen()
	 if status != cvl.CVL_SUCCESS {
		 log.Errorf("CVL validation session creation failed(%v).", status)
		 err = fmt.Errorf("%v", errStr)
		 return errStr
	 }

	 for tname, tblList := range ordTblListMap {
		 sortedTblList, status := cvSess.SortDepTables(tblList)
		 if status != cvl.CVL_SUCCESS {
			 log.Warningf("Failure in cvlSess.SortDepTables: %v", status)
			 cvl.ValidationSessClose(cvSess)
			 err = fmt.Errorf("%v", errStr)
			 return err
		 }
		 //sort.Sort(sort.Reverse(sortedTblList[:])) 
		 for i := len(sortedTblList)/2-1; i >= 0; i-- {
			 r := len(sortedTblList)-1-i
			 sortedTblList[i], sortedTblList[r] = sortedTblList[r], sortedTblList[i]
		 }
		 ordTblListMap[tname] = sortedTblList
	 }
	 cvl.ValidationSessClose(cvSess)
	 return err
 }

func xlateJsonTblInfoLoad(ordTblListMap map[string][]string, jsonFileName string) error {
    var tlist xTablelist

    jsonFile, err := os.Open(jsonFileName)
    if err != nil {
		errStr := fmt.Sprintf("Error: Unable to open table list file(%v)", jsonFileName)
		return errors.New(errStr)
    }
    defer jsonFile.Close()

    xfmrLogInfoAll("Successfully Opened users.json\r\n")

    byteValue, _ := ioutil.ReadAll(jsonFile)

    json.Unmarshal(byteValue, &tlist)
	tnMap := make(map[string]*gphNode)

    for i := 0; i < len(tlist.TblInfo); i++ {
		err := tblInfoAdd(tnMap, tlist.TblInfo[i].Name, tlist.TblInfo[i].Parent)
		if err != nil {
			log.Errorf("Failed to add table dependency(tbl:%v, par:%v) into tablenode list.(%v)\r\n",
		               tlist.TblInfo[i].Name, tlist.TblInfo[i].Parent, err)
			break;
		}
    }

	if err == nil {
		ordTblListCreate(ordTblListMap, tnMap)
		for tname, tlist := range ordTblListMap {
			ordTblListMap[tname] = append([]string{tname}, tlist...)
		}
		if len(ordTblListMap) > 0 {
			sortPerTblDeps(ordTblListMap)
		}
	}
	return nil
}
