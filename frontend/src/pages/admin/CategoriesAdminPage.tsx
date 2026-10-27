import { useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Pencil, Plus, Tags, Trash2 } from "lucide-react";
import { categoriesApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { number } from "@/lib/format";
import type { CategoryListItem } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, EmptyState, ErrorState, Field, Input, Skeleton, Table, Td, Textarea, Th } from "@/components/ui/primitives";
import { ConfirmDialog, Dialog } from "@/components/ui/dialog";
import { PageHeader } from "@/components/common";

export function CategoriesAdminPage() {
  const queryClient = useQueryClient();
  const categories = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list });
  const [editing, setEditing] = useState<CategoryListItem | "new" | null>(null);
  const [deleting, setDeleting] = useState<CategoryListItem | null>(null);

  const remove = useMutation({
    mutationFn: (id: string) => categoriesApi.remove(id),
    onSuccess: async () => {
      toast.success("Category deleted");
      setDeleting(null);
      await queryClient.invalidateQueries({ queryKey: ["categories"] });
    },
    onError: (error) => toast.error("Could not delete category", { description: errorMessage(error) }),
  });

  return (
    <div className="animate-in">
      <PageHeader
        title="Categories"
        description="Group products so customers can browse the storefront."
        actions={
          <Button onClick={() => setEditing("new")}>
            <Plus /> New category
          </Button>
        }
      />
      <Card>
        {categories.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(categories.error)} onRetry={() => categories.refetch()} />
          </div>
        ) : categories.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-48" />
          </div>
        ) : categories.data?.length === 0 ? (
          <EmptyState icon={<Tags />} title="No categories yet" description="Create one to start adding products." />
        ) : (
          <Table>
            <thead>
              <tr>
                <Th>Name</Th>
                <Th>Description</Th>
                <Th className="text-right">Products</Th>
                <Th className="text-right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {categories.data?.map((category) => (
                <tr key={category.id} className="hover:bg-muted/40" data-testid="admin-category-row">
                  <Td className="font-medium">
                    <Link to={`/shop?category=${category.id}`} target="_blank" className="hover:text-primary">
                      {category.name}
                    </Link>
                  </Td>
                  <Td className="max-w-md truncate text-subtle">{category.description || "—"}</Td>
                  <Td className="tabular text-right">{number(category.product_count)}</Td>
                  <Td>
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" onClick={() => setEditing(category)} aria-label={`Edit ${category.name}`}>
                        <Pencil />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleting(category)}
                        disabled={category.product_count > 0}
                        title={category.product_count > 0 ? "Move or delete its products first" : undefined}
                        aria-label={`Delete ${category.name}`}
                      >
                        <Trash2 className="text-danger" />
                      </Button>
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </Card>

      {editing && <CategoryDialog category={editing === "new" ? null : editing} onClose={() => setEditing(null)} />}

      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Delete category?"
        description={`"${deleting?.name}" will be removed. Categories that still have products cannot be deleted.`}
        confirmLabel="Delete category"
        destructive
        loading={remove.isPending}
        onConfirm={() => deleting && remove.mutate(deleting.id)}
      />
    </div>
  );
}

function CategoryDialog({ category, onClose }: { category: CategoryListItem | null; onClose: () => void }) {
  const queryClient = useQueryClient();
  const [name, setName] = useState(category?.name ?? "");
  const [description, setDescription] = useState(category?.description ?? "");
  const [touched, setTouched] = useState(false);
  const nameError = touched && (name.trim().length < 2 || name.trim().length > 100) ? "2–100 characters." : undefined;
  const descriptionError = description.length > 500 ? "At most 500 characters." : undefined;

  const save = useMutation({
    mutationFn: () => {
      const input = { name: name.trim(), description: description.trim() };
      return category ? categoriesApi.update(category.id, input) : categoriesApi.create(input);
    },
    onSuccess: async () => {
      toast.success(category ? "Category updated" : "Category created");
      await queryClient.invalidateQueries();
      onClose();
    },
    onError: (error) => toast.error("Could not save category", { description: errorMessage(error) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    setTouched(true);
    if (name.trim().length >= 2 && name.trim().length <= 100 && !descriptionError) save.mutate();
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={category ? "Edit category" : "New category"}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={save.isPending}>
            Cancel
          </Button>
          <Button type="submit" form="category-form" loading={save.isPending}>
            {category ? "Save changes" : "Create category"}
          </Button>
        </>
      }
    >
      <form id="category-form" onSubmit={submit} className="space-y-4" noValidate>
        <Field label="Name" htmlFor="c-name" error={nameError}>
          <Input id="c-name" value={name} onChange={(event) => setName(event.target.value)} aria-invalid={!!nameError} autoFocus />
        </Field>
        <Field label="Description" htmlFor="c-description" error={descriptionError} hint={`${description.length}/500`}>
          <Textarea id="c-description" value={description} onChange={(event) => setDescription(event.target.value)} />
        </Field>
      </form>
    </Dialog>
  );
}
