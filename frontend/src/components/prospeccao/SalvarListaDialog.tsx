import { useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { FormField, Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export function SalvarListaDialog({
  aberto,
  total,
  onClose,
  onSalvar,
}: {
  aberto: boolean;
  total: number;
  onClose: () => void;
  onSalvar: (nome: string, assessor: string) => Promise<void>;
}) {
  const [nome, setNome] = useState("");
  const [assessor, setAssessor] = useState("");
  const [salvando, setSalvando] = useState(false);
  const [erro, setErro] = useState<string | null>(null);

  async function confirmar() {
    if (!nome.trim() || !assessor.trim()) {
      setErro("Informe o nome da lista e o assessor.");
      return;
    }
    setErro(null);
    setSalvando(true);
    try {
      await onSalvar(nome.trim(), assessor.trim());
      setNome("");
      setAssessor("");
      onClose();
    } catch {
      setErro("Falha ao salvar a lista.");
    } finally {
      setSalvando(false);
    }
  }

  return (
    <Dialog open={aberto} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Salvar lista de prospecção</DialogTitle>
          <DialogDescription>{total} empresa(s) selecionada(s) — envie para um assessor.</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <FormField label="Nome da lista">
            <Input value={nome} onChange={(e) => setNome(e.target.value)} placeholder="Ex.: Supermercados Centro" />
          </FormField>
          <FormField label="Assessor">
            <Input value={assessor} onChange={(e) => setAssessor(e.target.value)} placeholder="Nome do assessor" />
          </FormField>
          {erro && <p className="text-sm text-danger">{erro}</p>}
          <Button onClick={confirmar} disabled={salvando}>
            {salvando ? "Salvando…" : "Salvar lista"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
