import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/hooks/useAuth";

// Sem token/usuário -> /login. Entra na Fase 3 (auth).
export function ExigeLogin() {
  const { usuario, carregando } = useAuth();
  if (carregando) return null;
  if (!usuario) return <Navigate to="/login" replace />;
  return <Outlet />;
}

// Logado mas ainda não aprovado -> /pendente.
export function ExigeAprovado() {
  const { aprovado } = useAuth();
  if (!aprovado) return <Navigate to="/pendente" replace />;
  return <Outlet />;
}

// Aprovado mas não admin -> volta pra Consulta.
export function ExigeAdmin() {
  const { isAdmin } = useAuth();
  if (!isAdmin) return <Navigate to="/consulta" replace />;
  return <Outlet />;
}
