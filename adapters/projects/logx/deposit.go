package logx

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/taikoxyz/trailblazer-adapters/adapters"
)

// Update this constant to reflect the event signature of SentTransferRemote
const (
	SentTransferRemoteSignature string = "SentTransferRemote(uint32,bytes32,uint256)"
	DepostiAddress string = "0x650e8941E4d90b70576fDF1b05dbDc962DA2cab8"
	
)

type SentTransferRemoteEvent struct {
	Destination uint32
	Recipient   [32]byte
	Amount      *big.Int
}

type TransferRemoteIndexer struct {
	client    *ethclient.Client
	addresses []common.Address
}

func NewTransferRemoteIndexer(client *ethclient.Client, addresses []common.Address) *TransferRemoteIndexer {
	return &TransferRemoteIndexer{
		client:    client,
		addresses: addresses,
	}
}

var _ adapters.LogIndexer[adapters.Whitelist] = &TransferRemoteIndexer{}

func (indexer *TransferRemoteIndexer) Addresses() []common.Address {
	return indexer.addresses
}

func (indexer *TransferRemoteIndexer) Index(ctx context.Context, logs ...types.Log) ([]adapters.Whitelist, error) {
	var whitelist []adapters.Whitelist

	for _, l := range logs {
		if !indexer.isTransferRemoteLog(l) {
			continue
		}

		var transferEvent SentTransferRemoteEvent

		// Define the ABI for the SentTransferRemote event
		transferRemoteABI, err := abi.JSON(strings.NewReader(`[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint32","name":"destination","type":"uint32"},{"indexed":true,"internalType":"bytes32","name":"recipient","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"}],"name":"SentTransferRemote","type":"event"}]`))
		if err != nil {
			return nil, err
		}

		// Unpack the event data
		err = transferRemoteABI.UnpackIntoInterface(&transferEvent, "SentTransferRemote", l.Data)
		if err != nil {
			return nil, err
		}

		block, err := indexer.client.BlockByNumber(ctx, big.NewInt(int64(l.BlockNumber)))
		if err != nil {
			return nil, err
		}

		// Capture the transfer details in the whitelist
		w := &adapters.Whitelist{
			User:        common.BytesToAddress(transferEvent.Recipient[:]),
			Time:        block.Time(),
			BlockNumber: block.NumberU64(),
			TxHash:      l.TxHash,
		}

		whitelist = append(whitelist, *w)
	}

	return whitelist, nil
}

func (indexer *TransferRemoteIndexer) isTransferRemoteLog(l types.Log) bool {
	// Compare the event signature hash with the log's topic
	return l.Topics[0].Hex() == crypto.Keccak256Hash([]byte(SentTransferRemoteSignature)).Hex()
}
