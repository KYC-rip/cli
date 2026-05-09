package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(WithBase(srv.URL), WithTimeout(2*time.Second)), srv
}

func TestCurrenciesParse(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/exchange/currencies" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]Currency{
			{ID: "btc", Ticker: "btc", Network: "Mainnet", Name: "Bitcoin", Engine: "trocador", Engines: []string{"trocador", "lizex"}},
			{ID: "usdt-trc20", Ticker: "usdt", Network: "TRC20", Name: "Tether", Memo: false},
		})
	}))
	got, err := c.Currencies(context.Background())
	if err != nil {
		t.Fatalf("Currencies err: %v", err)
	}
	if len(got) != 2 || got[0].Ticker != "btc" || got[1].Network != "TRC20" {
		t.Fatalf("unexpected payload: %+v", got)
	}
}

func TestEstimateParse(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("from") != "btc" || r.URL.Query().Get("to") != "usdt" {
			t.Fatalf("missing query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id":"q-1","rate":80000,"amount_from":0.01,"amount_to":800,
			"min":0.0003,"max":29,"provider":"Houdini_ChangeNow","engine":"houdini",
			"kyc_rating":"B","eta":30,
			"routes":[{"provider":"Houdini_ChangeNow","engine":"houdini","amount_to":800,"amount_from":0.01,"kyc":"B","log_policy":"B","eta":30,"fixed":false}]
		}`))
	}))
	q, err := c.Estimate(context.Background(), "btc", "", "usdt", "TRC20", 0.01)
	if err != nil {
		t.Fatalf("Estimate err: %v", err)
	}
	if q.AmountTo != 800 || len(q.Routes) != 1 || q.Routes[0].Engine != "houdini" {
		t.Fatalf("unexpected estimate: %+v", q)
	}
}

func TestCreateInjectsSource(t *testing.T) {
	var got CreateReq
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		_, _ = w.Write([]byte(`{"id":"abc","trade_id":"abc","status":"WAITING","engine":"houdini","provider":"Houdini_ChangeNow","fromTicker":"BTC","fromNetwork":"Mainnet","fromAmount":0.01,"toTicker":"USDT","toNetwork":"TRC20","toAmount":800,"depositAddress":"bc1qxyz","address_user":"TR..."}`))
	}))
	tr, err := c.Create(context.Background(), CreateReq{Provider: "p", Engine: "e", FromCurrency: "btc", ToCurrency: "usdt", AmountFrom: 0.01, AddressTo: "TR..."})
	if err != nil {
		t.Fatalf("Create err: %v", err)
	}
	if got.Source != "cli" {
		t.Fatalf("expected source=cli (auto-injected), got %q", got.Source)
	}
	if tr.ID != "abc" || tr.DepositAddress == "" {
		t.Fatalf("unexpected trade: %+v", tr)
	}
}

func TestCreatePropagatesError(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	_, err := c.Create(context.Background(), CreateReq{Provider: "p", FromCurrency: "btc", ToCurrency: "usdt", AmountFrom: 0.01, AddressTo: "x"})
	if err == nil {
		t.Fatal("expected error on 429")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected status in error, got: %v", err)
	}
}

func TestStatusParse(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v2/exchange/status/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"abc","status":"FINISHED","engine":"e","provider":"p","fromTicker":"BTC","fromNetwork":"Mainnet","fromAmount":0.01,"toTicker":"USDT","toNetwork":"TRC20","toAmount":800,"depositAddress":"d","address_user":"u","txIn":"hash1","txOut":"hash2"}`))
	}))
	tr, err := c.Status(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Status err: %v", err)
	}
	if tr.Status != "FINISHED" || tr.TxOut != "hash2" {
		t.Fatalf("unexpected: %+v", tr)
	}
}

func TestContextCancel(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.Currencies(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
