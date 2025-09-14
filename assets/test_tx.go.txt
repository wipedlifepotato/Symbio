package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	//"github.com/btcsuite/btcutil"
	"mFrelance/electrum"
)

func ProcessTx(client *electrum.Client, address string, txHash string) (*big.Float, error) {

	txHexRaw, err := client.GetTransaction(txHash)
	if err != nil {
		return nil, err
	}


	rawTx, err := hex.DecodeString(txHexRaw)
	if err != nil {
		return nil, err
	}

	msgTx := wire.NewMsgTx(wire.TxVersion)
	if err := msgTx.Deserialize(bytes.NewReader(rawTx)); err != nil {
		return nil, err
	}

	incoming := big.NewFloat(0)

	for _, txOut := range msgTx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, &chaincfg.TestNet3Params)
		if err != nil || len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			log.Print(addr.EncodeAddress())
			log.Print(txOut.Value)
			if addr.EncodeAddress() == address {
				val := new(big.Float).Quo(big.NewFloat(float64(txOut.Value)), big.NewFloat(1e8))
				incoming.Add(incoming, val)
			}
		}
	}

	return incoming, nil
}

func main() {
	address := "tb1q4fd0atukx96557ql07av5enl2u73ltdp06hqys"
	txHashes := []string{
		"3d46f85256989ab0209dd662b8db507b6f6c31eb17d22cda80e97dc9beb52e3e",
		"f621ab9e9c71bc35d59537d9fcb0a297577e61fb03dd87c2974b370d7ff77f40",
		"0bf0febde30368c46d66d77150bfb401478d8127ee89872e717d518c0faef5c6",
		"4d55fc900b45e858c51b794f66ebb844f56d89f6828a8f5e16164bc7e35c3289",
	}

	client := electrum.NewClient("Electrum", "Electrum", "127.0.0.1", 7777)

	for _, txHash := range txHashes {
		amount, err := ProcessTx(client, address, txHash)
		if err != nil {
			log.Println("Error processing tx:", txHash, err)
			continue
		}
		if amount.Cmp(big.NewFloat(0)) > 0 {
			fmt.Printf("Incoming TX %s -> %.8f BTC\n", txHash, amount)
		} else {
			fmt.Printf("TX %s has no outputs to our address\n", txHash)
		}
	}
}

