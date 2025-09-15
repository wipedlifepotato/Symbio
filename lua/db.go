package lua

import (
	"context"
	"fmt"
	"log"
	"mFrelance/db"
	"mFrelance/models"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	lua "github.com/yuin/gopher-lua"
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

func RegisterLuaPostgres(L *lua.LState, dbPg *sqlx.DB) {
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
			rows, err := dbPg.Queryx(query, params...)
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
			res, err := dbPg.Exec(query, params...)
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

	L.SetGlobal("create_task", L.NewFunction(func(L *lua.LState) int {
		tbl := L.ToTable(1)
		if tbl == nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString("expected table argument"))
			return 2
		}

		task := &models.Task{
			ClientID:    int64(tbl.RawGetInt(1).(lua.LNumber)),
			Title:       tbl.RawGetInt(2).String(),
			Description: tbl.RawGetInt(3).String(),
			Category:    tbl.RawGetInt(4).String(),
			Budget:      float64(tbl.RawGetInt(5).(lua.LNumber)),
			Currency:    tbl.RawGetInt(6).String(),
			Status:      tbl.RawGetInt(7).String(),
		}

		err := db.CreateTask(dbPg, task)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("get_task", L.NewFunction(func(L *lua.LState) int {
		id := L.ToInt64(1)
		task, err := db.GetTask(dbPg, id)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(task.ID))
		tbl.RawSetString("client_id", lua.LNumber(task.ClientID))
		tbl.RawSetString("title", lua.LString(task.Title))
		tbl.RawSetString("description", lua.LString(task.Description))
		tbl.RawSetString("category", lua.LString(task.Category))
		tbl.RawSetString("budget", lua.LNumber(task.Budget))
		tbl.RawSetString("currency", lua.LString(task.Currency))
		tbl.RawSetString("status", lua.LString(task.Status))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("create_chat_request", L.NewFunction(func(L *lua.LState) int {
		requesterID := L.ToInt64(1)
		requestedID := L.ToInt64(2)

		request := &models.ChatRequest{
			RequesterID: requesterID,
			RequestedID: requestedID,
			Status:      "pending",
			CreatedAt:   time.Now(),
		}

		err := db.CreateChatRequest(dbPg, request)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("update_chat_request", L.NewFunction(func(L *lua.LState) int {
		requesterID := L.ToInt64(1)
		requestedID := L.ToInt64(2)
		status := L.ToString(3)

		request := &models.ChatRequest{
			RequesterID: requesterID,
			RequestedID: requestedID,
			Status:      status,
		}

		err := db.UpdateChatRequest(dbPg, request)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("get_chat_messages", L.NewFunction(func(L *lua.LState) int {
		chatID := L.ToInt64(1)

		messages, err := db.GetChatMessages(dbPg, chatID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		for i, message := range messages {
			messageTable := L.NewTable()
			messageTable.RawSetString("id", lua.LNumber(message.ID))
			messageTable.RawSetString("chat_room_id", lua.LNumber(message.ChatRoomID))
			messageTable.RawSetString("sender_id", lua.LNumber(message.SenderID))
			messageTable.RawSetString("message", lua.LString(message.Message))
			messageTable.RawSetString("created_at", lua.LString(message.CreatedAt.String()))
			tbl.RawSetInt(i+1, messageTable)
		}
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("send_chat_message", L.NewFunction(func(L *lua.LState) int {
		chatID := L.ToInt64(1)
		senderID := L.ToInt64(2)
		messageText := L.ToString(3)

		message := &models.ChatMessage{
			ChatRoomID: chatID,
			SenderID:   senderID,
			Message:    messageText,
			CreatedAt:  time.Now(),
		}

		err := db.CreateChatMessage(dbPg, message)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))

	L.SetGlobal("get_chat_rooms_for_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToInt64(1)
		chatRooms, err := db.GetChatRoomsForUser(dbPg, userID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		for i, room := range chatRooms {
			roomTable := L.NewTable()
			roomTable.RawSetString("id", lua.LNumber(room.ID))
			roomTable.RawSetString("created_at", lua.LString(room.CreatedAt.String()))
			tbl.RawSetInt(i+1, roomTable)
		}

		L.Push(tbl)
		return 1
	}))
}
