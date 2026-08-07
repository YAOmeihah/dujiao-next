package orderwiring

import (
	"testing"

	ordertransport "github.com/dujiao-next/internal/modules/order/transport/http"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

func TestMapGuestOrderInputPreservesPhoneAndShippingAddress(t *testing.T) {
	shippingAddress := jsonmap.JSON{
		"receiver_phone": "18603751811",
		"detail_address": "测试地址",
	}

	got := mapGuestOrderInput(ordertransport.CreateGuestOrderInput{
		Phone:           "18603751811",
		OrderPassword:   "guest-password",
		ShippingAddress: shippingAddress,
	})

	if got.Phone != "18603751811" {
		t.Fatalf("guest phone was not preserved: got %q", got.Phone)
	}
	if got.ShippingAddress == nil {
		t.Fatal("shipping address was not preserved")
	}
	if got.ShippingAddress["detail_address"] != "测试地址" {
		t.Fatalf("shipping address was not preserved: got %#v", got.ShippingAddress)
	}
}

func TestMapOrderInputPreservesShippingAddress(t *testing.T) {
	shippingAddress := jsonmap.JSON{
		"receiver_phone": "18603751811",
		"detail_address": "测试地址",
	}

	got := mapOrderInput(ordertransport.CreateOrderInput{
		UserID:          42,
		ShippingAddress: shippingAddress,
	})

	if got.UserID != 42 {
		t.Fatalf("user ID was not preserved: got %d", got.UserID)
	}
	if got.ShippingAddress == nil {
		t.Fatal("shipping address was not preserved")
	}
	if got.ShippingAddress["detail_address"] != "测试地址" {
		t.Fatalf("shipping address was not preserved: got %#v", got.ShippingAddress)
	}
}
