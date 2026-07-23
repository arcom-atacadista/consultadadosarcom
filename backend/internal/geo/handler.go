package geo

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

const cacheTTL = 30 * 24 * time.Hour // endereço não muda de coordenada, cache longo

type Handler struct {
	client *Client
	rdb    *redis.Client
}

func NewHandler(client *Client, rdb *redis.Client) *Handler {
	return &Handler{client: client, rdb: rdb}
}

func (h *Handler) Geocode(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		httputil.WriteError(w, http.StatusBadRequest, "informe o endereço em ?q=")
		return
	}

	ctx := r.Context()
	chave := "geocode:" + q
	if cached, ok := h.buscarCache(ctx, chave); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	res, err := h.client.Geocode(ctx, q)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao geocodificar endereço")
		return
	}
	if res == nil {
		httputil.WriteJSON(w, http.StatusOK, nil)
		return
	}
	h.salvarCache(ctx, chave, res)
	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) buscarCache(ctx context.Context, chave string) (*Resultado, bool) {
	val, err := h.rdb.Get(ctx, chave).Result()
	if err != nil {
		return nil, false
	}
	var res Resultado
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return nil, false
	}
	return &res, true
}

func (h *Handler) salvarCache(ctx context.Context, chave string, res *Resultado) {
	b, err := json.Marshal(res)
	if err != nil {
		return
	}
	h.rdb.Set(ctx, chave, b, cacheTTL)
}
