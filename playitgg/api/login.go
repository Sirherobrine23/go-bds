package api

type TotpStatus struct {
	Status   string `json:"status"`
	EpochSec uint32 `json:"epoch_sec"`
}

type WebAuth struct {
	UpdateVersion uint32        `json:"update_version"`
	AccountID     uint64        `json:"account_id"`
	Timestamp     uint64        `json:"timestamp"`
	AccountStatus AccountStatus `json:"account_status"`
	TotpStatus    TotpStatus    `json:"totp_status"`
	AdminID       uint64        `json:"admin_id"`
}

type WebSession struct {
	SessionKey string  `json:"session_key"`
	Auth       WebAuth `json:"auth"`
}

func (w Api) Guest() (*WebSession, error) {
	d, _, err := requestAPI[*WebSession](w, "/login/guest", nil, nil)
	return d, err
}
