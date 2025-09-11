package main

import (
    "log"
    "mFrelance/config"
    "mFrelance/db"
    "mFrelance/server"
    "mFrelance/lua"
    "net/http"
)



func main() {
    luaFile := "lua/custom_handlers.lua"
    config.Init()
    db.Connect()
    db.Migrate(db.Postgres)
    db.ConnectRedis()


    L := lua.NewState(db.RedisClient, db.Postgres)
    defer lua.L.Close()
    
    if err := L.DoString(`
        print("Lua says hi")
        local result = helloGo("Alice")
        print("Lua got:", result)
    `); err != nil {
        panic(err)
    }

    s := server.New()
    lua.RegisterHttpHandler(lua.L, s.GetMux())
    
    if err := lua.L.DoFile(luaFile); err != nil {
    	panic(err)
    }
    lua.WatchLuaFile(lua.L, luaFile)

    
    s.Handle("/hello", server.HelloHandler)
    //log.Print(server.GenerateMnemonic())
    s.Handle("/register", func(w http.ResponseWriter, r *http.Request) {
        server.RegisterHandler(w, r, db.RedisClient)
    })
    s.Handle("/captcha", func(w http.ResponseWriter, r *http.Request) {
        server.CaptchaHandler(w, r, db.RedisClient)
    })
    s.Handle("/verify", func(w http.ResponseWriter, r *http.Request) {
        server.VerifyHandler(w, r, db.RedisClient)
    })

    log.Println("Starting server on :9999")
    if err := s.Start("9999"); err != nil {
        log.Fatal(err)
    }
}

