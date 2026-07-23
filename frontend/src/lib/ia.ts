import type { Empresa } from "@/lib/cnpj";

export type ContatosInsight = {
  telefoneOficial: string;
  telefoneOficial2: string;
  telefoneWeb: string;
  telefone: string;
  site: string;
  instagram: string;
  linkedin: string;
  email: string;
  fontes: { title: string; url: string; content: string }[];
};

export type Insight = {
  resumo: string;
  pontosFortes: string[];
  sinaisAtencao: string[];
  abordagemSugerida: string;
  perguntasQualificacao: string[];
  nivelConfianca: string;
  buscaWebRealizada: boolean;
  contatos: ContatosInsight;
};

export type ItemRanking = {
  cnpj: string;
  razao: string;
  posicao: number;
  motivo: string;
};

export type MensagemChat = { role: "user" | "assistant"; texto: string };

export function empresaParaResumo(e: Empresa) {
  return {
    cnpj: e.cnpj,
    razao: e.razao,
    situacao: e.situacao,
    porte: e.porte,
    atividade: e.atividade,
    municipio: e.municipio,
    uf: e.uf,
    capitalSocial: e.capitalSocial,
  };
}
