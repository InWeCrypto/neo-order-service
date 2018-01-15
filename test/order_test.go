package orderservice

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/dghubble/sling"

	"github.com/inwecrypto/neo-order-service/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateWallet(t *testing.T) {
	resp, err := http.Post("http://localhost:8000/wallet/xxxxx/test", "application/json", strings.NewReader("{}"))

	if assert.NoError(t, err) {
		assert.Equal(t, 200, resp.StatusCode)
	}

}

func TestDeleteWallet(t *testing.T) {

	req, err := http.NewRequest(http.MethodDelete, "http://localhost:8000/wallet/xxxxx/test", nil)

	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)

	if assert.NoError(t, err) {
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestCreateOrder(t *testing.T) {

	order, err := json.Marshal(&model.Order{
		Tx:    "xxxxxx",
		From:  "xxxxxxx",
		To:    "xxxxxxx",
		Asset: "xxxxxxxxxxx",
		Value: "1",
	})

	assert.NoError(t, err)

	resp, err := http.Post("http://localhost:8000/order", "application/json", bytes.NewReader(order))

	if assert.NoError(t, err) {
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestConfirmOrder(t *testing.T) {

	resp, err := http.Post("http://localhost:8000/order/xxxxxx", "application/json", bytes.NewReader([]byte{}))

	if assert.NoError(t, err) {
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestListOrder(t *testing.T) {
	request, err := sling.New().Get("http://localhost:8000/orders/AMpupnF6QweQXLfCtF4dR45FDdKbTXkLsr/0xc56f33fc6ecfcd0c225c4ab356fee59390af8560be0e930faebe74a6daff7c9b/0/10").Request()

	assert.NoError(t, err)

	var orders []*model.Order
	var errmsg interface{}

	_, err = sling.New().Do(request, &orders, &errmsg)
	assert.NoError(t, err)

	assert.NotZero(t, len(orders))

}
