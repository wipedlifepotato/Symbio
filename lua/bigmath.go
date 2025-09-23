package lua

import (
	"math/big"

	lua "github.com/yuin/gopher-lua"
)

func RegisterBigMath(L *lua.LState) {
	bigMath := L.NewTable()

	L.SetField(bigMath, "New", L.NewFunction(func(L *lua.LState) int {
		val := L.ToString(1)
		f, ok := new(big.Float).SetString(val)
		if !ok {
			f = big.NewFloat(0)
		}
		ud := L.NewUserData()
		ud.Value = f
		L.Push(ud)
		return 1
	}))

	L.SetField(bigMath, "Add", L.NewFunction(func(L *lua.LState) int {
		a := L.CheckUserData(1).Value.(*big.Float)
		b := L.CheckUserData(2).Value.(*big.Float)
		res := new(big.Float).Add(a, b)
		ud := L.NewUserData()
		ud.Value = res
		L.Push(ud)
		return 1
	}))

	L.SetField(bigMath, "Sub", L.NewFunction(func(L *lua.LState) int {
		a := L.CheckUserData(1).Value.(*big.Float)
		b := L.CheckUserData(2).Value.(*big.Float)
		res := new(big.Float).Sub(a, b)
		ud := L.NewUserData()
		ud.Value = res
		L.Push(ud)
		return 1
	}))

	L.SetField(bigMath, "Mul", L.NewFunction(func(L *lua.LState) int {
		a := L.CheckUserData(1).Value.(*big.Float)
		b := L.CheckUserData(2).Value.(*big.Float)
		res := new(big.Float).Mul(a, b)
		ud := L.NewUserData()
		ud.Value = res
		L.Push(ud)
		return 1
	}))

	L.SetField(bigMath, "Quo", L.NewFunction(func(L *lua.LState) int {
		a := L.CheckUserData(1).Value.(*big.Float)
		b := L.CheckUserData(2).Value.(*big.Float)
		res := new(big.Float).Quo(a, b)
		ud := L.NewUserData()
		ud.Value = res
		L.Push(ud)
		return 1
	}))

	L.SetField(bigMath, "String", L.NewFunction(func(L *lua.LState) int {
		a := L.CheckUserData(1).Value.(*big.Float)
		L.Push(lua.LString(a.Text('f', -1)))
		return 1
	}))

	L.SetGlobal("BigMath", bigMath)
}
