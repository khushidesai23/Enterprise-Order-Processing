import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { usersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { date, fullName, initials } from "@/lib/format";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardHeader, Field, Input, Skeleton } from "@/components/ui/primitives";
import { PageHeader } from "@/components/common";

export function AccountPage() {
  const { user, refresh } = useAuth();
  const profile = useQuery({ queryKey: ["user", user?.id], queryFn: () => usersApi.get(user!.id), enabled: !!user });
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");

  useEffect(() => {
    if (profile.data) {
      setFirstName(profile.data.first_name);
      setLastName(profile.data.last_name);
    }
  }, [profile.data]);

  const save = useMutation({
    mutationFn: () => usersApi.update(user!.id, { first_name: firstName.trim(), last_name: lastName.trim() }),
    onSuccess: async () => {
      toast.success("Profile updated");
      await Promise.all([refresh(), profile.refetch()]);
    },
    onError: (error) => toast.error("Could not update profile", { description: errorMessage(error) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (firstName.trim().length < 2) {
      toast.error("First name must be at least 2 characters");
      return;
    }
    save.mutate();
  }

  if (!user) return null;
  const dirty = profile.data && (firstName !== profile.data.first_name || lastName !== profile.data.last_name);

  return (
    <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6">
      <PageHeader title="My account" description="Manage your profile details." />
      <Card className="mb-6 flex items-center gap-4 p-5">
        <span className="flex size-14 items-center justify-center rounded-full bg-primary text-lg font-semibold text-white">
          {initials(fullName(user))}
        </span>
        <div className="min-w-0 flex-1">
          <p className="truncate text-lg font-semibold">{fullName(user)}</p>
          <p className="truncate text-sm text-subtle">{user.email}</p>
        </div>
        <Badge tone={user.role === "admin" ? "primary" : "neutral"}>{user.role === "admin" ? "Administrator" : "Customer"}</Badge>
      </Card>

      <Card>
        <CardHeader title="Profile" description={profile.data ? `Member since ${date(profile.data.created_at)}` : undefined} />
        {profile.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-24" />
          </div>
        ) : (
          <form onSubmit={submit} className="space-y-4 p-5">
            <div className="grid gap-4 sm:grid-cols-2">
              <Field label="First name" htmlFor="first_name">
                <Input id="first_name" value={firstName} onChange={(event) => setFirstName(event.target.value)} maxLength={100} />
              </Field>
              <Field label="Last name" htmlFor="last_name" hint="Leave unchanged to keep the current value.">
                <Input id="last_name" value={lastName} onChange={(event) => setLastName(event.target.value)} maxLength={100} />
              </Field>
            </div>
            <Field label="Email" htmlFor="email" hint="Email addresses cannot be changed.">
              <Input id="email" value={user.email} disabled />
            </Field>
            <div className="flex items-center justify-between pt-2">
              <Link to="/shop/orders" className="text-sm text-primary hover:underline">
                View my orders
              </Link>
              <Button type="submit" loading={save.isPending} disabled={!dirty}>
                Save changes
              </Button>
            </div>
          </form>
        )}
      </Card>
    </div>
  );
}
