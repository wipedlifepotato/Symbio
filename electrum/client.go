package electrum

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    //"log"
)

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

func (c *Client) call(method string, params ...interface{}) (json.RawMessage, error) {
    reqBody := map[string]interface{}{
        "jsonrpc": "2.0",
        "id":      "1",
        "method":  method,
    }

    if len(params) > 0 {
        reqBody["params"] = params
    }

    bodyBytes, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

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
