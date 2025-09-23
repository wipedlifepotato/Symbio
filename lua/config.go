package lua

import (
	"github.com/yuin/gopher-lua"
	"mFrelance/config"
	"math/big"
)

func RegisterConfigGlobals(L *lua.LState) {
	cfg := L.NewTable()
  
	stringFields := map[string]string{
		"PostgresHost":     config.AppConfig.PostgresHost,
		"PostgresPort":     config.AppConfig.PostgresPort,
		"PostgresUser":     config.AppConfig.PostgresUser,
		"PostgresPassword": config.AppConfig.PostgresPassword,
		"PostgresDB":       config.AppConfig.PostgresDB,
		"RedisHost":        config.AppConfig.RedisHost,
		"RedisPort":        config.AppConfig.RedisPort,
		"RedisPassword":    config.AppConfig.RedisPassword,
		"Port":             config.AppConfig.Port,
		"JWTToken":         config.AppConfig.JWTToken,
		"ListenAddr":       config.AppConfig.ListenAddr,
		"ElectrumHost":     config.AppConfig.ElectrumHost,
		"ElectrumPort":     config.AppConfig.ElectrumPort,
		"ElectrumUser":     config.AppConfig.ElectrumUser,
		"ElectrumPassword": config.AppConfig.ElectrumPassword,
		"MoneroHost":       config.AppConfig.MoneroHost,
		"MoneroPort":       config.AppConfig.MoneroPort,
		"MoneroUser":       config.AppConfig.MoneroUser,
		"MoneroPassword":   config.AppConfig.MoneroPassword,
		"MoneroAddress":    config.AppConfig.MoneroAddress,
		"BitcoinAddress":   config.AppConfig.BitcoinAddress,
	}

	for k, v := range stringFields {
		L.SetField(cfg, k, lua.LString(v))
	}

	floatFields := map[string]*big.Float{
		"MoneroCommission":  big.NewFloat(config.AppConfig.MoneroCommission),
		"BitcoinCommission": big.NewFloat(config.AppConfig.BitcoinCommission),
		"MaxProfiles":       big.NewFloat(float64(config.AppConfig.MaxProfiles)),
		"MaxAvatarSize":     big.NewFloat(float64(config.AppConfig.MaxAvatarSize)),
		"MaxAddrPerBlock":   big.NewFloat(float64(config.AppConfig.MaxAddrPerBlock)),
	}

	for k, v := range floatFields {
		f, _ := v.Float64()
		L.SetField(cfg, k, lua.LNumber(f))
	}
  
  // Captcha
	L.SetField(cfg, "CaptchaEnabled", lua.LBool(config.AppConfig.CaptchaEnabled))
	L.SetField(cfg, "CaptchaRateLimitPerMinute", lua.LNumber(config.AppConfig.CaptchaRateLimitPerMinute))
	L.SetField(cfg, "CaptchaRateLimitPerHour", lua.LNumber(config.AppConfig.CaptchaRateLimitPerHour))
	L.SetField(cfg, "CaptchaFontPath", lua.LString(config.AppConfig.CaptchaFontPath))

	L.SetGlobal("config", cfg)
}

