package lua

import (
	"mFrelance/config"
	"github.com/yuin/gopher-lua"
)

func RegisterConfigGlobals(L *lua.LState) {
	cfg := L.NewTable()

	L.SetField(cfg, "PostgresHost", lua.LString(config.AppConfig.PostgresHost))
	L.SetField(cfg, "PostgresPort", lua.LString(config.AppConfig.PostgresPort))
	L.SetField(cfg, "PostgresUser", lua.LString(config.AppConfig.PostgresUser))
	L.SetField(cfg, "PostgresPassword", lua.LString(config.AppConfig.PostgresPassword))
	L.SetField(cfg, "PostgresDB", lua.LString(config.AppConfig.PostgresDB))
	L.SetField(cfg, "RedisHost", lua.LString(config.AppConfig.RedisHost))
	L.SetField(cfg, "RedisPort", lua.LString(config.AppConfig.RedisPort))
	L.SetField(cfg, "RedisPassword", lua.LString(config.AppConfig.RedisPassword))
	L.SetField(cfg, "Port", lua.LString(config.AppConfig.Port))
	L.SetField(cfg, "JWTToken", lua.LString(config.AppConfig.JWTToken))
	L.SetField(cfg, "ListenAddr", lua.LString(config.AppConfig.ListenAddr))

	L.SetGlobal("config", cfg)
}
