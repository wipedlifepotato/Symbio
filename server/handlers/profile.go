package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    "mFrelance/config"
    "mFrelance/db"
    "mFrelance/models"
    "mFrelance/server"
)

// ProfileHandler godoc
// @Summary Get or update profile
// @Description Get current user profile (GET) or update profile (POST)
// @Tags profile
// @Accept json
// @Produce json
// @Success 200 {object} models.Profile
// @Failure 400 {string} string "invalid payload"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "db error"
// @Security BearerAuth
// @Router /profile [get]
// @Router /profile [post]
func ProfileHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        claims := server.GetUserFromContext(r)
        if claims == nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }
        maxAvatarSize := int(config.AppConfig.MaxAvatarSize) * 1024 * 1024
        switch r.Method {
        case "GET":
            profile, err := models.GetProfile(db.Postgres, claims.UserID)
            if err != nil {
                http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
                return
            }
            username, err := db.GetUsernameByID(db.Postgres, claims.UserID)
            if err != nil {
                http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
                return
            }
            server.SanitizeProfile(profile)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]any{
                "username": username,
                "profile":  profile,
            })
        case "POST":
            var p models.Profile
            if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
                log.Println("Decode error:", err)
                http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
                return
            }
            p.UserID = claims.UserID
            if len(p.Avatar) > maxAvatarSize {
                http.Error(w, "avatar too large", http.StatusBadRequest)
                return
            }
            server.SanitizeProfile(&p)
            if err := models.UpsertProfile(db.Postgres, &p); err != nil {
                http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
                return
            }
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        }
    }
}

// ProfilesHandler godoc
// @Summary List profiles
// @Description Returns paginated list of profiles
// @Tags profile
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} models.Profile
// @Failure 500 {string} string "db error"
// @Security BearerAuth
// @Router /profiles [get]
func ProfilesHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        limit := 50
        offset := 0
        if l := r.URL.Query().Get("limit"); l != "" {
            if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 500 {
                limit = v
            }
        }
        if limit > int(config.AppConfig.MaxProfiles) {
            limit = int(config.AppConfig.MaxProfiles)
        }
        if o := r.URL.Query().Get("offset"); o != "" {
            if v, err := strconv.Atoi(o); err == nil && v >= 0 {
                offset = v
            }
        }
        profiles, err := models.GetProfilesWithLimitOffset(db.Postgres, limit, offset)
        if err != nil {
            http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        json.NewEncoder(w).Encode(profiles)
    }
}


// ProfileByIDHandler godoc
// @Summary Get public profile by user_id
// @Description Returns sanitized profile and username by user_id
// @Tags profile
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} models.Profile
// @Failure 400 {string} string "invalid user_id"
// @Failure 404 {string} string "not found"
// @Router /profile/by_id [get]
func ProfileByIDHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userIDStr := r.URL.Query().Get("user_id")
        if userIDStr == "" {
            http.Error(w, "invalid user_id", http.StatusBadRequest)
            return
        }
        userID, err := strconv.ParseInt(userIDStr, 10, 64)
        if err != nil || userID <= 0 {
            http.Error(w, "invalid user_id", http.StatusBadRequest)
            return
        }

        prof, err := models.GetProfile(db.Postgres, userID)
        if err != nil {
            http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        if prof == nil {
            http.Error(w, "not found", http.StatusNotFound)
            return
        }
        username, err := db.GetUsernameByID(db.Postgres, userID)
        if err != nil {
            http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
            return
        }
        // включим username в ответе отдельно
        server.SanitizeProfile(prof)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "username": username,
            "profile":  prof,
        })
    }
}


