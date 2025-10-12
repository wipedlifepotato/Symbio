package main

import (
	"context"
	"github.com/gabstv/httpdigest"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/tetratelabs/wazero"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"io/ioutil"
	"log"
	"mFrelance/config"
	"mFrelance/db"
	_ "mFrelance/docs"
	"mFrelance/electrum"
	"mFrelance/lua"
	"mFrelance/server"
	serverhandlers "mFrelance/server/handlers"
	"net/http"
	"path/filepath"
	"strings"
)

// @title mFrelance API
// @version 1.0
// @description API for user registration, authentication, wallet operations and admin management
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	var useWasm bool
	useWasm = true

	ctxWasm := context.Background()
	config.Init()

	electrumClient := electrum.NewClient(
		config.AppConfig.ElectrumUser,
		config.AppConfig.ElectrumPassword,
		config.AppConfig.ElectrumHost,
		viper.GetInt("electrum.port"),
	)
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

	wazeroRuntime := wazero.NewRuntime(ctxWasm)
	defer wazeroRuntime.Close(ctxWasm)
	files, err := ioutil.ReadDir("modsWasm")
	if err != nil {
		useWasm = false
	}
	if useWasm {
		for _, f := range files {
			if filepath.Ext(f.Name()) != ".wasm" {
				continue
			}

			wasmPath := filepath.Join("modsWasm", f.Name())
			err = lua.LoadWasmModule(L, wasmPath)
			if err != nil {
				log.Fatal(err.Error())
			}

			log.Printf("Loaded WASM mod: %s", f.Name())
		}
	}
	if err := L.DoString(`
        print("Lua says hi")
        local result = helloGo("Alice")
        print("Lua got:", result)
    `); err != nil {
		panic(err)
	}

	s := server.New()
	lua.RegisterHttpHandler(lua.L, s.GetMux())

	files, err = ioutil.ReadDir("mods")
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
	s.Handle("/captcha/status", serverhandlers.CaptchaStatusHandler)
	s.Handle("/restoreuser", func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.RestoreHandler(w, r, db.RedisClient)
	})

	apiMux := http.NewServeMux()
	apiMux.Handle("/test", server.AuthMiddleware(http.HandlerFunc(serverhandlers.TestHandler)))
	apiMux.Handle("/ownID", server.AuthMiddleware(http.HandlerFunc(serverhandlers.OwnIdHandler())))

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
	apiMux.Handle("/admin/check", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.IsAdminHandler(w, r)
	})))
	apiMux.Handle("/admin/IIsAdmin", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.IsIAdminHandler(w, r)
	})))
	apiMux.Handle("/admin/block", server.AuthMiddleware(server.RequirePermission(server.PermUserBlock)(serverhandlers.BlockUserHandler)))
	apiMux.Handle("/admin/unblock", server.AuthMiddleware(server.RequirePermission(server.PermUserBlock)(serverhandlers.UnblockUserHandler)))
	apiMux.Handle("/admin/transactions", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminTransactionsHandler)))
	apiMux.Handle("/admin/wallets", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminWalletsHandler)))
	apiMux.Handle("/admin/update_balance", server.AuthMiddleware(server.RequirePermission(server.PermBalanceChange)(serverhandlers.AdminUpdateBalanceHandler)))
	apiMux.Handle("/admin/delete_user_tasks", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminDeleteUserTasksHandler)))
	apiMux.Handle("/admin/getRandomTicket", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminGetRandomTicketHandler)))
	apiMux.Handle("/admin/tickets", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.GetAllTicketsHandler)))
	apiMux.Handle("/admin/addUserToChatRoom", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.AdminAddUserToChatRoom)))
	apiMux.Handle("/admin/deleteChatRoom", server.AuthMiddleware(serverhandlers.RequireAdmin(serverhandlers.DeleteChatRoom)))

	// Task management routes
	apiMux.Handle("/tasks/create", server.AuthMiddleware(serverhandlers.CreateTaskHandler()))
	apiMux.Handle("/tasks", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetTasksHandler())))
	apiMux.Handle("/tasks/get", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetTaskHandler())))
	apiMux.Handle("/tasks/update", server.AuthMiddleware(http.HandlerFunc(serverhandlers.UpdateTaskHandler())))
	apiMux.Handle("/tasks/delete", server.AuthMiddleware(http.HandlerFunc(serverhandlers.DeleteTaskHandler())))

	// Task offers routes
	apiMux.Handle("/offers/create", server.AuthMiddleware(http.HandlerFunc(serverhandlers.CreateTaskOfferHandler())))
	apiMux.Handle("/offers", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetTaskOffersHandler())))
	apiMux.Handle("/offers/accept", server.AuthMiddleware(http.HandlerFunc(serverhandlers.AcceptTaskOfferHandler())))
	apiMux.Handle("/offers/update", server.AuthMiddleware(http.HandlerFunc(serverhandlers.UpdateTaskOfferHandler())))
	apiMux.Handle("/offers/delete", server.AuthMiddleware(http.HandlerFunc(serverhandlers.DeleteTaskOfferHandler())))
	apiMux.Handle("/tasks/complete", server.AuthMiddleware(http.HandlerFunc(serverhandlers.CompleteTaskHandler())))

	// Dispute routes
	apiMux.Handle("/disputes/create", server.AuthMiddleware(http.HandlerFunc(serverhandlers.CreateDisputeHandler())))
	apiMux.Handle("/disputes/get", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetDisputeHandler())))
	apiMux.Handle("/disputes/message", server.AuthMiddleware(http.HandlerFunc(serverhandlers.SendDisputeMessageHandler())))
	apiMux.Handle("/disputes/my", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetUserDisputesHandler())))

	// Review routes
	apiMux.Handle("/reviews/create", server.AuthMiddleware(http.HandlerFunc(serverhandlers.CreateReviewHandler())))
	apiMux.Handle("/reviews/user", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetReviewsByUserHandler())))
	apiMux.Handle("/reviews/task", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetReviewsByTaskHandler())))
	apiMux.Handle("/reviews/rating", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetUserRatingHandler())))

	// Admin dispute routes
	apiMux.Handle("/admin/disputes", server.AuthMiddleware(server.RequirePermission(server.PermDisputeManage)(http.HandlerFunc(serverhandlers.GetOpenDisputesHandler()))))
	apiMux.Handle("/admin/disputes/assign", server.AuthMiddleware(server.RequirePermission(server.PermDisputeManage)(http.HandlerFunc(serverhandlers.AssignDisputeHandler()))))
	apiMux.Handle("/admin/disputes/resolve", server.AuthMiddleware(server.RequirePermission(server.PermDisputeManage)(http.HandlerFunc(serverhandlers.ResolveDisputeHandler()))))
	apiMux.Handle("/admin/disputes/details", server.AuthMiddleware(server.RequirePermission(server.PermDisputeManage)(http.HandlerFunc(serverhandlers.GetDisputeDetailsHandler()))))
	apiMux.Handle("/api/disputes/details", server.AuthMiddleware(http.HandlerFunc(serverhandlers.GetDisputeDetailsForUserHandler())))

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
	apiMux.Handle("/chat/getChatRequests", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.GetChatRequestsHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/acceptChatRequest", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.AcceptChatRequestHandler().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/exitFromChat", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.ExitFromChat().ServeHTTP(w, r)
	})))
	apiMux.Handle("/chat/cancelChatRequest", server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverhandlers.CancelChatRequestHandler().ServeHTTP(w, r)
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
	go server.StartWalletSync(ctx, electrumClient, moneroClient, config.AppConfig.WalletSyncInterval)
	go server.StartTxBlockTransactions(ctx, electrumClient, config.AppConfig.TxBlockInterval)

	server.StartTxPoolFlusher(electrumClient, moneroClient, config.AppConfig.TxPoolFlushInterval, int(config.AppConfig.MaxAddrPerBlock))
	server.SetTxPoolBlocked(false)
	log.Println("Starting server on " + config.AppConfig.ListenAddr + ":" + config.AppConfig.Port)
	if err := s.Start(config.AppConfig.ListenAddr, config.AppConfig.Port); err != nil {
		log.Fatal(err)
	}

}
