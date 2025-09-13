package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
)

type PayToManyRequest struct {
    ID      string                 `json:"id"`
    JSONRPC string                 `json:"jsonrpc"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
}

func PayToManyElectrum(outputs [][2]interface{}) (string, error) {
    reqBody := PayToManyRequest{
        ID:      "1",
        JSONRPC: "2.0",
        Method:  "paytomany",
        Params: map[string]interface{}{
            "outputs": outputs,
            "rbf":     true,
        },
    }

    data, err := json.Marshal(reqBody)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %v", err)
    }

    req, err := http.NewRequest("POST", "http://127.0.0.1:7777", bytes.NewBuffer(data))
    if err != nil {
        return "", fmt.Errorf("failed to create request: %v", err)
    }

    req.SetBasicAuth("Electrum", "Electrum")
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()

    body, _ := ioutil.ReadAll(resp.Body)
    return string(body), nil
}

func main() {
    outputs := [][2]interface{}{
        {"2MzQCnSo839GFcyXNYeYGQD5wTzgN5exB96", 0.001},
        {"2Mydq5weSRT44Ej3ZLNykSFBzvnV8R8godU", 0.001},
    }

    res, err := PayToManyElectrum(outputs)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Response:", res)
}
