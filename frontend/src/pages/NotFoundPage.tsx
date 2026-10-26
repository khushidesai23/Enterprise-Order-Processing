import { Link } from "react-router-dom";
import { Compass } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/primitives";
import { pill } from "@/store/ui";

export function NotFoundPage({ admin = false }: { admin?: boolean }) {
  if (admin) {
    return (
      <div className="mx-auto max-w-xl px-4 py-20">
        <EmptyState
          icon={<Compass />}
          title="Page not found"
          description="The page you're looking for doesn't exist or has moved."
          action={
            <Link to="/admin">
              <Button variant="secondary">Back to overview</Button>
            </Link>
          }
        />
      </div>
    );
  }

  return (
    <div className="mx-auto flex max-w-[1600px] flex-col items-start px-5 py-20 sm:px-8">
      <p className="font-mono text-[11px] uppercase tracking-[0.2em] text-smoke">Error 404</p>
      <p className="mt-4 font-display text-[clamp(6rem,22vw,20rem)] font-extrabold leading-[0.8] tracking-[-0.07em]">
        4<span className="font-serif font-normal italic text-ultra">0</span>4
      </p>
      <p className="mt-6 max-w-md text-lg text-smoke">This page isn't in the index. It may have moved, or never existed.</p>
      <div className="mt-8 flex flex-wrap gap-2">
        <Link to="/shop" className={pill("ink", "lg")}>
          Back home
        </Link>
        <Link to="/shop/search" className={pill("line", "lg")}>
          Open the index
        </Link>
      </div>
    </div>
  );
}
