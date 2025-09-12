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
    s.Handle("/register", func(w http.ResponseWriter, r *http.Request) {
        server.RegisterHandler(w, r, db.RedisClient)
    })
    s.Handle("/auth", func(w http.ResponseWriter, r *http.Request) {
        server.AuthHandler(w, r, db.RedisClient)
    })
    s.Handle("/captcha", func(w http.ResponseWriter, r *http.Request) {
        server.CaptchaHandler(w, r, db.RedisClient)
    })
    s.Handle("/verify", func(w http.ResponseWriter, r *http.Request) {
        server.VerifyHandler(w, r, db.RedisClient)
    })
    s.Handle("/restoreuser", func(w http.ResponseWriter, r *http.Request) {
        server.RestoreHandler(w, r, db.RedisClient)
    })


    apiMux := http.NewServeMux()
    apiMux.Handle("/test", server.AuthMiddleware(http.HandlerFunc(server.TestHandler)))


    s.HandleHandler("/api/", http.StripPrefix("/api", apiMux))

    log.Println("Starting server on "+config.AppConfig.ListenAddr+":"+config.AppConfig.Port)
    if err := s.Start(config.AppConfig.ListenAddr, config.AppConfig.Port); err != nil {
        log.Fatal(err)
    }
}
