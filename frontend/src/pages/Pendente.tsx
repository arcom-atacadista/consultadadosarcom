import { Clock, XCircle } from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Alert } from "@/components/ui/alert";

export default function Pendente() {
  const { usuario } = useAuth();
  const negado = usuario?.status === "negado";

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center px-6">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-surface">
            {negado ? (
              <XCircle className="h-6 w-6 text-danger" strokeWidth={2} />
            ) : (
              <Clock className="h-6 w-6 text-verde-arcom" strokeWidth={2} />
            )}
          </div>
          <CardTitle>{negado ? "Acesso não liberado" : "Conta aguardando aprovação"}</CardTitle>
          <CardDescription>
            {negado
              ? "Um administrador negou o acesso a esta conta. Fale com o administrador se acha que isso é um engano."
              : "Um administrador precisa liberar seu acesso antes de você entrar no sistema."}
          </CardDescription>
        </CardHeader>
        <CardContent>
          {negado ? (
            <Alert variant="danger">Seu acesso foi negado.</Alert>
          ) : (
            <Alert variant="info">Você será avisado assim que sua conta for aprovada.</Alert>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
