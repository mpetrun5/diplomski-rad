package executor

import (
	"math/big"

	"github.com/mpetrun5/diplomski-rad/tss/signing"

	tssSigning "github.com/binance-chain/tss-lib/ecdsa/signing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	GasLimit = 200000
)

type Transactor interface {
	Transact(signedTx *types.Transaction) error
	GetNonce() (uint64, error)
	GetChainID() *big.Int
}

type Executor struct {
	transactor   Transactor
	signer       signing.Signing
	signatureChn chan *tssSigning.SignatureData
}

func NewExecutor(
	transactor Transactor,
	signer signing.Signing,
	signatureChn chan *tssSigning.SignatureData,
) *Executor {
	return &Executor{
		transactor:   transactor,
		signer:       signer,
		signatureChn: signatureChn,
	}
}

// Execute assembles signature and sends transaction on the Ethereum network
func (e *Executor) Execute(to common.Address, value *big.Int, data []byte) error {
	rawTx, err := e.createRawTransaction(to, value, data)
	if err != nil {
		return err
	}

	signer := types.NewEIP155Signer(big.NewInt(e.transactor.GetChainID().Int64()))
	msg := big.NewInt(0)
	txBytes := signer.Hash(rawTx)
	msg.SetBytes(txBytes[:])

	go e.signer.Initiate(msg)

	signatureData := <-e.signatureChn
	sig := signatureData.Signature.R
	sig = append(sig[:], signatureData.Signature.S[:]...)
	sig = append(sig[:], signatureData.Signature.SignatureRecovery...)
	signedTx, err := rawTx.WithSignature(signer, sig)
	if err != nil {
		return err
	}

	err = e.transactor.Transact(signedTx)
	if err != nil {
		return err
	}

	return err
}

func (e *Executor) createRawTransaction(to common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	nonce, err := e.transactor.GetNonce()
	if err != nil {
		return nil, err
	}

	return types.NewTransaction(
		nonce,
		to,
		big.NewInt(0),
		uint64(GasLimit),
		big.NewInt(10000000000),
		data,
	), nil
}
