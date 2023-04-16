package order

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fox-one/pkg/httputil/param"
	"github.com/pandodao/botastic/config"
	"github.com/pandodao/botastic/core"
	"github.com/pandodao/botastic/handler/render"
	"github.com/pandodao/botastic/session"
	"github.com/pandodao/lemon-checkout-go"
	"github.com/shopspring/decimal"
)

type CreateOrderRequest struct {
	Channel     string          `json:"channel"`
	Amount      decimal.Decimal `json:"amount"`
	VariantID   int64           `json:"variant_id"`
	RedirectURL string          `json:"redirect_url"`
}

type (
	LemonWebhookPayload struct {
		lemon.WebhookPayload
		Meta struct {
			TestMode   bool   `json:"test_mode"`
			EventName  string `json:"event_name"`
			CustomData struct {
				UserID  string `json:"user_id"`
				OrderID string `json:"order_id"`
			} `json:"custom_data"`
		} `json:"meta"`
	}
)

func CreateOrder(orderz core.OrderService, lemon config.Lemonsqueezy) http.HandlerFunc {
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

		ret := make(map[string]any)
		if body.Channel == "lemon" {
			var chosenVariant config.LemonsqueezyVariant
			for _, v := range lemon.Variants {
				if body.VariantID == v.ID {
					chosenVariant = v
					break
				}
			}

			if chosenVariant.ID == 0 {
				render.Error(w, http.StatusBadRequest, nil)
				return
			}

			paymentURL, err := orderz.CreateLemonOrder(ctx, user.ID, lemon.StoreID, chosenVariant.ID, decimal.NewFromFloat(chosenVariant.Amount), body.RedirectURL)
			if err != nil {
				render.Error(w, http.StatusInternalServerError, err)
				return
			}
			ret["payment_url"] = paymentURL

		} else if body.Channel == "mixpay" {
			code, err := orderz.CreateMixpayOrder(ctx, user.ID, body.Amount)
			if err != nil {
				render.Error(w, http.StatusInternalServerError, err)
				return
			}
			ret["code"] = code

		}

		render.JSON(w, ret)
	}
}

func GetVariants(lemon config.Lemonsqueezy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, lemon.Variants)
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

func HandleLemonCallback(orderz core.OrderService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body := &LemonWebhookPayload{}
		if err := param.Binding(r, body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		// @TODO verify signature here. FYI: https://docs.lemonsqueezy.com/api/webhooks#webhook-requests

		if body.Meta.EventName != "order_created" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		if body.Data.Attributes.Status != "paid" && body.Data.Attributes.Status != "cancelled" {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		orderID := body.Meta.CustomData.OrderID
		userID, _ := strconv.ParseUint(body.Meta.CustomData.UserID, 10, 64)

		if orderID == "" || userID == 0 {
			render.Error(w, http.StatusBadRequest, nil)
			return
		}

		lemonOrigAmount := decimal.NewFromInt(int64(body.Data.Attributes.TotalUSD + body.Data.Attributes.DiscountTotalUSD)).
			Div(decimal.NewFromInt(100))

		if err := orderz.HandleLemonCallback(ctx, orderID, userID, lemonOrigAmount, body.Data.Attributes.Status); err != nil {
			render.Error(w, http.StatusInternalServerError, err)
		}

		render.JSON(w, map[string]any{"code": "SUCCESS"})
	}
}
