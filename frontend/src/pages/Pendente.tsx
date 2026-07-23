import { Clock } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Alert } from "@/components/ui/alert";

export default function Pendente() {
  return (
    <div className="min-h-screen bg-surface flex items-center justify-center px-6">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="items-center">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-surface">
            <Clock className="h-6 w-6 text-verde-arcom" strokeWidth={2} />
          </div>
          <CardTitle>Conta aguardando aprovação</CardTitle>
          <CardDescription>
            Um administrador precisa liberar seu acesso antes de você entrar no sistema.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Alert variant="info">Você será avisado assim que sua conta for aprovada.</Alert>
        </CardContent>
      </Card>
    </div>
  );
}
