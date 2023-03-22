package order

import (
	"errors"
	"net/http"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"github.com/shopspring/decimal"
)

type CreateOrderRequest struct {
	Amount decimal.Decimal `json:"amount"`
}

func CreateOrder(orderz core.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user, _ := session.UserFrom(ctx)

		body := &CreateOrderRequest{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if body.Amount.LessThanOrEqual(decimal.Zero) {
			render.Error(w, http.StatusBadRequest, errors.New("invalid amount"))
			return
		}

		code, err := orderz.CreateMixpayOrder(ctx, user.ID, body.Amount)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, map[string]any{"code": code})
	}
}

func HandleMixpayCallback(orderz core.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var callbackData struct {
			OrderID string `json:"orderId"`
			TraceID string `json:"traceId"`
			PayeeID string `json:"payeeId"`
		}

		if err := param.Binding(r, &callbackData); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		if err := orderz.HandleMixpayCallback(r.Context(), callbackData.OrderID, callbackData.TraceID, callbackData.PayeeID); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
			return
		}

		render.JSON(w, map[string]any{"code": "SUCCESS"})
	}
}
