package wzcd

import (
	"io/ioutil"

	wzlib_database "github.com/infra-whizz/wzlib/database"
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

	return wpm.db.GetControllerAPI().GetKeysAPI().AddRSAPublicPEM(keypem, systemid, fqdn)
}

// RemovePEMKey from the database by a fingerprint
func (wpm *WzcPKIManager) RemovePEMKey(fingerprint string) error {
	return wpm.db.GetControllerAPI().GetKeysAPI().RemoveRSAPublicPEM(fingerprint)
}
