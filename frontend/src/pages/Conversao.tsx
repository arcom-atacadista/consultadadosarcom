import { useCallback, useEffect, useState } from "react";
import { RefreshCcw, TrendingUp, Download, Loader2 } from "lucide-react";
import { api } from "@/lib/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

type ItemRanking = { nome: string; total: number; convertido: number; verificado: boolean };
type ListaConversao = {
  id: string;
  nome: string;
  nomeUsuario: string;
  assessor: string;
  cidade: string;
  criadoEm: string;
  total: number;
  convertidos: number | null;
};
type Relatorio = {
  totalListas: number;
  totalEmpresas: number;
  totalConvertidas: number;
  taxaConversao: number;
  algumaVerificada: boolean;
  porAssessor: ItemRanking[];
  porQuemProspectou: ItemRanking[];
  porCidade: ItemRanking[];
  listas: ListaConversao[];
};

function RankingCard({ titulo, itens, coluna }: { titulo: string; itens: ItemRanking[]; coluna: string }) {
  if (!itens.length) return null;
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{titulo}</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{coluna}</TableHead>
              <TableHead className="text-center">Prospectadas</TableHead>
              <TableHead className="text-center">Viraram cliente</TableHead>
              <TableHead className="text-center">Conversão</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {itens.map((it) => (
              <TableRow key={it.nome}>
                <TableCell>{it.nome}</TableCell>
                <TableCell className="text-center">{it.total}</TableCell>
                <TableCell className="text-center font-semibold">{it.verificado ? it.convertido : "—"}</TableCell>
                <TableCell className="text-center">
                  {it.verificado && it.total ? `${((it.convertido / it.total) * 100).toFixed(0)}%` : "—"}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

export default function Conversao() {
  const [dias, setDias] = useState("30");
  const [relatorio, setRelatorio] = useState<Relatorio | null>(null);
  const [carregando, setCarregando] = useState(false);
  const [verificando, setVerificando] = useState(false);

  const carregar = useCallback(async (periodo: string) => {
    setCarregando(true);
    try {
      const { data } = await api.get("/conversao", { params: { dias: periodo } });
      setRelatorio(data);
    } finally {
      setCarregando(false);
    }
  }, []);

  useEffect(() => {
    carregar(dias);
  }, [dias, carregar]);

  async function verificar() {
    setVerificando(true);
    try {
      const { data } = await api.post("/conversao/verificar", null, { params: { dias } });
      setRelatorio(data);
    } finally {
      setVerificando(false);
    }
  }

  async function exportarCSV() {
    const resp = await api.get("/conversao/exportar", { params: { dias }, responseType: "blob" });
    const url = URL.createObjectURL(resp.data);
    const a = document.createElement("a");
    a.href = url;
    a.download = "conversao_prospeccao.csv";
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 8000);
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between flex-wrap gap-3">
        <div>
          <h1 className="text-xl font-black text-verde-escuro flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-verde-arcom" /> Conversão da prospecção
          </h1>
          <p className="text-sm text-arcom-gray max-w-2xl">
            De cada lista enviada ao assessor, quantos CNPJs viraram Cliente Arcom — por assessor, por quem
            prospectou e por cidade.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={dias} onValueChange={setDias}>
            <SelectTrigger className="w-44">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="0">Tudo</SelectItem>
              <SelectItem value="30">Últimos 30 dias</SelectItem>
              <SelectItem value="90">Últimos 90 dias</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={verificar} disabled={verificando}>
            {verificando ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCcw className="h-4 w-4" />}
            Verificar conversão agora
          </Button>
        </div>
      </div>

      {carregando && !relatorio ? (
        <p className="text-sm text-arcom-gray">Carregando…</p>
      ) : !relatorio || relatorio.totalListas === 0 ? (
        <p className="text-sm text-arcom-gray">
          Nenhuma lista salva no período. As listas aparecem aqui quando o time salva a seleção na Prospecção (com o
          assessor).
        </p>
      ) : (
        <>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <Card>
              <CardContent className="py-4">
                <div className="text-2xl font-black text-verde-escuro">{relatorio.totalListas}</div>
                <div className="text-xs font-bold uppercase text-arcom-gray">Listas enviadas</div>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="py-4">
                <div className="text-2xl font-black text-verde-escuro">{relatorio.totalEmpresas}</div>
                <div className="text-xs font-bold uppercase text-arcom-gray">Empresas prospectadas</div>
              </CardContent>
            </Card>
            <Card variant="accent">
              <CardContent className="py-4">
                <div className="text-2xl font-black text-verde-escuro">{relatorio.totalConvertidas}</div>
                <div className="text-xs font-bold uppercase text-verde-escuro/70">Viraram cliente</div>
              </CardContent>
            </Card>
            <Card variant="accent">
              <CardContent className="py-4">
                <div className="text-2xl font-black text-verde-escuro">{relatorio.taxaConversao.toFixed(1)}%</div>
                <div className="text-xs font-bold uppercase text-verde-escuro/70">Conversão</div>
              </CardContent>
            </Card>
          </div>

          {!relatorio.algumaVerificada && (
            <p className="text-sm text-arcom-gray">
              Clique em <strong>"Verificar conversão agora"</strong> pra reconsultar e calcular quantos viraram
              cliente.
            </p>
          )}

          <RankingCard titulo="Por assessor" itens={relatorio.porAssessor} coluna="Assessor" />
          <RankingCard titulo="Por quem prospectou" itens={relatorio.porQuemProspectou} coluna="Quem prospectou" />
          <RankingCard titulo="Por cidade" itens={relatorio.porCidade} coluna="Cidade" />

          <Card>
            <CardHeader className="flex-row items-center justify-between">
              <CardTitle className="text-base">Listas</CardTitle>
              <Button variant="ghost" size="sm" onClick={exportarCSV}>
                <Download className="h-4 w-4" /> Exportar (CSV)
              </Button>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Lista</TableHead>
                    <TableHead>Prospectou</TableHead>
                    <TableHead>Assessor</TableHead>
                    <TableHead>Cidade</TableHead>
                    <TableHead>Quando</TableHead>
                    <TableHead className="text-center">Conversão</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {relatorio.listas.map((l) => (
                    <TableRow key={l.id}>
                      <TableCell className="text-xs">{l.nome}</TableCell>
                      <TableCell className="text-xs text-arcom-gray">{l.nomeUsuario}</TableCell>
                      <TableCell className="text-xs">{l.assessor}</TableCell>
                      <TableCell className="text-xs text-arcom-gray">{l.cidade}</TableCell>
                      <TableCell className="text-xs text-arcom-gray">
                        {new Date(l.criadoEm).toLocaleString("pt-BR")}
                      </TableCell>
                      <TableCell className="text-center font-semibold">
                        {l.convertidos != null ? (
                          `${l.convertidos}/${l.total}${l.total ? ` (${((l.convertidos / l.total) * 100).toFixed(0)}%)` : ""}`
                        ) : (
                          <span className="text-arcom-gray">— ({l.total})</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  );
}
