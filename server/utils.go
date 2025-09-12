package server

import (
	   bip39 "github.com/tyler-smith/go-bip39"
	   "net/http"
	   "encoding/json"
	   "log"
	   "golang.org/x/crypto/bcrypt"
)
func writeErrorJSON(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func GenerateMnemonic() string {

    entropy, err := bip39.NewEntropy(128)
    if err != nil {
        log.Fatal(err)
    }


    mnemonic, err := bip39.NewMnemonic(entropy)
    if err != nil {
        log.Fatal(err)
    }

    return mnemonic
}

func HashPassword(password string) string {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    return string(hashed)
}


func VerifyPassword(password, hashed string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
    return err == nil
}
