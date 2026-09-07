package authn

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(username, issuer string) (secret string, otpauthURL string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: username,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func ValidateTOTP(secret, code string) bool {
	return totp.Validate(code, secret)
}

// CurrentTOTPCode computes the 6-digit code for a vaulted TOTP seed, plus how many seconds
// remain in the current 30s window. Used by the get_totp_code MCP tool and the equivalent
// REST endpoint, so the raw seed never has to leave the vault for routine logins.
func CurrentTOTPCode(secret string) (code string, secondsRemaining int, err error) {
	now := time.Now()
	code, err = totp.GenerateCode(secret, now)
	if err != nil {
		return "", 0, err
	}
	secondsRemaining = 30 - int(now.Unix()%30)
	return code, secondsRemaining, nil
}
