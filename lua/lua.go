package lua

import (
	"fmt"
	"log"
	"mFrelance/auth"
	"mFrelance/db"
	"mFrelance/electrum"
	"mFrelance/models"
	"mFrelance/server"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	lua "github.com/yuin/gopher-lua"

	//"mFrelance/server"
	"math/big"

	"gitlab.com/moneropay/go-monero/walletrpc"
)

var L *lua.LState

type LuaVM struct {
	L *lua.LState
}

func WatchLuaFile(L *lua.LState, path string) {
	log.Printf("[LUA] Start monitoring %s", path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				switch {
				case event.Op&fsnotify.Write == fsnotify.Write, event.Op&fsnotify.Create == fsnotify.Create:
					log.Println("[LUA] Reloading Lua handlers:", event.Name)
					if err := L.DoFile(path); err != nil {
						log.Println("Error reloading Lua:", err)
					}

				case event.Op&fsnotify.Remove == fsnotify.Remove, event.Op&fsnotify.Rename == fsnotify.Rename:
					log.Println("[LUA] File removed or renamed, re-adding watcher:", event.Name)
					time.Sleep(100 * time.Millisecond)
					watcher.Remove(path)
					err := watcher.Add(path)
					if err != nil {
						log.Println("Error re-adding watcher:", err)
					}
					if err := L.DoFile(path); err != nil {
						log.Println("Error reloading Lua:", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("[LUA] Watcher error:", err)
			}
		}
	}()
}
func RegisterLuaDisputes(L *lua.LState) {
	L.SetGlobal("create_dispute", L.NewFunction(func(L *lua.LState) int {
		taskID := L.CheckInt64(1)
		openedBy := L.CheckInt64(2)
		dispute := &models.Dispute{
			TaskID:    taskID,
			OpenedBy:  openedBy,
			Status:    "open",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.CreateDispute(dispute); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(dispute.ID))
		return 1
	}))

	L.SetGlobal("get_dispute", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckInt64(1)
		dispute, err := db.GetDisputeByID(id)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(dispute.ID))
		tbl.RawSetString("task_id", lua.LNumber(dispute.TaskID))
		tbl.RawSetString("opened_by", lua.LNumber(dispute.OpenedBy))
		if dispute.AssignedAdmin != nil {
		    tbl.RawSetString("assigned_admin", lua.LNumber(*dispute.AssignedAdmin))
		} else {
		    tbl.RawSetString("assigned_admin", lua.LNil)
		}
		tbl.RawSetString("status", lua.LString(dispute.Status))
		if dispute.Resolution != nil {
			tbl.RawSetString("resolution", lua.LString(*dispute.Resolution))
		} else {
			tbl.RawSetString("resolution", lua.LNil)
		}
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("get_dispute_messages", L.NewFunction(func(L *lua.LState) int {
		disputeID := L.CheckInt64(1)
		messages, err := db.GetDisputeMessages(disputeID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		for _, m := range messages {
			msgTbl := L.NewTable()
			msgTbl.RawSetString("id", lua.LNumber(m.ID))
			msgTbl.RawSetString("dispute_id", lua.LNumber(m.DisputeID))
			msgTbl.RawSetString("sender_id", lua.LNumber(m.SenderID))
			msgTbl.RawSetString("message", lua.LString(m.Message))
			msgTbl.RawSetString("created_at", lua.LString(m.CreatedAt.Format(time.RFC3339)))
			tbl.Append(msgTbl)
		}
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("get_dispute_messages_paged", L.NewFunction(func(L *lua.LState) int {
		disputeID := L.CheckInt64(1)
		limit := L.ToInt(2)
		offset := L.ToInt(3)
		messages, err := db.GetDisputeMessagesPaged(disputeID, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		for _, m := range messages {
			msgTbl := L.NewTable()
			msgTbl.RawSetString("id", lua.LNumber(m.ID))
			msgTbl.RawSetString("dispute_id", lua.LNumber(m.DisputeID))
			msgTbl.RawSetString("sender_id", lua.LNumber(m.SenderID))
			msgTbl.RawSetString("message", lua.LString(m.Message))
			msgTbl.RawSetString("created_at", lua.LString(m.CreatedAt.Format(time.RFC3339)))
			tbl.Append(msgTbl)
		}
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("update_dispute_status", L.NewFunction(func(L *lua.LState) int {
		id := L.CheckInt64(1)
		status := L.CheckString(2)
		var resolution *string
		if L.Get(3) != lua.LNil {
		    r := L.CheckString(3)
		    resolution = &r
		}
		if err := db.UpdateDisputeStatus(id, status, resolution); err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LTrue)
		return 1
	}))
}

func RegisterLuaEscrow(L *lua.LState) {
	L.SetGlobal("create_escrow", L.NewFunction(func(L *lua.LState) int {
		taskID := L.CheckInt64(1)
		clientID := L.CheckInt64(2)
		freelancerID := L.CheckInt64(3)
		amount := L.CheckString(4)
		currency := L.CheckString(5)

		amt, ok := new(big.Float).SetString(amount)
		if !ok {
		    L.Push(lua.LNil)
		    L.Push(lua.LString("invalid amount"))
		    return 2
		}
		f, _ := amt.Float64()
		escrow := &models.EscrowBalance{
		    TaskID: taskID,
		    ClientID: clientID,
		    FreelancerID: freelancerID,
		    Amount: f,
		    Currency: currency,
		    Status: "pending",
		    CreatedAt: time.Now(),
		}
		if err := db.CreateEscrowBalance(escrow); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(escrow.ID))
		return 1
	}))

	L.SetGlobal("get_escrow_by_task", L.NewFunction(func(L *lua.LState) int {
		taskID := L.CheckInt64(1)
		escrow, err := db.GetEscrowBalanceByTaskID(taskID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(escrow.ID))
		tbl.RawSetString("task_id", lua.LNumber(escrow.TaskID))
		tbl.RawSetString("client_id", lua.LNumber(escrow.ClientID))
		tbl.RawSetString("freelancer_id", lua.LNumber(escrow.FreelancerID))
		tbl.RawSetString("amount", lua.LString(fmt.Sprintf("%f", escrow.Amount)))
		tbl.RawSetString("currency", lua.LString(escrow.Currency))
		tbl.RawSetString("status", lua.LString(escrow.Status))
		L.Push(tbl)
		return 1
	}))
}

func RegisterLuaReviews(L *lua.LState) {
	L.SetGlobal("create_review", L.NewFunction(func(L *lua.LState) int {
		taskID := L.CheckInt64(1)
		reviewerID := L.CheckInt64(2)
		reviewedID := L.CheckInt64(3)
		rating := L.CheckInt(4)
		comment := L.CheckString(5)

		review := &models.Review{
			TaskID:     taskID,
			ReviewerID: reviewerID,
			ReviewedID: reviewedID,
			Rating:     rating,
			Comment:    comment,
			CreatedAt:  time.Now(),
		}
		if err := db.CreateReview(review); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(review.ID))
		return 1
	}))

	L.SetGlobal("get_reviews_by_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.CheckInt64(1)
		reviews, err := db.GetReviewsByUserID(userID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		for _, r := range reviews {
			rTbl := L.NewTable()
			rTbl.RawSetString("id", lua.LNumber(r.ID))
			rTbl.RawSetString("task_id", lua.LNumber(r.TaskID))
			rTbl.RawSetString("reviewer_id", lua.LNumber(r.ReviewerID))
			rTbl.RawSetString("reviewed_id", lua.LNumber(r.ReviewedID))
			rTbl.RawSetString("rating", lua.LNumber(r.Rating))
			rTbl.RawSetString("comment", lua.LString(r.Comment))
			rTbl.RawSetString("created_at", lua.LString(r.CreatedAt.Format(time.RFC3339)))
			tbl.Append(rTbl)
		}
		L.Push(tbl)
		return 1
	}))
}

func RegisterJWTLua(L *lua.LState) {
	L.SetGlobal("get_user_from_jwt", L.NewFunction(func(L *lua.LState) int {
		token := L.ToString(1)
		claims, err := auth.ParseJWT(token)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		tbl.RawSetString("user_id", lua.LNumber(claims.UserID))
		tbl.RawSetString("username", lua.LString(claims.Username))
		L.Push(tbl)
		return 1
	}))
}

func RegisterElectrumLua(L *lua.LState, client *electrum.Client) {

	L.SetGlobal("electrum_create_address", L.NewFunction(func(L *lua.LState) int {
		addr, err := client.CreateAddress()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(addr))
		return 1
	}))

	L.SetGlobal("electrum_set_withdraw_blocked", L.NewFunction(func(L *lua.LState) int {
		blocked := L.CheckBool(1)
		server.SetTxPoolBlocked(blocked)
		return 0
	}))

	L.SetGlobal("electrum_is_withdraw_blocked", L.NewFunction(func(L *lua.LState) int {
		if server.IsTxPoolBlocked() {
			L.Push(lua.LTrue)
		} else {
			L.Push(lua.LFalse)
		}
		return 1
	}))
	L.SetGlobal("electrum_pay_to_many", L.NewFunction(func(L *lua.LState) int {

		tbl := L.CheckTable(1)
		var outputs [][2]string

		tbl.ForEach(func(_ lua.LValue, value lua.LValue) {
			subTbl, ok := value.(*lua.LTable)
			if !ok {
				return
			}
			addr := subTbl.RawGetInt(1).String()
			amt := subTbl.RawGetInt(2).String()
			outputs = append(outputs, [2]string{addr, amt})
		})

		txID, err := client.PayToMany(outputs)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(txID))
		return 1
	}))

	L.SetGlobal("electrum_get_balance", L.NewFunction(func(L *lua.LState) int {
		addr := L.ToString(1)
		bal, err := client.GetBalance(addr)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		f, _ := bal.Float64()
		L.Push(lua.LNumber(f))
		return 1
	}))

	L.SetGlobal("electrum_pay_to", L.NewFunction(func(L *lua.LState) int {
		destination := L.ToString(1)
		amount := L.ToString(2)
		if destination == "" || amount == "" {
			L.Push(lua.LNil)
			L.Push(lua.LString("destination and amount required"))
			return 2
		}

		txid, err := client.PayTo(destination, amount)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LString(txid))
		return 1
	}))

	L.SetGlobal("electrum_list_addresses", L.NewFunction(func(L *lua.LState) int {
		addrs, err := client.ListAddresses()
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tbl := L.NewTable()
		for _, a := range addrs {
			tbl.Append(lua.LString(a))
		}
		L.Push(tbl)
		return 1
	}))
}

func RegisterLuaHelpers(L *lua.LState, rdb *redis.Client, psql *sqlx.DB) {

	L.SetGlobal("generate_mnemonic", L.NewFunction(func(L *lua.LState) int {
		mnemonic := server.GenerateMnemonic()
		L.Push(lua.LString(mnemonic))
		return 1
	}))
	L.SetGlobal("get_transactions", L.NewFunction(func(L *lua.LState) int {
		walletID := L.ToInt64(1)
		limit := L.ToInt(2)
		offset := L.ToInt(3)

		txs, err := models.GetTransactionsByWallet(db.Postgres, walletID, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		for _, tx := range txs {
			txTbl := L.NewTable()
			txTbl.RawSetString("id", lua.LNumber(tx.ID))
			txTbl.RawSetString("from_wallet_id", lua.LNumber(tx.FromWalletID.Int64))
			txTbl.RawSetString("to_wallet_id", lua.LNumber(tx.ToWalletID.Int64))
			txTbl.RawSetString("to_address", lua.LString(tx.ToAddress.String))
			txTbl.RawSetString("task_id", lua.LNumber(tx.TaskID.Int64))
			txTbl.RawSetString("amount", lua.LString(tx.Amount))
			txTbl.RawSetString("currency", lua.LString(tx.Currency))
			txTbl.RawSetString("confirmed", lua.LBool(tx.Confirmed))
			txTbl.RawSetString("created_at", lua.LString(tx.CreatedAt.Format(time.RFC3339)))
			tbl.Append(txTbl)
		}

		L.Push(tbl)
		return 1
	}))
	L.SetGlobal("get_wallet", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToString(1)
		currency := L.ToString(2)

		var id int64
		var balanceStr, address string
		err := psql.QueryRow(`SELECT id, balance::text, address FROM wallets WHERE user_id=$1 AND currency=$2 LIMIT 1`, userID, currency).
			Scan(&id, &balanceStr, &address)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(id))
		tbl.RawSetString("balance", lua.LString(balanceStr))
		tbl.RawSetString("address", lua.LString(address))

		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("set_balance", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToString(1)
		currency := L.ToString(2)
		newBalance := L.ToString(3)

		_, err := psql.Exec(`UPDATE wallets SET balance=$1 WHERE user_id=$2 AND currency=$3`, newBalance, userID, currency)
		if err != nil {
			L.Push(lua.LFalse)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LTrue)
		return 1
	}))
	L.SetGlobal("get_balance", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToString(1)
		currency := L.ToString(2)

		var balanceStr string
		err := psql.QueryRow(`SELECT balance::text FROM wallets WHERE user_id=$1 AND currency=$2 LIMIT 1`, userID, currency).Scan(&balanceStr)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(balanceStr))
		return 1
	}))

	L.SetGlobal("add_balance", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToString(1)
		currency := L.ToString(2)
		amountStr := L.ToString(3)

		amount, ok := new(big.Float).SetString(amountStr)
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LString("invalid amount"))
			return 2
		}

		_, err := psql.Exec(`UPDATE wallets SET balance = balance + $1 WHERE user_id=$2 AND currency=$3`, amount.Text('f', 8), userID, currency)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("sub_balance", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToString(1)
		currency := L.ToString(2)
		amountStr := L.ToString(3)

		amount, ok := new(big.Float).SetString(amountStr)
		if !ok {
			L.Push(lua.LNil)
			L.Push(lua.LString("invalid amount"))
			return 2
		}

		var balanceStr string
		err := psql.QueryRow(`SELECT balance::text FROM wallets WHERE user_id=$1 AND currency=$2 LIMIT 1`, userID, currency).Scan(&balanceStr)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		balance, _ := new(big.Float).SetString(balanceStr)
		if balance.Cmp(amount) < 0 {
			L.Push(lua.LNil)
			L.Push(lua.LString("insufficient balance"))
			return 2
		}

		newBalance := new(big.Float).Sub(balance, amount)
		_, err = psql.Exec(`UPDATE wallets SET balance=$1 WHERE user_id=$2 AND currency=$3`, newBalance.Text('f', 8), userID, currency)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		L.Push(lua.LBool(true))
		return 1
	}))
	L.SetGlobal("is_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		admin, err := db.IsAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(admin))
		return 1
	}))

	L.SetGlobal("make_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		err := db.MakeAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("remove_admin", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		err := db.RemoveAdmin(psql, userID)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("add_permission", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		perm := L.ToInt(2)
		err := db.AddPermission(psql, userID, perm)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("remove_permission", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		perm := L.ToInt(2)
		err := db.RemovePermission(psql, userID, perm)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("set_permissions", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		permissions := L.ToInt(2)
		err := db.SetPermissions(psql, userID, permissions)
		if err != nil {
			L.Push(lua.LBool(false))
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("get_user", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		userID, passwordHash, err := db.GetUserByUsername(psql, username)
		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(userID))
		tbl.RawSetString("password_hash", lua.LString(passwordHash))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("block_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToInt64(1)
		err := db.BlockUser(psql, userID)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("get_profile", L.NewFunction(func(L *lua.LState) int {
		userID := L.CheckInt64(1)
		profile, err := models.GetProfile(psql, userID)
		if err != nil {
			L.RaiseError("GetProfile error: %v", err)
			return 0
		}
		tbl := L.NewTable()
		tbl.RawSetString("user_id", lua.LNumber(profile.UserID))
		tbl.RawSetString("full_name", lua.LString(profile.FullName))
		tbl.RawSetString("bio", lua.LString(profile.Bio))

		skillsTbl := L.NewTable()
		for _, s := range profile.Skills {
			skillsTbl.Append(lua.LString(s))
		}
		tbl.RawSetString("skills", skillsTbl)

		tbl.RawSetString("avatar", lua.LString(profile.Avatar))
		tbl.RawSetString("rating", lua.LNumber(profile.Rating))
		tbl.RawSetString("completed_tasks", lua.LNumber(profile.CompletedTasks))

		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("upsert_profile", L.NewFunction(func(L *lua.LState) int {
		userID := L.CheckInt64(1)
		fullName := L.CheckString(2)
		bio := L.CheckString(3)
		skillsTbl := L.CheckTable(4)
		avatar := L.CheckString(5)

		skills := make(models.JSONStrings, 0)
		skillsTbl.ForEach(func(_, value lua.LValue) {
			if s, ok := value.(lua.LString); ok {
				skills = append(skills, string(s))
			}
		})

		profile := &models.Profile{
			UserID:   userID,
			FullName: fullName,
			Bio:      bio,
			Skills:   skills,
			Avatar:   avatar,
		}

		if err := models.UpsertProfile(psql, profile); err != nil {
			L.RaiseError("UpsertProfile error: %v", err)
			return 0
		}

		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("get_profiles", L.NewFunction(func(L *lua.LState) int {
		limit := L.CheckInt(1)
		offset := L.CheckInt(2)

		profiles, err := models.GetProfilesWithLimitOffset(psql, limit, offset)
		if err != nil {
			L.RaiseError("GetProfilesWithLimitOffset error: %v", err)
			return 0
		}

		tbl := L.NewTable()
		for _, p := range profiles {
			pTbl := L.NewTable()
			pTbl.RawSetString("user_id", lua.LNumber(p.UserID))
			pTbl.RawSetString("full_name", lua.LString(p.FullName))
			pTbl.RawSetString("bio", lua.LString(p.Bio))

			skillsTbl := L.NewTable()
			for _, s := range p.Skills {
				skillsTbl.Append(lua.LString(s))
			}
			pTbl.RawSetString("skills", skillsTbl)

			pTbl.RawSetString("avatar", lua.LString(p.Avatar))
			pTbl.RawSetString("rating", lua.LNumber(p.Rating))
			pTbl.RawSetString("completed_tasks", lua.LNumber(p.CompletedTasks))

			tbl.Append(pTbl)
		}

		L.Push(tbl)
		return 1
	}))
	// unblock_user(userID)
	L.SetGlobal("unblock_user", L.NewFunction(func(L *lua.LState) int {
		userID := L.ToInt64(1)
		err := db.UnblockUser(psql, userID)
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	L.SetGlobal("is_user_blocked", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		userID, _, err := db.GetUserByUsername(psql, username)
		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		blocked, err := db.IsUserBlocked(psql, userID)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LBool(blocked))
		return 1
	}))

	L.SetGlobal("verify_password", L.NewFunction(func(L *lua.LState) int {
		password := L.ToString(1)
		hashed := L.ToString(2)

		res := server.VerifyPassword(password, hashed)
		L.Push(lua.LBool(res))
		return 1
	}))

	L.SetGlobal("change_password", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		newPassword := L.ToString(2)

		hashed := server.HashPassword(newPassword)

		err := db.ChangeUserPassword(psql, username, string(hashed))
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbl := L.NewTable()
		tbl.RawSetString("username", lua.LString(username))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("restore_user", L.NewFunction(func(L *lua.LState) int {
		username := L.ToString(1)
		mnemonic := L.ToString(2)
		userID, usernameOut, err := db.RestoreUser(psql, username, mnemonic)

		if err != nil || userID == 0 {
			L.Push(lua.LNil)
			return 1
		}
		tbl := L.NewTable()
		tbl.RawSetString("id", lua.LNumber(userID))
		tbl.RawSetString("username", lua.LString(usernameOut))
		L.Push(tbl)
		return 1
	}))

	L.SetGlobal("generate_jwt", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		username := L.ToString(2)
		token, err := auth.GenerateJWT(userID, username)
		if err != nil {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(lua.LString(token))
		return 1
	}))

	// Создать чат
	L.SetGlobal("create_chat_room", L.NewFunction(func(L *lua.LState) int {
		room, err := db.CreateChatRoom(psql)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(room.ID))
		return 1
	}))

	// Добавить пользователя в чат
	L.SetGlobal("add_user_to_chat", L.NewFunction(func(L *lua.LState) int {
		userID := int64(L.ToInt(1))
		chatRoomID := int64(L.ToInt(2))
		err := db.AddUserToChatRoom(psql, userID, chatRoomID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// Получить участников чата
	L.SetGlobal("get_chat_participants", L.NewFunction(func(L *lua.LState) int {
		chatRoomID := int64(L.ToInt(1))
		participants, err := db.GetChatParticipants(psql, chatRoomID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, p := range participants {
			tb.Append(lua.LNumber(p.UserID))
		}
		L.Push(tb)
		return 1
	}))

	// Отправить сообщение в чат
	L.SetGlobal("create_chat_message", L.NewFunction(func(L *lua.LState) int {
		chatRoomID := int64(L.ToInt(1))
		senderID := int64(L.ToInt(2))
		msg := L.ToString(3)
		message := &models.ChatMessage{
			ChatRoomID: chatRoomID,
			SenderID:   senderID,
			Message:    msg,
			CreatedAt:  time.Now(),
		}
		err := db.CreateChatMessage(psql, message)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// Получить сообщения чата
	L.SetGlobal("get_chat_messages", L.NewFunction(func(L *lua.LState) int {
		chatRoomID := int64(L.ToInt(1))
		messages, err := db.GetChatMessages(psql, chatRoomID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, m := range messages {
			msgTbl := L.NewTable()
			msgTbl.RawSetString("sender_id", lua.LNumber(m.SenderID))
			msgTbl.RawSetString("message", lua.LString(m.Message))
			msgTbl.RawSetString("created_at", lua.LString(m.CreatedAt.Format(time.RFC3339)))
			tb.Append(msgTbl)
		}
		L.Push(tb)
		return 1
	}))

	L.SetGlobal("get_chat_messages_paged", L.NewFunction(func(L *lua.LState) int {
		chatRoomID := int64(L.ToInt(1))
		limit := L.ToInt(2)
		offset := L.ToInt(3)
		messages, err := db.GetChatMessagesPaged(psql, chatRoomID, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, m := range messages {
			msgTbl := L.NewTable()
			msgTbl.RawSetString("sender_id", lua.LNumber(m.SenderID))
			msgTbl.RawSetString("message", lua.LString(m.Message))
			msgTbl.RawSetString("created_at", lua.LString(m.CreatedAt.Format(time.RFC3339)))
			tb.Append(msgTbl)
		}
		L.Push(tb)
		return 1
	}))

	// Создать запрос на чат
	L.SetGlobal("create_chat_request", L.NewFunction(func(L *lua.LState) int {
		requesterID := int64(L.ToInt(1))
		requestedID := int64(L.ToInt(2))
		req := &models.ChatRequest{
			RequesterID: requesterID,
			RequestedID: requestedID,
			Status:      "pending",
			CreatedAt:   time.Now(),
		}
		err := db.CreateChatRequest(psql, req)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// Принять запрос на чат
	L.SetGlobal("accept_chat_request", L.NewFunction(func(L *lua.LState) int {
		requesterID := int64(L.ToInt(1))
		requestedID := int64(L.ToInt(2))
		err := db.AcceptChatRequest(psql, requesterID, requestedID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// Отменить запрос на чат
	L.SetGlobal("delete_chat_request", L.NewFunction(func(L *lua.LState) int {
		requesterID := int64(L.ToInt(1))
		requestedID := int64(L.ToInt(2))
		err := db.DeleteChatRequest(psql, requesterID, requestedID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))

	// Удалить участника чата
	L.SetGlobal("delete_chat_participant", L.NewFunction(func(L *lua.LState) int {
		chatRoomID := int64(L.ToInt(1))
		userID := int64(L.ToInt(2))
		err := db.DeleteChatParticipant(psql, chatRoomID, userID)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LBool(true))
		return 1
	}))
}

var registeredPaths = make(map[string]bool)

func RegisterHttpHandler(L *lua.LState, mux *http.ServeMux) {
	L.SetGlobal("register_handler", L.NewFunction(func(L *lua.LState) int {
		path := L.ToString(1)
		handlerFn := L.ToFunction(2)

		if registeredPaths[path] {
			return 0
		}
		registeredPaths[path] = true

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			reqTable := L.NewTable()
			reqTable.RawSetString("method", lua.LString(r.Method))

			paramsTable := L.NewTable()
			for key, values := range r.Form {
				if len(values) > 0 {
					paramsTable.RawSetString(key, lua.LString(values[0]))
				}
			}
			reqTable.RawSetString("params", paramsTable)

			if err := L.CallByParam(lua.P{
				Fn:      handlerFn,
				NRet:    1,
				Protect: true,
			}, reqTable); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			ret := L.Get(-1)
			L.Pop(1)
			w.Write([]byte(ret.String()))
		})
		return 0
	}))
}
func RegisterLuaTasks(L *lua.LState, psql *sqlx.DB) {
	L.SetGlobal("count_open_tasks", L.NewFunction(func(L *lua.LState) int {
		n, err := db.CountOpenTasks(psql)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(n))
		return 1
	}))

	L.SetGlobal("count_tasks_by_client_and_status", L.NewFunction(func(L *lua.LState) int {
		clientID := int64(L.ToInt(1))
		status := L.ToString(2)
		n, err := db.CountTasksByClientAndStatus(psql, clientID, status)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LNumber(n))
		return 1
	}))

	L.SetGlobal("get_tasks_by_client_paged", L.NewFunction(func(L *lua.LState) int {
		clientID := int64(L.ToInt(1))
		limit := L.ToInt(2)
		offset := L.ToInt(3)
		tasks, err := db.GetTasksByClientIDPaged(psql, clientID, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, t := range tasks {
			taskTbl := L.NewTable()
			taskTbl.RawSetString("id", lua.LNumber(t.ID))
			taskTbl.RawSetString("title", lua.LString(t.Title))
			taskTbl.RawSetString("status", lua.LString(t.Status))
			taskTbl.RawSetString("client_id", lua.LNumber(t.ClientID))
			taskTbl.RawSetString("created_at", lua.LString(t.CreatedAt.Format(time.RFC3339)))
			tb.Append(taskTbl)
		}
		L.Push(tb)
		return 1
	}))

	L.SetGlobal("get_open_tasks_paged", L.NewFunction(func(L *lua.LState) int {
		limit := L.ToInt(1)
		offset := L.ToInt(2)
		tasks, err := db.GetOpenTasksPaged(psql, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, t := range tasks {
			taskTbl := L.NewTable()
			taskTbl.RawSetString("id", lua.LNumber(t.ID))
			taskTbl.RawSetString("title", lua.LString(t.Title))
			taskTbl.RawSetString("status", lua.LString(t.Status))
			taskTbl.RawSetString("client_id", lua.LNumber(t.ClientID))
			taskTbl.RawSetString("created_at", lua.LString(t.CreatedAt.Format(time.RFC3339)))
			tb.Append(taskTbl)
		}
		L.Push(tb)
		return 1
	}))

	L.SetGlobal("get_tasks_by_client_and_status_paged", L.NewFunction(func(L *lua.LState) int {
		clientID := int64(L.ToInt(1))
		status := L.ToString(2)
		limit := L.ToInt(3)
		offset := L.ToInt(4)
		tasks, err := db.GetTasksByClientIDAndStatusPaged(psql, clientID, status, limit, offset)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		tb := L.NewTable()
		for _, t := range tasks {
			taskTbl := L.NewTable()
			taskTbl.RawSetString("id", lua.LNumber(t.ID))
			taskTbl.RawSetString("title", lua.LString(t.Title))
			taskTbl.RawSetString("status", lua.LString(t.Status))
			taskTbl.RawSetString("client_id", lua.LNumber(t.ClientID))
			taskTbl.RawSetString("created_at", lua.LString(t.CreatedAt.Format(time.RFC3339)))
			tb.Append(taskTbl)
		}
		L.Push(tb)
		return 1
	}))
}

func luaInit(l *lua.LState, rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) {
	//l := lua.NewState()
	l.SetGlobal("helloGo", L.NewFunction(HelloLua))
	RegisterLuaRedis(l, rdb)
	RegisterLuaPostgres(l, psql)
	RegisterConfigGlobals(l)
	RegisterLuaHelpers(l, rdb, psql)
	RegisterElectrumLua(l, eClient)
	RegisterJWTLua(l)
	RegisterMoneroLua(l, mClient)
	RegisterLuaDisputes(l)
	RegisterLuaEscrow(l)
	RegisterLuaReviews(l)
	RegisterBigMath(l)
	RegisterLuaTasks(l, psql)
	return
}

func NewVM(rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) *LuaVM {
	l := lua.NewState()
	luaInit(l, rdb, psql, eClient, mClient)

	return &LuaVM{L: l}
}

func (vm *LuaVM) Close() {
	vm.L.Close()
}

func HelloLua(L *lua.LState) int {
	name := L.ToString(1)
	fmt.Println("Hello from Go,", name)
	L.Push(lua.LString("Hi, " + name))
	return 1
}

func NewState(rdb *redis.Client, psql *sqlx.DB, eClient *electrum.Client, mClient *walletrpc.Client) *lua.LState {
	L = lua.NewState()
	luaInit(L, rdb, psql, eClient, mClient)
	return L
}
