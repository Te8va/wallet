package domain

type OperationType string

const (
	DEPOSIT  OperationType = "DEPOSIT"
	WITHDRAW OperationType = "WITHDRAW"
)

type Wallet struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

type WalletRequest struct {
	WalletID      string        `json:"valletId"`
	OperationType OperationType `json:"operationType"`
	Amount        int64         `json:"amount"`
}

type BalanceResponse struct {
	WalletID string `json:"wallet_id"`
	Balance  int64  `json:"balance"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
