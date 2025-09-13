package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
)

type PaymentRecord struct {
	Time    string       `json:"time"`
	Outputs [][2]string  `json:"outputs"`
	FeeBTC  float64      `json:"fee_btc"`
	Error   string       `json:"error"`
}
// TODO: delete it.
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run pending_req.go <paymentFile.json> <ElectrumHost:Port>")
		return
	}

	paymentFile := os.Args[1]
	electrumURL := os.Args[2]

	data, err := os.ReadFile(paymentFile)
	if err != nil {
		panic(err)
	}

	var records []PaymentRecord
	if err := json.Unmarshal(data, &records); err != nil {
		panic(err)
	}

	sumMap := make(map[string]*big.Float)

	for i, r := range records {
		if r.Error != "pending" {
			continue
		}

		for _, out := range r.Outputs {
			addr := out[0]
			amt, _ := new(big.Float).SetString(out[1])
			if existing, ok := sumMap[addr]; ok {
				sumMap[addr] = new(big.Float).Add(existing, amt)
			} else {
				sumMap[addr] = amt
			}
		}

		records[i].Error = "processing"
	}

	f, err := os.Create("PendingReq.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintf(f, "curl -s -u Electrum:Electrum -X POST http://%s -H \"Content-Type: application/json\" -d '{\n", electrumURL)
	fmt.Fprintln(f, `  "id": "1",`)
	fmt.Fprintln(f, `  "jsonrpc": "2.0",`)
	fmt.Fprintln(f, `  "method": "paytomany",`)
	fmt.Fprint(f, `  "params": {"outputs":[`)

	first := true
	for addr, amt := range sumMap {
		if !first {
			fmt.Fprint(f, ", ")
		}
		fmt.Fprintf(f, `[\"%s\", %s]`, addr, amt.Text('f', 8))
		first = false
	}

	fmt.Fprintln(f, "], \"rbf\": true}")
	fmt.Fprintln(f, "}'")

	fileData, _ := json.MarshalIndent(records, "", "  ")
	if err := os.WriteFile(paymentFile, fileData, 0644); err != nil {
		fmt.Println("Failed to save file:", err)
	}

}

