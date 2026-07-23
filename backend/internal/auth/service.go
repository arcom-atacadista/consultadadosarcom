package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/usuarios"
)

var (
	ErrCredenciaisInvalidas = errors.New("e-mail ou senha inválidos")
	ErrEmailEmUso           = errors.New("e-mail já cadastrado")
)

type Service struct {
	repo            *usuarios.Repo
	jwtSecret       string
	superAdminEmail string
}

func NewService(repo *usuarios.Repo, jwtSecret, superAdminEmail string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret, superAdminEmail: superAdminEmail}
}

// Registrar cria a conta sempre pendente e sem admin — igual à regra que
// hoje vive nas firestore.rules (create só aceita status='pendente',
// isAdmin=false). Ver docs/migracao/07-seguranca.md.
func (s *Service) Registrar(ctx context.Context, email, senha, nome string) (*usuarios.Usuario, error) {
	email = normalizarEmail(email)
	if _, err := s.repo.BuscarPorEmail(ctx, email); err == nil {
		return nil, ErrEmailEmUso
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &usuarios.Usuario{
		ID:        uuid.NewString(),
		Email:     email,
		SenhaHash: string(hash),
		Nome:      nome,
		Status:    usuarios.StatusPendente,
		IsAdmin:   false,
	}
	if err := s.repo.Criar(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login autentica e sempre devolve um token se a senha bater — aprovação vira
// autorização (checada pelo middleware), não impede o login em si. Isso deixa
// o front levar quem está pendente pra tela de espera com um token válido.
func (s *Service) Login(ctx context.Context, email, senha string) (string, *usuarios.Usuario, error) {
	email = normalizarEmail(email)
	u, err := s.repo.BuscarPorEmail(ctx, email)
	if err != nil {
		return "", nil, ErrCredenciaisInvalidas
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(senha)); err != nil {
		return "", nil, ErrCredenciaisInvalidas
	}

	isAdmin, aprovado := s.efetivos(u)
	token, err := gerarToken(s.jwtSecret, Claims{
		UID: u.ID, Email: u.Email, Nome: u.Nome, IsAdmin: isAdmin, Aprovado: aprovado,
	})
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *Service) TrocarSenha(ctx context.Context, uid, atual, nova string) error {
	u, err := s.repo.BuscarPorID(ctx, uid)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(atual)); err != nil {
		return ErrCredenciaisInvalidas
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(nova), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.AtualizarSenha(ctx, uid, string(hash))
}

// efetivos traduz emailAdmin()/isAdmin()/isAprovado() das firestore.rules:
// o super-admin por e-mail tem admin+aprovado mesmo sem nunca ter sido
// promovido no banco (era o "funciona mesmo sem doc" do Firestore).
func (s *Service) efetivos(u *usuarios.Usuario) (isAdmin, aprovado bool) {
	isAdmin = u.IsAdmin || (s.superAdminEmail != "" && u.Email == normalizarEmail(s.superAdminEmail))
	aprovado = isAdmin || u.Status == usuarios.StatusAprovado
	return isAdmin, aprovado
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
