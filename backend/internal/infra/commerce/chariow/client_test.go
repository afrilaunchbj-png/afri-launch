package chariow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"afrilaunch/backend/internal/application/port"
	"afrilaunch/backend/internal/domain"
)

func testServer(t *testing.T, fn http.HandlerFunc) *Provider {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(fn))
	t.Cleanup(srv.Close)
	return New(srv.URL)
}

func TestGetStoreOK(t *testing.T) {
	p := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk_live_1" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "success",
			"data":    map[string]any{"id": "str_1", "name": "Ma Boutique", "url": "https://mystore.chariow.com", "currency": "XOF"},
			"errors":  []string{},
		})
	})
	store, err := p.GetStore(context.Background(), "sk_live_1")
	if err != nil {
		t.Fatal(err)
	}
	if store.ID != "str_1" || store.Name != "Ma Boutique" || store.URL == "" {
		t.Fatalf("store = %+v", store)
	}
}

func TestGetStoreUnauthorized(t *testing.T) {
	p := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "Unauthorised", "errors": []string{"Invalid API key"}})
	})
	if _, err := p.GetStore(context.Background(), "sk_bad"); err != domain.ErrCommerceInvalidCredentials {
		t.Fatalf("err = %v", err)
	}
}

func TestListProductsAndCreateCheckout(t *testing.T) {
	p := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/products":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"data": []map[string]any{
						{"id": "prd_1", "slug": "ebook", "name": "Ebook", "type": "downloadable", "is_free": false,
							"pricing": map[string]any{"current_price": map[string]any{"value": 99.0, "currency": "USD"}}},
					},
					"pagination": map[string]any{"next_cursor": nil, "has_more": false},
				},
			})
		case "/checkout":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["product_id"] != "prd_1" || body["custom_metadata"] == nil {
				t.Errorf("body = %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"step": "payment",
					"purchase": map[string]any{"id": "sal_1", "status": "awaiting_payment",
						"amount": map[string]any{"value": 99.0, "currency": "USD"}},
					"payment": map[string]any{"checkout_url": "https://checkout.example/c"},
				},
			})
		default:
			http.Error(w, "nope", http.StatusNotFound)
		}
	})

	products, _, err := p.ListProducts(context.Background(), "sk_live_1", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].ID != "prd_1" || products[0].PriceMinor != 9900 {
		t.Fatalf("products = %+v", products)
	}

	res, err := p.CreateCheckout(context.Background(), "sk_live_1", port.CommerceCheckoutInput{
		ProductID: "prd_1", BuyerEmail: "a@b.c", BuyerFirstName: "A", BuyerLastName: "B",
		BuyerPhoneNumber: "90000000", BuyerPhoneCountry: "BJ",
		Metadata: map[string]string{"saas_user_id": "u1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.SaleID != "sal_1" || res.CheckoutURL != "https://checkout.example/c" || res.Step != "payment" {
		t.Fatalf("res = %+v", res)
	}
}

func TestListProductsArrayShape(t *testing.T) {
	// Forme réelle de l'API : data est un tableau direct.
	p := testServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"id": "prd_9", "slug": "kit", "name": "Kit", "type": "downloadable", "is_free": true,
					"pricing": map[string]any{"current_price": map[string]any{"value": 0, "currency": "XOF"}}},
			},
			"pagination": map[string]any{"next_cursor": nil, "has_more": false},
		})
	})
	products, next, err := p.ListProducts(context.Background(), "sk_live_1", "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(products) != 1 || products[0].ID != "prd_9" || !products[0].IsFree || next != "" {
		t.Fatalf("products = %+v next=%q", products, next)
	}
}
