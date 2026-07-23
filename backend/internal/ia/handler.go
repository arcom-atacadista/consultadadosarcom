package ia

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/auth"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/cnpj"
	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

var validate = validator.New()

type Handler struct {
	insight *InsightService
	ranking *RankingService
	chat    *ChatService
}

func NewHandler(insight *InsightService, ranking *RankingService, chat *ChatService) *Handler {
	return &Handler{insight: insight, ranking: ranking, chat: chat}
}

type insightInput struct {
	Empresa cnpj.Empresa `json:"empresa" validate:"required"`
}

func (h *Handler) Insight(w http.ResponseWriter, r *http.Request) {
	var in insightInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if in.Empresa.Razao == "" {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos a razão social da empresa")
		return
	}
	insight, err := h.insight.Gerar(r.Context(), in.Empresa)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao gerar insight: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, insight)
}

type rankingInput struct {
	Empresas []EmpresaResumo `json:"empresas" validate:"required,min=1"`
}

func (h *Handler) Ranking(w http.ResponseWriter, r *http.Request) {
	var in rankingInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe ao menos 1 empresa")
		return
	}
	ranking, err := h.ranking.Gerar(r.Context(), in.Empresas)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao gerar ranking: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{"ranking": ranking})
}

type chatInput struct {
	Mensagem         string          `json:"mensagem" validate:"required"`
	Historico        []MensagemChat  `json:"historico"`
	EmpresasContexto []EmpresaResumo `json:"empresasContexto"`
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	var in chatInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if err := validate.Struct(in); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "informe a mensagem")
		return
	}
	claims, _ := auth.FromContext(r.Context())
	isAdmin := claims != nil && claims.IsAdmin

	resposta, err := h.chat.Responder(r.Context(), in.Mensagem, in.Historico, in.EmpresasContexto, isAdmin)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "falha ao conversar com a IA: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"resposta": resposta})
}
