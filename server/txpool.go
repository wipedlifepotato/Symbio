package server

import (
	"math/big"
	"sync"
)

var txPoolBlocked struct {
	sync.RWMutex
	Blocked bool
}

func SetTxPoolBlocked(block bool) {
	txPoolBlocked.Lock()
	defer txPoolBlocked.Unlock()
	txPoolBlocked.Blocked = block
}

func IsTxPoolBlocked() bool {
	txPoolBlocked.RLock()
	defer txPoolBlocked.RUnlock()
	return txPoolBlocked.Blocked
}

// AddToTxPool aggregates outputs for batch payout in goroutines.go
func AddToTxPool(address string, amount *big.Float, currency string) {
    txPool.Lock()
    defer txPool.Unlock()
    txPool.outputs[address] = append(txPool.outputs[address], TxPoolItem{
        Amount:   amount,
        Currency: currency,
    })
}
