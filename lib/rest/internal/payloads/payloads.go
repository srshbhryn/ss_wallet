package payloads

import (
	"fmt"
	"time"
	"wallet/lib/core"
	"wallet/lib/deposits"
	"wallet/lib/withdraws"
	withdraws_enums "wallet/lib/withdraws/enums"

	"github.com/google/uuid"
)

type Wallet = core.Wallet
type Transaction = core.Transaction
type Withdraw = withdraws.Withdrawal
type Deposit = deposits.Deposit

type CreateWithdrawRequest struct {
	UserID   uuid.UUID                `json:"user_id"`
	IBan     string                   `json:"iban"`
	Amount   int64                    `json:"amount"`
	BankType withdraws_enums.BankType `json:"bank_type"`
}

type CreateDepositRequest struct {
	UserID  uuid.UUID  `json:"user_id"`
	Amount  int64      `json:"amount"`
	ApplyAt *time.Time `json:"apply_at,omitempty"`
}

type TransactionHistoryResponse struct {
	HasMore      bool          `json:"has_more"`
	Transactions []Transaction `json:"transactions"`
}

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Response struct {
	Data  any            `json:"data,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

func CreateRequiredParamResponse(param string) Response {
	return Response{
		Error: &ErrorResponse{
			Code:    fmt.Sprintf("no_%s_provided", param),
			Message: fmt.Sprintf("parameter %s is not provided", param),
		},
	}
}

func CreateInvalidParamResponse(param string) Response {
	return Response{
		Error: &ErrorResponse{
			Code:    fmt.Sprintf("invalid_%s", param),
			Message: fmt.Sprintf("parameter %s is invalid", param),
		},
	}
}

func CreateCallSupportResponse(traceID string) Response {
	return Response{
		Error: &ErrorResponse{
			Code:    "unexpected_error",
			Message: fmt.Sprintf("call support, trace id: '%s'", traceID),
		},
	}
}

func CreateInvalidPayloadResponse(err error) Response {
	return Response{
		Error: &ErrorResponse{
			Code:    "invalid_response",
			Message: "Invalid request payload: " + err.Error(),
		},
	}
}
