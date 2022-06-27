package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/libp2p/go-libp2p-core/peer"
)

type SaveDataStorage struct {
	filename string
	mu       sync.Mutex
}

type Keyshare struct {
	Key       keygen.LocalPartySaveData
	Peers     []peer.ID
	Threshold int
}

func NewSaveDataStorage(filename string) *SaveDataStorage {
	return &SaveDataStorage{
		filename: filename,
	}
}

func (s *SaveDataStorage) LockShare() {
	s.mu.Lock()
}

func (s *SaveDataStorage) UnlockShare() {
	s.mu.Unlock()
}

// SetSaveData overwrites old keyshare with new one. If no key
// exists, file will be created.
// Should be protected in code with a lock.
func (s *SaveDataStorage) SetSaveData(saveData Keyshare) error {
	f, err := os.OpenFile(s.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	dataBytes, err := json.Marshal(&saveData)
	if err != nil {
		return err
	}

	_, err = f.Write(dataBytes)
	return err
}

// GetSaveData retrieves keyshare. Will lock execution if a
// resharing or key generation is in progress.
func (s *SaveDataStorage) GetSaveData() (Keyshare, error) {
	s.LockShare()
	defer s.UnlockShare()

	dataBytes, err := ioutil.ReadFile(s.filename)
	if err != nil {
		return Keyshare{}, err
	}

	saveData := Keyshare{}
	err = json.Unmarshal(dataBytes, &saveData)
	if err != nil {
		return Keyshare{}, err
	}

	return saveData, nil
}
