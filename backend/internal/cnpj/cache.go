package cnpj

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheTTL = 24 * time.Hour

// Cache guarda o resultado por 24h — reconsultar o mesmo CNPJ não gasta uma
// nova chamada de API (docs/migracao/03 §3). É best-effort: uma falha do
// Redis nunca impede a consulta, só faz cair na API de novo.
type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func chaveCache(fonte, cnpjLimpo string) string {
	return "cnpj:" + fonte + ":" + cnpjLimpo
}

func (c *Cache) Buscar(ctx context.Context, fonte, cnpjLimpo string) (*Empresa, bool) {
	val, err := c.rdb.Get(ctx, chaveCache(fonte, cnpjLimpo)).Result()
	if err != nil {
		return nil, false
	}
	var e Empresa
	if err := json.Unmarshal([]byte(val), &e); err != nil {
		return nil, false
	}
	return &e, true
}

func (c *Cache) Salvar(ctx context.Context, fonte string, e Empresa) {
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	c.rdb.Set(ctx, chaveCache(fonte, e.CNPJ), b, cacheTTL)
}
