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
       "encoding/base64"
	   "strings"
       "image"
       _ "image/gif"
       _ "image/jpeg"
       _ "image/png"
       "errors"
)
const MAX_MESSAGE_SIZE = 256;

func writeErrorJSON(w http.ResponseWriter, msg string, code int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func ValidateMessage(message string) error {
    if message == "" {
        return errors.New("Message is null")
    }
    if len(message) > MAX_MESSAGE_SIZE && !IsBase64(message) {
        return errors.New("Very big message")
    }
    if IsBase64(message) && !IsBase64Image(message) {
        return errors.New("Message is base64 but not image")
    }
    return nil
}

func ValidateTicketField(subject, message string) error {
    subject = strings.TrimSpace(subject)
    message = strings.TrimSpace(message)
    err := ValidateMessage(message)
    if err != nil {
        log.Println("Uncorrect message")
        return err
    }
    if subject == "" {
        return errors.New("Subject is null")
    }

    if IsBase64(subject) {
        return errors.New("Subject can't be base64")
    }
    matched, err := regexp.MatchString(`^[\p{L}\p{N}\s\-\_]+$`, subject)
    if err != nil {
        return err
    }
    if !matched {
        return errors.New("Subject incorrect")
    }

    return nil
}

func IsBase64Image(s string) bool {
    s = strings.TrimSpace(s)
    decoded, err := base64.StdEncoding.DecodeString(s)
    if err != nil {
        return false
    }
    _, _, err = image.DecodeConfig(strings.NewReader(string(decoded)))
    return err == nil
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

func IsBase64(s string) bool {
	s = strings.TrimSpace(s)
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}