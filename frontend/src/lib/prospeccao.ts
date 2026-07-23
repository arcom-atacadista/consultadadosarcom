export type Prospect = {
  cnpj: string;
  razao: string;
  nomeFantasia: string;
  atividade: string;
  cnaeCodigo: string;
  bairro: string;
  endereco: string;
  cep: string;
  telefone: string;
  email: string;
  porte: string;
  dataInicio: string;
  latitude: number | null;
  longitude: number | null;
  cidade: string;
  uf: string;
  score: number;
  temperatura: "quente" | "morno" | "frio";
  sinais: string[];
  redeGrande: boolean;
  nova: boolean;
  mesesAtivo?: number;
};

export type CidadeFiltro = { cidade: string; uf: string };

// Catálogo de ramos (CNAE -> rótulo), sem emoji — Design System ARCOM.
export const RAMO_INFO: Record<string, string> = {
  "4711302": "Supermercados",
  "4771701": "Farmácias",
  "5611201": "Restaurantes",
  "5611203": "Lanchonetes",
  "5611204": "Pizzarias",
  "4772500": "Perfumarias",
  "4520001": "Oficinas Mecânicas",
  "4520006": "Borracharias",
  "4744099": "Material de Construção",
  "4530703": "Auto Peças",
  "4721102": "Padarias",
  "4722901": "Açougues",
  "4635403": "Distribuidoras de Bebidas",
  "9313100": "Academias",
  "4751201": "Lojas de Informática",
  "4752100": "Lojas de Celular",
  "4781400": "Lojas de Roupas",
  "4782201": "Lojas de Calçados",
  "4930202": "Transportadoras",
  "4731800": "Postos de Combustível",
  "8630504": "Dentistas",
  "8630503": "Clínicas Médicas",
  "7500100": "Veterinárias",
  "4789004": "Pet Shops",
  "6821801": "Imobiliárias",
  "6920601": "Contabilidades",
  "6911701": "Advocacia",
};

// Ramos oferecidos no filtro de busca (CNAE -> rótulo).
export const RAMOS_BUSCA: { value: string; label: string }[] = [
  { value: "MISTO", label: "Misto (comércio de bairro em geral)" },
  { value: "4711302", label: "Supermercados" },
  { value: "4771701", label: "Farmácias" },
  { value: "5611201", label: "Restaurantes" },
  { value: "4721102", label: "Padarias" },
  { value: "4722901", label: "Açougues" },
  { value: "4520001", label: "Oficinas Mecânicas" },
  { value: "4731800", label: "Postos de Combustível" },
  { value: "8630503", label: "Clínicas Médicas" },
  { value: "6821801", label: "Imobiliárias" },
  { value: "4781400", label: "Lojas de Roupas" },
  { value: "4789004", label: "Pet Shops" },
];

export function ramoDoCnae(codigo: string, atividade: string): string {
  return RAMO_INFO[codigo] ?? (atividade && atividade !== "—" ? atividade : "Outros ramos");
}

export function parseCidades(texto: string): CidadeFiltro[] {
  return texto
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean)
    .map((linha) => {
      const [cidade, uf] = linha.split(",").map((s) => s.trim());
      return { cidade: cidade ?? "", uf: (uf ?? "").toUpperCase() };
    });
}

export function telMovel(tel: string): boolean {
  const d = (tel || "").replace(/\D/g, "");
  if (d.length < 11) return false;
  const num = d.slice(2);
  return num.length === 9 && num[0] === "9";
}

type Ponto = { lat: number; lng: number };

export function distanciaKm(a: Ponto, b: Ponto): number {
  const R = 6371;
  const rad = Math.PI / 180;
  const dLat = (b.lat - a.lat) * rad;
  const dLng = (b.lng - a.lng) * rad;
  const s =
    Math.sin(dLat / 2) ** 2 + Math.cos(a.lat * rad) * Math.cos(b.lat * rad) * Math.sin(dLng / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(s));
}

// Ordena por vizinho-mais-próximo a partir de uma origem (se null, começa no 1º ponto).
export function rotaProximidade<T extends { latitude: number | null; longitude: number | null }>(
  pontos: T[],
  origem: Ponto | null,
): T[] {
  const xy = (p: T): Ponto => ({ lat: p.latitude ?? 0, lng: p.longitude ?? 0 });
  const restantes = pontos.slice();
  const ordem: T[] = [];
  let atual = origem;
  if (!atual) {
    const primeiro = restantes.shift();
    if (primeiro) {
      ordem.push(primeiro);
      atual = xy(primeiro);
    }
  }
  while (restantes.length) {
    let iMin = 0;
    let dMin = Infinity;
    restantes.forEach((p, i) => {
      const d = distanciaKm(atual!, xy(p));
      if (d < dMin) {
        dMin = d;
        iMin = i;
      }
    });
    const [prox] = restantes.splice(iMin, 1);
    ordem.push(prox);
    atual = xy(prox);
  }
  return ordem;
}
