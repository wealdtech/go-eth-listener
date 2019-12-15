package listener

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/peterbourgon/diskv/v3"
	log "github.com/sirupsen/logrus"
	"github.com/wealdtech/go-eth-listener/shared"
)

var zero = big.NewInt(0)

var checkpointBlock = big.NewInt(0)
var chainID *big.Int

var d *diskv.Diskv

// TODO set up the checkpoint path
func init() {
	d = diskv.New(diskv.Options{
		BasePath: "checkpoint",
	})
}

func initCheckpoint(actx *shared.AppContext) bool {
	var err error
	checkpointBlock, err = readCheckpoint(actx.ChainID)
	if err != nil {
		if err.Error() == "no checkpoint" {
			checkpointBlock = zero
			return true
		}
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to obtain checkpoint")
		return false
	}
	log.WithFields(log.Fields{"checkpoint": checkpointBlock}).Info("Obtained checkpoint")
	return false
}

// writeCheckpoint writes the current checkpoint value for a chain ID
func writeCheckpoint(chainID *big.Int, value *big.Int) error {
	return d.Write(checkpointKey(chainID), value.Bytes())
}

// readCheckpoint reads the current checkpoint value for a chain ID
func readCheckpoint(chainID *big.Int) (*big.Int, error) {
	var checkpoint *big.Int
	bytes, err := d.Read(checkpointKey(chainID))
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return nil, errors.New("no checkpoint")
		}
		return nil, err
	}
	checkpoint = new(big.Int).SetBytes(bytes)
	return checkpoint, nil
}

// checkpointKey is a helper to set a checkpoint key
func checkpointKey(chainID *big.Int) string {
	return fmt.Sprintf("Checkpoint %v", chainID)
}
