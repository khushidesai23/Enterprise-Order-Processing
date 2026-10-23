import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { setUnauthorizedHandler, tokenStore } from "./api";
import { authApi, usersApi } from "./endpoints";
import type { UserInfo } from "./types";

interface AuthState {
  user: UserInfo | null;
  // True until the stored token (if any) has been checked against /auth/me.
  loading: boolean;
  isAdmin: boolean;
  // True after the user signed out themselves (as opposed to an expired
  // session), so guarded pages send them to the shop, not to sign in.
  signedOut: boolean;
  login: (email: string, password: string) => Promise<UserInfo>;
  register: (input: { first_name: string; last_name: string; email: string; password: string }) => Promise<UserInfo>;
  logout: () => void;
  refresh: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(() => tokenStore.get() !== null);
  const [signedOut, setSignedOut] = useState(false);

  const clearSession = useCallback(() => {
    tokenStore.set(null);
    setUser(null);
    queryClient.clear();
  }, [queryClient]);

  const logout = useCallback(() => {
    setSignedOut(true);
    clearSession();
  }, [clearSession]);

  const refresh = useCallback(async () => {
    if (!tokenStore.get()) {
      setUser(null);
      setLoading(false);
      return;
    }
    try {
      setUser(await authApi.me());
    } catch {
      tokenStore.set(null);
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(clearSession);
    void refresh();
    return () => setUnauthorizedHandler(null);
  }, [clearSession, refresh]);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await authApi.login(email.trim(), password);
      tokenStore.set(result.token);
      queryClient.clear();
      setSignedOut(false);
      setUser(result.user);
      return result.user;
    },
    [queryClient],
  );

  const register = useCallback(
    async (input: { first_name: string; last_name: string; email: string; password: string }) => {
      await usersApi.register(input);
      return login(input.email, input.password);
    },
    [login],
  );

  const value = useMemo<AuthState>(
    () => ({ user, loading, isAdmin: user?.role === "admin", signedOut, login, register, logout, refresh }),
    [user, loading, signedOut, login, register, logout, refresh],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used inside AuthProvider");
  return context;
}
