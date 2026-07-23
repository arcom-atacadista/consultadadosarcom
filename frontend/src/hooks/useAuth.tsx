import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";
import { api } from "@/lib/api";

export type Usuario = {
  id: string;
  email: string;
  nome: string;
  status: "pendente" | "aprovado";
  isAdmin: boolean;
  criadoEm: string;
};

type AuthState = {
  usuario: Usuario | null;
  isAdmin: boolean;
  aprovado: boolean;
  carregando: boolean;
  erro: string | null;
  login: (email: string, senha: string) => Promise<{ aprovado: boolean; isAdmin: boolean }>;
  registrar: (email: string, senha: string, nome: string) => Promise<void>;
  logout: () => void;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [usuario, setUsuario] = useState<Usuario | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [aprovado, setAprovado] = useState(false);
  const [carregando, setCarregando] = useState(true);
  const [erro, setErro] = useState<string | null>(null);

  const aplicarMe = (data: { usuario: Usuario; isAdmin: boolean; aprovado: boolean }) => {
    setUsuario(data.usuario);
    setIsAdmin(data.isAdmin);
    setAprovado(data.aprovado);
  };

  const limpar = () => {
    localStorage.removeItem("token");
    setUsuario(null);
    setIsAdmin(false);
    setAprovado(false);
  };

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      setCarregando(false);
      return;
    }
    api
      .get("/auth/me")
      .then(({ data }) => aplicarMe(data))
      .catch(() => limpar())
      .finally(() => setCarregando(false));
  }, []);

  async function login(email: string, senha: string) {
    setErro(null);
    try {
      const { data } = await api.post("/auth/login", { email, senha });
      localStorage.setItem("token", data.token);
      const me = await api.get("/auth/me");
      aplicarMe(me.data);
      return { aprovado: me.data.aprovado as boolean, isAdmin: me.data.isAdmin as boolean };
    } catch {
      setErro("E-mail ou senha inválidos.");
      throw new Error("login falhou");
    }
  }

  async function registrar(email: string, senha: string, nome: string) {
    setErro(null);
    try {
      await api.post("/auth/register", { email, senha, nome });
    } catch (e: any) {
      setErro(e?.response?.data?.erro ?? "Não foi possível criar a conta.");
      throw e;
    }
  }

  function logout() {
    limpar();
  }

  return (
    <AuthContext.Provider
      value={{ usuario, isAdmin, aprovado, carregando, erro, login, registrar, logout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth precisa estar dentro de <AuthProvider>");
  return ctx;
}
