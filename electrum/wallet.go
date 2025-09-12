package electrum

import (
    "encoding/json"
    "fmt"
    "math/big"
)

type AddressBalance struct {
    Confirmed   string `json:"confirmed"`
    Unconfirmed string `json:"unconfirmed"`
}

var balance AddressBalance

func (c *Client) ListAddresses() ([]string, error) {
    if err := c.LoadWallet(); err != nil {
        return nil, err
    }
    res, err := c.call("listaddresses")
    if err != nil {
        return nil, err
    }
    var addresses []string
    if err := json.Unmarshal(res, &addresses); err != nil {
        return nil, err
    }
    return addresses, nil
}

func SatoshiToBTC(balance AddressBalance) (*big.Float, error) {
    confirmed, _, err := big.ParseFloat(balance.Confirmed, 10, 256, big.ToNearestEven)
    if err != nil {
        return nil, fmt.Errorf("invalid confirmed balance: %v", err)
    }

    unconfirmed, _, err := big.ParseFloat(balance.Unconfirmed, 10, 256, big.ToNearestEven)
    if err != nil {
        return nil, fmt.Errorf("invalid unconfirmed balance: %v", err)
    }

    sum := new(big.Float).Add(confirmed, unconfirmed)
    return sum, nil
}

func OnlyConfirmedSatoshiToBTC(balance AddressBalance) (*big.Float, error) {
    confirmed, _, err := big.ParseFloat(balance.Confirmed, 10, 256, big.ToNearestEven)
    if err != nil {
        return nil, fmt.Errorf("invalid confirmed balance: %v", err)
    }
    return confirmed, nil
}

func AddBTC(a, b *big.Float) *big.Float {
    return new(big.Float).Add(a, b)
}

func DelBTC(a, b *big.Float) *big.Float {
    return new(big.Float).Sub(a, b)
}

func (c *Client) GetBalance(address string) (*big.Float, error) {
    res, err := c.call("getaddressbalance", address)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(res, &balance); err != nil {
        return nil, err
    }
    return SatoshiToBTC(balance)
}

func (c *Client) GetAllBalances(addresses []string) (map[string]*big.Float, error) {
    result := make(map[string]*big.Float)
    for _, addr := range addresses {
        bal, err := c.GetBalance(addr)
        if err != nil {
            return nil, err
        }
        result[addr] = bal
    }
    return result, nil
}

