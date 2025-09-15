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
func AddToTxPool(address string, amount *big.Float) {
    txPool.Lock()
    if existing, ok := txPool.outputs[address]; ok {
        txPool.outputs[address] = new(big.Float).Add(existing, amount)
    } else {
        txPool.outputs[address] = new(big.Float).Set(amount)
    }
    txPool.Unlock()
}


