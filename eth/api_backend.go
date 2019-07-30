package eth

import (
	"context"
	"math/big"
	// "fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/downloader"
	"github.com/ethereum/go-ethereum/eth/gasprice"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
		"github.com/ethereum/go-ethereum/log"
)

// EthAPIBackend implements ethapi.Backend for full nodes
type EthAPIBackend struct {
	eth *Ethereum
	gpo *gasprice.Oracle
}

// ChainConfig returns the active chain configuration.
func (b *EthAPIBackend) ChainConfig() *params.ChainConfig {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.ChainConfig"," get chain config")

	return b.eth.chainConfig
}

func (b *EthAPIBackend) CurrentBlock() *types.Block {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.CurrentBlock"," get current block")

	return b.eth.blockchain.CurrentBlock()
}

func (b *EthAPIBackend) SetHead(number uint64) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SetHead"," set head with number: ",number)

	b.eth.protocolManager.downloader.Cancel()
	b.eth.blockchain.SetHead(number)
}

func (b *EthAPIBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.HeaderByNumber"," set head with blockNr: ",blockNr)

	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.eth.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.eth.blockchain.CurrentBlock().Header(), nil
	}
	return b.eth.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *EthAPIBackend) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.HeaderByHash"," set head with hash: ",hash)
	return b.eth.blockchain.GetHeaderByHash(hash), nil
}

func (b *EthAPIBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.BlockByNumber"," set block with blockNr : ",blockNr)

	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.eth.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.eth.blockchain.CurrentBlock(), nil
	}
	return b.eth.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *EthAPIBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.StateAndHeaderByNumber"," set block with blockNr : ",blockNr)

	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.eth.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.eth.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *EthAPIBackend) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetBlock"," get block")

	return b.eth.blockchain.GetBlockByHash(hash), nil
}

func (b *EthAPIBackend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetReceipts"," get GetReceipts")

	return b.eth.blockchain.GetReceiptsByHash(hash), nil
}

func (b *EthAPIBackend) GetLogs(ctx context.Context, hash common.Hash) ([][]*types.Log, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetLogs"," get GetLogs: ", hash)

	receipts := b.eth.blockchain.GetReceiptsByHash(hash)
	if receipts == nil {
		log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetLogs"," receipts is nil ")
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *EthAPIBackend) GetTd(blockHash common.Hash) *big.Int {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetTd"," get GetTd")

	return b.eth.blockchain.GetTdByHash(blockHash)
}

func (b *EthAPIBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header) (*vm.EVM, func() error, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetEVM"," get GetEVM")

	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.eth.BlockChain(), nil)
	return vm.NewEVM(context, state, b.eth.chainConfig, *b.eth.blockchain.GetVMConfig()), vmError, nil
}

func (b *EthAPIBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeRemovedLogsEvent"," get SubscribeRemovedLogsEvent")
	return b.eth.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *EthAPIBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeChainEvent"," get SubscribeChainEvent")
	return b.eth.BlockChain().SubscribeChainEvent(ch)
}

func (b *EthAPIBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeChainHeadEvent"," get SubscribeChainHeadEvent")
	return b.eth.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *EthAPIBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeChainSideEvent"," get SubscribeChainSideEvent")
	return b.eth.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *EthAPIBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeLogsEvent"," get SubscribeLogsEvent")
	return b.eth.BlockChain().SubscribeLogsEvent(ch)
}

func (b *EthAPIBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SendTx"," call b.eth.txPool.AddLocal(signedTx)")
	return b.eth.txPool.AddLocal(signedTx)
}

func (b *EthAPIBackend) GetPoolTransactions() (types.Transactions, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetPoolTransactions"," get pool transactions")

	pending, err := b.eth.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *EthAPIBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetPoolTransaction"," get pool transaction")

	return b.eth.txPool.Get(hash)
}

func (b *EthAPIBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.GetPoolNonce"," get nonce of address = ",addr)
	return b.eth.txPool.State().GetNonce(addr), nil
}

func (b *EthAPIBackend) Stats() (pending int, queued int) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.Stats"," get stats")

	return b.eth.txPool.Stats()
}

func (b *EthAPIBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.TxPoolContent"," get TxPoolContent")
	return b.eth.TxPool().Content()
}

func (b *EthAPIBackend) SubscribeNewTxsEvent(ch chan<- core.NewTxsEvent) event.Subscription {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SubscribeNewTxsEvent"," get SubscribeNewTxsEvent")
	return b.eth.TxPool().SubscribeNewTxsEvent(ch)
}

func (b *EthAPIBackend) Downloader() *downloader.Downloader {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.Downloader"," get Downloader")

	return b.eth.Downloader()
}

func (b *EthAPIBackend) ProtocolVersion() int {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.ProtocolVersion"," get ProtocolVersion")

	return b.eth.EthVersion()
}

func (b *EthAPIBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.SuggestPrice"," get SuggestPrice")

	return b.gpo.SuggestPrice(ctx)
}

func (b *EthAPIBackend) ChainDb() ethdb.Database {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.ChainDb"," get ChainDb")

	return b.eth.ChainDb()
}

func (b *EthAPIBackend) EventMux() *event.TypeMux {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.EventMux"," get EventMux")

	return b.eth.EventMux()
}

func (b *EthAPIBackend) AccountManager() *accounts.Manager {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.AccountManager"," get AccountManager")


	return b.eth.AccountManager()
}

func (b *EthAPIBackend) BloomStatus() (uint64, uint64) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.BloomStatus"," get BloomStatus")

	sections, _, _ := b.eth.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *EthAPIBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	log.Debug("binhnt.eth.api_backend","EthAPIBackend.ServiceFilter"," get ServiceFilter")

	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.eth.bloomRequests)
	}
}
