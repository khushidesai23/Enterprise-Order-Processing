import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import * as Dropdown from "@radix-ui/react-dropdown-menu";
import { toast } from "sonner";
import { MoreHorizontal, Search, ShieldCheck, ShieldOff, Trash2, UserCheck, UserX, Users } from "lucide-react";
import { usersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { date, fullName, initials, number } from "@/lib/format";
import type { Role, User } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, EmptyState, ErrorState, Input, Select, Skeleton, Table, Td, Th } from "@/components/ui/primitives";
import { ConfirmDialog } from "@/components/ui/dialog";
import { PageHeader, StatCard } from "@/components/common";

type Pending =
  | { kind: "role"; user: User; role: Role }
  | { kind: "active"; user: User; active: boolean }
  | { kind: "delete"; user: User };

export function CustomersAdminPage() {
  const { user: me } = useAuth();
  const queryClient = useQueryClient();
  const users = useQuery({ queryKey: ["users"], queryFn: usersApi.list });
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState<Role | "">("");
  const [pending, setPending] = useState<Pending | null>(null);

  const visible = useMemo(() => {
    const needle = search.trim().toLowerCase();
    return (users.data ?? []).filter(
      (user) =>
        (!roleFilter || user.role === roleFilter) &&
        (!needle || fullName(user).toLowerCase().includes(needle) || user.email.toLowerCase().includes(needle)),
    );
  }, [users.data, search, roleFilter]);

  const act = useMutation({
    mutationFn: (action: Pending) => {
      if (action.kind === "role") return usersApi.updateRole(action.user.id, action.role);
      if (action.kind === "active") return usersApi.update(action.user.id, { is_active: action.active });
      return usersApi.remove(action.user.id);
    },
    onSuccess: async (_, action) => {
      const name = fullName(action.user);
      toast.success(
        action.kind === "role"
          ? `${name} is now ${action.role === "admin" ? "an admin" : "a customer"}`
          : action.kind === "active"
            ? `${name} ${action.active ? "reactivated" : "deactivated"}`
            : `${name} deleted`,
        action.kind === "role" ? { description: "The change applies at their next sign-in." } : undefined,
      );
      setPending(null);
      await queryClient.invalidateQueries({ queryKey: ["users"] });
    },
    onError: (error) => toast.error("Action failed", { description: errorMessage(error) }),
  });

  const all = users.data ?? [];
  const admins = all.filter((user) => user.role === "admin").length;
  const inactive = all.filter((user) => !user.is_active).length;

  const copy = (() => {
    if (!pending) return { title: "", description: "", confirm: "", destructive: false };
    const name = fullName(pending.user);
    if (pending.kind === "role")
      return pending.role === "admin"
        ? { title: "Grant admin access?", description: `${name} will be able to manage the catalogue, inventory, orders, refunds and users.`, confirm: "Make admin", destructive: false }
        : { title: "Remove admin access?", description: `${name} will become a regular customer.`, confirm: "Remove admin", destructive: true };
    if (pending.kind === "active")
      return pending.active
        ? { title: "Reactivate account?", description: `${name} will be able to sign in again.`, confirm: "Reactivate", destructive: false }
        : { title: "Deactivate account?", description: `${name} will no longer be able to sign in. Their orders are kept.`, confirm: "Deactivate", destructive: true };
    return { title: "Delete user?", description: `${name}'s account is removed. Their past orders remain for reporting.`, confirm: "Delete user", destructive: true };
  })();

  const itemClass = "flex cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-sm outline-none data-[highlighted]:bg-muted [&_svg]:size-4";

  return (
    <div className="animate-in">
      <PageHeader title="Customers" description="Manage accounts, access and roles." />

      <div className="mb-6 grid gap-4 sm:grid-cols-3">
        <StatCard label="Total accounts" value={users.data ? number(all.length) : "–"} icon={<Users />} />
        <StatCard label="Administrators" value={users.data ? number(admins) : "–"} icon={<ShieldCheck />} tone="info" />
        <StatCard label="Deactivated" value={users.data ? number(inactive) : "–"} icon={<UserX />} tone="warning" />
      </div>

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint" />
            <Input className="pl-9" placeholder="Search by name or email" value={search} onChange={(event) => setSearch(event.target.value)} aria-label="Search users" />
          </div>
          <Select className="sm:w-44" value={roleFilter} onChange={(event) => setRoleFilter(event.target.value as Role | "")} aria-label="Filter by role">
            <option value="">All roles</option>
            <option value="customer">Customers</option>
            <option value="admin">Admins</option>
          </Select>
        </div>

        {users.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(users.error)} onRetry={() => users.refetch()} />
          </div>
        ) : users.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-64" />
          </div>
        ) : visible.length === 0 ? (
          <EmptyState icon={<Users />} title="No users match" />
        ) : (
          <Table>
            <thead>
              <tr>
                <Th>User</Th>
                <Th>Role</Th>
                <Th>Status</Th>
                <Th>Joined</Th>
                <Th className="text-right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {visible.map((user) => {
                const self = user.id === me?.id;
                return (
                  <tr key={user.id} className="hover:bg-muted/40" data-testid="admin-user-row">
                    <Td>
                      <div className="flex items-center gap-3">
                        <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary-soft text-xs font-semibold text-primary">
                          {initials(fullName(user))}
                        </span>
                        <div className="min-w-0">
                          <p className="truncate font-medium">
                            {fullName(user)} {self && <span className="text-xs font-normal text-subtle">(you)</span>}
                          </p>
                          <p className="truncate text-xs text-subtle">{user.email}</p>
                        </div>
                      </div>
                    </Td>
                    <Td>
                      <Badge tone={user.role === "admin" ? "primary" : "neutral"}>{user.role === "admin" ? "Admin" : "Customer"}</Badge>
                    </Td>
                    <Td>
                      <Badge tone={user.is_active ? "success" : "warning"}>{user.is_active ? "Active" : "Deactivated"}</Badge>
                    </Td>
                    <Td className="text-subtle">{date(user.created_at)}</Td>
                    <Td className="text-right">
                      {self ? (
                        <span className="text-xs text-faint">—</span>
                      ) : (
                        <Dropdown.Root>
                          <Dropdown.Trigger asChild>
                            <Button variant="ghost" size="icon" aria-label={`Actions for ${fullName(user)}`}>
                              <MoreHorizontal />
                            </Button>
                          </Dropdown.Trigger>
                          <Dropdown.Portal>
                            <Dropdown.Content align="end" sideOffset={4} className="animate-in z-50 w-52 rounded-xl border border-border bg-surface p-1.5 shadow-pop">
                              {user.role === "admin" ? (
                                <Dropdown.Item className={itemClass} onSelect={() => setPending({ kind: "role", user, role: "customer" })}>
                                  <ShieldOff className="text-subtle" /> Remove admin
                                </Dropdown.Item>
                              ) : (
                                <Dropdown.Item className={itemClass} onSelect={() => setPending({ kind: "role", user, role: "admin" })}>
                                  <ShieldCheck className="text-subtle" /> Make admin
                                </Dropdown.Item>
                              )}
                              <Dropdown.Item className={itemClass} onSelect={() => setPending({ kind: "active", user, active: !user.is_active })}>
                                {user.is_active ? <UserX className="text-subtle" /> : <UserCheck className="text-subtle" />}
                                {user.is_active ? "Deactivate" : "Reactivate"}
                              </Dropdown.Item>
                              <Dropdown.Separator className="my-1 h-px bg-border" />
                              <Dropdown.Item className={`${itemClass} text-danger`} onSelect={() => setPending({ kind: "delete", user })}>
                                <Trash2 /> Delete user
                              </Dropdown.Item>
                            </Dropdown.Content>
                          </Dropdown.Portal>
                        </Dropdown.Root>
                      )}
                    </Td>
                  </tr>
                );
              })}
            </tbody>
          </Table>
        )}
      </Card>

      <ConfirmDialog
        open={!!pending}
        onOpenChange={(open) => !open && setPending(null)}
        title={copy.title}
        description={copy.description}
        confirmLabel={copy.confirm}
        destructive={copy.destructive}
        loading={act.isPending}
        onConfirm={() => pending && act.mutate(pending)}
      />
    </div>
  );
}
