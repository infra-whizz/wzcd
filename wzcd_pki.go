package wzcd

import (
	"fmt"
	"io/ioutil"

	wzlib_database "github.com/infra-whizz/wzlib/database"
	wzlib_database_controller "github.com/infra-whizz/wzlib/database/controller"
	wzlib_logger "github.com/infra-whizz/wzlib/logger"
)

type WzcPKIManager struct {
	db *wzlib_database.WzDBH
	wzlib_logger.WzLogger
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

// Get PEM key from the file (private or public)
func (wpm *WzcPKIManager) getPEMKeyFromFile(filePath string, systemid string, fqdn string) ([]byte, error) {
	keypem, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if systemid == "" {
		return nil, fmt.Errorf("Machine ID is not defined")
	}

	if fqdn == "" {
		return nil, fmt.Errorf("FQDN of the machine required. NOTE: it should be the same FQDN as the remote seeing it.")
	}

	return keypem, nil
}

// RegisterClusterPrivatePEMKey registers PEM private for the entire cluster
func (wpm *WzcPKIManager) RegisterClusterPEMKeyPair(pubkeyFilePath string, privkeyFilePath string) error {
	// Remove current cluster key, if any.
	fingerprint := wpm.GetClusterPublicPEMKeyFingerprint()
	if fingerprint != "" {
		if err := wpm.RemovePEMKey(fingerprint); err != nil {
			return err
		}
	}

	pub, err := wpm.getPEMKeyFromFile(pubkeyFilePath,
		wzlib_database_controller.OWNER_APP_CLUSTER, wzlib_database_controller.OWNER_APP_CLUSTER)
	if err != nil {
		return err
	}

	priv, err := wpm.getPEMKeyFromFile(privkeyFilePath,
		wzlib_database_controller.OWNER_APP_CLUSTER, wzlib_database_controller.OWNER_APP_CLUSTER)
	if err != nil {
		return err
	}

	return wpm.db.GetControllerAPI().GetKeysAPI().AddRSAKeypairPEM(
		pub, priv, wzlib_database_controller.OWNER_APP_CLUSTER, wzlib_database_controller.OWNER_APP_CLUSTER,
		wzlib_database_controller.OWNER_APP_CLUSTER)
}

// GetClusterPublicPEMKeyFingerprint returns current fingerprint for the cluster
func (wpm *WzcPKIManager) GetClusterPublicPEMKeyFingerprint() string {
	fingerprint := ""
	keys := wpm.db.GetControllerAPI().GetKeysAPI().ListRSAPublicPEMByOwner(wzlib_database_controller.OWNER_APP_CLUSTER)
	if len(keys) > 1 {
		wpm.GetLogger().Warningf("More than one cluster keypair found: %d", len(keys))
		wpm.GetLogger().Warningf("Removing all the keys, please re-add them all again!")
		for _, key := range keys {
			if err := wpm.db.GetControllerAPI().GetKeysAPI().RemoveRSAKeyPEM(key.RsaFp); err != nil {
				wpm.GetLogger().Fatalln("Unable to remove RSA key:", err.Error())
			}
			wpm.GetLogger().Warningf("Removed RSA key container: %s (%s as %s)", key.RsaFp, key.MachineId, key.Fqdn)
		}
		return fingerprint

	} else if len(keys) == 1 {
		return keys[0].RsaFp
	}

	return fingerprint
}

// RegisterPEMKey from the file on the system and tie up to a specific systemid and fqdn
func (wpm *WzcPKIManager) RegisterPEMKey(filePath string, systemid string, fqdn string) error {
	keypem, err := wpm.getPEMKeyFromFile(filePath, systemid, fqdn)
	if err != nil {
		return err
	}
	return wpm.db.GetControllerAPI().GetKeysAPI().AddRSAPublicPEM(
		keypem, systemid, fqdn, wzlib_database_controller.OWNER_APP_REMOTE)
}

// RemovePEMKey from the database by a fingerprint
func (wpm *WzcPKIManager) RemovePEMKey(fingerprint string) error {
	return wpm.db.GetControllerAPI().GetKeysAPI().RemoveRSAKeyPEM(fingerprint)
}

// ListRemotePEMKeys registered in the database
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
