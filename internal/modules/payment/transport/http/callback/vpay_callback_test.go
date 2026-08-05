package paymentcallbackhttp

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/constants"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/dujiao-next/internal/shared/money"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type vpayCallbackServiceStub struct {
	calls int
}

func (s *vpayCallbackServiceStub) HandleSyncCallback(_ *paymentdomain.PaymentChannel, _ map[string][]string, _ []byte) (*paymentdomain.Payment, error) {
	s.calls++
	return &paymentdomain.Payment{Status: constants.PaymentStatusSuccess}, nil
}

func (s *vpayCallbackServiceStub) HandleWechatWebhook(_ WechatWebhookInput) (*paymentdomain.Payment, string, error) {
	return nil, "", nil
}

type vpayPaymentLookupStub struct {
	payment *paymentdomain.Payment
}

func (s vpayPaymentLookupStub) GetByGatewayOrderNo(string) (*paymentdomain.Payment, error) {
	return s.payment, nil
}

func (s vpayPaymentLookupStub) GetLatestByProviderRef(string) (*paymentdomain.Payment, error) {
	return nil, nil
}

type vpayChannelLookupStub struct {
	channel *paymentdomain.PaymentChannel
}

func (s vpayChannelLookupStub) GetByID(uint) (*paymentdomain.PaymentChannel, error) {
	return s.channel, nil
}

func TestHandleVpayCallbackPreservesReallyPriceVerification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &vpayCallbackServiceStub{}
	handler := &Handler{
		service: service,
		payments: vpayPaymentLookupStub{payment: &paymentdomain.Payment{
			ID:              1,
			ChannelID:       2,
			Amount:          money.FromDecimal(decimal.RequireFromString("99.00")),
			ProviderPayload: jsonmap.JSON{"really_price": "99.01"},
		}},
		channels: vpayChannelLookupStub{channel: &paymentdomain.PaymentChannel{ID: 2, ProviderType: constants.PaymentProviderVpay}},
	}

	for name, reallyPrice := range map[string]string{"match": "99.01", "mismatch": "99.02"} {
		t.Run(name, func(t *testing.T) {
			service.calls = 0
			query := signedVpayQuery("PAY-1001", "ORDER-1001", "2", "99.00", reallyPrice, "secret-key")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payments/callback?"+query, nil)

			if handled := handler.handleVpayCallback(c); !handled {
				t.Fatal("expected VPay callback to be matched")
			}
			if reallyPrice == "99.01" {
				if service.calls != 1 || strings.TrimSpace(w.Body.String()) != constants.VpayCallbackSuccess {
					t.Fatalf("expected successful callback, calls=%d body=%q", service.calls, w.Body.String())
				}
			} else if service.calls != 0 || strings.TrimSpace(w.Body.String()) != constants.VpayCallbackFail {
				t.Fatalf("expected reallyPrice mismatch rejection, calls=%d body=%q", service.calls, w.Body.String())
			}
		})
	}
}

func signedVpayQuery(payID, param, payType, price, reallyPrice, key string) string {
	sum := md5.Sum([]byte(payID + param + payType + price + reallyPrice + key))
	return "payId=" + payID + "&param=" + param + "&type=" + payType + "&price=" + price +
		"&reallyPrice=" + reallyPrice + "&sign=" + strings.ToLower(hex.EncodeToString(sum[:]))
}
