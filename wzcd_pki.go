package wzcd

import (
	"fmt"
	"io/ioutil"

	wzlib_database "github.com/infra-whizz/wzlib/database"
	wzlib_database_controller "github.com/infra-whizz/wzlib/database/controller"
)

type WzcPKIManager struct {
	db *wzlib_database.WzDBH
}

func NewWzcPKIManager() *WzcPKIManager {
	wpm := new(WzcPKIManager)
	return wpm
}

// SetDbh sets a database connectivity
func (wpm *WzcPKIManager) SetDbh(dbh *wzlib_database.WzDBH) *WzcPKIManager {
	if wpm.db == nil {
		wpm.db = dbh
	}
	return wpm
}

// RegisterPEMKey from the file on the system and tie up to a specific systemid and fqdn
func (wpm *WzcPKIManager) RegisterPEMKey(filePath string, systemid string, fqdn string) error {
	keypem, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if systemid == "" {
		return fmt.Errorf("Machine ID is not defined")
	}

	if fqdn == "" {
		return fmt.Errorf("FQDN of the machine required. NOTE: it should be the same FQDN as the remote seeing it.")
	}

	return wpm.db.GetControllerAPI().GetKeysAPI().AddRSAPublicPEM(
		keypem, systemid, fqdn, wzlib_database_controller.OWNER_APP_REMOTE)
}

// RemovePEMKey from the database by a fingerprint
func (wpm *WzcPKIManager) RemovePEMKey(fingerprint string) error {
	return wpm.db.GetControllerAPI().GetKeysAPI().RemoveRSAPublicPEM(fingerprint)
}

func (wpm *WzcPKIManager) ListRemotePEMKeys() {
	keys := wpm.db.GetControllerAPI().GetKeysAPI().ListRSAPublicPEMByOwner(wzlib_database_controller.OWNER_APP_REMOTE)
	if len(keys) > 0 {
		fmt.Println("Below are hosts and their PEM keys. They are used to issue signed commands to the entire Whizz cluster.")
		fmt.Println("List of keys:")
		for idx, k := range keys {
			fmt.Printf("  %d. %s\n\t  System ID: %s\n\tFingerprint: %s\n\n", idx+1, k.Fqdn, k.MachineId, k.RsaFp)
		}
	} else {
		fmt.Println("No keys registered at the moment.")
	}
}
