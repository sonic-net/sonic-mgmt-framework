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

/*
Package translib defines the functions to be used to authorize 

an incoming user. It also includes caching of the UserDB data

needed to authorize the user.

*/

package translib

import (
	//"strconv"
	//log "github.com/golang/glog"
)

//TODO:define maps for storing the UserDB cache

func init() {
	//TODO:Allocate the maps and populate them here
}

func isAuthorizedForSet(req SetRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	for _, r := range req.User.Roles {
        if r == "admin" {
            return true
        }
    }
	return false
}
func isAuthorizedForBulk(req BulkRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	for _, r := range req.User.Roles {
        if r == "admin" {
            return true
        }
    }
	return false
}

func isAuthorizedForGet(req GetRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	return  true
}
func isAuthorizedForSubscribe(req SubscribeRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	return  true
}
func isAuthorizedForIsSubscribe(req IsSubscribeRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	return  true
}

func isAuthorizedForAction(req ActionRequest) bool {
	if !req.AuthEnabled {
		return true
	}
	return  true
}
