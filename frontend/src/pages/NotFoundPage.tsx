import { Link } from "react-router-dom";
import { Compass } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/primitives";

export function NotFoundPage({ admin = false }: { admin?: boolean }) {
  return (
    <div className="mx-auto max-w-xl px-4 py-20">
      <EmptyState
        icon={<Compass />}
        title="Page not found"
        description="The page you're looking for doesn't exist or has moved."
        action={
          <Link to={admin ? "/admin" : "/shop"}>
            <Button variant="secondary">{admin ? "Back to overview" : "Back to shop"}</Button>
          </Link>
        }
      />
    </div>
  );
}
