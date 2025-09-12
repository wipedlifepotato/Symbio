package main

import (
    "io/ioutil"
    "log"
    "path/filepath"
    "strings"
    "mFrelance/config"
    "mFrelance/db"
    "mFrelance/server"
    "mFrelance/lua"
    "net/http"
    "mFrelance/electrum"
    //"mFrelance/monero"
    "github.com/gabstv/httpdigest"
    "gitlab.com/moneropay/go-monero/walletrpc"
    "fmt"
    "context"
)

func main() {
    config.Init()
    electrumClient := electrum.NewClient(
        config.AppConfig.ElectrumUser,
        config.AppConfig.ElectrumPassword,
        config.AppConfig.ElectrumHost,
        config.MustAtoi(config.AppConfig.ElectrumPort),
    )
    moneroClient := walletrpc.New(walletrpc.Config{
		Address: "http://"+config.AppConfig.MoneroHost+":"+config.AppConfig.MoneroPort+"/json_rpc",
		Client: &http.Client{
			Transport: httpdigest.New(config.AppConfig.MoneroUser, config.AppConfig.MoneroPassword), // Remove if no auth.
		},
    })
    resp, err := moneroClient.GetBalance(context.Background(), &walletrpc.GetBalanceRequest{})
    if err != nil {
		log.Fatal(err)
    }
    fmt.Println("Total balance:", walletrpc.XMRToDecimal(resp.Balance))
    fmt.Println("Unlocked balance:", walletrpc.XMRToDecimal(resp.UnlockedBalance))

    if err := electrumClient.LoadWallet(); err != nil {
        log.Fatal("Failed to load wallet:", err)
    }
    addresses, err := electrumClient.ListAddresses()
    if err != nil {
        log.Fatal("Failed to list addresses:", err)
    }
    if len(addresses) == 0 {
        panic("electrum does not work")
    }

    db.Connect()
    db.Migrate(db.Postgres)
    db.ConnectRedis()

    L := lua.NewState(db.RedisClient, db.Postgres, electrumClient, moneroClient)
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

    files, err := ioutil.ReadDir("mods")
    if err != nil {
        log.Fatal("Failed to read mods directory:", err)
    }

    for _, f := range files {
        if !f.IsDir() && strings.HasSuffix(f.Name(), ".lua") {
            path := filepath.Join("mods", f.Name())
            log.Println("Loading Lua module:", path)
            if err := lua.L.DoFile(path); err != nil {
                log.Fatalf("Error loading %s: %v", path, err)
            }
            lua.WatchLuaFile(lua.L, path)
        }
    }

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

    log.Println("Starting server on " + config.AppConfig.ListenAddr + ":" + config.AppConfig.Port)
    if err := s.Start(config.AppConfig.ListenAddr, config.AppConfig.Port); err != nil {
        log.Fatal(err)
    }
}

