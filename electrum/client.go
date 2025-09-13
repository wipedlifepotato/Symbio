package electrum

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "log"
    "strconv"
    "time"
    "strings"
    "errors"

)
type PaymentRecord struct {
    Time        string      `json:"time"`
    Outputs     [][2]string `json:"outputs"`
    FeeBTC      float64     `json:"fee_btc"`
    RawTx       string      `json:"raw_tx,omitempty"`
    Broadcasted bool        `json:"broadcasted"`
    Error       string      `json:"error,omitempty"`
}

func savePaymentRecord(record PaymentRecord) error {
    fileName := "payments_log.json"

    var records []PaymentRecord

    data, err := ioutil.ReadFile(fileName)
    if err == nil {
        json.Unmarshal(data, &records)
    }

    records = append(records, record)

    newData, _ := json.MarshalIndent(records, "", "  ")
    return ioutil.WriteFile(fileName, newData, 0644)
}

type Client struct {
    URL      string
    User     string
    Password string
}

func NewClient(user, password, host string, port int) *Client {
    return &Client{
        URL:      fmt.Sprintf("http://%s:%d", host, port),
        User:     user,
        Password: password,
    }
}

type RPCResponse struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      string          `json:"id"`
    Result  json.RawMessage `json:"result"`
    Error   interface{}     `json:"error"`
}

func (c *Client) LoadWallet() error {
    _, err := c.call("load_wallet")
    return err
}

type txStatusResponse struct {
	Confirmations int64 `json:"confirmations"`
}

func (c *Client) Get_tx_status(txId string) (int64, error) {
	data, err := c.call("get_tx_status", txId)
	if err != nil {
		return -1, err
	}

	var status txStatusResponse
	if err := json.Unmarshal(data, &status); err != nil {
		return -1, err
	}

	if status.Confirmations < 0 {
		return -1, errors.New("confirmations field missing or invalid")
	}

	return status.Confirmations, nil
}

func (c *Client) call(method string, params ...interface{}) (json.RawMessage, error) {
    reqBody := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      "1",
        "method":  method,
    }

    if len(params) > 0 {
        reqBody["params"] = params
    }
    log.Print(reqBody)
    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }
    log.Println(string(bodyBytes))
    req, err := http.NewRequest("POST", c.URL, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, err
    }

    req.SetBasicAuth(c.User, c.Password)
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    var rpcResp RPCResponse
    //log.Print(string(data))
    if err := json.Unmarshal(data, &rpcResp); err != nil {
        return nil, err
    }

    if rpcResp.Error != nil {
        return nil, fmt.Errorf("RPC error: %v", rpcResp.Error)
    }

    return rpcResp.Result, nil
}


func (c *Client) PayTo(destination string, amount string) (string, error) {

    res, err := c.call("payto", destination, amount)
    if err != nil {
        return "", fmt.Errorf("failed to create transaction: %v", err)
    }
    log.Print(string(res))
    resBroadcast, err := c.call("broadcast", res)
    if err != nil {
        return "", fmt.Errorf("failed to broadcast transaction: %v", err)
    }

    return string(resBroadcast), nil
}
// Comission sometimes is 0?
func (c *Client) PayToMany(outputs [][2]string) (string, error) {
	const (
		minFeeRateSats = 1.5
		maxFeeRateSats = 11.2
		txBaseSize     = 200
		txOutSize      = 34
	)

	feeRate := minFeeRateSats
	maxAttempts := 24

	for attempt := 0; attempt < maxAttempts; attempt++ {

		outList := make([][]interface{}, len(outputs))
		for i, out := range outputs {
			amt, err := strconv.ParseFloat(out[1], 64)
			if err != nil {
				return "", fmt.Errorf("invalid amount '%s': %v", out[1], err)
			}
			outList[i] = []interface{}{out[0], amt}
		}

		txSizeBytes := txBaseSize + txOutSize*len(outputs)
		feeBTC := feeRate * float64(txSizeBytes) / 1e8

		reqBody := map[string]interface{}{
			"id":      "1",
			"jsonrpc": "2.0",
			"method":  "paytomany",
			"params": map[string]interface{}{
				"outputs": outList,
				"rbf":     true,
				"fee":     feeBTC,
			},
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", c.URL, bytes.NewReader(bodyBytes))
		req.SetBasicAuth(c.User, c.Password)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		respData, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var rpcResp RPCResponse
		_ = json.Unmarshal(respData, &rpcResp)

		if rpcResp.Error != nil {
			record := PaymentRecord{
				Time:    time.Now().Format(time.RFC3339),
				Outputs: outputs,
				FeeBTC:  feeBTC,
				Error:   parseRPCError(rpcResp.Error),
			}
			savePaymentRecord(record)

			if strings.Contains(record.Error, "fee") {
				feeRate *= 2.4
				if feeRate > maxFeeRateSats {
					return "", fmt.Errorf("fee_rate exceeded %.2f sat/byte", maxFeeRateSats)
				}
				log.Printf("Fee too low, increasing to %.2f sat/byte, retrying...", feeRate)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			return "", fmt.Errorf("RPC error: %v", record.Error)
		}

		rawTx := rpcResp.Result

		// Broadcast
		broadcastReq := map[string]interface{}{
			"id":      "1",
			"jsonrpc": "2.0",
			"method":  "broadcast",
			"params":  []interface{}{rawTx},
		}
		broadcastBytes, _ := json.Marshal(broadcastReq)
		breq, _ := http.NewRequest("POST", c.URL, bytes.NewReader(broadcastBytes))
		breq.SetBasicAuth(c.User, c.Password)
		breq.Header.Set("Content-Type", "application/json")

		bresp, _ := http.DefaultClient.Do(breq)
		brespData, _ := ioutil.ReadAll(bresp.Body)
		bresp.Body.Close()

		var bRpcResp RPCResponse
		_ = json.Unmarshal(brespData, &bRpcResp)

		if bRpcResp.Error != nil {
			record := PaymentRecord{
				Time:    time.Now().Format(time.RFC3339),
				Outputs: outputs,
				FeeBTC:  feeBTC,
				Error:   parseRPCError(bRpcResp.Error),
			}
			savePaymentRecord(record)

			if strings.Contains(record.Error, "fee") {
				feeRate *= 1.2
				if feeRate > maxFeeRateSats {
					return "", fmt.Errorf("broadcast fee_rate exceeded %.2f sat/byte", maxFeeRateSats)
				}
				log.Printf("Broadcast failed: fee too low, increasing to %.2f sat/byte, retrying...", feeRate)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			return "", fmt.Errorf("broadcast RPC error: %v", record.Error)
		}

		record := PaymentRecord{
			Time:    time.Now().Format(time.RFC3339),
			Outputs: outputs,
			FeeBTC:  feeBTC,
			Error:   "",
		}
		savePaymentRecord(record)

		return string(bRpcResp.Result), nil
	}

	return "", fmt.Errorf("failed to broadcast transaction after %d attempts", maxAttempts)
}

func parseRPCError(err interface{}) string {
	if err == nil {
		return ""
	}
	if errMap, ok := err.(map[string]interface{}); ok {
		if msg, ok := errMap["message"].(string); ok {
			return msg
		}
		return fmt.Sprintf("%v", err)
	}
	return fmt.Sprintf("%v", err)
}


type Transaction struct {
    Txid   string  `json:"tx_hash"`
    Amount float64 `json:"amount"`
    Confirmations int `json:"confirmations"`
}

func (c *Client) ListTransactions(address string) ([]Transaction, error) {
    res, err := c.call("getaddresshistory", address)
    if err != nil {
        return nil, err
    }
    var txs []Transaction
    if err := json.Unmarshal(res, &txs); err != nil {
        return nil, err
    }
    return txs, nil
}

