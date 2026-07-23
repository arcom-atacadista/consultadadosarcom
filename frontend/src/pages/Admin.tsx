import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ShieldCheck,
  RefreshCw,
  Users,
  Clock,
  LogIn,
  Search,
  Target,
  FileText,
  Check,
  X,
  Shield,
  ShieldOff,
  Trash2,
} from "lucide-react";
import { api } from "@/lib/api";
import type { Dashboard, Usuario, Atividade } from "@/lib/admin";
import { ATIVIDADE_META } from "@/lib/admin";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const DASH_THROTTLE_MS = 5 * 60 * 1000; // mesmo teto do app antigo: 5 min entre refreshes automáticos

function KPI({ icon: Icon, valor, label }: { icon: any; valor: string | number; label: string }) {
  return (
    <Card>
      <CardContent className="flex flex-col gap-1 py-4">
        <div className="flex items-center gap-2 text-arcom-gray">
          <Icon className="h-4 w-4" />
          <span className="text-xs font-bold uppercase tracking-wide">{label}</span>
        </div>
        <div className="text-2xl font-black text-verde-escuro">{valor}</div>
      </CardContent>
    </Card>
  );
}

function badgeStatus(status: string) {
  if (status === "aprovado") return <Badge variant="accent">Aprovado</Badge>;
  if (status === "negado") return <Badge variant="danger">Negado</Badge>;
  return <Badge variant="outline">Pendente</Badge>;
}

export default function Admin() {
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [usuarios, setUsuarios] = useState<Usuario[]>([]);
  const [atividades, setAtividades] = useState<Atividade[]>([]);
  const [filtroAtividade, setFiltroAtividade] = useState<string>("");
  const [carregando, setCarregando] = useState(false);
  const [ultimaCarga, setUltimaCarga] = useState(0);

  const carregarDashboard = useCallback(async (forcar = false) => {
    if (!forcar && ultimaCarga && Date.now() - ultimaCarga < DASH_THROTTLE_MS) return;
    setCarregando(true);
    try {
      const [{ data: dash }, { data: users }] = await Promise.all([
        api.get("/admin/dashboard"),
        api.get("/usuarios"),
      ]);
      setDashboard(dash);
      setUsuarios(users ?? []);
      setUltimaCarga(Date.now());
    } finally {
      setCarregando(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const carregarAtividades = useCallback(async (tipo: string) => {
    const { data } = await api.get("/admin/atividades", { params: tipo ? { tipo } : {} });
    setAtividades(data ?? []);
  }, []);

  useEffect(() => {
    carregarDashboard(true);
  }, [carregarDashboard]);

  useEffect(() => {
    carregarAtividades(filtroAtividade);
  }, [filtroAtividade, carregarAtividades]);

  async function definirStatus(id: string, status: string) {
    await api.patch(`/usuarios/${id}`, { status });
    await carregarDashboard(true);
  }

  async function definirAdmin(id: string, isAdmin: boolean) {
    if (!isAdmin && !window.confirm("Remover o acesso de admin deste usuário?")) return;
    await api.patch(`/usuarios/${id}`, { isAdmin });
    await carregarDashboard(true);
  }

  async function limparAntigos() {
    if (!window.confirm("Apagar só os registros de LOGIN com mais de 90 dias? Nenhum dado de CNPJ é tocado.")) return;
    const { data } = await api.delete("/admin/atividades/antigos", { params: { dias: 90 } });
    window.alert(`${data.apagados} registro(s) de login antigos apagados.`);
    await carregarAtividades(filtroAtividade);
  }

  const porUsuarioOrdenado = useMemo(() => dashboard?.porUsuario ?? [], [dashboard]);

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-black text-verde-escuro flex items-center gap-2">
            <ShieldCheck className="h-5 w-5 text-verde-arcom" /> Administração
          </h1>
          <p className="text-sm text-arcom-gray">Contas, atividades e presença online.</p>
        </div>
        <Button variant="secondary" size="sm" onClick={() => carregarDashboard(true)} disabled={carregando}>
          <RefreshCw className={carregando ? "h-4 w-4 animate-spin" : "h-4 w-4"} /> Atualizar
        </Button>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-7 gap-3">
        <KPI icon={Users} valor={dashboard?.online ?? "—"} label="Online agora" />
        <KPI icon={Users} valor={dashboard?.contas ?? "—"} label="Contas criadas" />
        <KPI icon={Clock} valor={dashboard?.contasPendentes ?? "—"} label="Contas pendentes" />
        <KPI icon={LogIn} valor={dashboard?.logins ?? "—"} label="Logins" />
        <KPI icon={Search} valor={dashboard?.consultas ?? "—"} label="Consultas" />
        <KPI icon={Target} valor={dashboard?.prospeccoes ?? "—"} label="Prospecções" />
        <KPI icon={FileText} valor={dashboard?.pdfs ?? "—"} label="PDFs gerados" />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Online agora</CardTitle>
          </CardHeader>
          <CardContent>
            {!dashboard?.onlineLista.length ? (
              <p className="text-sm text-arcom-gray">Ninguém online agora.</p>
            ) : (
              <div className="flex flex-col gap-2">
                {dashboard.onlineLista.map((u) => (
                  <div key={u.email} className="flex items-center gap-2 text-sm">
                    <span className="h-2 w-2 rounded-full bg-verde-arcom shadow-[0_0_6px_theme(colors.verde-arcom)]" />
                    <strong className="text-verde-escuro">{u.nome || "—"}</strong>
                    <span className="text-arcom-gray">{u.email}</span>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Atividade por usuário</CardTitle>
          </CardHeader>
          <CardContent>
            {porUsuarioOrdenado.length === 0 ? (
              <p className="text-sm text-arcom-gray">Sem atividade recente.</p>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Usuário</TableHead>
                    <TableHead className="text-center">Consultas</TableHead>
                    <TableHead className="text-center">Prospecções</TableHead>
                    <TableHead className="text-center">PDFs</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {porUsuarioOrdenado.map((u) => (
                    <TableRow key={u.email}>
                      <TableCell>
                        <strong className="text-verde-escuro">{u.nome || "—"}</strong>
                        <br />
                        <span className="text-xs text-arcom-gray">{u.email}</span>
                      </TableCell>
                      <TableCell className="text-center">{u.consultas}</TableCell>
                      <TableCell className="text-center">{u.prospeccoes}</TableCell>
                      <TableCell className="text-center">{u.pdfIndicacao}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Contas</CardTitle>
          <CardDescription>Aprovar, negar, promover ou remover acesso de admin.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nome</TableHead>
                <TableHead>E-mail</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Admin</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {usuarios.map((u) => (
                <TableRow key={u.id}>
                  <TableCell>{u.nome}</TableCell>
                  <TableCell className="text-xs text-arcom-gray">{u.email}</TableCell>
                  <TableCell>{badgeStatus(u.status)}</TableCell>
                  <TableCell>{u.isAdmin ? <Badge variant="brand">Admin</Badge> : "—"}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      {u.status !== "aprovado" && (
                        <button title="Aprovar" onClick={() => definirStatus(u.id, "aprovado")}>
                          <Check className="h-4 w-4 text-verde-arcom hover:brightness-75" />
                        </button>
                      )}
                      {u.status !== "negado" && (
                        <button title="Negar" onClick={() => definirStatus(u.id, "negado")}>
                          <X className="h-4 w-4 text-danger hover:brightness-75" />
                        </button>
                      )}
                      {u.status === "aprovado" &&
                        (u.isAdmin ? (
                          <button title="Remover admin" onClick={() => definirAdmin(u.id, false)}>
                            <ShieldOff className="h-4 w-4 text-arcom-gray hover:text-danger" />
                          </button>
                        ) : (
                          <button title="Tornar admin" onClick={() => definirAdmin(u.id, true)}>
                            <Shield className="h-4 w-4 text-arcom-gray hover:text-verde-arcom" />
                          </button>
                        ))}
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">Atividades recentes</CardTitle>
            <CardDescription>Últimos eventos do site.</CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Select
              value={filtroAtividade || "todos"}
              onValueChange={(v) => setFiltroAtividade(v === "todos" ? "" : v)}
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="Todos os tipos" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="todos">Todos os tipos</SelectItem>
                {Object.entries(ATIVIDADE_META).map(([tipo, meta]) => (
                  <SelectItem key={tipo} value={tipo}>
                    {meta.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button variant="ghost" size="sm" onClick={limparAntigos}>
              <Trash2 className="h-4 w-4" /> Limpar logins antigos
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          {atividades.length === 0 ? (
            <p className="text-sm text-arcom-gray">Nenhuma atividade encontrada.</p>
          ) : (
            atividades.map((a) => {
              const meta = ATIVIDADE_META[a.tipo] ?? { label: a.tipo, cor: "text-arcom-gray" };
              return (
                <div key={a.id} className="border border-surface-border rounded-md px-3 py-2">
                  <p className="text-sm">
                    <strong className={meta.cor}>{meta.label}</strong>{" "}
                    <span className="text-verde-escuro">— {a.nome || a.email || "—"}</span>
                  </p>
                  <p className="text-xs text-arcom-gray">
                    {a.detalhe} · {new Date(a.criadoEm).toLocaleString("pt-BR")}
                  </p>
                </div>
              );
            })
          )}
        </CardContent>
      </Card>
    </div>
  );
}
