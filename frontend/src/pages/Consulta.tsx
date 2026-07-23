import { useMemo, useRef, useState } from "react";
import { Search, Star, History, Loader2, Sparkles, Trophy } from "lucide-react";
import { api } from "@/lib/api";
import { cnpjValido, formatarCNPJ, parseCNPJs, type Empresa } from "@/lib/cnpj";
import { useCnpjHistorico } from "@/hooks/useCnpjHistorico";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert } from "@/components/ui/alert";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { InsightDialog } from "@/components/ia/InsightDialog";
import { RankingDialog } from "@/components/ia/RankingDialog";

function badgeSituacao(situacao: string) {
  if (situacao === "ATIVA") return <Badge variant="brand">{situacao}</Badge>;
  if (!situacao) return <Badge variant="default">—</Badge>;
  return <Badge variant="danger">{situacao}</Badge>;
}

function badgeClienteArcom(valor: string) {
  if (valor === "Sim") return <Badge variant="accent">Sim</Badge>;
  if (valor === "Não") return <Badge variant="outline">Não</Badge>;
  return <Badge variant="default">—</Badge>;
}

export default function Consulta() {
  const [texto, setTexto] = useState("");
  const [fonte, setFonte] = useState<"arcom" | "brasilapi">("arcom");
  const [carregando, setCarregando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [resultados, setResultados] = useState<Empresa[]>([]);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const { favoritos, historico, isFavorito, alternarFavorito, registrarHistorico, limparHistorico } =
    useCnpjHistorico();

  const [insightEmpresa, setInsightEmpresa] = useState<Empresa | null>(null);
  const [rankingEmpresas, setRankingEmpresas] = useState<Empresa[] | null>(null);

  const cnpjsDigitados = useMemo(() => parseCNPJs(texto), [texto]);
  const invalidos = useMemo(() => cnpjsDigitados.filter((c) => !cnpjValido(c)), [cnpjsDigitados]);

  async function consultar() {
    const validos = cnpjsDigitados.filter(cnpjValido);
    if (validos.length === 0) {
      setErro("Cole ao menos um CNPJ válido.");
      return;
    }
    setErro(null);
    setCarregando(true);
    try {
      const { data } = await api.post("/cnpj/consultar", { cnpjs: validos, fonte });
      const lista: Empresa[] = data.resultados;
      setResultados(lista);
      registrarHistorico(
        lista.filter((e) => e.encontrado).map((e) => ({ cnpj: e.cnpj, razao: e.razao })),
      );
    } catch (e: any) {
      setErro(e?.response?.data?.erro ?? "Falha ao consultar CNPJ.");
    } finally {
      setCarregando(false);
    }
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      e.preventDefault();
      consultar();
    }
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[360px_1fr] gap-6">
      <div className="flex flex-col gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Consultar CNPJ</CardTitle>
            <CardDescription>
              Cole um ou vários CNPJs (um por linha). <kbd className="text-xs">Ctrl/Cmd+Enter</kbd> consulta.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">
            <Textarea
              ref={textareaRef}
              value={texto}
              onChange={(e) => setTexto(e.target.value)}
              onKeyDown={onKeyDown}
              placeholder={"11.222.333/0001-81\n19.131.243/0001-97"}
              rows={6}
            />
            {invalidos.length > 0 && (
              <p className="text-xs text-danger">
                {invalidos.length} CNPJ(s) com dígito inválido serão ignorados.
              </p>
            )}
            <Select value={fonte} onValueChange={(v) => setFonte(v as "arcom" | "brasilapi")}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="arcom">Consulta CNPJ Arcom</SelectItem>
                <SelectItem value="brasilapi">Brasil API</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={consultar} disabled={carregando}>
              {carregando ? <Loader2 className="h-4 w-4 animate-spin" /> : <Search className="h-4 w-4" />}
              {carregando ? "Consultando…" : "Consultar"}
            </Button>
            {erro && <Alert variant="danger">{erro}</Alert>}
          </CardContent>
        </Card>

        {favoritos.length > 0 && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-base">
                <Star className="h-4 w-4 text-verde-lima" /> Favoritos
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {favoritos.map((f) => (
                <button
                  key={f.cnpj}
                  onClick={() => setTexto((t) => (t ? t + "\n" : "") + f.cnpj)}
                  className="text-left text-sm text-verde-escuro hover:text-verde-arcom"
                >
                  {formatarCNPJ(f.cnpj)} — {f.razao}
                </button>
              ))}
            </CardContent>
          </Card>
        )}

        {historico.length > 0 && (
          <Card>
            <CardHeader className="flex-row items-center justify-between">
              <CardTitle className="flex items-center gap-2 text-base">
                <History className="h-4 w-4 text-arcom-gray" /> Histórico
              </CardTitle>
              <Button variant="ghost" size="sm" onClick={limparHistorico}>
                Limpar
              </Button>
            </CardHeader>
            <CardContent className="flex flex-col gap-2">
              {historico.slice(0, 10).map((h) => (
                <button
                  key={h.cnpj + h.ts}
                  onClick={() => setTexto((t) => (t ? t + "\n" : "") + h.cnpj)}
                  className="text-left text-sm text-arcom-gray hover:text-verde-arcom"
                >
                  {formatarCNPJ(h.cnpj)} — {h.razao}
                </button>
              ))}
            </CardContent>
          </Card>
        )}
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle>Resultados</CardTitle>
            <CardDescription>{resultados.length} CNPJ(s) consultado(s).</CardDescription>
          </div>
          {resultados.some((e) => e.encontrado) && (
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setRankingEmpresas(resultados.filter((e) => e.encontrado))}
            >
              <Trophy className="h-4 w-4" /> Ranking de leads (IA)
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {resultados.length === 0 ? (
            <p className="text-sm text-arcom-gray">Nenhuma consulta ainda.</p>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>CNPJ</TableHead>
                  <TableHead>Razão social</TableHead>
                  <TableHead>Situação</TableHead>
                  <TableHead>Cliente Arcom</TableHead>
                  <TableHead>UF/Município</TableHead>
                  <TableHead>Telefone</TableHead>
                  <TableHead />
                </TableRow>
              </TableHeader>
              <TableBody>
                {resultados.map((e) => (
                  <TableRow key={e.cnpj}>
                    <TableCell className="font-mono text-xs">{formatarCNPJ(e.cnpj)}</TableCell>
                    {!e.encontrado ? (
                      <TableCell colSpan={5} className="text-danger">
                        {e.erro}
                      </TableCell>
                    ) : (
                      <>
                        <TableCell>{e.razao}</TableCell>
                        <TableCell>{badgeSituacao(e.situacao)}</TableCell>
                        <TableCell>{badgeClienteArcom(e.clienteArcom)}</TableCell>
                        <TableCell>{e.uf ? `${e.municipio}/${e.uf}` : "—"}</TableCell>
                        <TableCell>{e.telefone}</TableCell>
                      </>
                    )}
                    <TableCell>
                      {e.encontrado && (
                        <div className="flex items-center gap-2">
                          <button onClick={() => alternarFavorito(e.cnpj, e.razao)} aria-label="Favoritar">
                            <Star
                              className={
                                isFavorito(e.cnpj)
                                  ? "h-4 w-4 fill-verde-lima text-verde-lima"
                                  : "h-4 w-4 text-arcom-gray"
                              }
                            />
                          </button>
                          <button onClick={() => setInsightEmpresa(e)} title="Gerar Insight IA">
                            <Sparkles className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                          </button>
                        </div>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <InsightDialog empresa={insightEmpresa} onClose={() => setInsightEmpresa(null)} />
      <RankingDialog empresas={rankingEmpresas} onClose={() => setRankingEmpresas(null)} />
    </div>
  );
}
