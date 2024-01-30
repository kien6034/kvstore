package main

import (
	"bytes"
	"log"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v3"
)

type KVStoreApplication struct {
	db           *badger.DB
	onGoingBlock *badger.Txn
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{
		db: db,
	}
}

func (app *KVStoreApplication) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (app *KVStoreApplication) Query(query abcitypes.RequestQuery) abcitypes.ResponseQuery {
	return abcitypes.ResponseQuery{}
}

func (app *KVStoreApplication) CheckTx(tx abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(tx.Tx)

	// any response code other than 0 is considered a failure
	return abcitypes.ResponseCheckTx{Code: code}
}

func (app *KVStoreApplication) InitChain(chain abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (app *KVStoreApplication) PrepareProposal(proposal abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	return abcitypes.ResponsePrepareProposal{}
}

func (app *KVStoreApplication) ProcessProposal(proposal abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{}
}

// is called once to indicate that the application is about to receive a block
func (app *KVStoreApplication) BeginBlock(block abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	// create a new Badger transaction. All application transactions in the current block will be executed under this badger tx
	app.onGoingBlock = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

// add key/value to the database transaction everytime the app receives a new application transaction
func (app *KVStoreApplication) DeliverTx(tx abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	// validate against, since it is not guranteed that the tx valid if transaction state is used to determine transaction validity
	if code := app.isValid(tx.Tx); code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.SplitN(tx.Tx, []byte("="), 2)
	key, value := parts[0], parts[1]

	if err := app.onGoingBlock.Set(key, value); err != nil {
		log.Panicf("Error writing to database, unable to execute tx: %v", err)
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}

// is to inform the app that full block has been delivered, and give the application the chance to do additonal processing before commit the state
func (app *KVStoreApplication) EndBlock(block abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

// make permanent changes to the application state
func (app *KVStoreApplication) Commit() abcitypes.ResponseCommit {
	if err := app.onGoingBlock.Commit(); err != nil {
		log.Panicf("Error writing to database, unable to commit block: %v", err)
	}
	return abcitypes.ResponseCommit{Data: []byte{}}
}

func (app *KVStoreApplication) ListSnapshots(snapshots abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (app *KVStoreApplication) OfferSnapshot(snapshot abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (app *KVStoreApplication) LoadSnapshotChunk(chunk abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (app *KVStoreApplication) ApplySnapshotChunk(chunk abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (app *KVStoreApplication) isValid(tx []byte) uint32 {
	// transaction is a string of with the form key=value.
	// most basic checks is to check if the tx conforms to the key=value
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	return 0
}
