package transformer

import (
    "errors"
    "strings"
    "encoding/json"
    "translib/ocbinds"
    "translib/db"
    "os/exec"
    log "github.com/golang/glog"
)

func getBgpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp, string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
    var err error

    if len(pathInfo.Vars) <  3 {
        return nil, "", errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return nil, "", errors.New("vrf name is missing")
    }
    if strings.Contains(bgpId,"BGP") == false {
        return nil, "", errors.New("BGP ID is missing")
    }
    if len(protoName) == 0 {
        return nil, "", errors.New("Protocol Name is missing")
    }

	deviceObj := (*inParams.ygRoot).(*ocbinds.Device)
    netInstsObj := deviceObj.NetworkInstances

    if netInstsObj.NetworkInstance == nil {
        return nil, "", errors.New("Network-instances container missing")
    }

    netInstObj := netInstsObj.NetworkInstance[niName]
    if netInstObj == nil {
        return nil, "", errors.New("Network-instance obj missing")
    }

    if netInstObj.Protocols == nil || len(netInstObj.Protocols.Protocol) == 0 {
        return nil, "", errors.New("Network-instance protocols-container missing or protocol-list empty")
    }

    var protoKey ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Key
    protoKey.Identifier = ocbinds.OpenconfigPolicyTypes_INSTALL_PROTOCOL_TYPE_BGP
    protoKey.Name = protoName
    protoInstObj := netInstObj.Protocols.Protocol[protoKey]
    if protoInstObj == nil {
        return nil, "", errors.New("Network-instance BGP-Protocol obj missing")
    }
    return protoInstObj.Bgp, niName, err
}

func exec_vtysh_cmd (vtysh_cmd string) (map[string]interface{}, error) {
    var err error
    oper_err := errors.New("Opertational error")

    log.Infof("Going to execute vtysh cmd ==> \"%s\"", vtysh_cmd)

    cmd := exec.Command("/usr/bin/docker", "exec", "bgp", "vtysh", "-c", vtysh_cmd)
    out_stream, err := cmd.StdoutPipe()
    if err != nil {
        log.Errorf("Can't get stdout pipe: %s\n", err)
        return nil, oper_err
    }

    err = cmd.Start()
    if err != nil {
        log.Errorf("cmd.Start() failed with %s\n", err)
        return nil, oper_err
    }

    var outputJson map[string]interface{}
    err = json.NewDecoder(out_stream).Decode(&outputJson)
    if err != nil {
        log.Errorf("Not able to decode vtysh json output: %s\n", err)
        return nil, oper_err
    }

    err = cmd.Wait()
    if err != nil {
        log.Errorf("Command execution completion failed with %s\n", err)
        return nil, oper_err
    }

    log.Infof("Successfully executed vtysh-cmd ==> \"%s\"", vtysh_cmd)

    if outputJson == nil {
        log.Errorf("VTYSH output empty !!!")
        return nil, oper_err
    }

    return outputJson, err
}


func fake_rib_json_output_blob (rib_type string) ([]byte) {
    outJsonBlob := ``

    switch rib_type {
        case "ipv4-loc-rib":
            outJsonBlob = `{
                "vrfId": 0,
                "vrfName": "default",
                "tableVersion": 2,
                "routerId": "10.20.30.40",
                "defaultLocPrf": 100,
                "localAS": 400,
                "routes": {
                    "70.80.90.0/24": {
                        "prefix":"70.80.90.0\/24",
                        "advertisedTo":{
                            "10.10.10.1":{
                            },
                            "11.1.1.1":{
                            }
                        },
                        "paths":[
                        {
                            "pathId":0,
                            "aspath":{
                                "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                "segments":[
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        200
                                    ]
                                },
                                {
                                    "type":"as-set",
                                    "list":[
                                        111,
                                    333,
                                    555,
                                    777
                                    ]
                                },
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        222,
                                    444,
                                    666,
                                    888
                                    ]
                                },
                                {
                                    "type":"as-set",
                                    "list":[
                                        100000,
                                    200000,
                                    300000,
                                    400000
                                    ]
                                }
                                ],
                                "length":7
                            },
                            "aggregatorAs":5000,
                            "aggregatorId":"10.20.30.40",
                            "originatorId":"0.0.0.0",
                            "localPref":200,
                            "origin":"IGP",
                            "med":37,
                            "metric":37,
                            "valid":true,
                            "atomicAggregate":true,
                            "bestpath":{
                                "overall":true
                            },
                            "community":{
                                "string":"1000:2000 3000:4000",
                                "list":[
                                    "1000:2000",
                                "3000:4000"
                                ]
                            },
                            "extendedCommunity":{
                                "string":"RT:10:3369254580 RT:20:2358681770"
                            },
                            "clusterList":{
                                "list":[
                                    "10.20.30.40",
                                "50.60.70.80",
                                "90.100.110.120"
                                ]
                            },
                            "lastUpdate":{
                                "epoch":1495225039,
                                "string":"Fri May 19 20:17:19 2017\n"
                            },
                            "nexthops":[
                            {
                                "ip":"10.10.10.1",
                                "afi":"ipv4",
                                "metric":0,
                                "accessible":true,
                                "used":true
                            }
                            ],
                            "peer":{
                                "peerId":"10.10.10.1",
                                "routerId":"25.98.0.1",
                                "type":"external"
                            }
                        }
                        ]
                    }
                    ,
                        "70.80.91.0/24": {
                            "prefix":"70.80.91.0\/24",
                            "advertisedTo":{
                                "10.10.10.1":{
                                },
                                "11.1.1.1":{
                                }
                            },
                            "paths":[
                            {
                                "pathId":0,
                                "aspath":{
                                    "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                    "segments":[
                                    {
                                        "type":"as-sequence",
                                        "list":[
                                            200
                                        ]
                                    },
                                    {
                                        "type":"as-set",
                                        "list":[
                                            111,
                                        333,
                                        555,
                                        777
                                        ]
                                    },
                                    {
                                        "type":"as-sequence",
                                        "list":[
                                            222,
                                        444,
                                        666,
                                        888
                                        ]
                                    },
                                    {
                                        "type":"as-set",
                                        "list":[
                                            100000,
                                        200000,
                                        300000,
                                        400000
                                        ]
                                    }
                                    ],
                                    "length":7
                                },
                                "aggregatorAs":5000,
                                "aggregatorId":"10.20.30.40",
                                "originatorId":"0.0.0.0",
                                "localPref":200,
                                "origin":"IGP",
                                "med":37,
                                "metric":37,
                                "valid":true,
                                "atomicAggregate":true,
                                "bestpath":{
                                    "overall":true
                                },
                                "community":{
                                    "string":"1000:2000 3000:4000",
                                    "list":[
                                        "1000:2000",
                                    "3000:4000"
                                    ]
                                },
                                "extendedCommunity":{
                                    "string":"RT:10:3369254580 RT:20:2358681770"
                                },
                                "clusterList":{
                                    "list":[
                                        "11.21.31.41",
                                    "51.61.71.81",
                                    "90.100.110.120"
                                    ]
                                },
                                "lastUpdate":{
                                    "epoch":1495225039,
                                    "string":"Fri May 19 20:17:19 2017\n"
                                },
                                "nexthops":[
                                {
                                    "ip":"10.10.10.1",
                                    "afi":"ipv4",
                                    "metric":0,
                                    "accessible":true,
                                    "used":true
                                }
                                ],
                                "peer":{
                                    "peerId":"10.10.10.1",
                                    "routerId":"25.98.0.1",
                                    "type":"external"
                                }
                            }
                            ]
                        }
                }
            }`

        case "ipv6-loc-rib" :
            outJsonBlob = `{                                    
                "vrfId": 0,                         
                "vrfName": "default",               
                "tableVersion": 6,                  
                "routerId": "11.1.1.4",             
                "defaultLocPrf": 100,               
                "localAS": 400,                     
                "routes": {                         
                    "3001::/64": {                       
                        "prefix":"3001::\/64",             
                        "advertisedTo":{                   
                            "2001::1":{                      
                            }                                
                        },                                 
                        "paths":[                          
                        {                                
                            "pathId":0,                    
                            "aspath":{                     
                                "string":"100 {10000,20000} 30000 40000",
                                "segments":[                             
                                {                                      
                                    "type":"as-sequence",                
                                    "list":[                             
                                        100                                
                                    ]                                    
                                },                                     
                                {                                      
                                    "type":"as-set",                     
                                    "list":[                             
                                        10000,                             
                                    20000                              
                                    ]                                    
                                },                                     
                                {                                      
                                    "type":"as-sequence",                
                                    "list":[                             
                                        30000,                             
                                    40000                              
                                    ]                                    
                                }                                      
                                ],                                       
                                "length":4                               
                            },                                         
                            "origin":"IGP",                            
                            "valid":true,                              
                            "bestpath":{                               
                                "overall":true                           
                            },                                         
                            "lastUpdate":{                             
                                "epoch":1495309393,                      
                                "string":"Sat May 20 19:43:13 2017\n"    
                            },                                         
                            "nexthops":[                               
                            {                                        
                                "ip":"2001::1",                        
                                "afi":"ipv6",                          
                                "scope":"global",                      
                                "metric":0,                            
                                "accessible":true                      
                            },                                       
                            {                                        
                                "ip":"fe80::200:1ff:fe4a:417c",        
                                "afi":"ipv6",                          
                                "scope":"link-local",                  
                                "accessible":true,                     
                                "used":true                            
                            }                                        
                            ],                                         
                            "peer":{                                   
                                "peerId":"2001::1",                      
                                "routerId":"25.98.0.2",                  
                                "type":"external"                        
                            }                                          
                        }                                            
                        ]                                              
                    }                                                
                    ,                                                
                        "3001:0:0:1::/64": {                             
                            "prefix":"3001:0:0:1::\/64",                   
                            "advertisedTo":{                               
                                "2001::1":{                                  
                                }                                            
                            },                                             
                            "paths":[                                      
                            {                                            
                                "pathId":0,                                
                                "aspath":{                                 
                                    "string":"100 {10000,20000} 30000 40000",
                                    "segments":[                             
                                    {                                      
                                        "type":"as-sequence",                
                                        "list":[                             
                                            100                                
                                        ]                                    
                                    },                                     
                                    {                                      
                                        "type":"as-set",                     
                                        "list":[                             
                                            10000,                             
                                        20000                              
                                        ]                                    
                                    },                                     
                                    {                                      
                                        "type":"as-sequence",                
                                        "list":[                             
                                            30000,                             
                                        40000                              
                                        ]                                    
                                    }                                      
                                    ],
                                    "length":4
                                },
                                "origin":"IGP",
                                "valid":true,
                                "bestpath":{
                                    "overall":true
                                },
                                "lastUpdate":{
                                    "epoch":1495309393,
                                    "string":"Sat May 20 19:43:13 2017\n"
                                },
                                "nexthops":[
                                {
                                    "ip":"2001::1",
                                    "afi":"ipv6",
                                    "scope":"global",
                                    "metric":0,
                                    "accessible":true
                                },
                                {
                                    "ip":"fe80::200:1ff:fe4a:417c",
                                    "afi":"ipv6",
                                    "scope":"link-local",
                                    "accessible":true,
                                    "used":true
                                }
                                ],
                                "peer":{
                                    "peerId":"2001::1",
                                    "routerId":"25.98.0.2",
                                    "type":"external"
                                }
                            }
                            ]
                        }
                }
            }`

        case "ipv4-adj-rib-in-pre" :
            outJsonBlob = `{
                "vrfId":0,                                                 
                    "vrfName":"default",                                       
                    "bgpTableVersion":0,                                       
                    "bgpLocalRouterId":"10.20.30.40",                          
                    "defaultLocPrf":100,                                       
                    "localAS":400,                                             
                    "routes":{                                                 
                        "105.205.0.0\/24":{                                      
                            "addrPrefix":"105.205.0.0",                            
                            "prefixLen":24,                                        
                            "prefix":"105.205.0.0\/24",                            
                            "pathId":0,                                            
                            "nextHop":"11.1.1.1",                                  
                            "localPref":35,                                        
                            "locPrf":35,                                           
                            "weight":0,
                            "med":1,
                            "aspath":{                                             
                                "string":"{3000,5000}",                              
                                "segments":[                                         
                                {                                                  
                                    "type":"as-set",                                 
                                    "list":[                                         
                                        3000,                                          
                                    5000                                           
                                    ]                                                
                                }                                                  
                                ],                                                   
                                "length":1                                           
                            },                                                     
                            "origin":"EGP",                                        
                            "clusterList":{                                        
                                "list":[                                             
                                    "10.20.30.40",                                     
                                "50.60.70.80",                                     
                                "90.100.110.120"                                   
                                ]                                                    
                            },                                                     
                            "lastUpdate":{                                         
                                "epoch":1495219239,                                  
                                "string":"Fri May 19 18:40:39 2017\n"                
                            },                                                     
                            "appliedStatusSymbols":{                               
                                "*":true,                                            
                                ">":true                                             
                            }                                                      
                        },                                                       
                        "105.205.1.0\/24":{                                      
                            "addrPrefix":"105.205.1.0",                            
                            "prefixLen":24,                                        
                            "prefix":"105.205.1.0\/24",                            
                            "pathId":0,                                            
                            "nextHop":"11.1.1.1",                                  
                            "localPref":35,                                        
                            "locPrf":35,                                           
                            "weight":0,                                            
                            "aspath":{                                             
                                "string":"{3000,5000}",                              
                                "segments":[                                         
                                {                                                  
                                    "type":"as-set",                                 
                                    "list":[                                         
                                        3000,                                          
                                    5000                                           
                                    ]                                                
                                }                                                  
                                ],                                                   
                                "length":1                                           
                            },                                                     
                            "origin":"EGP",                                        
                            "clusterList":{                                        
                                "list":[                                             
                                    "10.20.30.40",                                     
                                "50.60.70.80",                                     
                                "90.100.110.120"                                   
                                ]                                                    
                            },                                                     
                            "lastUpdate":{                                         
                                "epoch":1495219239,                                  
                                "string":"Fri May 19 18:40:39 2017\n"                
                            },                                                     
                            "appliedStatusSymbols":{                               
                                "*":true,                                            
                                ">":true                                             
                            }                                                      
                        }                                                        
                    },                                                         
                    "totalPrefixCounter":2,                                    
                    "filteredPrefixCounter":0                                  
            }`

        case "ipv6-adj-rib-in-pre" :
            outJsonBlob = `{
                "vrfId":0,                                                   
                "vrfName":"default",                                         
                "bgpTableVersion":0,                                         
                "bgpLocalRouterId":"11.1.1.4",                               
                "defaultLocPrf":100,                                         
                "localAS":400,                                               
                "routes":{                                                   
                    "3001::\/64":{                                             
                        "addrPrefix":"3001::",                                   
                        "prefixLen":64,                                          
                        "prefix":"3001::\/64",                                   
                        "pathId":0,                                              
                        "nextHopGlobal":"2001::1",                               
                        "weight":0,                                              
                        "aspath":{                                               
                            "string":"100 {10000,20000} 30000 40000",              
                            "segments":[                                           
                            {                                                    
                                "type":"as-sequence",                              
                                "list":[                                           
                                100                                              
                                ]                                                  
                            },                                                   
                            {                                                    
                                "type":"as-set",                                   
                                "list":[                                           
                                10000,                                           
                                20000                                            
                                ]                                                  
                            },                                                   
                            {                                                    
                                "type":"as-sequence",                              
                                "list":[                                           
                                30000,                                           
                                40000                                            
                                ]                                                  
                            }                                                    
                            ],                                                     
                            "length":4                                             
                        },                                                       
                        "origin":"IGP",                                          
                        "lastUpdate":{                                           
                            "epoch":1495309393,                                    
                            "string":"Sat May 20 19:43:13 2017\n"                  
                        },                                                       
                        "appliedStatusSymbols":{                                 
                            "*":true,                                              
                            ">":true                                               
                        }                                                        
                    },                                                         
                    "3001:0:0:1::\/64":{                                       
                        "addrPrefix":"3001:0:0:1::",                             
                        "prefixLen":64,                                          
                        "prefix":"3001:0:0:1::\/64",                             
                        "pathId":0,                                              
                        "nextHopGlobal":"2001::1",                               
                        "weight":0,                                              
                        "aspath":{                                               
                            "string":"100 {10000,20000} 30000 40000",              
                            "segments":[                                           
                            {
                                "type":"as-sequence",
                                "list":[
                                100
                                ]
                            },
                            {
                                "type":"as-set",
                                "list":[
                                10000,
                                20000
                                ]
                            },
                            {
                                "type":"as-sequence",
                                "list":[
                                30000,
                                40000
                                ]
                            }
                            ],
                            "length":4
                        },
                        "origin":"IGP",
                        "lastUpdate":{
                            "epoch":1495309393,
                            "string":"Sat May 20 19:43:13 2017\n"
                        },
                        "appliedStatusSymbols":{
                            "*":true,
                            ">":true
                        }
                    }
                },
                "totalPrefixCounter":2,
                "filteredPrefixCounter":0
            }`

        case "ipv4-adj-rib-in-post":
            outJsonBlob = `{
                "vrfId": 0,                                               
                    "vrfName": "default",                                     
                    "tableVersion": 2,                                        
                    "routerId": "10.20.30.40",                                
                    "defaultLocPrf": 100,                                     
                    "localAS": 400,                                           
                    "routes": {                                               
                        "70.80.90.0/24": {                                         
                            "prefix":"70.80.90.0\/24",                               
                            "advertisedTo":{                                         
                                "10.10.10.1":{                                         
                                },                                                     
                                "11.1.1.1":{                                           
                                }                                                      
                            },                                                       
                            "paths":[                                                
                            {                                                      
                                "pathId":0,                                          
                                "aspath":{                                           
                                    "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                    "segments":[                                                                   
                                    {                                                                            
                                        "type":"as-sequence",                                                      
                                        "list":[                                                                   
                                            200                                                                      
                                        ]                                                                          
                                    },                                                                           
                                    {                                                                            
                                        "type":"as-set",                                                           
                                        "list":[                                                                   
                                            111,                                                                     
                                        333,                                                                     
                                        555,                                                                     
                                        777                                                                      
                                        ]                                                                          
                                    },                                                                           
                                    {                                                                            
                                        "type":"as-sequence",                                                      
                                        "list":[                                                                   
                                            222,                                                                     
                                        444,                                                                     
                                        666,                                                                     
                                        888                                                                      
                                        ]                                                                          
                                    },                                                                           
                                    {                                                                            
                                        "type":"as-set",                                                           
                                        "list":[                                                                   
                                            100000,                                                                  
                                        200000,                                                                  
                                        300000,                                                                  
                                        400000                                                                   
                                        ]                                                                          
                                    }                                                                            
                                    ],                                                                             
                                    "length":7                                                                     
                                },                                                                               
                                "aggregatorAs":5000,                                                             
                                "aggregatorId":"10.20.30.40",                                                    
                                "origin":"IGP",                                                                  
                                "med":37,                                                                        
                                "metric":37,                                                                     
                                "valid":true,                                                                    
                                "localPref":100,
                                "atomicAggregate":true,                                                          
                                "bestpath":{                                                                     
                                    "overall":true                                                                 
                                },                                                                               
                                "community":{                                                                    
                                    "string":"1000:2000 3000:4000",                                                
                                    "list":[                                                                       
                                        "1000:2000",                                                                 
                                    "3000:4000"                                                                  
                                    ]                                                                              
                                },                                                                               
                                "extendedCommunity":{                                                            
                                    "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                },                                                                               
                                "lastUpdate":{                                                                   
                                    "epoch":1495225039,                                                            
                                    "string":"Fri May 19 20:17:19 2017\n"                                          
                                },                                                                               
                                "nexthops":[                                                                     
                                {                                                                              
                                    "ip":"10.10.10.1",                                                           
                                    "afi":"ipv4",                                                                
                                    "metric":0,                                                                  
                                    "accessible":true,                                                           
                                    "used":true                                                                  
                                }                                                                              
                                ],                                                                               
                                "peer":{                                                                         
                                    "peerId":"10.10.10.1",                                                         
                                    "routerId":"25.98.0.1",                                                        
                                    "type":"external"                                                              
                                }                                                                                
                            }                                                                                  
                            ]                                                                                    
                        }                                                                                      
                        ,                                                                                      
                            "70.80.91.0/24": {                                                                     
                                "prefix":"70.80.91.0\/24",                                                           
                                "advertisedTo":{                                                                     
                                    "10.10.10.1":{                                                                     
                                    },                                                                                 
                                    "11.1.1.1":{                                                                       
                                    }                                                                                  
                                },                                                                                   
                                "paths":[                                                                            
                                {                                                                                  
                                    "pathId":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                200                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "origin":"IGP",                                                                  
                                    "med":37,                                                                        
                                    "metric":37,                                                                     
                                    "valid":true,
                                    "atomicAggregate":true,
                                    "bestpath":{
                                        "overall":true
                                    },
                                    "community":{
                                        "string":"1000:2000 3000:4000",
                                        "list":[
                                            "1000:2000",
                                        "3000:4000"
                                        ]
                                    },
                                    "extendedCommunity":{
                                        "string":"RT:10:3369254580 RT:20:2358681770"
                                    },
                                    "lastUpdate":{
                                        "epoch":1495225039,
                                        "string":"Fri May 19 20:17:19 2017\n"
                                    },
                                    "nexthops":[
                                    {
                                        "ip":"10.10.10.1",
                                        "afi":"ipv4",
                                        "metric":0,
                                        "accessible":true,
                                        "used":true
                                    }
                                    ],
                                    "peer":{
                                        "peerId":"10.10.10.1",
                                        "routerId":"25.98.0.1",
                                        "type":"external"
                                    }
                                }
                                ]
                            }
                    }  }`

        case "ipv6-adj-rib-in-post":
            outJsonBlob = `{
                "vrfId": 0,                                          
                    "vrfName": "default",                                
                    "tableVersion": 6,                                   
                    "routerId": "11.1.1.4",                              
                    "defaultLocPrf": 100,                                
                    "localAS": 400,                                      
                    "routes": {                                          
                        "3001::/64": {                                        
                            "prefix":"3001::\/64",                              
                            "advertisedTo":{                                    
                                "2001::1":{                                       
                                }                                                 
                            },                                                  
                            "paths":[                                           
                            {                                                 
                                "pathId":0,                                     
                                "aspath":{                                      
                                    "string":"100 {10000,20000} 30000 40000",     
                                    "segments":[                                  
                                    {                                           
                                        "type":"as-sequence",                     
                                        "list":[                                  
                                            100                                     
                                        ]                                         
                                    },                                          
                                    {                                           
                                        "type":"as-set",                          
                                        "list":[                                  
                                            10000,                                  
                                        20000                                   
                                        ]                                         
                                    },                                          
                                    {                                           
                                        "type":"as-sequence",                     
                                        "list":[                                  
                                            30000,                                  
                                        40000                                   
                                        ]                                         
                                    }                                           
                                    ],                                            
                                    "length":4                                    
                                },                                              
                                "origin":"IGP",                                 
                                "valid":true,    
                                "med":37,                                                                        
                                "metric":37,                                                                     
                                "atomicAggregate":true,
                                "localPref":100,
                                "bestpath":{                                    
                                    "overall":true                                
                                },                                              
                                "lastUpdate":{                                  
                                    "epoch":1495309392,                           
                                    "string":"Sat May 20 19:43:12 2017\n"         
                                },                                              
                                "nexthops":[                                    
                                {                                             
                                    "ip":"2001::1",                             
                                    "afi":"ipv6",                               
                                    "scope":"global",                           
                                    "metric":0,                                 
                                    "accessible":true                           
                                },                                            
                                {                                             
                                    "ip":"fe80::200:1ff:fe4a:417c",             
                                    "afi":"ipv6",                               
                                    "scope":"link-local",                       
                                    "accessible":true,                          
                                    "used":true                                 
                                }                                             
                                ],                                              
                                "peer":{                                        
                                    "peerId":"2001::1",                           
                                    "routerId":"25.98.0.2",                       
                                    "type":"external"                             
                                }                                               
                            }                                                 
                            ]                                                   
                        }                                                     
                        ,                                                     
                            "3001:0:0:1::/64": {                                  
                                "prefix":"3001:0:0:1::\/64",                        
                                "advertisedTo":{                                    
                                    "2001::1":{                                       
                                    }                                                 
                                },                                                  
                                "paths":[                                           
                                {                                                 
                                    "pathId":0,                                     
                                    "aspath":{                                      
                                        "string":"100 {10000,20000} 30000 40000",     
                                        "segments":[                                  
                                        {                                           
                                            "type":"as-sequence",                     
                                            "list":[                                  
                                                100                                     
                                            ]                                         
                                        },                                          
                                        {                                           
                                            "type":"as-set",                          
                                            "list":[                                  
                                                10000,                                  
                                            20000                                   
                                            ]                                         
                                        },                                          
                                        {                                           
                                            "type":"as-sequence",                     
                                            "list":[                                  
                                                30000,                                  
                                            40000                                   
                                            ]                                         
                                        }                                           
                                        ],
                                        "length":4
                                    },
                                    "origin":"IGP",
                                    "valid":true,
                                    "bestpath":{
                                        "overall":true
                                    },
                                    "lastUpdate":{
                                        "epoch":1495309392,
                                        "string":"Sat May 20 19:43:12 2017\n"
                                    },
                                    "nexthops":[
                                    {
                                        "ip":"2001::1",
                                        "afi":"ipv6",
                                        "scope":"global",
                                        "metric":0,
                                        "accessible":true
                                    },
                                    {
                                        "ip":"fe80::200:1ff:fe4a:417c",
                                        "afi":"ipv6",
                                        "scope":"link-local",
                                        "accessible":true,
                                        "used":true
                                    }
                                    ],
                                    "peer":{
                                        "peerId":"2001::1",
                                        "routerId":"25.98.0.2",
                                        "type":"external"
                                    }
                                }
                                ]
                            }
                    }  }`

        case "ipv4-adj-rib-out-post":
            outJsonBlob = `{
                "vrfId":0,
                "vrfName":"default",
                "bgpTableVersion":2,
                "bgpLocalRouterId":"10.20.30.40",
                "defaultLocPrf":100,
                "localAS":400,
                "routes":{
                    "70.80.90.0\/24":{
                        "addrPrefix":"70.80.90.0",
                        "prefixLen":24,
                        "prefix":"70.80.90.0\/24",
                        "pathId":0,
                        "nextHop":"10.10.10.1",
                        "metric":37,
                        "localPref":100,
                        "locPrf":100,
                        "weight":0,
                        "aspath":{
                            "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                            "segments":[
                            {
                                "type":"as-sequence",
                                "list":[
                                    200
                                ]
                            },
                            {
                                "type":"as-set",
                                "list":[
                                    111,
                                333,
                                555,
                                777
                                ]
                            },
                            {
                                "type":"as-sequence",
                                "list":[
                                    222,
                                444,
                                666,
                                888
                                ]
                            },
                            {
                                "type":"as-set",
                                "list":[
                                    100000,
                                200000,
                                300000,
                                400000
                                ]
                            }
                            ],
                            "length":7
                        },
                        "origin":"IGP",
                        "aggregatorAs":5000,
                        "aggregatorId":"10.20.30.40",
                        "atomicAggregate":true,
                        "community":{
                            "string":"1000:2000 3000:4000",
                            "list":[
                                "1000:2000",
                            "3000:4000"
                            ]
                        },
                        "extendedCommunity":{
                            "string":"RT:10:3369254580 RT:20:2358681770"
                        },
                        "lastUpdate":{
                            "epoch":1495225039,
                            "string":"Fri May 19 20:17:19 2017\n"
                        },
                        "appliedStatusSymbols":{
                            "*":true,
                            ">":true
                        }
                    },
                    "70.80.91.0\/24":{
                        "addrPrefix":"70.80.91.0",
                        "prefixLen":24,
                        "prefix":"70.80.91.0\/24",
                        "pathId":0,
                        "nextHop":"10.10.10.1",
                        "metric":37,
                        "localPref":100,
                        "locPrf":100,
                        "weight":0,
                        "aspath":{
                            "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                            "segments":[
                            {
                                "type":"as-sequence",
                                "list":[
                                    200
                                ]
                            },
                            {
                                "type":"as-set",
                                "list":[
                                    111,
                                333,
                                555,
                                777
                                ]
                            },
                            {
                                "type":"as-sequence",
                                "list":[
                                    222,
                                444,
                                666,
                                888
                                ]
                            },
                            {
                                "type":"as-set",
                                "list":[
                                    100000,
                                200000,
                                300000,
                                400000
                                ]
                            }
                            ],
                            "length":7
                        },
                        "origin":"IGP",
                        "aggregatorAs":5000,
                        "aggregatorId":"10.20.30.40",
                        "atomicAggregate":true,
                        "community":{
                            "string":"1000:2000 3000:4000",
                            "list":[
                                "1000:2000",
                            "3000:4000"
                            ]
                        },
                        "extendedCommunity":{
                            "string":"RT:10:3369254580 RT:20:2358681770"
                        },
                        "lastUpdate":{
                            "epoch":1495225039,
                            "string":"Fri May 19 20:17:19 2017\n"
                        },
                        "appliedStatusSymbols":{
                            "*":true,
                            ">":true
                        }
                    }
                },
                "totalPrefixCounter":2,
                "filteredPrefixCounter":0
            }`

        case "ipv6-adj-rib-out-post":
            outJsonBlob = `{
                "vrfId":0,                                                       
                    "vrfName":"default",                                             
                    "bgpTableVersion":6,                                             
                    "bgpLocalRouterId":"11.1.1.4",                                   
                    "defaultLocPrf":100,                                             
                    "localAS":400,                                                   
                    "routes":{                                                       
                        "3001::\/64":{                                                 
                            "addrPrefix":"3001::",                                       
                            "prefixLen":64,                                              
                            "prefix":"3001::\/64",                                       
                            "pathId":0,                                                  
                            "nextHopGlobal":"::",                                        
                            "weight":0,                                                  
                            "aspath":{                                                   
                                "string":"100 {10000,20000} 30000 40000",                  
                                "segments":[                                               
                                {                                                        
                                    "type":"as-sequence",                                  
                                    "list":[                                               
                                        100                                                  
                                    ]                                                      
                                },                                                       
                                {                                                        
                                    "type":"as-set",                                       
                                    "list":[                                               
                                        10000,                                               
                                    20000                                                
                                    ]                                                      
                                },                                                       
                                {                                                        
                                    "type":"as-sequence",                                  
                                    "list":[                                               
                                        30000,                                               
                                    40000                                                
                                    ]                                                      
                                }                                                        
                                ],                                                         
                                "length":4                                                 
                            },                                                           
                            "origin":"IGP",                                              
                            "lastUpdate":{                                               
                                "epoch":1495309738,                                        
                                "string":"Sat May 20 19:48:58 2017\n"                      
                            },                                                           
                            "appliedStatusSymbols":{                                     
                                "*":true,                                                  
                                ">":true                                                   
                            }                                                            
                        },                                                             
                        "3001:0:0:1::\/64":{                                           
                            "addrPrefix":"3001:0:0:1::",                                 
                            "prefixLen":64,                                              
                            "prefix":"3001:0:0:1::\/64",                                 
                            "pathId":0,                                                  
                            "nextHopGlobal":"::",                                        
                            "weight":0,                                                  
                            "aspath":{                                                   
                                "string":"100 {10000,20000} 30000 40000",                  
                                "segments":[                                               
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        100
                                    ]
                                },
                                {
                                    "type":"as-set",
                                    "list":[
                                        10000,
                                    20000
                                    ]
                                },
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        30000,
                                    40000
                                    ]
                                }
                                ],
                                "length":4
                            },
                            "origin":"IGP",
                            "lastUpdate":{
                                "epoch":1495309738,
                                "string":"Sat May 20 19:48:58 2017\n"
                            },
                            "appliedStatusSymbols":{
                                "*":true,
                                ">":true
                            }
                        }
                    },
                    "totalPrefixCounter":2,
                    "filteredPrefixCounter":0
            }`

        case "ipv4-all-nbrs-adj-rib":
            outJsonBlob = `{
                "10.10.10.1":[                                   
                    {                                                
                        "vrfId":0,                                     
                            "vrfName":"default",                           
                            "bgpTableVersion":4,                           
                            "bgpLocalRouterId":"10.20.30.40",              
                            "defaultLocPrf":100,                           
                            "localAS":200,                                 
                            "advertisedRoutes":{                           
                                "70.80.90.0\/24":{                           
                                    "addrPrefix":"70.80.90.0",                 
                                    "prefixLen":24,                            
                                    "prefix":"70.80.90.0\/24",                 
                                    "pathId":0,                                
                                    "nextHop":"10.10.10.1",                    
                                    "weight":0,                                
                                    "aspath":{                                 
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "med":37,                                                                        
                                    "localPref":200,
                                    "originatorId":"0.0.0.0",
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                },                                                                                 
                                "70.80.91.0\/24":{                                                                 
                                    "addrPrefix":"70.80.91.0",                                                       
                                    "prefixLen":24,                                                                  
                                    "prefix":"70.80.91.0\/24",                                                       
                                    "pathId":0,                                                                      
                                    "nextHop":"10.10.10.1",                                                          
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                },                                                                                 
                                "105.205.0.0\/24":{                                                                
                                    "addrPrefix":"105.205.0.0",                                                      
                                    "prefixLen":24,                                                                  
                                    "prefix":"105.205.0.0\/24",                                                      
                                    "pathId":0,                                                                      
                                    "nextHop":"11.1.1.1",                                                            
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"400 {3000,5000}",                                                    
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                400                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                3000,                                                                    
                                            5000                                                                     
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":2                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                },                                                                                 
                                "105.205.1.0\/24":{                                                                
                                    "addrPrefix":"105.205.1.0",                                                      
                                    "prefixLen":24,                                                                  
                                    "prefix":"105.205.1.0\/24",                                                      
                                    "pathId":0,                                                                      
                                    "nextHop":"11.1.1.1",                                                            
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"400 {3000,5000}",                                                    
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                400                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                3000,                                                                    
                                            5000                                                                     
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":2                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                }                                                                                  
                            },                                                                                   
                            "totalPrefixCounter":4,                                                              
                            "filteredPrefixCounter":0                                                            
                    }                                                                                      
                    ,{                                                                                     
                        "vrfId":0,                                                                           
                            "vrfName":"default",                                                                 
                            "bgpTableVersion":0,                                                                 
                            "bgpLocalRouterId":"10.20.30.40",                                                    
                            "defaultLocPrf":100,                                                                 
                            "localAS":200,                                                                       
                            "receivedRoutes":{                                                                   
                                "70.80.90.0\/24":{                                                                 
                                    "addrPrefix":"70.80.90.0",                                                       
                                    "prefixLen":24,                                                                  
                                    "prefix":"70.80.90.0\/24",                                                       
                                    "pathId":0,                                                                      
                                    "nextHop":"10.10.10.1",                                                          
                                    "metric":37,                                                                     
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                },                                                                                 
                                "70.80.91.0\/24":{                                                                 
                                    "addrPrefix":"70.80.91.0",                                                       
                                    "prefixLen":24,                                                                  
                                    "prefix":"70.80.91.0\/24",                                                       
                                    "pathId":0,                                                                      
                                    "nextHop":"10.10.10.1",                                                          
                                    "metric":37,                                                                     
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                }                                                                                  
                            },                                                                                   
                            "totalPrefixCounter":2,                                                              
                            "filteredPrefixCounter":0                                                            
                    }                                                                                      
                    ,{                                                                                     
                        "vrfId": 0,                                                                           
                            "vrfName": "default",                                                                 
                            "tableVersion": 4,                                                                    
                            "routerId": "10.20.30.40",                                                            
                            "defaultLocPrf": 100,                                                                 
                            "localAS": 200,                                                                       
                            "routes": {                                                                           
                                "70.80.90.0/24": {                                                                     
                                    "prefix":"70.80.90.0\/24",                                                           
                                    "advertisedTo":{                                                                     
                                        "10.10.10.1":{                                                                     
                                        },                                                                                 
                                        "11.1.1.1":{                                                                       
                                        }                                                                                  
                                    },                                                                                   
                                    "paths":[                                                                            
                                    {                                                                                  
                                        "pathId":0,                                                                      
                                        "aspath":{                                                                       
                                            "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                            "segments":[                                                                   
                                            {                                                                            
                                                "type":"as-sequence",                                                      
                                                "list":[                                                                   
                                                    100                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-set",                                                           
                                                "list":[                                                                   
                                                    111,                                                                     
                                                333,                                                                     
                                                555,                                                                     
                                                777                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-sequence",                                                      
                                                "list":[                                                                   
                                                    222,                                                                     
                                                444,                                                                     
                                                666,                                                                     
                                                888                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-set",                                                           
                                                "list":[                                                                   
                                                    100000,                                                                  
                                                200000,                                                                  
                                                300000,                                                                  
                                                400000                                                                   
                                                ]                                                                          
                                            }                                                                            
                                            ],                                                                             
                                            "length":7                                                                     
                                        },                                                                               
                                        "aggregatorAs":5000,                                                             
                                        "aggregatorId":"10.20.30.40",                                                    
                                        "origin":"IGP",                                                                  
                                        "med":37,                                                                        
                                        "metric":37,                                                                     
                                        "valid":true,                                                                    
                                        "atomicAggregate":true,                                                          
                                        "bestpath":{                                                                     
                                            "overall":true                                                                 
                                        },                                                                               
                                        "community":{                                                                    
                                            "string":"1000:2000 3000:4000",                                                
                                            "list":[                                                                       
                                                "1000:2000",                                                                 
                                            "3000:4000"                                                                  
                                            ]                                                                              
                                        },                                                                               
                                        "extendedCommunity":{                                                            
                                            "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                        },                                                                               
                                        "lastUpdate":{                                                                   
                                            "epoch":1495740584,                                                            
                                            "string":"Thu May 25 19:29:44 2017\n"                                          
                                        },                                                                               
                                        "nexthops":[                                                                     
                                        {                                                                              
                                            "ip":"10.10.10.1",                                                           
                                            "afi":"ipv4",                                                                
                                            "metric":0,                                                                  
                                            "accessible":true,                                                           
                                            "used":true                                                                  
                                        }                                                                              
                                        ],                                                                               
                                        "peer":{                                                                         
                                            "peerId":"10.10.10.1",                                                         
                                            "routerId":"25.98.0.1",                                                        
                                            "type":"external"                                                              
                                        }                                                                                
                                    }                                                                                  
                                    ]                                                                                    
                                }                                                                                  
                                ,                                                                                      
                                    "70.80.91.0/24": {                                                                     
                                        "prefix":"70.80.91.0\/24",                                                           
                                        "advertisedTo":{                                                                     
                                            "10.10.10.1":{                                                                     
                                            },                                                                                 
                                            "11.1.1.1":{                                                                       
                                            }                                                                                  
                                        },                                                                                   
                                        "paths":[                                                                            
                                        {                                                                                  
                                            "pathId":0,                                                                      
                                            "aspath":{                                                                       
                                                "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                                "segments":[                                                                   
                                                {                                                                            
                                                    "type":"as-sequence",                                                      
                                                    "list":[                                                                   
                                                        100                                                                      
                                                    ]                                                                          
                                                },                                                                           
                                                {                                                                            
                                                    "type":"as-set",                                                           
                                                    "list":[                                                                   
                                                        111,                                                                     
                                                    333,                                                                     
                                                    555,                                                                     
                                                    777                                                                      
                                                    ]                                                                          
                                                },                                                                           
                                                {                                                                            
                                                    "type":"as-sequence",                                                      
                                                    "list":[                                                                   
                                                        222,                                                                     
                                                    444,                                                                     
                                                    666,                                                                     
                                                    888                                                                      
                                                    ]                                                                          
                                                },                                                                           
                                                {                                                                            
                                                    "type":"as-set",                                                           
                                                    "list":[                                                                   
                                                        100000,                                                                  
                                                    200000,                                                                  
                                                    300000,                                                                  
                                                    400000                                                                   
                                                    ]                                                                          
                                                }                                                                            
                                                ],                                                                             
                                                "length":7                                                                     
                                            },                                                                               
                                            "aggregatorAs":5000,                                                             
                                            "aggregatorId":"10.20.30.40",                                                    
                                            "origin":"IGP",                                                                  
                                            "med":37,                                                                        
                                            "metric":37,                                                                     
                                            "valid":true,                                                                    
                                            "atomicAggregate":true,                                                          
                                            "bestpath":{                                                                     
                                                "overall":true                                                                 
                                            },                                                                               
                                            "community":{                                                                    
                                                "string":"1000:2000 3000:4000",                                                
                                                "list":[                                                                       
                                                    "1000:2000",                                                                 
                                                "3000:4000"                                                                  
                                                ]                                                                              
                                            },                                                                               
                                            "extendedCommunity":{                                                            
                                                "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                            },                                                                               
                                            "lastUpdate":{                                                                   
                                                "epoch":1495740584,                                                            
                                                "string":"Thu May 25 19:29:44 2017\n"                                          
                                            },                                                                               
                                            "nexthops":[                                                                     
                                            {                                                                              
                                                "ip":"10.10.10.1",                                                           
                                                "afi":"ipv4",                                                                
                                                "metric":0,                                                                  
                                                "accessible":true,                                                           
                                                "used":true                                                                  
                                            }                                                                              
                                            ],                                                                               
                                            "peer":{                                                                         
                                                "peerId":"10.10.10.1",                                                         
                                                "routerId":"25.98.0.1",                                                        
                                                "type":"external"                                                              
                                            }                                                                                
                                        }                                                                                  
                                        ]                                                                                    
                                    }                                                                                      
                            }
                    }                                                                                  
                ],                                                                                     
                "11.1.1.1":[                                                                           
                {                                                
                    "vrfId":0,                                     
                    "vrfName":"default",                           
                    "bgpTableVersion":4,                           
                    "bgpLocalRouterId":"10.20.30.40",              
                    "defaultLocPrf":100,                           
                    "localAS":200,                                 
                    "advertisedRoutes":{                           
                        "70.80.90.0\/24":{                           
                            "addrPrefix":"70.80.90.0",                 
                            "prefixLen":24,                            
                            "prefix":"70.80.90.0\/24",                 
                            "pathId":0,                                
                            "nextHop":"10.10.10.1",                    
                            "weight":0,                                
                            "aspath":{                                 
                                "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                "segments":[                                                                   
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        100                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        111,                                                                     
                                    333,                                                                     
                                    555,                                                                     
                                    777                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        222,                                                                     
                                    444,                                                                     
                                    666,                                                                     
                                    888                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        100000,                                                                  
                                    200000,                                                                  
                                    300000,                                                                  
                                    400000                                                                   
                                    ]                                                                          
                                }                                                                            
                                ],                                                                             
                                "length":7                                                                     
                            },                                                                               
                            "origin":"IGP",                                                                  
                            "aggregatorAs":5000,                                                             
                            "aggregatorId":"10.20.30.40",                                                    
                            "atomicAggregate":true,                                                          
                            "med":37,                                                                        
                            "localPref":200,
                            "originatorId":"0.0.0.0",
                            "community":{                                                                    
                                "string":"1000:2000 3000:4000",                                                
                                "list":[                                                                       
                                    "1000:2000",                                                                 
                                "3000:4000"                                                                  
                                ]                                                                              
                            },                                                                               
                            "extendedCommunity":{                                                            
                                "string":"RT:10:3369254580 RT:20:2358681770"                                   
                            },                                                                               
                            "lastUpdate":{                                                                   
                                "epoch":1495740584,                                                            
                                "string":"Thu May 25 19:29:44 2017\n"                                          
                            },                                                                               
                            "appliedStatusSymbols":{                                                         
                                "*":true,                                                                      
                                ">":true                                                                       
                            }                                                                                
                        },                                                                                 
                        "70.80.91.0\/24":{                                                                 
                            "addrPrefix":"70.80.91.0",                                                       
                            "prefixLen":24,                                                                  
                            "prefix":"70.80.91.0\/24",                                                       
                            "pathId":0,                                                                      
                            "nextHop":"10.10.10.1",                                                          
                            "weight":0,                                                                      
                            "aspath":{                                                                       
                                "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                "segments":[                                                                   
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        100                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        111,                                                                     
                                    333,                                                                     
                                    555,                                                                     
                                    777                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        222,                                                                     
                                    444,                                                                     
                                    666,                                                                     
                                    888                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        100000,                                                                  
                                    200000,                                                                  
                                    300000,                                                                  
                                    400000                                                                   
                                    ]                                                                          
                                }                                                                            
                                ],                                                                             
                                "length":7                                                                     
                            },                                                                               
                            "origin":"IGP",                                                                  
                            "aggregatorAs":5000,                                                             
                            "aggregatorId":"10.20.30.40",                                                    
                            "atomicAggregate":true,                                                          
                            "community":{                                                                    
                                "string":"1000:2000 3000:4000",                                                
                                "list":[                                                                       
                                    "1000:2000",                                                                 
                                "3000:4000"                                                                  
                                ]                                                                              
                            },                                                                               
                            "extendedCommunity":{                                                            
                                "string":"RT:10:3369254580 RT:20:2358681770"                                   
                            },                                                                               
                            "lastUpdate":{                                                                   
                                "epoch":1495740584,                                                            
                                "string":"Thu May 25 19:29:44 2017\n"                                          
                            },                                                                               
                            "appliedStatusSymbols":{                                                         
                                "*":true,                                                                      
                                ">":true                                                                       
                            }                                                                                
                        },                                                                                 
                        "105.205.0.0\/24":{                                                                
                            "addrPrefix":"105.205.0.0",                                                      
                            "prefixLen":24,                                                                  
                            "prefix":"105.205.0.0\/24",                                                      
                            "pathId":0,                                                                      
                            "nextHop":"11.1.1.1",                                                            
                            "weight":0,                                                                      
                            "aspath":{                                                                       
                                "string":"400 {3000,5000}",                                                    
                                "segments":[                                                                   
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        400                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        3000,                                                                    
                                    5000                                                                     
                                    ]                                                                          
                                }                                                                            
                                ],                                                                             
                                "length":2                                                                     
                            },                                                                               
                            "origin":"IGP",                                                                  
                            "lastUpdate":{                                                                   
                                "epoch":1495740584,                                                            
                                "string":"Thu May 25 19:29:44 2017\n"                                          
                            },                                                                               
                            "appliedStatusSymbols":{                                                         
                                "*":true,                                                                      
                                ">":true                                                                       
                            }                                                                                
                        },                                                                                 
                        "105.205.1.0\/24":{                                                                
                            "addrPrefix":"105.205.1.0",                                                      
                            "prefixLen":24,                                                                  
                            "prefix":"105.205.1.0\/24",                                                      
                            "pathId":0,                                                                      
                            "nextHop":"11.1.1.1",                                                            
                            "weight":0,                                                                      
                            "aspath":{                                                                       
                                "string":"400 {3000,5000}",                                                    
                                "segments":[                                                                   
                                {                                                                            
                                    "type":"as-sequence",                                                      
                                    "list":[                                                                   
                                        400                                                                      
                                    ]                                                                          
                                },                                                                           
                                {                                                                            
                                    "type":"as-set",                                                           
                                    "list":[                                                                   
                                        3000,                                                                    
                                    5000                                                                     
                                    ]                                                                          
                                }                                                                            
                                ],                                                                             
                                "length":2                                                                     
                            },                                                                               
                            "origin":"IGP",                                                                  
                            "lastUpdate":{                                                                   
                                "epoch":1495740584,                                                            
                                "string":"Thu May 25 19:29:44 2017\n"                                          
                            },                                                                               
                            "appliedStatusSymbols":{                                                         
                                "*":true,                                                                      
                                ">":true                                                                       
                            }                                                                                
                        }                                                                                  
                    },                                                                                   
                    "totalPrefixCounter":4,                                                              
                    "filteredPrefixCounter":0                                                            
                }                                                                                      
                    ,{                                                                                     
                        "vrfId":0,                                                                           
                            "vrfName":"default",                                                                 
                            "bgpTableVersion":0,                                                                 
                            "bgpLocalRouterId":"10.20.30.40",                                                    
                            "defaultLocPrf":100,                                                                 
                            "localAS":200,                                                                       
                            "receivedRoutes":{                                                                   
                                "70.80.90.0\/24":{                                                                 
                                    "addrPrefix":"70.80.90.0",                                                       
                                    "prefixLen":24,                                                                  
                                    "prefix":"70.80.90.0\/24",                                                       
                                    "pathId":0,                                                                      
                                    "nextHop":"10.10.10.1",                                                          
                                    "metric":37,                                                                     
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                },                                                                                 
                                "70.80.91.0\/24":{                                                                 
                                    "addrPrefix":"70.80.91.0",                                                       
                                    "prefixLen":24,                                                                  
                                    "prefix":"70.80.91.0\/24",                                                       
                                    "pathId":0,                                                                      
                                    "nextHop":"10.10.10.1",                                                          
                                    "metric":37,                                                                     
                                    "weight":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"100 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                100                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "atomicAggregate":true,                                                          
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "appliedStatusSymbols":{                                                         
                                        "*":true,                                                                      
                                        ">":true                                                                       
                                    }                                                                                
                                }                                                                                  
                            },                                                                                   
                            "totalPrefixCounter":2,                                                              
                            "filteredPrefixCounter":0                                                            
                    }
                    ,{                                                                                      
                        "vrfId": 0,                                                                           
                        "vrfName": "default",                                                                 
                        "tableVersion": 4,                                                                    
                        "routerId": "10.20.30.40",                                                            
                        "defaultLocPrf": 100,                                                                 
                        "localAS": 200,                                                                       
                        "routes": {                                                                           
                            "105.205.0.0/24": {                                                                    
                                "prefix":"105.205.0.0\/24",                                                          
                                "advertisedTo":{                                                                     
                                    "10.10.10.1":{                                                                     
                                    },                                                                                 
                                    "11.1.1.1":{                                                                       
                                    }                                                                                  
                                },                                                                                   
                                "paths":[                                                                            
                                {                                                                                  
                                    "pathId":0,                                                                      
                                    "aspath":{                                                                       
                                        "string":"400 {3000,5000}",                                                    
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                400                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                3000,                                                                    
                                            5000                                                                     
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":2                                                                     
                                    },                                                                               
                                    "origin":"IGP",                                                                  
                                    "valid":true,                                                                    
                                    "bestpath":{                                                                     
                                        "overall":true                                                                 
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495740584,                                                            
                                        "string":"Thu May 25 19:29:44 2017\n"                                          
                                    },                                                                               
                                    "nexthops":[                                                                     
                                    {                                                                              
                                        "ip":"11.1.1.1",                                                             
                                        "afi":"ipv4",                                                                
                                        "metric":0,                                                                  
                                        "accessible":true,                                                           
                                        "used":true                                                                  
                                    }                                                                              
                                    ],                                                                               
                                    "peer":{                                                                         
                                        "peerId":"11.1.1.1",                                                           
                                        "routerId":"25.99.0.2",                                                        
                                        "type":"external"                                                              
                                    }                                                                                
                                }                                                                                  
                                ]                                                                                    
                            }                                                                                      
                            ,                                                                                      
                                "105.205.1.0/24": {                                                                    
                                    "prefix":"105.205.1.0\/24",                                                          
                                    "advertisedTo":{                                                                     
                                        "10.10.10.1":{                                                                     
                                        },                                                                                 
                                        "11.1.1.1":{                                                                       
                                        }                                                                                  
                                    },                                                                                   
                                    "paths":[                                                                            
                                    {                                                                                  
                                        "pathId":0,                                                                      
                                        "aspath":{                                                                       
                                            "string":"400 {3000,5000}",                                                    
                                            "segments":[                                                                   
                                            {                                                                            
                                                "type":"as-sequence",                                                      
                                                "list":[                                                                   
                                                    400                                                                      
                                                ]                                                                          
                                            },
                                            {
                                                "type":"as-set",
                                                "list":[
                                                    3000,
                                                5000
                                                ]
                                            }
                                            ],
                                            "length":2
                                        },
                                        "origin":"IGP",
                                        "valid":true,
                                        "bestpath":{
                                            "overall":true
                                        },
                                        "lastUpdate":{
                                            "epoch":1495740584,
                                            "string":"Thu May 25 19:29:44 2017\n"
                                        },
                                        "nexthops":[
                                        {
                                            "ip":"11.1.1.1",
                                            "afi":"ipv4",
                                            "metric":0,
                                            "accessible":true,
                                            "used":true
                                        }
                                        ],
                                        "peer":{
                                            "peerId":"11.1.1.1",
                                            "routerId":"25.99.0.2",
                                            "type":"external"
                                        }
                                    }
                                    ]
                                }
                        }
                    }
                ]
            }`

        case "ipv6-all-nbrs-adj-rib":
            outJsonBlob = `{
                "1001::1":[                                   
                {                                                
                    "vrfId":0,                                                       
                        "vrfName":"default",                                             
                        "bgpTableVersion":6,                                             
                        "bgpLocalRouterId":"11.1.1.4",                                   
                        "defaultLocPrf":100,                                             
                        "localAS":400,                                                   
                        "routes": {                                               
                            "70.80.90.0/24": {                                         
                                "prefix":"70.80.90.0\/24",                               
                                "advertisedTo":{                                         
                                    "10.10.10.1":{                                         
                                    },                                                     
                                    "11.1.1.1":{                                           
                                    }                                                      
                                },                                                       
                                "paths":[                                                
                                {                                                      
                                    "pathId":0,                                          
                                    "aspath":{                                           
                                        "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                        "segments":[                                                                   
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                200                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                111,                                                                     
                                            333,                                                                     
                                            555,                                                                     
                                            777                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-sequence",                                                      
                                            "list":[                                                                   
                                                222,                                                                     
                                            444,                                                                     
                                            666,                                                                     
                                            888                                                                      
                                            ]                                                                          
                                        },                                                                           
                                        {                                                                            
                                            "type":"as-set",                                                           
                                            "list":[                                                                   
                                                100000,                                                                  
                                            200000,                                                                  
                                            300000,                                                                  
                                            400000                                                                   
                                            ]                                                                          
                                        }                                                                            
                                        ],                                                                             
                                        "length":7                                                                     
                                    },                                                                               
                                    "aggregatorAs":5000,                                                             
                                    "aggregatorId":"10.20.30.40",                                                    
                                    "origin":"IGP",                                                                  
                                    "med":37,                                                                        
                                    "metric":37,                                                                     
                                    "valid":true,                                                                    
                                    "atomicAggregate":true,                                                          
                                    "bestpath":{                                                                     
                                        "overall":true                                                                 
                                    },                                                                               
                                    "community":{                                                                    
                                        "string":"1000:2000 3000:4000",                                                
                                        "list":[                                                                       
                                            "1000:2000",                                                                 
                                        "3000:4000"                                                                  
                                        ]                                                                              
                                    },                                                                               
                                    "extendedCommunity":{                                                            
                                        "string":"RT:10:3369254580 RT:20:2358681770"                                   
                                    },                                                                               
                                    "lastUpdate":{                                                                   
                                        "epoch":1495225039,                                                            
                                        "string":"Fri May 19 20:17:19 2017\n"                                          
                                    },                                                                               
                                    "nexthops":[                                                                     
                                    {                                                                              
                                        "ip":"10.10.10.1",                                                           
                                        "afi":"ipv4",                                                                
                                        "metric":0,                                                                  
                                        "accessible":true,                                                           
                                        "used":true                                                                  
                                    }                                                                              
                                    ],                                                                               
                                    "peer":{                                                                         
                                        "peerId":"10.10.10.1",                                                         
                                        "routerId":"25.98.0.1",                                                        
                                        "type":"external"                                                              
                                    }                                                                                
                                }                                                                                  
                                ]                                                                                    
                            }                                                                                      
                            ,                                                                                      
                                "70.80.91.0/24": {                                                                     
                                    "prefix":"70.80.91.0\/24",                                                           
                                    "advertisedTo":{                                                                     
                                        "10.10.10.1":{                                                                     
                                        },                                                                                 
                                        "11.1.1.1":{                                                                       
                                        }                                                                                  
                                    },                                                                                   
                                    "paths":[                                                                            
                                    {                                                                                  
                                        "pathId":0,                                                                      
                                        "aspath":{                                                                       
                                            "string":"200 {111,333,555,777} 222 444 666 888 {100000,200000,300000,400000}",
                                            "segments":[                                                                   
                                            {                                                                            
                                                "type":"as-sequence",                                                      
                                                "list":[                                                                   
                                                    200                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-set",                                                           
                                                "list":[                                                                   
                                                    111,                                                                     
                                                333,                                                                     
                                                555,                                                                     
                                                777                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-sequence",                                                      
                                                "list":[                                                                   
                                                    222,                                                                     
                                                444,                                                                     
                                                666,                                                                     
                                                888                                                                      
                                                ]                                                                          
                                            },                                                                           
                                            {                                                                            
                                                "type":"as-set",                                                           
                                                "list":[                                                                   
                                                    100000,                                                                  
                                                200000,                                                                  
                                                300000,                                                                  
                                                400000                                                                   
                                                ]                                                                          
                                            }                                                                            
                                            ],                                                                             
                                            "length":7                                                                     
                                        },                                                                               
                                        "aggregatorAs":5000,                                                             
                                        "aggregatorId":"10.20.30.40",                                                    
                                        "origin":"IGP",                                                                  
                                        "med":37,                                                                        
                                        "metric":37,                                                                     
                                        "valid":true,
                                        "atomicAggregate":true,
                                        "bestpath":{
                                            "overall":true
                                        },
                                        "community":{
                                            "string":"1000:2000 3000:4000",
                                            "list":[
                                                "1000:2000",
                                            "3000:4000"
                                            ]
                                        },
                                        "extendedCommunity":{
                                            "string":"RT:10:3369254580 RT:20:2358681770"
                                        },
                                        "lastUpdate":{
                                            "epoch":1495225039,
                                            "string":"Fri May 19 20:17:19 2017\n"
                                        },
                                        "nexthops":[
                                        {
                                            "ip":"10.10.10.1",
                                            "afi":"ipv4",
                                            "metric":0,
                                            "accessible":true,
                                            "used":true
                                        }
                                        ],
                                        "peer":{
                                            "peerId":"10.10.10.1",
                                            "routerId":"25.98.0.1",
                                            "type":"external"
                                        }
                                    }
                                    ]
                                }
                        }
                },
                {
                    "vrfId":0,                                                   
                    "vrfName":"default",                                         
                    "bgpTableVersion":0,                                         
                    "bgpLocalRouterId":"11.1.1.4",                               
                    "defaultLocPrf":100,                                         
                    "localAS":400,                                               
                    "receivedRoutes":{                                                   
                        "3001::\/64":{                                             
                            "addrPrefix":"3001::",                                   
                            "prefixLen":64,                                          
                            "prefix":"3001::\/64",                                   
                            "pathId":0,                                              
                            "nextHopGlobal":"2001::1",                               
                            "weight":0,                                              
                            "aspath":{                                               
                                "string":"100 {10000,20000} 30000 40000",              
                                "segments":[                                           
                                {                                                    
                                    "type":"as-sequence",                              
                                    "list":[                                           
                                        100                                              
                                    ]                                                  
                                },                                                   
                                {                                                    
                                    "type":"as-set",                                   
                                    "list":[                                           
                                        10000,                                           
                                    20000                                            
                                    ]                                                  
                                },                                                   
                                {                                                    
                                    "type":"as-sequence",                              
                                    "list":[                                           
                                        30000,                                           
                                    40000                                            
                                    ]                                                  
                                }                                                    
                                ],                                                     
                                "length":4                                             
                            },                                                       
                            "origin":"IGP",                                          
                            "lastUpdate":{                                           
                                "epoch":1495309393,                                    
                                "string":"Sat May 20 19:43:13 2017\n"                  
                            },                                                       
                            "appliedStatusSymbols":{                                 
                                "*":true,                                              
                                ">":true                                               
                            }                                                        
                        },                                                         
                        "3001:0:0:1::\/64":{                                       
                            "addrPrefix":"3001:0:0:1::",                             
                            "prefixLen":64,                                          
                            "prefix":"3001:0:0:1::\/64",                             
                            "pathId":0,                                              
                            "nextHopGlobal":"2001::1",                               
                            "weight":0,                                              
                            "aspath":{                                               
                                "string":"100 {10000,20000} 30000 40000",              
                                "segments":[                                           
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        100
                                    ]
                                },
                                {
                                    "type":"as-set",
                                    "list":[
                                        10000,
                                    20000
                                    ]
                                },
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        30000,
                                    40000
                                    ]
                                }
                                ],
                                "length":4
                            },
                            "origin":"IGP",
                            "lastUpdate":{
                                "epoch":1495309393,
                                "string":"Sat May 20 19:43:13 2017\n"
                            },
                            "appliedStatusSymbols":{
                                "*":true,
                                ">":true
                            }
                        }
                    },
                    "totalPrefixCounter":2,
                    "filteredPrefixCounter":0
                },
                {
                    "vrfId":0,                                                       
                    "vrfName":"default",                                             
                    "bgpTableVersion":6,                                             
                    "bgpLocalRouterId":"11.1.1.4",                                   
                    "defaultLocPrf":100,                                             
                    "localAS":400,                                                   
                    "advertisedRoutes":{                                                       
                        "3001::\/64":{                                                 
                            "addrPrefix":"3001::",                                       
                            "prefixLen":64,                                              
                            "prefix":"3001::\/64",                                       
                            "pathId":0,                                                  
                            "nextHopGlobal":"::",                                        
                            "weight":0,                                                  
                            "aspath":{                                                   
                                "string":"100 {10000,20000} 30000 40000",                  
                                "segments":[                                               
                                {                                                        
                                    "type":"as-sequence",                                  
                                    "list":[                                               
                                        100                                                  
                                    ]                                                      
                                },                                                       
                                {                                                        
                                    "type":"as-set",                                       
                                    "list":[                                               
                                        10000,                                               
                                    20000                                                
                                    ]                                                      
                                },                                                       
                                {                                                        
                                    "type":"as-sequence",                                  
                                    "list":[                                               
                                        30000,                                               
                                    40000                                                
                                    ]                                                      
                                }                                                        
                                ],                                                         
                                "length":4                                                 
                            },                                                           
                            "origin":"IGP",                                              
                            "lastUpdate":{                                               
                                "epoch":1495309738,                                        
                                "string":"Sat May 20 19:48:58 2017\n"                      
                            },                                                           
                            "appliedStatusSymbols":{                                     
                                "*":true,                                                  
                                ">":true                                                   
                            }                                                            
                        },                                                             
                        "3001:0:0:1::\/64":{                                           
                            "addrPrefix":"3001:0:0:1::",                                 
                            "prefixLen":64,                                              
                            "prefix":"3001:0:0:1::\/64",                                 
                            "pathId":0,                                                  
                            "nextHopGlobal":"::",                                        
                            "weight":0,                                                  
                            "aspath":{                                                   
                                "string":"100 {10000,20000} 30000 40000",                  
                                "segments":[                                               
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        100
                                    ]
                                },
                                {
                                    "type":"as-set",
                                    "list":[
                                        10000,
                                    20000
                                    ]
                                },
                                {
                                    "type":"as-sequence",
                                    "list":[
                                        30000,
                                    40000
                                    ]
                                }
                                ],
                                "length":4
                            },
                            "origin":"IGP",
                            "lastUpdate":{
                                "epoch":1495309738,
                                "string":"Sat May 20 19:48:58 2017\n"
                            },
                            "appliedStatusSymbols":{
                                "*":true,
                                ">":true
                            }
                        }
                    },
                    "totalPrefixCounter":2,
                    "filteredPrefixCounter":0
                }
                ]
            }`
    }

    return []byte(outJsonBlob)
}

func fake_rib_exec_vtysh_cmd (vtysh_cmd string, rib_type string) (map[string]interface{}, error) {
    var err error
    var outputJson map[string]interface{}

    if err = json.Unmarshal(fake_rib_json_output_blob (rib_type), &outputJson) ; err != nil {
        return nil, err
    }

    return outputJson, err
}

func init () {
    XlateFuncBind("YangToDb_network_instance_protocol_key_xfmr", YangToDb_network_instance_protocol_key_xfmr)
    XlateFuncBind("DbToYang_network_instance_protocol_key_xfmr", DbToYang_network_instance_protocol_key_xfmr)
    XlateFuncBind("YangToDb_bgp_gbl_tbl_key_xfmr", YangToDb_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_tbl_key_xfmr", DbToYang_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("YangToDb_bgp_gbl_afi_safi_field_xfmr", YangToDb_bgp_gbl_afi_safi_field_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_afi_safi_field_xfmr", DbToYang_bgp_gbl_afi_safi_field_xfmr)
	XlateFuncBind("YangToDb_bgp_dyn_neigh_listen_key_xfmr", YangToDb_bgp_dyn_neigh_listen_key_xfmr)
	XlateFuncBind("DbToYang_bgp_dyn_neigh_listen_key_xfmr", DbToYang_bgp_dyn_neigh_listen_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_gbl_afi_safi_key_xfmr", YangToDb_bgp_gbl_afi_safi_key_xfmr)
	XlateFuncBind("DbToYang_bgp_gbl_afi_safi_key_xfmr", DbToYang_bgp_gbl_afi_safi_key_xfmr) 
	XlateFuncBind("YangToDb_bgp_dyn_neigh_listen_field_xfmr", YangToDb_bgp_dyn_neigh_listen_field_xfmr)
	XlateFuncBind("DbToYang_bgp_dyn_neigh_listen_field_xfmr", DbToYang_bgp_dyn_neigh_listen_field_xfmr) 
    XlateFuncBind("YangToDb_bgp_global_subtree_xfmr", YangToDb_bgp_global_subtree_xfmr)
}

var YangToDb_bgp_gbl_afi_safi_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error

    log.Info("YangToDb_bgp_gbl_afi_safi_field_xfmr")
    rmap["NULL"] = "NULL"
    
    return rmap, err
}

var DbToYang_bgp_gbl_afi_safi_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_field_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")
	afi := ""

	switch mpathKey[1] {
	case "ipv4_unicast":
		afi = "IPV4_UNICAST"
	case "ipv6_unicast":
		afi = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afi = "L2VPN_EVPN"
	}

    rmap["afi-safi-name"] = afi

    return rmap, err
}

var YangToDb_bgp_dyn_neigh_listen_field_xfmr FieldXfmrYangToDb = func(inParams XfmrParams) (map[string]string, error) {
    rmap := make(map[string]string)
    var err error

    log.Info("YangToDb_bgp_dyn_neigh_listen_field_xfmr")
    rmap["NULL"] = "NULL"
    
    return rmap, err
}

var DbToYang_bgp_dyn_neigh_listen_field_xfmr FieldXfmrDbtoYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    
    entry_key := inParams.key
    log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: ", entry_key)

    dynKey := strings.Split(entry_key, "|")

    rmap["prefix"] = dynKey[1]

    return rmap, err
}

var YangToDb_network_instance_protocol_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    return "", nil 
}

var DbToYang_network_instance_protocol_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error

    pathInfo := NewPathInfo(inParams.uri)

    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")

    rmap["name"] = protoName; 
    rmap["identifier"] = bgpId; 
    return rmap, err
}

var YangToDb_bgp_gbl_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")

    if len(pathInfo.Vars) <  3 {
        return niName, errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return niName, errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return niName, errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return niName, errors.New("Protocol Name is missing")
    }

    log.Info("URI VRF ", niName)

    return niName, err
}

var DbToYang_bgp_gbl_tbl_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    var err error
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_tbl_key: ", entry_key)

    rmap["name"] = entry_key
    return rmap, err
}

var YangToDb_bgp_dyn_neigh_listen_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
	log.Info("YangToDb_bgp_dyn_neigh_listen_key_xfmr key: ", inParams.uri)

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
	prefix := pathInfo.Var("prefix")

    if len(pathInfo.Vars) < 4 {
        return "", errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return "", errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return "", errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return "", errors.New("Protocol Name is missing")
    }

	key := niName + "|" + prefix
	
	log.Info("YangToDb_bgp_dyn_neigh_listen_key_xfmr key: ", key)

    return key, nil
}

var DbToYang_bgp_dyn_neigh_listen_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: ", entry_key)

    dynKey := strings.Split(entry_key, "|")

    rmap["prefix"] = dynKey[1]

	log.Info("DbToYang_bgp_dyn_neigh_listen_key_xfmr: rmap:", rmap)
    return rmap, nil
}

var YangToDb_bgp_gbl_afi_safi_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {

    pathInfo := NewPathInfo(inParams.uri)

    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
	afName := pathInfo.Var("afi-safi-name")
	afi := ""
    var err error

    if len(pathInfo.Vars) < 4 {
        return afi, errors.New("Invalid Key length")
    }

    if len(niName) == 0 {
        return afi, errors.New("vrf name is missing")
    }

    if strings.Contains(bgpId,"BGP") == false {
        return afi, errors.New("BGP ID is missing")
    }
    
    if len(protoName) == 0 {
        return afi, errors.New("Protocol Name is missing")
    }

	if strings.Contains(afName, "IPV4_UNICAST") {
		afi = "ipv4_unicast"
	} else if strings.Contains(afName, "IPV6_UNICAST") {
		afi = "ipv6_unicast"
	} else if strings.Contains(afName, "L2VPN_EVPN") {
		afi = "l2vpn_evpn"
	} else {
		log.Info("Unsupported AFI type " + afName)
        return afi, errors.New("Unsupported AFI type " + afName)
	}

    if strings.Contains(afName, "IPV4_UNICAST") {
        afName = "IPV4_UNICAST"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV4_UNICAST supported only on ipv4-config container")
		    log.Info("IPV4_UNICAST supported only on ipv4-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "IPV6_UNICAST") {
        afName = "IPV6_UNICAST"
        if strings.Contains(inParams.uri, "ipv4-unicast") ||
           strings.Contains(inParams.uri, "l2vpn-evpn") {
		    err = errors.New("IPV6_UNICAST supported only on ipv6-config container")
		    log.Info("IPV6_UNICAST supported only on ipv6-config container: ", afName);
		    return afName, err
        }
    } else if strings.Contains(afName, "L2VPN_EVPN") {
        afName = "L2VPN_EVPN"
        if strings.Contains(inParams.uri, "ipv6-unicast") ||
           strings.Contains(inParams.uri, "ipv4-unicast") {
		    err = errors.New("L2VPN_EVPN supported only on l2vpn-evpn container")
		    log.Info("L2VPN_EVPN supported only on l2vpn-evpn container: ", afName);
		    return afName, err
        }
    } else  {
	    err = errors.New("Unsupported AFI SAFI")
	    log.Info("Unsupported AFI SAFI ", afName);
	    return afName, err
    }

	key := niName + "|" + afi
	
	log.Info("AFI key: ", key)

    return key, nil
}

var DbToYang_bgp_gbl_afi_safi_key_xfmr KeyXfmrDbToYang = func(inParams XfmrParams) (map[string]interface{}, error) {
    rmap := make(map[string]interface{})
    entry_key := inParams.key
    log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: ", entry_key)

    mpathKey := strings.Split(entry_key, "|")
	afi := ""

	switch mpathKey[1] {
	case "ipv4_unicast":
		afi = "IPV4_UNICAST"
	case "ipv6_unicast":
		afi = "IPV6_UNICAST"
	case "l2vpn_evpn":
		afi = "L2VPN_EVPN"
	}

    rmap["afi-safi-name"] = afi

	log.Info("DbToYang_bgp_gbl_afi_safi_key_xfmr: rmap:", rmap)
    return rmap, nil
}

var YangToDb_bgp_global_subtree_xfmr SubTreeXfmrYangToDb = func(inParams XfmrParams) (map[string]map[string]db.Value, error) {
    var err error
	log.Info("YangToDb_bgp_global_subtree_xfmr:", inParams.oper)
    if inParams.oper == DELETE {
        return nil, errors.New("Invalid request")
    }
    return nil, err
}
