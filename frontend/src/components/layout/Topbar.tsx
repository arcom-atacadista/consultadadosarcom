import { useLocation, useNavigate } from "react-router-dom";
import { LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/hooks/useAuth";

const TITLES: Record<string, string> = {
  "/consulta": "Consulta de CNPJ",
  "/prospeccao": "Prospecção",
  "/enriquecimento": "Enriquecimento",
  "/conversao": "Conversão",
  "/admin": "Administração",
};

export function Topbar() {
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const { logout, usuario } = useAuth();
  const titulo = TITLES[pathname] ?? "CDA";

  return (
    <header className="flex items-center justify-between border-b border-surface-border bg-white px-6 py-4">
      <h1 className="font-bold text-xl text-verde-escuro">{titulo}</h1>
      <div className="flex items-center gap-3">
        {usuario && <span className="text-sm text-arcom-gray">{usuario.nome}</span>}
        <Button
          variant="ghost"
          size="sm"
          onClick={() => {
            logout();
            navigate("/login");
          }}
        >
          <LogOut className="h-4 w-4" strokeWidth={2} />
          Sair
        </Button>
      </div>
    </header>
  );
}
