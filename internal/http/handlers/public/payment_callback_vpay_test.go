package public

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/models"
	"github.com/dujiao-next/internal/provider"
	"github.com/dujiao-next/internal/repository"
	"github.com/dujiao-next/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type vpayCallbackFixture struct {
	orderRepo   repository.OrderRepository
	paymentRepo repository.PaymentRepository
	handler     *Handler
	order       *models.Order
	payment     *models.Payment
}

func newVpayCallbackFixture(t *testing.T) *vpayCallbackFixture {
	t.Helper()

	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:payment_callback_vpay_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.ProductSKU{},
		&models.Order{},
		&models.OrderItem{},
		&models.Fulfillment{},
		&models.PaymentChannel{},
		&models.Payment{},
	); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	user := &models.User{
		Email:        "vpay-callback@example.com",
		PasswordHash: "hash",
		Status:       constants.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	order := &models.Order{
		OrderNo:                 "DJVPAYCALLBACK001",
		UserID:                  user.ID,
		Status:                  constants.OrderStatusPendingPayment,
		Currency:                "CNY",
		OriginalAmount:          models.NewMoneyFromDecimal(decimal.NewFromInt(99)),
		DiscountAmount:          models.NewMoneyFromDecimal(decimal.Zero),
		PromotionDiscountAmount: models.NewMoneyFromDecimal(decimal.Zero),
		TotalAmount:             models.NewMoneyFromDecimal(decimal.NewFromInt(99)),
		WalletPaidAmount:        models.NewMoneyFromDecimal(decimal.Zero),
		OnlinePaidAmount:        models.NewMoneyFromDecimal(decimal.NewFromInt(99)),
		RefundedAmount:          models.NewMoneyFromDecimal(decimal.Zero),
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	channel := &models.PaymentChannel{
		Name:            "VPay",
		ProviderType:    constants.PaymentProviderVpay,
		ChannelType:     constants.PaymentChannelTypeWechat,
		InteractionMode: constants.PaymentInteractionRedirect,
		FeeRate:         models.NewMoneyFromDecimal(decimal.Zero),
		ConfigJSON: models.JSON{
			"gateway_url": "https://vpay.example.com",
			"sign_key":    "vpay-secret",
			"notify_url":  "https://api.example.com/api/v1/payments/callback",
			"return_url":  "https://shop.example.com/pay",
		},
		IsActive:  true,
		SortOrder: 10,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(channel).Error; err != nil {
		t.Fatalf("create channel failed: %v", err)
	}
	payment := &models.Payment{
		OrderID:         order.ID,
		ChannelID:       channel.ID,
		ProviderType:    channel.ProviderType,
		ChannelType:     channel.ChannelType,
		InteractionMode: channel.InteractionMode,
		Amount:          models.NewMoneyFromDecimal(decimal.NewFromInt(99)),
		FeeRate:         models.NewMoneyFromDecimal(decimal.Zero),
		FeeAmount:       models.NewMoneyFromDecimal(decimal.Zero),
		Currency:        "CNY",
		Status:          constants.PaymentStatusPending,
		ProviderRef:     "VPAY-ORDER-1001",
		GatewayOrderNo:  "DJPVPAY1001",
		ProviderPayload: models.JSON{
			"really_price": "99.01",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(payment).Error; err != nil {
		t.Fatalf("create payment failed: %v", err)
	}

	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	channelRepo := repository.NewPaymentChannelRepository(db)
	productRepo := repository.NewProductRepository(db)
	productSKURepo := repository.NewProductSKURepository(db)
	paymentService := service.NewPaymentService(service.PaymentServiceOptions{
		OrderRepo:      orderRepo,
		ProductRepo:    productRepo,
		ProductSKURepo: productSKURepo,
		PaymentRepo:    paymentRepo,
		ChannelRepo:    channelRepo,
		ExpireMinutes:  15,
	})

	return &vpayCallbackFixture{
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		handler: &Handler{Container: &provider.Container{
			OrderRepo:          orderRepo,
			PaymentRepo:        paymentRepo,
			PaymentChannelRepo: channelRepo,
			PaymentService:     paymentService,
		}},
		order:   order,
		payment: payment,
	}
}

func TestPaymentCallbackHandlesVpay(t *testing.T) {
	fixture := newVpayCallbackFixture(t)
	query := vpaySignedQuery("DJPVPAY1001", "DJVPAYCALLBACK001", "1", "99.00", "99.01", "vpay-secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/callback?"+query, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	fixture.handler.PaymentCallback(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != constants.VpayCallbackSuccess {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}

	updatedPayment, err := fixture.paymentRepo.GetByID(fixture.payment.ID)
	if err != nil {
		t.Fatalf("reload payment failed: %v", err)
	}
	if updatedPayment == nil || updatedPayment.Status != constants.PaymentStatusSuccess {
		t.Fatalf("payment status not updated: %+v", updatedPayment)
	}
	updatedOrder, err := fixture.orderRepo.GetByID(fixture.order.ID)
	if err != nil {
		t.Fatalf("reload order failed: %v", err)
	}
	if updatedOrder == nil || updatedOrder.Status != constants.OrderStatusPaid {
		t.Fatalf("order status not updated: %+v", updatedOrder)
	}
}

func TestPaymentCallbackRejectsVpayReallyPriceMismatch(t *testing.T) {
	fixture := newVpayCallbackFixture(t)
	query := vpaySignedQuery("DJPVPAY1001", "DJVPAYCALLBACK001", "1", "99.00", "99.02", "vpay-secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/callback?"+query, nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	fixture.handler.PaymentCallback(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != constants.VpayCallbackFail {
		t.Fatalf("unexpected response body: %s", w.Body.String())
	}

	updatedPayment, err := fixture.paymentRepo.GetByID(fixture.payment.ID)
	if err != nil {
		t.Fatalf("reload payment failed: %v", err)
	}
	if updatedPayment == nil || updatedPayment.Status != constants.PaymentStatusPending {
		t.Fatalf("payment status should stay pending: %+v", updatedPayment)
	}
}

func vpaySignedQuery(payID, param, payType, price, reallyPrice, key string) string {
	values := make([]string, 0, 6)
	values = append(values,
		"payId="+payID,
		"param="+param,
		"type="+payType,
		"price="+price,
		"reallyPrice="+reallyPrice,
		"sign="+md5HexLower(payID+param+payType+price+reallyPrice+key),
	)
	return strings.Join(values, "&")
}

func md5HexLower(raw string) string {
	sum := md5.Sum([]byte(raw))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}
