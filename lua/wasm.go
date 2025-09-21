package lua

import (
	"context"
	"encoding/binary"
	"github.com/wasmerio/wasmer-go/wasmer"
	lua_gopher "github.com/yuin/gopher-lua"
	"io/ioutil"
)

func LoadWasmModule(L *lua_gopher.LState, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	module, err := wasmer.NewModule(store, bytes)
	if err != nil {
		return err
	}

	importObject := wasmer.NewImportObject()
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return err
	}

	mem, err := instance.Exports.GetMemory("memory")
	if err != nil {
		return err
	}

	L.SetGlobal("wasm_memory", L.NewUserData())
	L.SetGlobal("wasm_call_bytes", L.NewFunction(func(L *lua_gopher.LState) int {
		fnName := L.ToString(1)
		fn, err := instance.Exports.GetFunction(fnName)
		if err != nil {
			L.Push(lua_gopher.LString("no such function"))
			return 1
		}

		var buf []byte
		for i := 2; i <= L.GetTop(); i++ {
			val := L.Get(i)
			switch v := val.(type) {
			case lua_gopher.LNumber:
				// int32 -> 4 байта
				b := make([]byte, 4)
				binary.LittleEndian.PutUint32(b, uint32(v))
				buf = append(buf, b...)
			case lua_gopher.LString:
				buf = append(buf, []byte(v)...)
			default:
				// TODO
			}
		}

		ptr := 0
		copy(mem.Data()[ptr:], buf)

		_, err = fn(ptr, len(buf))
		if err != nil {
			L.Push(lua_gopher.LString("error: " + err.Error()))
			return 1
		}

		result := mem.Data()[ptr : ptr+len(buf)]
		L.Push(lua_gopher.LString(string(result)))
		return 1
	}))

	L.SetGlobal("wasm_call", L.NewFunction(func(L *lua_gopher.LState) int {
		fnName := L.ToString(1)
		fn, err := instance.Exports.GetFunction(fnName)
		if err != nil {
			L.Push(lua_gopher.LString("no such function"))
			return 1
		}

		input := L.ToInt(2)
		res, err := fn(context.Background(), int32(input))
		if err != nil {
			L.Push(lua_gopher.LString("error: " + err.Error()))
			return 1
		}

		L.Push(lua_gopher.LNumber(res.(int32)))
		return 1
	}))

	return nil
}
