package mixpay

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	host string
}

func New() *Client {
	return &Client{
		host: "https://api.mixpay.me",
	}
}

type Error struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`

	StatusCode int    `json:"-"`
	Raw        string `json:"-"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("statusCode: %d, code: %d, message: %s", e.StatusCode, e.Code, e.Message)
}

type CreateOneTimePaymentRequest struct {
	// required
	PayeeId           string
	QuoteAmount       string
	QuoteAssetId      string
	SettlementAssetId string
	OrderId           string

	// optional
	StrictMode       bool
	PaymentAssetId   string
	Remark           string
	ExpireSeconds    int64
	TraceId          string
	SettlementMemo   string
	ReturnTo         string
	FailedReturnTo   string
	CallbackUrl      string
	ExpiredTimestamp int64
}

type CreateOneTimePaymentResponse struct {
	Code string `json:"code"`
}

func (c *Client) CreateOneTimePayment(ctx context.Context, req CreateOneTimePaymentRequest) (*CreateOneTimePaymentResponse, error) {
	values := url.Values{}
	values.Set("payeeId", req.PayeeId)
	values.Set("quoteAmount", req.QuoteAmount)
	values.Set("quoteAssetId", req.QuoteAssetId)
	values.Set("settlementAssetId", req.SettlementAssetId)
	values.Set("orderId", req.OrderId)

	if req.StrictMode {
		values.Set("strictMode", "true")
	}
	if req.PaymentAssetId != "" {
		values.Set("paymentAssetId", req.PaymentAssetId)
	}
	if req.Remark != "" {
		values.Set("remark", req.Remark)
	}
	if req.ExpireSeconds != 0 {
		values.Set("expireSeconds", strconv.FormatInt(req.ExpireSeconds, 10))
	}
	if req.TraceId != "" {
		values.Set("traceId", req.TraceId)
	}
	if req.SettlementMemo != "" {
		values.Set("settlementMemo", req.SettlementMemo)
	}
	if req.ReturnTo != "" {
		values.Set("returnTo", req.ReturnTo)
	}
	if req.FailedReturnTo != "" {
		values.Set("failedReturnTo", req.FailedReturnTo)
	}
	if req.CallbackUrl != "" {
		values.Set("callbackUrl", req.CallbackUrl)
	}
	if req.ExpiredTimestamp != 0 {
		values.Set("expiredTimestamp", strconv.FormatInt(req.ExpiredTimestamp, 10))
	}

	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.host+"/v1/one_time_payment", strings.NewReader(values.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Error
		Data CreateOneTimePaymentResponse `json:"data"`
	}

	result.StatusCode = resp.StatusCode
	data, _ := io.ReadAll(resp.Body)
	result.Raw = string(data)

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, &result.Error
	}

	if !result.Success {
		return nil, &result.Error
	}

	return &result.Data, nil
}

type GetPaymentResultRequest struct {
	TraceId string
	OrderId string
	PayeeId string
}

type GetPaymentResultResponse struct {
	Raw              string `json:"-"`
	Status           string `json:"status"`
	QuoteAmount      string `json:"quoteAmount"`
	QuoteSymbol      string `json:"quoteSymbol"`
	QuoteAssetID     string `json:"quoteAssetId"`
	PaymentAmount    string `json:"paymentAmount"`
	PaymentSymbol    string `json:"paymentSymbol"`
	PaymentAssetID   string `json:"paymentAssetId"`
	Payee            string `json:"payee"`
	PayeeID          string `json:"payeeId"`
	PayeeMixinNumber string `json:"payeeMixinNumber"`
	PayeeAvatarURL   string `json:"payeeAvatarUrl"`
	Txid             string `json:"txid"`
	// Date             string `json:"date"`
	SurplusAmount string `json:"surplusAmount"`
	SurplusStatus string `json:"surplusStatus"`
	Confirmations int    `json:"confirmations"`
	PayableAmount string `json:"payableAmount"`
	FailureCode   string `json:"failureCode"`
	FailureReason string `json:"failureReason"`
	ReturnTo      string `json:"returnTo"`
	TraceID       string `json:"traceId"`
}

func (c *Client) GetPaymentResult(ctx context.Context, req GetPaymentResultRequest) (*GetPaymentResultResponse, error) {
	value := url.Values{}
	if req.TraceId != "" {
		value.Set("traceId", req.TraceId)
	}
	if req.OrderId != "" {
		value.Set("orderId", req.OrderId)
	}
	if req.PayeeId != "" {
		value.Set("payeeId", req.PayeeId)
	}

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.host+"/v1/payments_result"+"?"+value.Encode(), nil)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Error
		Data GetPaymentResultResponse `json:"data"`
	}

	result.StatusCode = resp.StatusCode
	data, _ := io.ReadAll(resp.Body)
	result.Raw = string(data)

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, &result.Error
	}

	if !result.Success {
		return nil, &result.Error
	}

	result.Data.Raw = result.Raw
	return &result.Data, nil
}
