import { useForm } from "react-hook-form";
import { useNavigate } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input, FormField } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

type LoginForm = { email: string; senha: string };

export default function Login() {
  const navigate = useNavigate();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>();

  // POST /api/auth/login entra na Fase 3 (auth). Por ora só valida o formulário
  // e demonstra a navegação até a tela de espera de aprovação.
  const onSubmit = () => navigate("/pendente");

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center px-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-verde-escuro">Entrar</CardTitle>
          <CardDescription>Acesso restrito à equipe ARCOM.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="flex flex-col gap-4" onSubmit={handleSubmit(onSubmit)}>
            <FormField label="E-mail" error={errors.email?.message}>
              <Label htmlFor="email" className="sr-only">E-mail</Label>
              <Input
                id="email"
                type="email"
                placeholder="voce@arcom.com.br"
                error={!!errors.email}
                {...register("email", { required: "Informe o e-mail" })}
              />
            </FormField>
            <FormField label="Senha" error={errors.senha?.message}>
              <Label htmlFor="senha" className="sr-only">Senha</Label>
              <Input
                id="senha"
                type="password"
                placeholder="••••••••"
                error={!!errors.senha}
                {...register("senha", { required: "Informe a senha" })}
              />
            </FormField>
            <Button type="submit" className="mt-2">Entrar</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
