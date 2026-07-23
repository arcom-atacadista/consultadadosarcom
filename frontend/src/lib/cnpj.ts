export type Socio = {
  nome_socio: string;
  cpf: string;
  qualificacao_socio: string;
  faixa_etaria: string;
  data_entrada_sociedade: string;
};

export type Empresa = {
  cnpj: string;
  encontrado: boolean;
  erro?: string;
  situacao: string;
  dataSituacaoCadastral: string;
  motivoSituacaoCadastral: string;
  razao: string;
  nomeFantasia: string;
  porte: string;
  clienteArcom: string;
  natureza: string;
  atividade: string;
  matrizFilial: string;
  uf: string;
  municipio: string;
  dataInicio: string;
  capitalSocial: string;
  endereco: string;
  cep: string;
  telefone: string;
  email: string;
  socios: Socio[];
  simples: string;
  mei: string;
  latitude: number | null;
  longitude: number | null;
  api: string;
};

export function limparCNPJ(cnpj: string): string {
  return cnpj.replace(/\D/g, "");
}

// Mesma validação de dígito verificador do backend — só evita gastar uma
// chamada de API com CNPJ obviamente inválido (a autoridade continua sendo
// o backend, que revalida tudo).
export function cnpjValido(cnpj: string): boolean {
  const c = limparCNPJ(cnpj);
  if (c.length !== 14) return false;
  if (/^(\d)\1{13}$/.test(c)) return false;

  const digitos = c.split("").map(Number);
  const dv = (nums: number[], pesos: number[]) => {
    const soma = nums.reduce((acc, n, i) => acc + n * pesos[i], 0);
    const resto = soma % 11;
    return resto < 2 ? 0 : 11 - resto;
  };
  const d1 = dv(digitos.slice(0, 12), [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2]);
  if (d1 !== digitos[12]) return false;
  const d2 = dv(digitos.slice(0, 13), [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2]);
  return d2 === digitos[13];
}

export function parseCNPJs(texto: string): string[] {
  return texto
    .split(/[\s,;]+/)
    .map((s) => limparCNPJ(s))
    .filter((s) => s.length > 0);
}

export function formatarCNPJ(cnpj: string): string {
  const c = limparCNPJ(cnpj);
  if (c.length !== 14) return cnpj;
  return `${c.slice(0, 2)}.${c.slice(2, 5)}.${c.slice(5, 8)}/${c.slice(8, 12)}-${c.slice(12)}`;
}
