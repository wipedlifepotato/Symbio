package lua

import (
	"mFrelance/config"
	"github.com/yuin/gopher-lua"
)

func RegisterConfigGlobals(L *lua.LState) {
	cfg := L.NewTable()

	// Postgres
	L.SetField(cfg, "PostgresHost", lua.LString(config.AppConfig.PostgresHost))
	L.SetField(cfg, "PostgresPort", lua.LString(config.AppConfig.PostgresPort))
	L.SetField(cfg, "PostgresUser", lua.LString(config.AppConfig.PostgresUser))
	L.SetField(cfg, "PostgresPassword", lua.LString(config.AppConfig.PostgresPassword))
	L.SetField(cfg, "PostgresDB", lua.LString(config.AppConfig.PostgresDB))

	// Redis
	L.SetField(cfg, "RedisHost", lua.LString(config.AppConfig.RedisHost))
	L.SetField(cfg, "RedisPort", lua.LString(config.AppConfig.RedisPort))
	L.SetField(cfg, "RedisPassword", lua.LString(config.AppConfig.RedisPassword))

	// Server
	L.SetField(cfg, "Port", lua.LString(config.AppConfig.Port))
	L.SetField(cfg, "JWTToken", lua.LString(config.AppConfig.JWTToken))
	L.SetField(cfg, "ListenAddr", lua.LString(config.AppConfig.ListenAddr))

	// Electrum
	L.SetField(cfg, "ElectrumHost", lua.LString(config.AppConfig.ElectrumHost))
	L.SetField(cfg, "ElectrumPort", lua.LString(config.AppConfig.ElectrumPort))
	L.SetField(cfg, "ElectrumUser", lua.LString(config.AppConfig.ElectrumUser))
	L.SetField(cfg, "ElectrumPassword", lua.LString(config.AppConfig.ElectrumPassword))

	// Monero
	L.SetField(cfg, "MoneroHost", lua.LString(config.AppConfig.MoneroHost))
	L.SetField(cfg, "MoneroPort", lua.LString(config.AppConfig.MoneroPort))
	L.SetField(cfg, "MoneroUser", lua.LString(config.AppConfig.MoneroUser))
	L.SetField(cfg, "MoneroPassword", lua.LString(config.AppConfig.MoneroPassword))
	L.SetField(cfg, "MoneroAddress", lua.LString(config.AppConfig.MoneroAddress))
	L.SetField(cfg, "MoneroCommission", lua.LNumber(config.AppConfig.MoneroCommission))

	// Bitcoin
	L.SetField(cfg, "BitcoinAddress", lua.LString(config.AppConfig.BitcoinAddress))
	L.SetField(cfg, "BitcoinCommission", lua.LNumber(config.AppConfig.BitcoinCommission))
	// Constants
	L.SetField(cfg, "MaxProfiles", lua.LNumber(config.AppConfig.MaxProfiles))
	L.SetField(cfg, "MaxAvatarSize", lua.LNumber(config.AppConfig.MaxAvatarSize))	
	L.SetField(cfg, "MaxAddrPerBlock", lua.LNumber(config.AppConfig.MaxAddrPerBlock))	

	L.SetGlobal("config", cfg)
}

