import { useLocation } from "react-router-dom";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";

const TITLES: Record<string, string> = {
  "/consulta": "Consulta de CNPJ",
  "/prospeccao": "Prospecção",
  "/enriquecimento": "Enriquecimento",
  "/conversao": "Conversão",
  "/admin": "Administração",
};

export function Topbar() {
  const { pathname } = useLocation();
  const titulo = TITLES[pathname] ?? "CDA";

  return (
    <header className="flex items-center justify-between border-b border-surface-border bg-white px-6 py-4">
      <h1 className="font-bold text-xl text-verde-escuro">{titulo}</h1>
      <Button variant="ghost" size="sm">
        <LogOut className="h-4 w-4" strokeWidth={2} />
        Sair
      </Button>
    </header>
  );
}
