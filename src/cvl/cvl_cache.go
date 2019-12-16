////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or //
//  its subsidiaries.                                                         //
//                                                                            //
//  Licensed under the Apache License, Version 2.0 (the "License");           //
//  you may not use this file except in compliance with the License.          //
//  You may obtain a copy of the License at                                   //
//                                                                            //
//     http://www.apache.org/licenses/LICENSE-2.0                             //
//                                                                            //
//  Unless required by applicable law or agreed to in writing, software       //
//  distributed under the License is distributed on an "AS IS" BASIS,         //
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  //
//  See the License for the specific language governing permissions and       //
//  limitations under the License.                                            //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

package cvl
import (
	"fmt"
	"encoding/json"
	"github.com/go-redis/redis"
	. "cvl/internal/util"
	"cvl/internal/yparser"
	"time"
	"runtime"
	"github.com/antchfx/jsonquery"
	"github.com/antchfx/xmlquery"
)

func (c *CVL) addTableEntryToCache(tableName string, redisKey string) {
	if (tableName == "" || redisKey == "") {
		return
	}

	if (c.tmpDbCache[tableName] == nil) {
		c.tmpDbCache[tableName] = map[string]interface{}{redisKey: nil}
	} else {
		tblMap := c.tmpDbCache[tableName]
		tblMap.(map[string]interface{})[redisKey] =nil
		c.tmpDbCache[tableName] = tblMap
	}
}

//Add the data which are referring this key
func (c *CVL) updateDeleteDataToCache(tableName string, redisKey string) {
	if _, existing := c.tmpDbCache[tableName]; existing == false {
		return
	} else {
		tblMap := c.tmpDbCache[tableName]
		if _, existing := tblMap.(map[string]interface{})[redisKey]; existing == true {
			delete(tblMap.(map[string]interface{}), redisKey)
			c.tmpDbCache[tableName] = tblMap
		}
	}
}

// Fetch dependent data from validated data cache,
// Returns the data and flag to indicate that if requested data 
// is found in update request, the data should be merged with Redis data
func (c *CVL) fetchDataFromRequestCache(tableName string, key string) (d map[string]string, m bool) {
	defer func() {
		pd := &d
		pm := &m

		TRACE_LOG(TRACE_CACHE,
		"Returning data from request cache, data = %v, merge needed = %v",
		*pd, *pm)
	}()

	cfgDataArr := c.requestCache[tableName][key]
	if (cfgDataArr != nil) {
		for _, cfgReqData := range cfgDataArr {
			//Delete request doesn't have depedent data
			if (cfgReqData.reqData.VOp == OP_CREATE) {
				return cfgReqData.reqData.Data, false
			} else	if (cfgReqData.reqData.VOp == OP_UPDATE) {
				return cfgReqData.reqData.Data, true
			}
		}
	}

	return nil, false
}

//Fetch given table entries using pipeline
func (c *CVL) fetchTableDataToTmpCache(tableName string, dbKeys map[string]interface{}) int {

	TRACE_LOG(TRACE_CACHE, "\n%v, Entered fetchTableDataToTmpCache", time.Now())

	totalCount := len(dbKeys)
	if (totalCount == 0) {
		//No entry to be fetched
		return 0
	}

	entryFetched := 0
	bulkCount := 0
	bulkKeys := []string{}
	for dbKey, val := range dbKeys { //for all keys

		 if (val != nil) { //skip entry already fetched
                        mapTable := c.tmpDbCache[tableName]
                        delete(mapTable.(map[string]interface{}), dbKey) //delete entry already fetched
                        totalCount = totalCount - 1
                        if(bulkCount != totalCount) {
                                //If some entries are remaining go back to 'for' loop
                                continue
                        }
                } else {
                        //Accumulate entries to be fetched
                        bulkKeys = append(bulkKeys, dbKey)
                        bulkCount = bulkCount + 1
                }

                if(bulkCount != totalCount) && ((bulkCount % MAX_BULK_ENTRIES_IN_PIPELINE) != 0) {
                        //If some entries are remaining and bulk bucket is not filled,
                        //go back to 'for' loop
                        continue
                }

		mCmd := map[string]*redis.StringStringMapCmd{}

		pipe := redisClient.Pipeline()

		for _, dbKey := range bulkKeys {

			redisKey := tableName + modelInfo.tableInfo[tableName].redisKeyDelim + dbKey
			//Check in validated cache first and add as dependent data
			if entry, mergeNeeded := c.fetchDataFromRequestCache(tableName, dbKey); (entry != nil) {
				entryFetched = entryFetched + 1
				//Entry found in validated cache, so skip fetching from Redis
				//if merging is not required with Redis DB
				if (mergeNeeded == false) {
					fieldMap := c.checkFieldMap(&entry)
					c.tmpDbCache[tableName].(map[string]interface{})[dbKey] = fieldMap
					continue
				}
				c.tmpDbCache[tableName].(map[string]interface{})[dbKey] = entry
			}

			//Otherwise fetch it from Redis
			mCmd[dbKey] = pipe.HGetAll(redisKey) //write into pipeline
			if mCmd[dbKey] == nil {
				CVL_LOG(ERROR, "Failed pipe.HGetAll('%s')", redisKey)
			}
		}

		_, err := pipe.Exec()
		if err != nil {
			CVL_LOG(ERROR, "Failed to fetch details for table %s", tableName)
			return 0
		}
		pipe.Close()
		bulkKeys = nil

		mapTable := c.tmpDbCache[tableName]

		for key, val := range mCmd {
			res, err := val.Result()

			if (mapTable == nil) {
				break
			}

			if (err != nil || len(res) == 0) {
				//no data found, don't keep blank entry
				delete(mapTable.(map[string]interface{}), key)
				continue
			}
			//exclude table name and delim
			keyOnly := key

			if (len(mapTable.(map[string]interface{})) > 0) && (mapTable.(map[string]interface{})[keyOnly] != nil) {
				tmpFieldMap := (mapTable.(map[string]interface{})[keyOnly]).(map[string]string)
				//merge with validated cache data
				mergeMap(res, tmpFieldMap)
				fieldMap := c.checkFieldMap(&res)
				mapTable.(map[string]interface{})[keyOnly] = fieldMap
			} else {
				fieldMap := c.checkFieldMap(&res)
				mapTable.(map[string]interface{})[keyOnly] = fieldMap
			}

			entryFetched = entryFetched + 1
		}

		runtime.Gosched()
	}

	TRACE_LOG(TRACE_CACHE,"\n%v, Exiting fetchTableDataToTmpCache", time.Now())

	return entryFetched
}

//populate redis data to cache
func (c *CVL) fetchDataToTmpCache() *yparser.YParserNode {
	TRACE_LOG(TRACE_CACHE, "\n%v, Entered fetchToTmpCache", time.Now())

	entryToFetch := 0
	var root *yparser.YParserNode = nil
	var errObj yparser.YParserError

	for entryToFetch = 1; entryToFetch > 0; { //Force to enter the loop for first time
		//Repeat until all entries are fetched 
		entryToFetch = 0
		for tableName, dbKeys := range c.tmpDbCache { //for each table
			entryToFetch = entryToFetch + c.fetchTableDataToTmpCache(tableName, dbKeys.(map[string]interface{}))
		} //for each table

		//If no table entry delete the table  itself
		for tableName, dbKeys := range c.tmpDbCache { //for each table
			if (len(dbKeys.(map[string]interface{}))  == 0) {
				 delete(c.tmpDbCache, tableName)
				 continue
			}
		}

		if (entryToFetch == 0) {
			//No more entry to fetch
			break
		}

		if (Tracing == true) {
			jsonDataBytes, _ := json.Marshal(c.tmpDbCache)
			jsonData := string(jsonDataBytes)
			TRACE_LOG(TRACE_CACHE, "Top Node=%v\n", jsonData)
		}

		data, err := jsonquery.ParseJsonMap(&c.tmpDbCache)

		if (err != nil) {
			return nil
		}

		//Build yang tree for each table and cache it
		for jsonNode := data.FirstChild; jsonNode != nil; jsonNode=jsonNode.NextSibling {
			TRACE_LOG(TRACE_CACHE, "Top Node=%v\n", jsonNode.Data)
			//Visit each top level list in a loop for creating table data
			topNode, _ := c.generateTableData(true, jsonNode)
			if (root == nil) {
				root = topNode
			} else {
				if root, errObj = c.yp.MergeSubtree(root, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
					return nil
				}
			}

			//Generate YANG data for Yang Validator
			topYangNode, cvlYErrObj := c.generateYangListData(jsonNode, true)
			if  topYangNode == nil {
				cvlYErrObj.ErrCode = CVL_SYNTAX_ERROR
				CVL_LOG(ERROR, "Unable to translate cache data to YANG format")
				return nil
			}

			//Create a full document and merge with main YANG data
			doc := &xmlquery.Node{Type: xmlquery.DocumentNode}
			doc.FirstChild = topYangNode
			doc.LastChild = topYangNode
			topYangNode.Parent = doc

			//TRACE_LOG(TRACE_CACHE, "Before cache merge = %s", c.yv.root.OutputXML(false))
			if c.mergeYangData(c.yv.root, doc) != CVL_SUCCESS {
				CVL_LOG(ERROR, "Unable to merge translated YANG data while " +
				"translating from cahce data to YANG format")
				cvlYErrObj.ErrCode = CVL_SYNTAX_ERROR
				return nil
			}
			//TRACE_LOG(TRACE_CACHE, "After cache merge = %s", c.yv.root.OutputXML(false))
		}
	} // until all dependent data is fetched

	if root != nil && Tracing == true {
		dumpStr := c.yp.NodeDump(root)
		TRACE_LOG(TRACE_CACHE, "Dependent Data = %v\n", dumpStr)
	}

	TRACE_LOG(TRACE_CACHE, "\n%v, Exiting fetchToTmpCache", time.Now())
	return root
}


func (c *CVL) clearTmpDbCache() {
	for key, _ := range c.tmpDbCache {
		delete(c.tmpDbCache, key)
	}
}

//Get table entry from cache for redis key
func dbCacheEntryGet(tableName, key string) (*yparser.YParserNode, CVLRetCode) {
	//First check if the table is cached
	topNode, _ := dbCacheGet(tableName)


	if (topNode != nil) {
		//Convert to Yang keys
		keyValuePair := getRedisToYangKeys(tableName, key)

		//Find if the entry is cached
		keyCompStr := ""
		for _, keyValItem := range keyValuePair {
			keyCompStr = keyCompStr + fmt.Sprintf("[%s='%s']",
			keyValItem.key, keyValItem.values[0])
		}

		entryNode := yparser.FindNode(topNode, fmt.Sprintf("//%s:%s/%s%s",
		modelInfo.tableInfo[tableName].modelName,
		modelInfo.tableInfo[tableName].modelName,
		tableName, keyCompStr))

		if (entryNode != nil) {
			return entryNode, CVL_SUCCESS
		}
	}

	return nil, CVL_ERROR
}

//Get the data from global cache
func dbCacheGet(tableName string) (*yparser.YParserNode, CVLRetCode) {

	TRACE_LOG(TRACE_CACHE, "Updating global cache for table %s", tableName)
	dbCacheTmp, existing := cvg.db[tableName]

	if (existing == false) {
		return  nil, CVL_FAILURE //not even empty cache present
	}

	if (dbCacheTmp.root != nil) {
		if (dbCacheTmp.expiry != 0) {
			//If cache is destroyable (i.e. expiry != 0), check if it has already expired.
			//If not expired update the time stamp
			if (time.Now().After(dbCacheTmp.startTime.Add(time.Second * time.Duration(dbCacheTmp.expiry)))) {
				//Cache expired, clear the cache
				dbCacheClear(tableName)

				return nil, CVL_ERROR
			}

			//Since the cache is used actively, update the timestamp
			dbCacheTmp.startTime = time.Now()
			cvg.db[tableName] = dbCacheTmp
		}

		return dbCacheTmp.root, CVL_SUCCESS
	} else {
		return  nil, CVL_SUCCESS // return success for no entry in Redis db and hencec empty cache
	}
}

//Get the table data from redis and cache it in yang node format
//expiry =0 never expire the cache
func dbCacheSet(update bool, tableName string, expiry uint16) CVLRetCode {

	cvg.mutex.Lock()

	//Get the data from redis and save it
	tableKeys, err:= redisClient.Keys(tableName +
	modelInfo.tableInfo[tableName].redisKeyDelim + "*").Result()

	if (err != nil) {
		cvg.mutex.Unlock()
		return CVL_FAILURE
	}

	TRACE_LOG(TRACE_CACHE, "Building global cache for table %s", tableName)

	tablePrefixLen := len(tableName + modelInfo.tableInfo[tableName].redisKeyDelim)
	for _, tableKey := range tableKeys {
		tableKey = tableKey[tablePrefixLen:] //remove table prefix
		if (cvg.cv.tmpDbCache[tableName] == nil) {
			cvg.cv.tmpDbCache[tableName] = map[string]interface{}{tableKey: nil}
		} else {
			tblMap := cvg.cv.tmpDbCache[tableName]
			tblMap.(map[string]interface{})[tableKey] =nil
			cvg.cv.tmpDbCache[tableName] = tblMap
		}
	}

	cvg.db[tableName] = dbCachedData{startTime:time.Now(), expiry: expiry,
	root: cvg.cv.fetchDataToTmpCache()}

	if (Tracing == true) {
		TRACE_LOG(TRACE_CACHE, "Cached Data = %v\n", cvg.cv.yp.NodeDump(cvg.db[tableName].root))
	}

	cvg.mutex.Unlock()

	//install keyspace notification for updating the cache
	if (update == false) {
		installDbChgNotif()
	}


	return CVL_SUCCESS
}

//Receive all updates for all tables on a single channel
func installDbChgNotif() {
	if (len(cvg.db) > 1) { //notif running for at least one table added previously
		cvg.stopChan <- 1 //stop active notification 
	}

	subList := make([]string, 0)
	for tableName, _ := range cvg.db {
		subList = append(subList,
		fmt.Sprintf("__keyspace@%d__:%s%s*", modelInfo.tableInfo[tableName].dbNum,
		tableName, modelInfo.tableInfo[tableName].redisKeyDelim))

	}

	//Listen on multiple channels
	cvg.pubsub = redisClient.PSubscribe(subList...)

	go func() {
		keySpacePrefixLen := len("__keyspace@4__:")

		notifCh := cvg.pubsub.Channel()
		for {
			select  {
			case <-cvg.stopChan:
				//stop this routine
				return
			case msg:= <-notifCh:
				//Handle update
				tbl, key := splitRedisKey(msg.Channel[keySpacePrefixLen:])
				if (tbl != "" && key != "") {
					dbCacheUpdate(tbl, key, msg.Payload)
				}
			}
		}
	}()
}

func dbCacheUpdate(tableName, key, op string) CVLRetCode {
	TRACE_LOG(TRACE_CACHE, "Updating global cache for table %s with key %s", tableName, key)

	//Find the node
	//Delete the entry in yang tree 

	cvg.mutex.Lock()

	node, _:= dbCacheEntryGet(tableName, key)
	if (node != nil) {
		//unlink and free the node
		cvg.cv.yp.FreeNode(node)
	}

	//Clear json map cache if any
	cvg.cv.clearTmpDbCache()

	tableKeys := []string {key}
	switch op {
	case "hset", "hmset", "hdel":
		//Get the entry from DB
		for _, tableKey := range tableKeys {
			cvg.cv.tmpDbCache[tableName] = map[string]interface{}{tableKey: nil}
		}

		//Get the translated Yang tree
		topNode := cvg.cv.fetchDataToTmpCache()

		//Merge the subtree with existing yang tree
		var errObj yparser.YParserError
		if (cvg.db[tableName].root != nil) {
			if topNode, errObj = cvg.cv.yp.MergeSubtree(cvg.db[tableName].root, topNode); errObj.ErrCode != yparser.YP_SUCCESS {
				cvg.mutex.Unlock()
				return CVL_ERROR
			}
		}

		//Update DB map
		db := cvg.db[tableName]
		db.root = topNode
		cvg.db[tableName] = db

	case "del":
		//NOP, already deleted the entry
	}

	cvg.mutex.Unlock()

	return CVL_SUCCESS
}

//Clear cache data for given table
func dbCacheClear(tableName string) CVLRetCode {
	cvg.cv.yp.FreeNode(cvg.db[tableName].root)
	delete(cvg.db, tableName)

	TRACE_LOG(TRACE_CACHE, "Clearing global cache for table %s", tableName)

	return CVL_SUCCESS
}

