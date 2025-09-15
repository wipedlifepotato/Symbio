package main

import (
	"io/ioutil"
	"log"
	"mFrelance/config"
	"mFrelance/db"
	"mFrelance/electrum"
	"mFrelance/lua"
	"mFrelance/server"
	serverhandlers "mFrelance/server/handlers"
	"net/http"
	"path/filepath"
	"strings"

	//"mFrelance/monero"
	"github.com/gabstv/httpdigest"
	"gitlab.com/moneropay/go-monero/walletrpc"

	//"fmt"
	"context"
	_ "mFrelance/docs"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title mFrelance API
// @version 1.0
// @description API for user registration, authentication, wallet operations and admin management
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	config.Init()

	electrumClient := electrum.NewClient(
		config.AppConfig.ElectrumUser,
		config.AppConfig.ElectrumPassword,
		config.AppConfig.ElectrumHost,
		config.MustAtoi(config.AppConfig.ElectrumPort),
	)
	log.Print(electrumClient.GetAllBalances([]string{"tb1qljppje9qdhtp39nk2lkgm53vmslmyd4cw3g4sr"}))
	moneroClient := walletrpc.New(walletrpc.Config{
		Address: "http://" + config.AppConfig.MoneroHost + ":" + config.AppConfig.MoneroPort + "/json_rpc",
		Client: &http.Client{
			Transport: httpdigest.New(config.AppConfig.MoneroUser, config.AppConfig.MoneroPassword), // Remove if no auth.
		},
	})

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

	s.Handle("/hello", serverhandlers.HelloHandler)
	s.Handle("/register", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.RegisterHandler(w, r, db.RedisClient)
	})
	s.Handle("/auth", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.AuthHandler(w, r, db.RedisClient)
	})
	s.Handle("/captcha", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.CaptchaHandler(w, r, db.RedisClient)
	})
	s.Handle("/verify", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.VerifyHandler(w, r, db.RedisClient)
	})
	s.Handle("/restoreuser", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.RestoreHandler(w, r, db.RedisClient)
	})

	apiMux := http.NewServeMux()
	apiMux.Handle("/test", server.AuthMiddleware(http.HandlerFunc(serverhandlers.TestHandler)))
	apiMux.Handle("/wallet", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.WalletHandler(w, r, moneroClient, electrumClient)
	})))

	//apiMux.Handle("/wallet/update", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		server.UpdateBalanceHandler(w, r, moneroClient, electrumClient)
	//    })))
	apiMux.Handle("/wallet/moneroSend", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.SendMoneroHandler(w, r, moneroClient)
	})))
	apiMux.Handle("/wallet/bitcoinSend", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.SendElectrumHandler(w, r, electrumClient)
	})))

	apiMux.Handle("/admin/make", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.MakeAdminHandler)))
	apiMux.Handle("/admin/remove", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.RemoveAdminHandler)))
	apiMux.Handle("/admin/check", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.IsAdminHandler)))
	apiMux.Handle("/admin/block", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.BlockUserHandler)))
	apiMux.Handle("/admin/unblock", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.UnblockUserHandler)))
	apiMux.Handle("/admin/transactions", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminTransactionsHandler)))
	apiMux.Handle("/admin/wallets", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminWalletsHandler)))
	apiMux.Handle("/admin/update_balance", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminUpdateBalanceHandler)))
	apiMux.Handle("/admin/getRandomTicket", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminGetRandomTicketHandler)))

	apiMux.Handle("/ticket/my", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetMyTicketsHandler)))
	apiMux.Handle("/ticket/messages", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetTicketMessagesHandler)))
	apiMux.Handle("/ticket/write", server.AuthMiddleware(http.HandlerFunc(serverhandlers.WriteToTicketHandler)))
	apiMux.Handle("/ticket/exit", server.AuthMiddleware(http.HandlerFunc(serverhandlers.ExitFromTicketHandler)))
	apiMux.Handle("/ticket/close", server.AuthMiddleware(http.HandlerFunc(serverhandlers.CloseTicketHandler)))

	apiMux.Handle("/ticket/createTicket", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.CreateTicket(w, r)
	})))

	apiMux.Handle("/chat/createChatRequest", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.CreateChatRequestHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/UpdateChatRequest", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.UpdateChatRequestHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/getChatRoomsForUser", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.GetChatRoomsForUserHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/getChatMessages", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.GetChatMessagesHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/sendMessage", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.SendMessageHandler().ServeHTTP(w, r)
	})))
	s.HandleHandler("/api/", http.StripPrefix("/api", apiMux))
	s.Handle("/profile", func(w http.ResponseWriter, r *http.Request) {
		server.AuthMiddleware(serverhandlers.ProfileHandler()).ServeHTTP(w, r)
	})
	s.Handle("/profiles", func(w http.ResponseWriter, r *http.Request) {
		server.AuthMiddleware(serverhandlers.ProfilesHandler()).ServeHTTP(w, r)
	})
	s.Handle("/profile/by_id", func(w http.ResponseWriter, r *http.Request) {
		server.AuthMiddleware(serverhandlers.ProfileByIDHandler()).ServeHTTP(w, r)
	})

	s.Handle("/swagger/", httpSwagger.WrapHandler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go server.StartWalletSync(ctx, electrumClient, moneroClient, 30*time.Second)
	go server.StartTxBlockTransactions(ctx, electrumClient, 15*time.Second)

	server.StartTxPoolFlusher(electrumClient, 15*time.Second, int(config.AppConfig.MaxAddrPerBlock))
	server.SetTxPoolBlocked(false)
	log.Println("Starting server on " + config.AppConfig.ListenAddr + ":" + config.AppConfig.Port)
	if err := s.Start(config.AppConfig.ListenAddr, config.AppConfig.Port); err != nil {
		log.Fatal(err)
	}

}
