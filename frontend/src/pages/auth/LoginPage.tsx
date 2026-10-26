import { useState, type FormEvent } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { StoreButton as Button } from "@/store/Field";
import { StoreField as Field, StoreInput as Input } from "@/store/Field";
import { AuthShell } from "./AuthShell";

export function LoginPage() {
  const { login, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from;
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  if (user) return <Navigate to={from ?? (user.role === "admin" ? "/admin" : "/shop")} replace />;

  async function submit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setSubmitting(true);
    try {
      const signedIn = await login(email, password);
      toast.success(`Welcome back, ${signedIn.first_name}`);
      navigate(from ?? (signedIn.role === "admin" ? "/admin" : "/shop"), { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell
      title={<>Welcome <span className="font-serif font-normal italic tracking-[-0.02em] text-ultra">back.</span></>}
      subtitle="Track orders, check out faster and manage your account."
      footer={
        <>
          New here?{" "}
          <Link to="/register" state={location.state} className="font-medium text-ink underline underline-offset-4 hover:text-ultra">
            Create an account
          </Link>
        </>
      }
    >
      <form onSubmit={submit} className="space-y-4" noValidate>
        {error && (
          <div role="alert" className="rounded-2xl border border-signal/30 bg-signal/10 px-4 py-3 text-sm text-signal">
            {error}
          </div>
        )}
        <Field label="Email" htmlFor="email">
          <Input
            id="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            required
            autoFocus
          />
        </Field>
        <Field label="Password" htmlFor="password">
          <Input
            id="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            required
          />
        </Field>
        <Button type="submit" className="w-full" size="lg" loading={submitting} disabled={!email || !password}>
          Sign in
        </Button>
      </form>
    </AuthShell>
  );
}
