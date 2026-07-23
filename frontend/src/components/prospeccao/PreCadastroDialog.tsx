import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Input, FormField } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { api } from "@/lib/api";
import { formatarCNPJ } from "@/lib/cnpj";
import type { Prospect } from "@/lib/prospeccao";

export function PreCadastroDialog({
  prospect,
  onClose,
}: {
  prospect: Prospect | null;
  onClose: () => void;
}) {
  const [contato, setContato] = useState(prospect?.telefone ?? "");
  const [notas, setNotas] = useState("");
  const [enviando, setEnviando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);
  const [ok, setOk] = useState(false);

  async function salvar() {
    if (!prospect) return;
    setEnviando(true);
    setErro(null);
    try {
      await api.post("/precadastros", {
        cnpj: prospect.cnpj,
        razao: prospect.razao,
        endereco: prospect.endereco,
        contato,
        notas,
      });
      setOk(true);
    } catch {
      setErro("Não foi possível salvar o pré-cadastro.");
    } finally {
      setEnviando(false);
    }
  }

  function fechar() {
    setOk(false);
    setErro(null);
    setNotas("");
    onClose();
  }

  return (
    <Dialog open={!!prospect} onOpenChange={(open) => !open && fechar()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Pré-cadastrar cliente</DialogTitle>
          <DialogDescription>
            {prospect && `${prospect.razao} — ${formatarCNPJ(prospect.cnpj)}`}
          </DialogDescription>
        </DialogHeader>
        {ok ? (
          <Alert variant="success">Pré-cadastro salvo com sucesso.</Alert>
        ) : (
          <div className="flex flex-col gap-4 pt-2">
            {erro && <Alert variant="danger">{erro}</Alert>}
            <FormField label="Contato">
              <Input value={contato} onChange={(e) => setContato(e.target.value)} placeholder="Telefone/e-mail" />
            </FormField>
            <FormField label="Notas">
              <Textarea value={notas} onChange={(e) => setNotas(e.target.value)} rows={3} />
            </FormField>
            <Button onClick={salvar} disabled={enviando}>
              {enviando ? "Salvando…" : "Salvar pré-cadastro"}
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
