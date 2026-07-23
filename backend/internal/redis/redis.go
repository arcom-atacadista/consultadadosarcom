// Package redis conecta no Redis usado para cache (consulta de CNPJ) e,
// em fases futuras, filas (asynq) e presença.
package redis

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

func Connect(url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parsear REDIS_URL: %w", err)
	}
	return redis.NewClient(opt), nil
}
