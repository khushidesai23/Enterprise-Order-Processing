import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { ArrowUpRight } from "lucide-react";
import { ordersApi, usersApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { date } from "@/lib/format";
import { useWishlist } from "@/lib/saved";
import { StoreButton, StoreField, StoreInput } from "@/store/Field";
import { Eyebrow, MaskText, Skel } from "@/store/ui";

export function AccountPage() {
  const { user, refresh } = useAuth();
  const wishlist = useWishlist();
  const profile = useQuery({ queryKey: ["user", user?.id], queryFn: () => usersApi.get(user!.id), enabled: !!user });
  const orders = useQuery({ queryKey: ["orders", "me", 1], queryFn: () => ordersApi.mine(1, 20), enabled: !!user });
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
      toast("Profile updated");
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
  const recent = orders.data ?? [];
  const open = recent.filter((order) => !["DELIVERED", "CANCELLED"].includes(order.status)).length;

  return (
    <div className="mx-auto max-w-[1200px] px-5 pt-6 sm:px-8">
      <Eyebrow>Account · member since {profile.data ? date(profile.data.created_at) : "—"}</Eyebrow>
      <h1 className="mt-3 font-display text-[clamp(3rem,8vw,7rem)] font-extrabold leading-[0.88] tracking-[-0.055em]">
        <MaskText text="Hi," /> <MaskText text={`${user.first_name}.`} delay={0.1} className="font-serif font-normal italic tracking-[-0.02em] text-ultra" />
      </h1>

      <div className="mt-10 grid gap-4 sm:grid-cols-3">
        {[
          { label: "Open orders", value: orders.isLoading ? "—" : String(open), to: "/shop/orders" },
          { label: "Recent orders", value: orders.isLoading ? "—" : String(recent.length), to: "/shop/orders" },
          { label: "Wishlist", value: String(wishlist.items.length), to: "/shop/wishlist" },
        ].map((card) => (
          <Link key={card.label} to={card.to} className="group flex items-end justify-between rounded-[2rem] bg-bone-2/70 p-6 transition-colors hover:bg-ink hover:text-bone">
            <span>
              <span className="block font-mono text-[11px] uppercase tracking-[0.16em] opacity-60">{card.label}</span>
              <span className="mt-2 block font-display text-5xl font-extrabold tracking-tight">{card.value}</span>
            </span>
            <ArrowUpRight className="size-6 transition-transform group-hover:-translate-y-1 group-hover:translate-x-1" />
          </Link>
        ))}
      </div>

      <section className="mt-6 rounded-[2rem] bg-bone-2/70 p-6 sm:p-10">
        <Eyebrow>Profile</Eyebrow>
        {profile.isLoading ? (
          <Skel className="mt-6 h-40" />
        ) : (
          <form onSubmit={submit} className="mt-6 grid gap-5 sm:grid-cols-2">
            <StoreField label="First name" htmlFor="first_name">
              <StoreInput id="first_name" value={firstName} onChange={(event) => setFirstName(event.target.value)} maxLength={100} />
            </StoreField>
            <StoreField label="Last name" htmlFor="last_name">
              <StoreInput id="last_name" value={lastName} onChange={(event) => setLastName(event.target.value)} maxLength={100} />
            </StoreField>
            <StoreField label="Email" htmlFor="email" hint="Email addresses cannot be changed.">
              <StoreInput id="email" value={user.email} disabled />
            </StoreField>
            <div className="flex items-end justify-end">
              <StoreButton type="submit" loading={save.isPending} disabled={!dirty}>
                Save changes
              </StoreButton>
            </div>
          </form>
        )}
      </section>
    </div>
  );
}
