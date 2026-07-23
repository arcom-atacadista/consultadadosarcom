import { useState } from "react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Input, FormField } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Alert } from "@/components/ui/alert";
import { useAuth } from "@/hooks/useAuth";

type RegistrarForm = { nome: string; email: string; senha: string };

export default function Registrar() {
  const navigate = useNavigate();
  const { registrar } = useAuth();
  const [erro, setErro] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegistrarForm>();

  const onSubmit = async (values: RegistrarForm) => {
    setErro(null);
    setEnviando(true);
    try {
      await registrar(values.email, values.senha, values.nome);
      navigate("/login");
    } catch (e: any) {
      setErro(e?.response?.data?.erro ?? "Não foi possível criar a conta.");
    } finally {
      setEnviando(false);
    }
  };

  return (
    <div className="min-h-screen bg-surface flex items-center justify-center px-6">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle className="text-verde-escuro">Criar conta</CardTitle>
          <CardDescription>
            Sua conta fica pendente até um administrador aprovar o acesso.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="flex flex-col gap-4" onSubmit={handleSubmit(onSubmit)}>
            {erro && <Alert variant="danger">{erro}</Alert>}
            <FormField label="Nome" error={errors.nome?.message}>
              <Label htmlFor="nome" className="sr-only">Nome</Label>
              <Input
                id="nome"
                placeholder="Seu nome"
                error={!!errors.nome}
                {...register("nome", { required: "Informe seu nome" })}
              />
            </FormField>
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
            <FormField label="Senha" error={errors.senha?.message} hint="Mínimo de 8 caracteres.">
              <Label htmlFor="senha" className="sr-only">Senha</Label>
              <Input
                id="senha"
                type="password"
                placeholder="••••••••"
                error={!!errors.senha}
                {...register("senha", {
                  required: "Informe uma senha",
                  minLength: { value: 8, message: "Mínimo de 8 caracteres" },
                })}
              />
            </FormField>
            <Button type="submit" className="mt-2" disabled={enviando}>
              {enviando ? "Criando…" : "Criar conta"}
            </Button>
          </form>
          <p className="mt-4 text-center text-sm text-arcom-gray">
            Já tem conta?{" "}
            <Link to="/login" className="font-semibold text-verde-arcom hover:underline">
              Entrar
            </Link>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
