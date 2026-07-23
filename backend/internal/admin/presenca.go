package admin

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

// PresencaOnlineTTL replica a janela "online = visto nos últimos 3 min" do
// app antigo — só que aqui o próprio TTL do Redis já resolve a expiração,
// sem precisar de consulta com filtro por timestamp.
const PresencaOnlineTTL = 3 * time.Minute

type presencaValor struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	TS    int64  `json:"ts"` // unix millis, só pra ordenar por "mais recente primeiro"
}

type PresencaItem struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
}

type PresencaService struct {
	rdb *redis.Client
}

func NewPresencaService(rdb *redis.Client) *PresencaService {
	return &PresencaService{rdb: rdb}
}

func chavePresenca(uid string) string { return "presenca:" + uid }

// Heartbeat grava a presença do usuário com TTL — chamado a cada 60s pelo
// frontend enquanto a aba está visível (mesmo intervalo do app antigo).
func (s *PresencaService) Heartbeat(ctx context.Context, uid, nome, email string) error {
	v, _ := json.Marshal(presencaValor{Nome: nome, Email: email, TS: time.Now().UnixMilli()})
	return s.rdb.Set(ctx, chavePresenca(uid), v, PresencaOnlineTTL).Err()
}

// Remover apaga a presença no logout (mesmo comportamento do pararPresenca
// do app antigo, que deletava o doc na saída).
func (s *PresencaService) Remover(ctx context.Context, uid string) error {
	return s.rdb.Del(ctx, chavePresenca(uid)).Err()
}

// Online lista quem está com presença viva agora, mais recente primeiro.
func (s *PresencaService) Online(ctx context.Context) ([]PresencaItem, error) {
	var chaves []string
	iter := s.rdb.Scan(ctx, 0, "presenca:*", 200).Iterator()
	for iter.Next(ctx) {
		chaves = append(chaves, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	if len(chaves) == 0 {
		return []PresencaItem{}, nil
	}
	valores, err := s.rdb.MGet(ctx, chaves...).Result()
	if err != nil {
		return nil, err
	}
	type comTS struct {
		item PresencaItem
		ts   int64
	}
	var vistos []comTS
	for _, raw := range valores {
		s, ok := raw.(string)
		if !ok {
			continue
		}
		var pv presencaValor
		if json.Unmarshal([]byte(s), &pv) != nil {
			continue
		}
		vistos = append(vistos, comTS{item: PresencaItem{Nome: pv.Nome, Email: pv.Email}, ts: pv.TS})
	}
	sort.Slice(vistos, func(i, j int) bool { return vistos[i].ts > vistos[j].ts })
	itens := make([]PresencaItem, len(vistos))
	for i, v := range vistos {
		itens[i] = v.item
	}
	return itens, nil
}
