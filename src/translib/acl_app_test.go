package translib

import (
    "fmt"
    "errors"
    "rest/server"
    "swagger"
    "strings"
    "testing"
    "net/http"
    "net/http/httptest"
    "github.com/gorilla/mux"
    "os"
    //"io/ioutil"
    db "translib/db"
)

var router *mux.Router

func init() {
    swagger.Load()
    router = server.NewRouter()
    fmt.Println("+++++  Init acl_app_test  +++++")
}

func TestMain(m *testing.M) {
    if err := clearAclDataFromDb(); err != nil {
        os.Exit(-1)
    }
    fmt.Println("+++++  Removed All Acl Data from Db  +++++")

    ret := m.Run()

    /*if err := clearAclDataFromDb(); err != nil {
        os.Exit(-1)
    }*/

    os.Exit(ret)
}

// This will test GET on /openconfig-acl:acl
func TestTopLevelPath(t *testing.T) {
    url := "/restconf/data/openconfig-acl:acl"

    t.Run("EmptyResponseTopLevel", testGetData(url, emptyJson, http.StatusOK))

    t.Run("PostTopLevel", testSetData(url, requestBulkPostJson, "POST", http.StatusCreated))

    t.Run("GetTopLevel", testGetData(url, responseBulkPostJson, http.StatusOK))

    // Delete all bindings before deleting at top level
    t.Run("DeleteAllBindingsTopLevel", testDeleteData("/restconf/data/openconfig-acl:acl/interfaces"))
    t.Run("DeleteTopLevel", testDeleteData(url))

    t.Run("VerifyTopLevelDelete", testGetData(url, emptyJson, http.StatusOK))
}

func TestSingleAclOperations(t *testing.T) {
    url := "/restconf/data/openconfig-acl:acl/acl-sets/acl-set=MyACL3,ACL_IPV4"

    t.Run("PatchSingleAcl", testSetData(url, requestOneAclPatchJson, "PATCH", http.StatusNoContent))

    t.Run("GetSingleAcl", testGetData(url, responseOneAclJson, http.StatusOK))

    t.Run("DeleteSingleAcl", testDeleteData(url))

    t.Run("VerifySingleAclDelete", testGetData(url, notFoundErrorJson, http.StatusNotFound))
}

func TestSingleRuleOperations(t *testing.T) {
    aclUrl := "/restconf/data/openconfig-acl:acl/acl-sets/acl-set=MyACL5,ACL_IPV4"
    ruleUrl := "/restconf/data/openconfig-acl:acl/acl-sets/acl-set=MyACL5,ACL_IPV4/acl-entries/acl-entry=8"

    t.Run("PostSingleAcl", testSetData(aclUrl, requestOneAclPostJson, "POST", http.StatusCreated))
    t.Run("GetSingleAcl", testGetData(aclUrl, responseOneAclPostJson, http.StatusOK))

    t.Run("PostSingleRule", testSetData(ruleUrl, requestOneRulePostJson, "POST", http.StatusCreated))

    t.Run("GetSingleRule", testGetData(ruleUrl, responseOneRuleJson, http.StatusOK))

    t.Run("DeleteSingleRule", testDeleteData(ruleUrl))
    t.Run("VerifySingleRuleDelete", testGetData(ruleUrl, notFoundErrorJson, http.StatusNotFound))

    t.Run("DeleteSingleAcl", testDeleteData(aclUrl))
    t.Run("VerifySingleAclDelete", testGetData(aclUrl, notFoundErrorJson, http.StatusNotFound))
}

func testGetData(url string, expectedRespJson string, expHttpStatus int) func(*testing.T) {
    return func(t *testing.T) {
        req := httptest.NewRequest("GET", url, nil)
        resp := executeRequest(req)
        checkResponseReturnStatus(t, expHttpStatus, resp.Code)

        /*err := ioutil.WriteFile("TmpResp.json", resp.Body.Bytes(), 0644)
        if err != nil {
            fmt.Println(err)
        }*/
        /*jsonRespVal, err := ioutil.ReadFile(expectedRespJsonFileName)
        if err != nil {
            t.Errorf("Error occured reading file: %s. Error: %v", expectedRespJsonFileName, err)
        }*/
        if resp != nil && resp.Body.String() != expectedRespJson {
            t.Errorf("Response received not matching with expected for Url: %s", url)
        }
    }
}

func testSetData(url string, jsonPayload string, oper string, expHttpStatus int) func(*testing.T) {
    return func(t *testing.T) {
        /*jsonVal, err := ioutil.ReadFile(jsonFileName)
        if err != nil {
            t.Errorf("Error occured reading file: %s. Error: %v", jsonFileName, err)
        }*/
        //fmt.Println("Set Data from File: \n" + string(jsonVal))
        //req := httptest.NewRequest(oper, url, bytes.NewReader(jsonVal))
        req := httptest.NewRequest(oper, url, strings.NewReader(jsonPayload))
        req.Header.Set("Content-Type", "application/yang-data+json")
        req.Header.Set("accept", "application/yang-data+json")
        resp := executeRequest(req)
        checkResponseReturnStatus(t, expHttpStatus, resp.Code)
    }
}

func testDeleteData(url string) func(*testing.T) {
    return func(t *testing.T) {
        req := httptest.NewRequest("DELETE", url, nil)
        resp := executeRequest(req)
        checkResponseReturnStatus(t, http.StatusNoContent, resp.Code)
    }
}

// THis will delete ACL table and Rules Table from DB
func clearAclDataFromDb() error {
    var err error
    ruleTable := db.TableSpec{Name: "ACL_RULE"}
    aclTable := db.TableSpec{Name: "ACL_TABLE"}

    d := getConfigDb()
    if d == nil {
        err = errors.New("Failed to connect to config Db")
        return err
    }
    if err = d.DeleteTable(&ruleTable); err != nil {
        err = errors.New("Failed to delete Rules Table")
        return err
    }
    if err = d.DeleteTable(&aclTable); err != nil {
        err = errors.New("Failed to delete Acl Table")
        return err
    }
    return err
}

/*******  These are utilities functions  ********/
func getConfigDb() *db.DB {
    configDb,_ := db.NewDB(db.Options {
        DBNo              : db.ConfigDB,
        InitIndicator     : "CONFIG_DB_INITIALIZED",
        TableNameSeparator: "|",
        KeySeparator      : "|",
    })

    return configDb
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    response := httptest.NewRecorder()
    router.ServeHTTP(response, req)

    return response
}

func checkResponseReturnStatus(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/
var emptyJson string = "{}"

var notFoundErrorJson string = "{\"ietf-restconf:errors\":{\"error\":[{\"error-type\":\"application\",\"error-tag\":\"invalid-value\",\"error-message\":\"Entry not found\"}]}}"

var requestBulkPostJson string = "{\"acl-sets\":{\"acl-set\":[{\"name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"config\":{\"name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"description\":\"Description for MyACL1\"},\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"config\":{\"sequence-id\":1,\"description\":\"Description for MyACL1 Rule Seq 1\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.1/32\",\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":101,\"destination-port\":201}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":2,\"config\":{\"sequence-id\":2,\"description\":\"Description for MyACL1 Rule Seq 2\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.2/32\",\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":102,\"destination-port\":202}},\"actions\":{\"config\":{\"forwarding-action\":\"DROP\"}}},{\"sequence-id\":3,\"config\":{\"sequence-id\":3,\"description\":\"Description for MyACL1 Rule Seq 3\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.3/32\",\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":103,\"destination-port\":203}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":4,\"config\":{\"sequence-id\":4,\"description\":\"Description for MyACL1 Rule Seq 4\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.4/32\",\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":104,\"destination-port\":204}},\"actions\":{\"config\":{\"forwarding-action\":\"DROP\"}}},{\"sequence-id\":5,\"config\":{\"sequence-id\":5,\"description\":\"Description for MyACL1 Rule Seq 5\"},\"ipv4\":{\"config\":{\"source-address\":\"11.1.1.5/32\",\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":105,\"destination-port\":205}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}}]}},{\"name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"config\":{\"name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"description\":\"Description for MyACL2\"},\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"config\":{\"sequence-id\":1,\"description\":\"Description for Rule Seq 1\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.1/32\",\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":101,\"destination-port\":201}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":2,\"config\":{\"sequence-id\":2,\"description\":\"Description for Rule Seq 2\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.2/32\",\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":102,\"destination-port\":202}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":3,\"config\":{\"sequence-id\":3,\"description\":\"Description for Rule Seq 3\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.3/32\",\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":103,\"destination-port\":203}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":4,\"config\":{\"sequence-id\":4,\"description\":\"Description for Rule Seq 4\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.4/32\",\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":104,\"destination-port\":204}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}},{\"sequence-id\":5,\"config\":{\"sequence-id\":5,\"description\":\"Description for Rule Seq 5\"},\"ipv4\":{\"config\":{\"source-address\":\"12.1.1.5/32\",\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"IP_TCP\"}},\"transport\":{\"config\":{\"source-port\":105,\"destination-port\":205}},\"actions\":{\"config\":{\"forwarding-action\":\"ACCEPT\"}}}]}}]},\"interfaces\":{\"interface\":[{\"id\":\"Ethernet0\",\"config\":{\"id\":\"Ethernet0\"},\"interface-ref\":{\"config\":{\"interface\":\"Ethernet0\"}},\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"set-name\":\"MyACL1\",\"type\":\"ACL_IPV4\",\"config\":{\"set-name\":\"MyACL1\",\"type\":\"ACL_IPV4\"}}]}},{\"id\":\"Ethernet4\",\"config\":{\"id\":\"Ethernet4\"},\"interface-ref\":{\"config\":{\"interface\":\"Ethernet4\"}},\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"set-name\":\"MyACL2\",\"type\":\"ACL_IPV4\",\"config\":{\"set-name\":\"MyACL2\",\"type\":\"ACL_IPV4\"}}]}}]}}" 

var responseBulkPostJson string = "{\"openconfig-acl:acl\":{\"acl-sets\":{\"acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"},\"state\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.2/32\"},\"state\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"},\"state\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"},\"state\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"},\"state\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL1\",\"name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL1\",\"state\":{\"description\":\"Description for MyACL1\",\"name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"},{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.1/32\"},\"state\":{\"destination-address\":\"22.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.2/32\"},\"state\":{\"destination-address\":\"22.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.3/32\"},\"state\":{\"destination-address\":\"22.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.4/32\"},\"state\":{\"destination-address\":\"22.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.5/32\"},\"state\":{\"destination-address\":\"22.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"12.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL2\",\"state\":{\"description\":\"Description for MyACL2\",\"name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"interfaces\":{\"interface\":[{\"config\":{\"id\":\"Ethernet0\"},\"id\":\"Ethernet0\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1}},{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}},{\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3}},{\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4}},{\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5}}]},\"config\":{\"set-name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL1\",\"state\":{\"set-name\":\"MyACL1\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"state\":{\"id\":\"Ethernet0\"}},{\"config\":{\"id\":\"Ethernet4\"},\"id\":\"Ethernet4\",\"ingress-acl-sets\":{\"ingress-acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1}},{\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2}},{\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3}},{\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4}},{\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5}}]},\"config\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"set-name\":\"MyACL2\",\"state\":{\"set-name\":\"MyACL2\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]},\"state\":{\"id\":\"Ethernet4\"}}]}}}"


var requestOneAclPatchJson string = "{ \"name\": \"MyACL3\", \"type\": \"ACL_IPV4\", \"config\": { \"name\": \"MyACL3\", \"type\": \"ACL_IPV4\", \"description\": \"Description for MyACL3\" }, \"acl-entries\": { \"acl-entry\": [ { \"sequence-id\": 1, \"config\": { \"sequence-id\": 1, \"description\": \"Description for MyACL3 Rule Seq 1\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.1/32\", \"destination-address\": \"21.1.1.1/32\", \"dscp\": 1, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 101, \"destination-port\": 201 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } }, { \"sequence-id\": 2, \"config\": { \"sequence-id\": 2, \"description\": \"Description for MyACL3 Rule Seq 2\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.2/32\", \"destination-address\": \"21.1.1.2/32\", \"dscp\": 2, \"protocol\": \"IP_UDP\" } }, \"transport\": { \"config\": { \"source-port\": 102, \"destination-port\": 202 } }, \"actions\": { \"config\": { \"forwarding-action\": \"DROP\" } } }, { \"sequence-id\": 3, \"config\": { \"sequence-id\": 3, \"description\": \"Description for MyACL3 Rule Seq 3\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.3/32\", \"destination-address\": \"21.1.1.3/32\", \"dscp\": 3, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 103, \"destination-port\": 203 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } }, { \"sequence-id\": 4, \"config\": { \"sequence-id\": 4, \"description\": \"Description for MyACL3 Rule Seq 4\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.4/32\", \"destination-address\": \"21.1.1.4/32\", \"dscp\": 4, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 104, \"destination-port\": 204 } }, \"actions\": { \"config\": { \"forwarding-action\": \"DROP\" } } }, { \"sequence-id\": 5, \"config\": { \"sequence-id\": 5, \"description\": \"Description for MyACL3 Rule Seq 5\" }, \"ipv4\": { \"config\": { \"source-address\": \"11.1.1.5/32\", \"destination-address\": \"21.1.1.5/32\", \"dscp\": 5, \"protocol\": \"IP_TCP\" } }, \"transport\": { \"config\": { \"source-port\": 105, \"destination-port\": 205 } }, \"actions\": { \"config\": { \"forwarding-action\": \"ACCEPT\" } } } ] }}"

var responseOneAclJson string = "{\"openconfig-acl:acl-set\":[{\"acl-entries\":{\"acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":1},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"},\"state\":{\"destination-address\":\"21.1.1.1/32\",\"dscp\":1,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.1/32\"}},\"sequence-id\":1,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":1},\"transport\":{\"config\":{\"destination-port\":201,\"source-port\":101}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":2},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_UDP\",\"source-address\":\"11.1.1.2/32\"},\"state\":{\"destination-address\":\"21.1.1.2/32\",\"dscp\":2,\"protocol\":\"openconfig-packet-match-types:IP_UDP\",\"source-address\":\"11.1.1.2/32\"}},\"sequence-id\":2,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":2},\"transport\":{\"config\":{\"destination-port\":202,\"source-port\":102}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":3},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"},\"state\":{\"destination-address\":\"21.1.1.3/32\",\"dscp\":3,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.3/32\"}},\"sequence-id\":3,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":3},\"transport\":{\"config\":{\"destination-port\":203,\"source-port\":103}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:DROP\"}},\"config\":{\"sequence-id\":4},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"},\"state\":{\"destination-address\":\"21.1.1.4/32\",\"dscp\":4,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.4/32\"}},\"sequence-id\":4,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":4},\"transport\":{\"config\":{\"destination-port\":204,\"source-port\":104}}},{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":5},\"ipv4\":{\"config\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"},\"state\":{\"destination-address\":\"21.1.1.5/32\",\"dscp\":5,\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"11.1.1.5/32\"}},\"sequence-id\":5,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":5},\"transport\":{\"config\":{\"destination-port\":205,\"source-port\":105}}}]},\"config\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL3\",\"state\":{\"description\":\"Description for MyACL3\",\"name\":\"MyACL3\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var requestOneAclPostJson string = "{\"config\": {\"name\": \"MyACL5\",\"type\": \"ACL_IPV4\",\"description\": \"Description for MyACL5\"}}"
var responseOneAclPostJson string = "{\"openconfig-acl:acl-set\":[{\"config\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"name\":\"MyACL5\",\"state\":{\"description\":\"Description for MyACL5\",\"name\":\"MyACL5\",\"type\":\"openconfig-acl:ACL_IPV4\"},\"type\":\"openconfig-acl:ACL_IPV4\"}]}"

var requestOneRulePostJson string = "{\"sequence-id\": 8,\"config\": {\"sequence-id\": 8,\"description\": \"Description for MyACL5 Rule Seq 8\"},\"ipv4\": {\"config\": {\"source-address\": \"4.4.4.4/24\",\"destination-address\": \"5.5.5.5/24\",\"protocol\": \"IP_TCP\"}},\"transport\": {\"config\": {\"source-port\": 101,\"destination-port\": 100,\"tcp-flags\": [\"TCP_FIN\",\"TCP_ACK\"]}},\"actions\": {\"config\": {\"forwarding-action\": \"ACCEPT\"}}}"

var responseOneRuleJson string = "{\"openconfig-acl:acl-entry\":[{\"actions\":{\"config\":{\"forwarding-action\":\"openconfig-acl:ACCEPT\"}},\"config\":{\"sequence-id\":8},\"ipv4\":{\"config\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"},\"state\":{\"destination-address\":\"5.5.5.5/24\",\"protocol\":\"openconfig-packet-match-types:IP_TCP\",\"source-address\":\"4.4.4.4/24\"}},\"sequence-id\":8,\"state\":{\"matched-octets\":\"0\",\"matched-packets\":\"0\",\"sequence-id\":8},\"transport\":{\"config\":{\"destination-port\":100,\"source-port\":101,\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]},\"state\":{\"tcp-flags\":[\"openconfig-packet-match-types:TCP_FIN\",\"openconfig-packet-match-types:TCP_ACK\"]}}}]}"
