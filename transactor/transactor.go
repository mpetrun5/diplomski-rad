package transactor

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Transactor struct {
	client  *ethclient.Client
	account common.Address
}

func NewTransactor(url string, account common.Address) (*Transactor, error) {
	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err)
	}

	return &Transactor{
		client:  client,
		account: account,
	}, nil
}

func (t *Transactor) GetNonce() (uint64, error) {
	nonce, err := t.client.PendingNonceAt(context.Background(), t.account)
	if err != nil {
		return 0, err
	}

	return nonce, nil
}

func (t *Transactor) GetChainID() *big.Int {
	return big.NewInt(5)
}

func (t *Transactor) Transact(signedTx *types.Transaction) error {
	return t.client.SendTransaction(context.Background(), signedTx)
}
