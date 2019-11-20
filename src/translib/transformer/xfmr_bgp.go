package transformer

import (
    "errors"
    "encoding/json"
    "translib/ocbinds"
    "os/exec"
    log "github.com/golang/glog"
)

func getBgpRoot (inParams XfmrParams) (*ocbinds.OpenconfigNetworkInstance_NetworkInstances_NetworkInstance_Protocols_Protocol_Bgp, string, error) {
    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    bgpId := pathInfo.Var("identifier")
    protoName := pathInfo.Var("name#2")
    var err error

    if len(niName) == 0 {
        return nil, "", errors.New("Network-instance-name missing")
    }

    if bgpId != "BGP" {
        return nil, "", errors.New("Protocol-id is not BGP!! Incoming Protocol-id:" + bgpId)
    }

    if len(protoName) == 0 {
        return nil, "", errors.New("Network-instance Protocol-name missing")
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

func fake_rib_exec_vtysh_cmd (vtysh_cmd string) (map[string]interface{}, error) {
    var err error
    var outputJson map[string]interface{}
//    outJsonBlob := `{
//        "vrfId": 0,
//        "vrfName": "default",
//        "tableVersion": 54,
//        "routerId": "200.9.0.4",
//        "defaultLocPrf": 100,
//        "localAS": 400,
//        "routes": {
//            "4.4.4.4\/32": {
//                "prefix":"4.4.4.4\/32",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"Local",
//                        "segments":[
//                        ],
//                        "length":0
//                    },
//                    "origin":"incomplete",
//                    "med":0,
//                    "metric":0,
//                    "weight":32768,
//                    "valid":true,
//                    "sourced":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011680,
//                        "string":"Fri May  5 19:14:40 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"0.0.0.0",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"0.0.0.0",
//                        "routerId":"200.9.0.4"
//                    }
//                }
//                ]
//            },
//            "10.10.10.0\/24": {
//                "prefix":"10.10.10.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"Local",
//                        "segments":[
//                        ],
//                        "length":0
//                    },
//                    "origin":"incomplete",
//                    "med":0,
//                    "metric":0,
//                    "weight":32768,
//                    "valid":true,
//                    "sourced":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011680,
//                        "string":"Fri May  5 19:14:40 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"0.0.0.0",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"0.0.0.0",
//                        "routerId":"200.9.0.4"
//                    }
//                }
//                ]
//            },
//            "10.59.128.0\/20": {
//                "prefix":"10.59.128.0\/20",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"Local",
//                        "segments":[
//                        ],
//                        "length":0
//                    },
//                    "origin":"incomplete",
//                    "med":0,
//                    "metric":0,
//                    "weight":32768,
//                    "valid":true,
//                    "sourced":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011680,
//                        "string":"Fri May  5 19:14:40 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"0.0.0.0",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"0.0.0.0",
//                        "routerId":"200.9.0.4"
//                    }
//                }
//                ]
//            },
//            "69.10.20.0\/24": {
//                "prefix":"69.10.20.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.21.0\/24": {
//                "prefix":"69.10.21.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                        ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                ],
//                "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.22.0\/24": {
//                "prefix":"69.10.22.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.23.0\/24": {
//                "prefix":"69.10.23.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.24.0\/24": {
//                "prefix":"69.10.24.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.25.0\/24": {
//                "prefix":"69.10.25.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.26.0\/24": {
//                "prefix":"69.10.26.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.27.0\/24": {
//                "prefix":"69.10.27.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.28.0\/24": {
//                "prefix":"69.10.28.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.29.0\/24": {
//                "prefix":"69.10.29.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.30.0\/24": {
//                "prefix":"69.10.30.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.31.0\/24": {
//                "prefix":"69.10.31.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.32.0\/24": {
//                "prefix":"69.10.32.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.33.0\/24": {
//                "prefix":"69.10.33.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.34.0\/24": {
//                "prefix":"69.10.34.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            },
//            "69.10.35.0\/24": {
//                "prefix":"69.10.35.0\/24",
//                "advertisedTo":{
//                    "10.10.10.1":{
//                    }
//                },
//                "paths":[
//                {
//                    "aspath":{
//                        "string":"200 {100,300,500}",
//                        "segments":[
//                        {
//                            "type":"as-sequence",
//                            "list":[
//                                200
//                            ]
//                        },
//                        {
//                            "type":"as-set",
//                            "list":[
//                                100,
//                            300,
//                            500
//                            ]
//                        }
//                        ],
//                        "length":2
//                    },
//                    "origin":"IGP",
//                    "valid":true,
//                    "bestpath":{
//                        "overall":true
//                    },
//                    "community":{
//                        "string":"500:700",
//                        "list":[
//                            "500:700"
//                        ]
//                    },
//                    "lastUpdate":{
//                        "epoch":1494011770,
//                        "string":"Fri May  5 19:16:10 2017\n"
//                    },
//                    "nexthops":[
//                    {
//                        "ip":"10.10.10.1",
//                        "afi":"ipv4",
//                        "metric":0,
//                        "accessible":true,
//                        "used":true
//                    }
//                    ],
//                    "peer":{
//                        "peerId":"10.10.10.1",
//                        "routerId":"25.98.0.1",
//                        "type":"external"
//                    }
//                }
//                ]
//            }
//        }
//    }`

    outJsonBlob := `{
        "vrfId": 0,
        "vrfName": "default",
        "tableVersion": 54,
        "routerId": "200.9.0.4",
        "defaultLocPrf": 100,
        "localAS": 400,
        "routes": {
            "4.4.4.4\/32": {
                "prefix":"4.4.4.4\/32",
                "advertisedTo":{
                    "10.10.10.1":{
                    }
                },
                "paths":{
                    "1" : {
                        "aspath":{
                            "string":"Local",
                            "segments":[
                            ],
                            "length":0
                        },
                        "origin":"incomplete",
                        "med":0,
                        "metric":0,
                        "weight":32768,
                        "valid":true,
                        "sourced":true,
                        "bestpath":{
                            "overall":true
                        },
                        "lastUpdate":{
                            "epoch":1494011680,
                            "string":"Fri May  5 19:14:40 2017\n"
                        },
                        "nexthops":[
                        {
                            "ip":"0.0.0.0",
                            "afi":"ipv4",
                            "metric":0,
                            "accessible":true,
                            "used":true
                        }
                        ],
                        "peer":{
                            "peerId":"0.0.0.0",
                            "routerId":"200.9.0.4"
                        }
                    },
                    "2" : {
                        "aspath":{
                            "string":"200 {100,300,500}",
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
                                    100,
                                300,
                                500
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
                        "community":{
                            "string":"500:700",
                            "list":[
                                "500:700"
                            ]
                        },
                        "lastUpdate":{
                            "epoch":1494011770,
                            "string":"Fri May  5 19:16:10 2017\n"
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
                }
            },
            "69.10.30.0\/24": {
                "prefix":"69.10.30.0\/24",
                "advertisedTo":{
                    "10.10.10.1":{
                    }
                },
                "paths":{
                    "1" : {
                        "aspath":{
                            "string":"200 {100,300,500}",
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
                                    100,
                                300,
                                500
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
                        "community":{
                            "string":"500:700",
                            "list":[
                                "500:700"
                            ]
                        },
                        "lastUpdate":{
                            "epoch":1494011770,
                            "string":"Fri May  5 19:16:10 2017\n"
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
                            "peerId":"10.10.10.2",
                            "routerId":"25.98.0.1",
                            "type":"external"
                        }
                    }
                }
            }
        }
    }`

    if err = json.Unmarshal([]byte(outJsonBlob), &outputJson) ; err != nil {
        return nil, err
    }

    return outputJson, err
}

func init () {
    XlateFuncBind("YangToDb_bgp_gbl_tbl_key_xfmr", YangToDb_bgp_gbl_tbl_key_xfmr)
    XlateFuncBind("DbToYang_bgp_gbl_tbl_key_xfmr", DbToYang_bgp_gbl_tbl_key_xfmr)
}

var YangToDb_bgp_gbl_tbl_key_xfmr KeyXfmrYangToDb = func(inParams XfmrParams) (string, error) {
    var err error

    pathInfo := NewPathInfo(inParams.uri)
    niName := pathInfo.Var("name")
    protoName := pathInfo.Var("name#2")

    if protoName != "bgp" {
        return niName, errors.New("Invalid protocol name : " + protoName)
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
