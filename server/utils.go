package server

import (
	   bip39 "github.com/tyler-smith/go-bip39"
	   "net/http"
	   "encoding/json"
	   "html"
	   "log"
	   "golang.org/x/crypto/bcrypt"
	   "regexp"
	   "mFrelance/models"
	       
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

func IsValidBTCAddress(address string) bool {
    re := regexp.MustCompile(`\b((bc|tb)(0([ac-hj-np-z02-9]{39}|[ac-hj-np-z02-9]{59})|1[ac-hj-np-z02-9]{8,87})|([13]|[mn2])[a-km-zA-HJ-NP-Z1-9]{25,39})\b`)
    return re.MatchString(address)
}


func SanitizeString(s string) string {
    return html.EscapeString(s) 
}

func SanitizeProfile(p *models.Profile) {
    p.FullName = SanitizeString(p.FullName)
    p.Bio = SanitizeString(p.Bio)
    for i, skill := range p.Skills {
        p.Skills[i] = SanitizeString(skill)
    }
    p.Avatar = SanitizeString(p.Avatar)
}
