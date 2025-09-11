package lua

import (
    "fmt"
    "github.com/yuin/gopher-lua"
    "github.com/go-redis/redis/v8"
    "github.com/jmoiron/sqlx"
    "context"
    "log"
    "time"
)

func RegisterLuaRedis(L *lua.LState, rdb *redis.Client) {
	L.SetGlobal("get_captcha", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		val, err := rdb.Get(context.Background(), "captcha:"+id).Result()
		if err != nil {
			L.Push(lua.LNil)
		} else {
			L.Push(lua.LString(val))
		}
		return 1
	}))

	L.SetGlobal("set_captcha", L.NewFunction(func(L *lua.LState) int {
		id := L.ToString(1)
		val := L.ToString(2)
		exp := L.ToInt(3) // seconds
		err := rdb.Set(context.Background(), "captcha:"+id, val, time.Second*time.Duration(exp)).Err()
		if err != nil {
			L.Push(lua.LFalse)
		} else {
			L.Push(lua.LTrue)
		}
		return 1
	}))
}


func RegisterLuaPostgres(L *lua.LState, db *sqlx.DB) {
	L.SetGlobal("pg_query", L.NewFunction(func(L *lua.LState) int {
		query := L.ToString(1)
		params := []interface{}{}


		if tbl := L.ToTable(2); tbl != nil {
			tbl.ForEach(func(_, value lua.LValue) {
				params = append(params, value.String())
			})
		}


		switch {
		case len(query) >= 6 && (query[:6] == "SELECT" || query[:6] == "select"):
			rows, err := db.Queryx(query, params...)
			if err != nil {
				L.Push(lua.LNil)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			defer rows.Close()

			results := lua.LTable{}
			for rows.Next() {
				rowMap := make(map[string]interface{})
				if err := rows.MapScan(rowMap); err != nil {
					log.Println("Row scan error:", err)
					continue
				}

				tbl := lua.LTable{}
				for k, v := range rowMap {
					tbl.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
				}
				results.Append(&tbl)
			}
			L.Push(&results)
			return 1

		case len(query) >= 6 && (query[:6] == "INSERT" || query[:6] == "insert" ||
			query[:6] == "UPDATE" || query[:6] == "update" ||
			query[:6] == "DELETE" || query[:6] == "delete"):
			res, err := db.Exec(query, params...)
			if err != nil {
				L.Push(lua.LFalse)
				L.Push(lua.LString(err.Error()))
				return 2
			}
			affected, _ := res.RowsAffected()
			L.Push(lua.LNumber(affected))
			return 1

		default:
			L.Push(lua.LNil)
			L.Push(lua.LString("Only SELECT/INSERT/UPDATE/DELETE allowed"))
			return 2
		}
	}))
}

