package internal

import (
	"errors"
	"net/http"
	"strconv"
	"time"
	"wallet/lib/rest/internal/payloads"
	"wallet/lib/utils"
	"wallet/lib/utils/logger"
	"wallet/lib/withdraws"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *server) GetBalanceHandler(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, payloads.CreateRequiredParamResponse("user_id"))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidParamResponse("user_id"))
		return
	}
	wallet, err := s.coreRepoFactory.New(nil).Wallet().GetOrCreate(ctx, userID)
	if err != nil {
		traceID, _ := ctx.Get("request_id")
		logger.Get().With("trace_id", traceID)
		ctx.JSON(http.StatusInternalServerError, payloads.CreateCallSupportResponse(traceID.(string)))
		return
	}
	ctx.JSON(http.StatusOK, payloads.Response{
		Data: wallet,
	})
}

func (s *server) getTransactionsHistoryHandler(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id") // TODO :  unify duplicate code for get user id
	if userIDStr == "" {
		ctx.JSON(http.StatusBadRequest, payloads.CreateRequiredParamResponse("user_id"))
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidParamResponse("user_id"))
		return
	}

	page, pageSize, err := getPageAndPageSize(ctx)
	if errors.Is(err, ErrInvalidPage) {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidParamResponse("page"))
		return
	}
	if errors.Is(err, ErrInvalidPageSize) {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidParamResponse("page_size"))
		return
	}
	transactions, hasMore, err := s.coreRepoFactory.New(nil).Transaction().Get(ctx, userID, page, pageSize)
	if err != nil { // TODO (*DUPLICATE CODE*) unify code for managing unexpected errors
		traceID, _ := ctx.Get("request_id")
		logger.Get().With("trace_id", traceID)
		ctx.JSON(http.StatusInternalServerError, payloads.CreateCallSupportResponse(traceID.(string)))
		return
	}
	ctx.JSON(http.StatusOK, payloads.Response{
		Data: payloads.TransactionHistoryResponse{
			HasMore:      hasMore,
			Transactions: transactions,
		},
	})

}

func (s *server) createWithdrawHandler(ctx *gin.Context) {
	var request payloads.CreateWithdrawRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidPayloadResponse(err))
		return
	}
	withdraw := payloads.Withdraw{
		WalletID: request.UserID,
		Bank:     request.BankType,
		Iban:     request.IBan,
		Amount:   request.Amount,
	}
	err := s.withdrawService.Create(ctx, &withdraw)
	if errors.Is(err, withdraws.ErrInsufficientBalance) {
		ctx.JSON(http.StatusBadRequest, payloads.Response{
			Error: &payloads.ErrorResponse{
				Code:    "insufficient_balance", // TODO move all errors code to single file as constants
				Message: "there is not enough available balance to create withdraw",
			},
		})
	}
	if err != nil {
		traceID, _ := ctx.Get("request_id")
		logger.Get().With("trace_id", traceID).Error("cant create withdrawal", "error", utils.Stringify(err))
		ctx.JSON(http.StatusInternalServerError, payloads.CreateCallSupportResponse(traceID.(string)))
		return
	}
	ctx.JSON(http.StatusCreated, payloads.Response{
		Data: withdraw,
	})
}

func (s *server) createDepositHandler(ctx *gin.Context) {
	var request payloads.CreateDepositRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, payloads.CreateInvalidPayloadResponse(err))
		return
	}
	if request.ApplyAt == nil {
		now := time.Now()
		request.ApplyAt = &now
	}
	deposit := payloads.Deposit{
		UserID:  request.UserID,
		Amount:  request.Amount,
		ApplyAt: *request.ApplyAt,
	}

	if err := s.depositService.Create(ctx, &deposit); err != nil {
		traceID, _ := ctx.Get("request_id")
		logger.Get().With("trace_id", traceID).Error("cant create withdrawal", "error", utils.Stringify(err))
		ctx.JSON(http.StatusInternalServerError, payloads.CreateCallSupportResponse(traceID.(string)))
		return
	}
	ctx.JSON(http.StatusCreated, payloads.Response{
		Data: deposit,
	})
}

var ErrInvalidPage = errors.New("invalid page")
var ErrInvalidPageSize = errors.New("invalid page")

func getPageAndPageSize(ctx *gin.Context) (int, int, error) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("page_size")
	if pageStr == "" {
		pageStr = "1"
	}
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidPage
	}
	pageSize, err := strconv.ParseInt(pageSizeStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidPageSize
	}
	return int(page), int(pageSize), nil
}
