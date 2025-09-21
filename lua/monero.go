package lua

import (
	"context"

	"strconv"

	"github.com/yuin/gopher-lua"
	"gitlab.com/moneropay/go-monero/walletrpc"
)

func RegisterMoneroLua(L *lua.LState, mClient *walletrpc.Client) {

	L.SetGlobal("monero_get_balance", L.NewFunction(func(L *lua.LState) int {
		resp, err := mClient.GetBalance(context.Background(), &walletrpc.GetBalanceRequest{})
		if err != nil {
			L.RaiseError("Monero GetBalance error: %v", err)
			return 0
		}

		totalStr := walletrpc.XMRToDecimal(resp.Balance)
		unlockedStr := walletrpc.XMRToDecimal(resp.UnlockedBalance)

		total, err := strconv.ParseFloat(totalStr, 64)
		if err != nil {
			L.RaiseError("Invalid balance value: %v", err)
			return 0
		}

		unlocked, err := strconv.ParseFloat(unlockedStr, 64)
		if err != nil {
			L.RaiseError("Invalid unlocked balance: %v", err)
			return 0
		}

		L.Push(lua.LNumber(total))
		L.Push(lua.LNumber(unlocked))
		return 2
	}))

	L.SetGlobal("monero_create_address", L.NewFunction(func(L *lua.LState) int {
		label := L.ToString(1)
		req := &walletrpc.CreateAddressRequest{AccountIndex: 0, Label: label}
		resp, err := mClient.CreateAddress(context.Background(), req)
		if err != nil {
			L.RaiseError("Monero CreateAddress error: %v", err)
			return 0
		}
		L.Push(lua.LString(resp.Address))
		return 1
	}))

	L.SetGlobal("monero_transfer", L.NewFunction(func(L *lua.LState) int {
		dest := L.ToString(1)
		amount := L.ToNumber(2)

		req := &walletrpc.TransferRequest{
			Destinations: []walletrpc.Destination{
				{Address: dest, Amount: uint64(amount * 1e12)},
			},
			AccountIndex: 0,
		}
		resp, err := mClient.Transfer(context.Background(), req)
		if err != nil {
			L.RaiseError("Monero Transfer error: %v", err)
			return 0
		}

		L.Push(lua.LString(resp.TxHash))
		return 1
	}))
	L.SetGlobal("monero_get_subaddress_info", L.NewFunction(func(L *lua.LState) int {
		account := uint64(L.ToInt(1))
		sub := uint64(L.ToInt(2))

		balResp, err := mClient.GetBalance(context.Background(), &walletrpc.GetBalanceRequest{
			AccountIndex:   account,
			AddressIndices: []uint64{sub},
		})
		if err != nil {
			L.RaiseError("Monero GetBalance error: %v", err)
			return 0
		}

		if len(balResp.PerSubaddress) == 0 {
			L.Push(lua.LNumber(0))
			L.Push(lua.LNumber(0))
			L.Push(lua.LString(""))
			return 3
		}

		subBal := balResp.PerSubaddress[0]
		total, _ := strconv.ParseFloat(walletrpc.XMRToDecimal(subBal.Balance), 64)
		unlocked, _ := strconv.ParseFloat(walletrpc.XMRToDecimal(subBal.UnlockedBalance), 64)

		addrResp, err := mClient.GetAddress(context.Background(), &walletrpc.GetAddressRequest{
			AccountIndex: account,
		})
		if err != nil || len(addrResp.Addresses) <= int(sub) {
			L.RaiseError("Monero GetAddress error: %v", err)
			return 0
		}

		address := addrResp.Addresses[sub].Address

		L.Push(lua.LNumber(total))
		L.Push(lua.LNumber(unlocked))
		L.Push(lua.LString(address))
		return 3
	}))

	L.SetGlobal("monero_get_subaddress_balance", L.NewFunction(func(L *lua.LState) int {
		account := uint64(L.ToInt(1))
		sub := uint64(L.ToInt(2))

		resp, err := mClient.GetBalance(context.Background(), &walletrpc.GetBalanceRequest{
			AccountIndex:   account,
			AddressIndices: []uint64{sub},
		})
		if err != nil {
			L.RaiseError("Monero GetBalance error: %v", err)
			return 0
		}

		if len(resp.PerSubaddress) == 0 {
			L.Push(lua.LNumber(0))
			L.Push(lua.LNumber(0))
			return 2
		}

		subBal := resp.PerSubaddress[0]
		total, _ := strconv.ParseFloat(walletrpc.XMRToDecimal(subBal.Balance), 64)
		unlocked, _ := strconv.ParseFloat(walletrpc.XMRToDecimal(subBal.UnlockedBalance), 64)

		L.Push(lua.LNumber(total))
		L.Push(lua.LNumber(unlocked))
		return 2
	}))
}
