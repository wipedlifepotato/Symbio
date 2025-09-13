package lua

import (
    "github.com/yuin/gopher-lua"
    "github.com/go-redis/redis/v8"
    "github.com/jmoiron/sqlx"
    "fmt"
    "net/http"
    "github.com/fsnotify/fsnotify"
    "log"
    "time"
    "mFrelance/server"
    "mFrelance/db"
    "mFrelance/auth"
    "mFrelance/electrum"
    //"mFrelance/server"
    "gitlab.com/moneropay/go-monero/walletrpc"

)
var L *lua.LState

type LuaVM struct {
	L *lua.LState
}

func WatchLuaFile(L *lua.LState, path string) {
    log.Printf("[LUA] Start monitoring %s", path)

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    err = watcher.Add(path)
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        defer watcher.Close()

        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }

                switch {
                case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
                    log.Println("[LUA] Reloading Lua handlers:", event.Name)
                    if err := L.DoFile(path); err != nil {
                        log.Println("Error reloading Lua:", err)
                    }

                case event.Op&fsnotify.Remove == fsnotify.Remove, event.Op&fsnotify.Rename == fsnotify.Rename:
                    log.Println("[LUA] File removed or renamed, re-adding watcher:", event.Name)
                    time.Sleep(100 * time.Millisecond) 
                    watcher.Remove(path) 
                    err := watcher.Add(path)
                    if err != nil {
                        log.Println("Error re-adding watcher:", err)
                    }
                    if err := L.DoFile(path); err != nil {
                        log.Println("Error reloading Lua:", err)
                    }
                }
            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                log.Println("[LUA] Watcher error:", err)
            }
        }
    }()
}

func RegisterJWTLua(L *lua.LState) {
    L.SetGlobal("get_user_from_jwt", L.NewFunction(func(L *lua.LState) int {
        token := L.ToString(1)
        claims, err := auth.ParseJWT(token)
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        tbl := L.NewTable()
        tbl.RawSetString("user_id", lua.LNumber(claims.UserID))
        tbl.RawSetString("username", lua.LString(claims.Username))
        L.Push(tbl)
        return 1
    }))
}

func RegisterElectrumLua(L *lua.LState, client *electrum.Client) {

    L.SetGlobal("electrum_create_address", L.NewFunction(func(L *lua.LState) int {
        addr, err := client.CreateAddress()
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        L.Push(lua.LString(addr))
        return 1
    }))

    L.SetGlobal("electrum_set_withdraw_blocked", L.NewFunction(func(L *lua.LState) int {
        blocked := L.CheckBool(1) 
        server.SetTxPoolBlocked(blocked)
        return 0 
    }))


    L.SetGlobal("electrum_is_withdraw_blocked", L.NewFunction(func(L *lua.LState) int {
        if server.IsTxPoolBlocked() {
            L.Push(lua.LTrue)
        } else {
            L.Push(lua.LFalse)
        }
        return 1
    }))
    L.SetGlobal("electrum_pay_to_many", L.NewFunction(func(L *lua.LState) int {

		tbl := L.CheckTable(1)
		var outputs [][2]string

		tbl.ForEach(func(_ lua.LValue, value lua.LValue) {
			subTbl, ok := value.(*lua.LTable)
			if !ok {
				return
			}
			addr := subTbl.RawGetInt(1).String()
			amt := subTbl.RawGetInt(2).String()
			outputs = append(outputs, [2]string{addr, amt})
		})

		txID, err := client.PayToMany(outputs)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(txID))
		return 1
    }))
    
    L.SetGlobal("electrum_get_balance", L.NewFunction(func(L *lua.LState) int {
        addr := L.ToString(1)
        bal, err := client.GetBalance(addr)
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        f, _ := bal.Float64()
        L.Push(lua.LNumber(f))
        return 1
    }))

    L.SetGlobal("electrum_pay_to", L.NewFunction(func(L *lua.LState) int {
	    destination := L.ToString(1)
	    amount := L.ToString(2)
	    if destination == "" || amount == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("destination and amount required"))
		return 2
	    }

	    txid, err := client.PayTo(destination, amount)
	    if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	    }

	    L.Push(lua.LString(txid))
	    return 1
    }))

    L.SetGlobal("electrum_list_addresses", L.NewFunction(func(L *lua.LState) int {
        addrs, err := client.ListAddresses()
        if err != nil {
            L.Push(lua.LNil)
            L.Push(lua.LString(err.Error()))
            return 2
        }
        tbl := L.NewTable()
        for _, a := range addrs {
            tbl.Append(lua.LString(a))
        }
        L.Push(tbl)
        return 1
    }))
}

func RegisterLuaHelpers(L *lua.LState, rdb *redis.Client, psql *sqlx.DB) {

	L.SetGlobal("generate_mnemonic", L.NewFunction(func(L *lua.LState) int {
		mnemonic := server.GenerateMnemonic()
		L.Push(lua.LString(mnemonic))
		return 1
	}))

	L.SetGlobal("is_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		admin, err := db.IsAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(admin))
		return 1
	}))


	L.SetGlobal("make_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		err := db.MakeAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("remove_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		err := db.RemoveAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))
	
	L.SetGlobal("get_user", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		userID, passwordHash, err := db.GetUserByUsername(psql, username)
		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(userID))
		tbl.RawSetString("password_hash", lua.LString(passwordHash))
		L.Push(tbl)
		return 1
	}))
	// block_user(userID)
	L.SetGlobal("block_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToInt64(1)
		err := db.BlockUser(psql, userID)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// unblock_user(userID)
	L.SetGlobal("unblock_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToInt64(1)
		err := db.UnblockUser(psql, userID)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// is_user_blocked(username)
	L.SetGlobal("is_user_blocked", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		userID, _, err := db.GetUserByUsername(psql, username)
		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		blocked, err := db.IsUserBlocked(psql, userID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LBool(blocked))
		return 1
	}))
	
	L.SetGlobal("verify_password", L.NewFunction(func(L *lua.LState) int {
	    password := L.ToString(1)
	    hashed := L.ToString(2)

	    res := server.VerifyPassword(password,hashed)
	    L.Push(lua.LBool(res))
	    return 1
	}))
	
	L.SetGlobal("change_password", L.NewFunction(func(L *lua.LState) int {
	    username := L.ToString(1)
	    newPassword := L.ToString(2)


	    hashed := server.HashPassword(newPassword)

	    err := db.ChangeUserPassword(psql, username, string(hashed))
	    if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	    }

	    tbl := L.NewTable()
	    tbl.RawSetString("username", lua.LString(username))
	    L.Push(tbl)
	    return 1
	}))
	
	L.SetGlobal("restore_user", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		mnemonic := L.ToString(2)
		userID, usernameOut, err := db.RestoreUser(psql, username, mnemonic)
		
		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(userID))
		tbl.RawSetString("username", lua.LString(usernameOut))
		L.Push(tbl)
		return 1
	}))


	L.SetGlobal("generate_jwt", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		username := L.ToString(2)
		token, err := auth.GenerateJWT(userID, username)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(token))
		return 1
	}))
}

var registeredPaths = make(map[string]bool)

func RegisterHttpHandler(L *lua.LState, mux *http.ServeMux) {
    L.SetGlobal("register_handler", L.NewFunction(func(L *lua.LState) int {
        path := L.ToString(1)
        handlerFn := L.ToFunction(2)

        if registeredPaths[path] {
            return 0
        }
        registeredPaths[path] = true

        mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
            r.ParseForm()
            reqTable := L.NewTable()
            reqTable.RawSetString("method", lua.LString(r.Method))

            paramsTable := L.NewTable()
            for key, values := range r.Form {
                if len(values) > 0 {
                    paramsTable.RawSetString(key, lua.LString(values[0]))
                }
            }
            reqTable.RawSetString("params", paramsTable)

            if err := L.CallByParam(lua.P{
                Fn:      handlerFn,
                NRet:    1,
                Protect: true,
            }, reqTable); err != nil {
                http.Error(w, err.Error(), 500)
                return
            }

            ret := L.Get(-1)
            L.Pop(1)
            w.Write([]byte(ret.String()))
        })
        return 0
    }))
}


func luaInit(l *lua.LState, rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) {
	//l := lua.NewState()
	l.SetGlobal("helloGo", L.NewFunction(HelloLua))
	RegisterLuaRedis(l, rdb)
	RegisterLuaPostgres(l, psql)
	RegisterConfigGlobals(l)
	RegisterLuaHelpers(l,rdb,psql)
	RegisterElectrumLua(l, eClient)
	RegisterJWTLua(l)
	RegisterMoneroLua(l, mClient)
	return
}

func NewVM(rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) *LuaVM {
	l := lua.NewState()
	luaInit(l, rdb, psql, eClient, mClient)
	
	return &LuaVM{L: l}
}

func (vm *LuaVM) Close() {
	vm.L.Close()
}

func HelloLua(L *lua.LState) int {
    name := L.ToString(1)
    fmt.Println("Hello from Go,", name)
    L.Push(lua.LString("Hi, " + name))
    return 1
}

func NewState(rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) *lua.LState {
	L = lua.NewState()
	luaInit(L, rdb, psql, eClient, mClient) 
	return L
}
