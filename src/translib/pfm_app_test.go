package translib

import (
	"errors"
	"fmt"
	"testing"
	db "translib/db"
)

// TLV EEPROM Data Types
const (
	_TLV_CODE_PRODUCT_NAME   = "0x21"
	_TLV_CODE_PART_NUMBER    = "0x22"
	_TLV_CODE_SERIAL_NUMBER  = "0x23"
	_TLV_CODE_MAC_BASE       = "0x24"
	_TLV_CODE_MANUF_DATE     = "0x25"
	_TLV_CODE_DEVICE_VERSION = "0x26"
	_TLV_CODE_LABEL_REVISION = "0x27"
	_TLV_CODE_PLATFORM_NAME  = "0x28"
	_TLV_CODE_ONIE_VERSION   = "0x29"
	_TLV_CODE_MAC_SIZE       = "0x2A"
	_TLV_CODE_MANUF_NAME     = "0x2B"
	_TLV_CODE_MANUF_COUNTRY  = "0x2C"
	_TLV_CODE_VENDOR_NAME    = "0x2D"
	_TLV_CODE_DIAG_VERSION   = "0x2E"
	_TLV_CODE_SERVICE_TAG    = "0x2F"
	_TLV_CODE_VENDOR_EXT     = "0xFD"
	_TLV_CODE_CRC_32         = "0xFE"
)

const (
	TEST_PRODUCT_NAME  = "6776-64X-O-AC-F"
	TEST_PART_NUMBER   = "FP123454321PF"
	TEST_PLATFORM_NAME = "x86_64-pfm_test-platform"
	TEST_SERVICE_TAG   = "6776X6776"
	TEST_MANUF_NAME    = "TestManufacture"
)

type EepromEntry struct {
	TlvType string
	Name    string
	Value   string
}

var testStateDbList = [...]EepromEntry{
	{_TLV_CODE_PRODUCT_NAME, "Product Name", TEST_PRODUCT_NAME},
	{_TLV_CODE_PLATFORM_NAME, "Platform Name", TEST_PLATFORM_NAME},
	{_TLV_CODE_SERVICE_TAG, "Service Tag", TEST_SERVICE_TAG},
	{_TLV_CODE_MANUF_NAME, "Manufacturer", TEST_MANUF_NAME},
	{_TLV_CODE_PART_NUMBER, "Part Number", TEST_PART_NUMBER},
}

func init() {
	fmt.Println("+++++  Init pfm_app_test  +++++")

	if err := clearPfmDataFromDb(); err == nil {
		fmt.Println("+++++  Removed All Platform Data from Db  +++++")
	} else {
		fmt.Printf("Failed to remove All Platform Data from Db: %v", err)
	}
}

// This will test GET on /openconfig-platform:components
func Test_PfmApp_TopLevelPath(t *testing.T) {
	url := "/openconfig-platform:components"

	t.Run("Default_Response_Top_Level", processGetRequest(url, bulkPfmShowDefaultResponse, false))

	//Set the factory DB with pre-defined EEPROM_INFO entry
	if err := createPfmFactoryDb(); err != nil {
		fmt.Printf("Failed to add Platform Data to Db: %v", err)
	}

	t.Run("Get_Full_Pfm_Tree_Top_Level", processGetRequest(url, bulkPfmShowAllJsonResponse, false))
}

// THis will delete Platform Table from DB
func clearPfmDataFromDb() error {
	var err error
	eepromTable := db.TableSpec{Name: "EEPROM_INFO"}

	d := getStateDB()
	if d == nil {
		err = errors.New("Failed to connect to state Db")
		return err
	}
	if err = d.DeleteTable(&eepromTable); err != nil {
		err = errors.New("Failed to delete Eeprom Table")
		return err
	}
	return err
}

func createPfmFactoryDb() error {
	var err error
	eepromTable := db.TableSpec{Name: "EEPROM_INFO"}

	d := getStateDB()
	if d == nil {
		err = errors.New("Failed to connect to state Db")
		return err
	}

	for _, dbItem := range testStateDbList {
		ca := make([]string, 1, 1)
		ca[0] = dbItem.TlvType

		akey := db.Key{Comp: ca}
		avalue := db.Value{Field: map[string]string{
			"Name":  dbItem.Name,
			"Value": dbItem.Value,
		},
		}
		d.SetEntry(&eepromTable, akey, avalue)
	}
	return err
}

func getStateDB() *db.DB {
	stateDb, _ := db.NewDB(db.Options{
		DBNo:               db.StateDB,
		InitIndicator:      "STATE_DB_INITIALIZED",
		TableNameSeparator: "|",
		KeySeparator:       "|",
	})

	return stateDb
}

/***************************************************************************/
///////////                  JSON Data for Tests              ///////////////
/***************************************************************************/
var bulkPfmShowDefaultResponse string = "{\"openconfig-platform:components\":{\"component\":[{\"name\":\"System Eeprom\",\"state\":{\"empty\":false,\"location\":\"Slot 1\",\"name\":\"System Eeprom\",\"oper-status\":\"openconfig-platform-types:ACTIVE\",\"removable\":false}}]}}"

var bulkPfmShowAllJsonResponse string = "{\"openconfig-platform:components\":{\"component\":[{\"name\":\"System Eeprom\",\"state\":{\"description\":\"" + TEST_PLATFORM_NAME + "\",\"empty\":false,\"id\":\"" + TEST_PRODUCT_NAME + "\",\"location\":\"Slot 1\",\"mfg-name\":\"" + TEST_MANUF_NAME + "\",\"name\":\"System Eeprom\",\"oper-status\":\"openconfig-platform-types:ACTIVE\",\"part-no\":\"" + TEST_PART_NUMBER + "\",\"removable\":false,\"serial-no\":\"" + TEST_SERVICE_TAG + "\"}}]}}"
