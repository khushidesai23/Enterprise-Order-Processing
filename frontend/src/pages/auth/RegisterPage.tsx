import { useState, type FormEvent } from "react";
import { Link, Navigate, useLocation, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Field, Input } from "@/components/ui/primitives";
import { AuthShell } from "./AuthShell";

interface Form {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
  confirm: string;
}

// Matches CreateUserRequest's validate tags in internal/modules/user/dto.go.
function validate(form: Form): Partial<Record<keyof Form, string>> {
  const errors: Partial<Record<keyof Form, string>> = {};
  if (form.first_name.trim().length < 2) errors.first_name = "At least 2 characters.";
  if (form.first_name.trim().length > 100) errors.first_name = "At most 100 characters.";
  if (form.last_name.trim().length > 100) errors.last_name = "At most 100 characters.";
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) errors.email = "Enter a valid email address.";
  if (form.password.length < 8) errors.password = "At least 8 characters.";
  if (form.confirm !== form.password) errors.confirm = "Passwords do not match.";
  return errors;
}

export function RegisterPage() {
  const { register, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from;
  const [form, setForm] = useState<Form>({ first_name: "", last_name: "", email: "", password: "", confirm: "" });
  const [touched, setTouched] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  if (user) return <Navigate to={from ?? "/shop"} replace />;

  const errors = touched ? validate(form) : {};
  const set = (key: keyof Form) => (event: React.ChangeEvent<HTMLInputElement>) =>
    setForm((current) => ({ ...current, [key]: event.target.value }));

  async function submit(event: FormEvent) {
    event.preventDefault();
    setTouched(true);
    setError(null);
    if (Object.keys(validate(form)).length > 0) return;
    setSubmitting(true);
    try {
      const created = await register({
        first_name: form.first_name.trim(),
        last_name: form.last_name.trim(),
        email: form.email.trim(),
        password: form.password,
      });
      toast.success(`Welcome to Meridian, ${created.first_name}!`);
      navigate(from ?? "/shop", { replace: true });
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell
      title="Create your account"
      subtitle="It takes less than a minute."
      footer={
        <>
          Already have an account?{" "}
          <Link to="/login" state={location.state} className="font-medium text-primary hover:underline">
            Sign in
          </Link>
        </>
      }
    >
      <form onSubmit={submit} className="space-y-4" noValidate>
        {error && (
          <div role="alert" className="rounded-lg border border-danger/20 bg-danger-soft px-3 py-2 text-sm text-danger">
            {error}
          </div>
        )}
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="First name" htmlFor="first_name" error={errors.first_name}>
            <Input id="first_name" autoComplete="given-name" value={form.first_name} onChange={set("first_name")} aria-invalid={!!errors.first_name} autoFocus />
          </Field>
          <Field label="Last name" htmlFor="last_name" error={errors.last_name}>
            <Input id="last_name" autoComplete="family-name" value={form.last_name} onChange={set("last_name")} aria-invalid={!!errors.last_name} />
          </Field>
        </div>
        <Field label="Email" htmlFor="email" error={errors.email}>
          <Input id="email" type="email" autoComplete="email" value={form.email} onChange={set("email")} aria-invalid={!!errors.email} />
        </Field>
        <Field label="Password" htmlFor="password" error={errors.password} hint="At least 8 characters.">
          <Input id="password" type="password" autoComplete="new-password" value={form.password} onChange={set("password")} aria-invalid={!!errors.password} />
        </Field>
        <Field label="Confirm password" htmlFor="confirm" error={errors.confirm}>
          <Input id="confirm" type="password" autoComplete="new-password" value={form.confirm} onChange={set("confirm")} aria-invalid={!!errors.confirm} />
        </Field>
        <Button type="submit" className="w-full" size="lg" loading={submitting}>
          Create account
        </Button>
      </form>
    </AuthShell>
  );
}
