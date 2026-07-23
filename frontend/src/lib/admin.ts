export type Usuario = {
  id: string;
  email: string;
  nome: string;
  status: "pendente" | "aprovado" | "negado";
  isAdmin: boolean;
  criadoEm: string;
};

export type PresencaItem = { nome: string; email: string };

export type PorUsuario = {
  nome: string;
  email: string;
  consultas: number;
  prospeccoes: number;
  pdfIndicacao: number;
};

export type Dashboard = {
  online: number;
  onlineLista: PresencaItem[];
  contas: number;
  contasPendentes: number;
  logins: number;
  consultas: number;
  prospeccoes: number;
  pdfs: number;
  pdfsCnpjs: string[] | null;
  porUsuario: PorUsuario[];
};

export type Atividade = {
  id: string;
  tipo: string;
  uid: string;
  nome: string;
  email: string;
  detalhe: string;
  criadoEm: string;
};

export const ATIVIDADE_META: Record<string, { label: string; cor: string }> = {
  conta_criada: { label: "Conta criada", cor: "text-[#A78BFA]" },
  login: { label: "Login", cor: "text-verde-arcom" },
  consulta: { label: "Consulta", cor: "text-[#22D3EE]" },
  prospeccao: { label: "Prospecção", cor: "text-verde-lima" },
  precadastro: { label: "Pré-cadastro", cor: "text-verde-arcom" },
  pdf_indicacao: { label: "PDF de indicação", cor: "text-danger" },
};
