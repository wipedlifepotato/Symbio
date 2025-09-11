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


func luaInit(l *lua.LState, rdb *redis.Client, psql *sqlx.DB) {
	//l := lua.NewState()
	l.SetGlobal("helloGo", L.NewFunction(HelloLua))
	RegisterLuaRedis(l, rdb)
	RegisterLuaPostgres(l, psql)
	return
}

func NewVM(rdb *redis.Client, psql *sqlx.DB) *LuaVM {
	l := lua.NewState()
	luaInit(l, rdb, psql)
	
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

func NewState(rdb *redis.Client, psql *sqlx.DB) *lua.LState {
	L = lua.NewState()
	luaInit(L, rdb, psql)
	return L
}
