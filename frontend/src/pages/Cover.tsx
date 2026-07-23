import { Link } from "react-router-dom";
import { Building2 } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function Cover() {
  return (
    <div className="min-h-screen bg-verde-escuro flex items-center justify-center px-6">
      <div className="text-center max-w-lg">
        <div className="mx-auto mb-6 flex h-16 w-16 items-center justify-center rounded-lg bg-verde-lima">
          <Building2 className="h-8 w-8 text-verde-escuro" strokeWidth={2} />
        </div>
        <h1 className="font-black text-4xl md:text-5xl text-white">CDA</h1>
        <p className="mt-3 text-white/70 text-lg">
          Consulta, prospecção e enriquecimento comercial de empresas via CNPJ —
          uso interno ARCOM.
        </p>
        <Button asChild size="lg" variant="accent" className="mt-8">
          <Link to="/login">Entrar</Link>
        </Button>
      </div>
    </div>
  );
}
