import { useEffect, useState, type FormEvent } from "react";
import { Link } from "react-router-dom";
import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { ExternalLink, Package, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { categoriesApi, inventoryApi, productsApi } from "@/lib/endpoints";
import { errorMessage } from "@/lib/api";
import { money, number } from "@/lib/format";
import type { ProductInput, ProductListItem } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, EmptyState, ErrorState, Field, Input, Select, Skeleton, Table, Td, Textarea, Th } from "@/components/ui/primitives";
import { StockBadge } from "@/components/ui/badge";
import { ConfirmDialog, Dialog } from "@/components/ui/dialog";
import { PageHeader, Pagination } from "@/components/common";
import { ProductImage } from "@/components/shop";
import { LOW_STOCK_THRESHOLD } from "./InventoryAdminPage";

interface FormState {
  name: string;
  description: string;
  sku: string;
  price: string;
  mrp: string;
  brand: string;
  image_url: string;
  images: string;
  category_id: string;
  initial_stock: string;
}

const EMPTY: FormState = {
  name: "",
  description: "",
  sku: "",
  price: "",
  mrp: "",
  brand: "",
  image_url: "",
  images: "",
  category_id: "",
  initial_stock: "0",
};

const PAGE_SIZE = 25;

function imageLines(value: string): string[] {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

// Matches CreateProductRequest / UpdateProductRequest binding tags.
function validate(form: FormState, creating: boolean): Partial<Record<keyof FormState, string>> {
  const errors: Partial<Record<keyof FormState, string>> = {};
  const name = form.name.trim();
  const sku = form.sku.trim();
  if (name.length < 2 || name.length > 255) errors.name = "2–255 characters.";
  if (form.description.length > 1000) errors.description = "At most 1000 characters.";
  if (sku.length < 2 || sku.length > 100) errors.sku = "2–100 characters.";
  const price = Number(form.price);
  if (!form.price || Number.isNaN(price) || price <= 0) errors.price = "Enter a price greater than zero.";
  if (!form.category_id) errors.category_id = "Choose a category.";
  if (form.mrp) {
    const mrp = Number(form.mrp);
    if (Number.isNaN(mrp) || mrp < 0) errors.mrp = "Enter zero or a positive amount.";
    else if (mrp > 0 && mrp < price) errors.mrp = "MRP cannot be below the price.";
  }
  if (form.brand.length > 100) errors.brand = "At most 100 characters.";
  if (form.image_url.length > 500) errors.image_url = "At most 500 characters.";
  const gallery = imageLines(form.images);
  if (gallery.length > 10) errors.images = "At most 10 images.";
  else if (gallery.some((url) => url.length > 500)) errors.images = "Each URL can be at most 500 characters.";
  if (creating) {
    const stock = Number(form.initial_stock);
    if (!Number.isInteger(stock) || stock < 0) errors.initial_stock = "Whole number, zero or more.";
  }
  return errors;
}

export function ProductsAdminPage() {
  const queryClient = useQueryClient();
  const categories = useQuery({ queryKey: ["categories"], queryFn: categoriesApi.list });
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [page, setPage] = useState(1);

  // Search after typing pauses.
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setQuery(search.trim());
      setPage(1);
    }, 300);
    return () => window.clearTimeout(timer);
  }, [search]);

  const products = useQuery({
    queryKey: ["search", "admin-products", query, categoryFilter, page],
    queryFn: () =>
      productsApi.search({ q: query || undefined, category_id: categoryFilter || undefined, sort: "newest", page, limit: PAGE_SIZE }),
    placeholderData: keepPreviousData,
  });
  const [editing, setEditing] = useState<ProductListItem | "new" | null>(null);
  const [deleting, setDeleting] = useState<ProductListItem | null>(null);

  const visible = products.data?.items ?? [];

  const remove = useMutation({
    mutationFn: (id: string) => productsApi.remove(id),
    onSuccess: async () => {
      toast.success("Product deleted");
      setDeleting(null);
      await queryClient.invalidateQueries();
    },
    onError: (error) => toast.error("Could not delete product", { description: errorMessage(error) }),
  });

  return (
    <div className="animate-in">
      <PageHeader
        title="Products"
        description="Manage the catalogue customers see in the storefront."
        actions={
          <Button onClick={() => setEditing("new")} disabled={!categories.data?.length}>
            <Plus /> New product
          </Button>
        }
      />
      {categories.data && categories.data.length === 0 && (
        <div className="mb-4 rounded-lg border border-warning/25 bg-warning-soft px-4 py-3 text-sm">
          Create a <Link to="/admin/categories" className="font-medium text-primary hover:underline">category</Link> before adding products.
        </div>
      )}

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row">
          <div className="relative flex-1">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-faint" />
            <Input className="pl-9" placeholder="Search by name or SKU" value={search} onChange={(event) => setSearch(event.target.value)} aria-label="Search products" />
          </div>
          <Select
            className="sm:w-56"
            value={categoryFilter}
            onChange={(event) => {
              setCategoryFilter(event.target.value);
              setPage(1);
            }}
            aria-label="Filter by category"
          >
            <option value="">All categories</option>
            {categories.data?.map((category) => (
              <option key={category.id} value={category.id}>
                {category.name}
              </option>
            ))}
          </Select>
        </div>

        {products.isError ? (
          <div className="p-5">
            <ErrorState message={errorMessage(products.error)} onRetry={() => products.refetch()} />
          </div>
        ) : products.isLoading ? (
          <div className="p-5">
            <Skeleton className="h-64" />
          </div>
        ) : visible.length === 0 ? (
          <EmptyState icon={<Package />} title={query || categoryFilter ? "No products match" : "No products yet"} />
        ) : (
          <Table>
            <thead>
              <tr>
                <Th>Product</Th>
                <Th>Category</Th>
                <Th className="text-right">Price</Th>
                <Th className="text-right">Available</Th>
                <Th className="text-right">Reserved</Th>
                <Th>Stock</Th>
                <Th className="text-right">Actions</Th>
              </tr>
            </thead>
            <tbody>
              {visible.map((product) => (
                <tr key={product.id} className="hover:bg-muted/40" data-testid="admin-product-row">
                  <Td>
                    <div className="flex items-center gap-3">
                      <ProductImage id={product.id} name={product.name} src={product.image_url} size="sm" className="size-10 shrink-0 rounded-md border border-border" />
                      <div className="min-w-0">
                        <p className="max-w-72 truncate font-medium">{product.name}</p>
                        <p className="text-xs text-subtle">
                          {product.sku}
                          {product.brand ? ` · ${product.brand}` : ""}
                        </p>
                      </div>
                    </div>
                  </Td>
                  <Td className="text-subtle">{product.category}</Td>
                  <Td className="tabular text-right font-medium">
                    {money(product.price)}
                    {product.mrp > product.price && <p className="text-xs font-normal text-subtle line-through">{money(product.mrp)}</p>}
                  </Td>
                  <Td className="tabular text-right">{number(product.available_quantity)}</Td>
                  <Td className="tabular text-right text-subtle">{number(product.reserved_quantity)}</Td>
                  <Td>
                    <StockBadge available={product.available_quantity} threshold={LOW_STOCK_THRESHOLD} />
                  </Td>
                  <Td>
                    <div className="flex justify-end gap-1">
                      <Link to={`/shop/products/${product.id}`} target="_blank" aria-label={`View ${product.name} in store`}>
                        <Button variant="ghost" size="icon">
                          <ExternalLink />
                        </Button>
                      </Link>
                      <Button variant="ghost" size="icon" onClick={() => setEditing(product)} aria-label={`Edit ${product.name}`}>
                        <Pencil />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setDeleting(product)} aria-label={`Delete ${product.name}`}>
                        <Trash2 className="text-danger" />
                      </Button>
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
        {products.data && products.data.total > 0 && (
          <div className="flex items-center justify-between border-t border-border">
            <span className="pl-4 text-sm text-subtle">{number(products.data.total)} products</span>
            <Pagination page={page} hasNext={page * PAGE_SIZE < products.data.total} onChange={setPage} loading={products.isFetching} />
          </div>
        )}
      </Card>

      {editing && (
        <ProductDialog
          product={editing === "new" ? null : editing}
          categories={categories.data ?? []}
          onClose={() => setEditing(null)}
        />
      )}

      <ConfirmDialog
        open={!!deleting}
        onOpenChange={(open) => !open && setDeleting(null)}
        title="Delete product?"
        description={`"${deleting?.name}" will be removed from the catalogue. Existing orders keep their history.`}
        confirmLabel="Delete product"
        destructive
        loading={remove.isPending}
        onConfirm={() => deleting && remove.mutate(deleting.id)}
      />
    </div>
  );
}

function ProductDialog({
  product,
  categories,
  onClose,
}: {
  product: ProductListItem | null;
  categories: { id: string; name: string }[];
  onClose: () => void;
}) {
  const queryClient = useQueryClient();
  const creating = !product;
  // The list item has no description; load the full product when editing.
  const details = useQuery({ queryKey: ["product", product?.id], queryFn: () => productsApi.get(product!.id), enabled: !creating });
  const [form, setForm] = useState<FormState>(() =>
    product
      ? {
          ...EMPTY,
          name: product.name,
          sku: product.sku,
          price: String(product.price),
          mrp: product.mrp ? String(product.mrp) : "",
          brand: product.brand,
          image_url: product.image_url,
          category_id: product.category_id,
        }
      : { ...EMPTY, category_id: categories[0]?.id ?? "" },
  );
  const [descriptionLoaded, setDescriptionLoaded] = useState(creating);
  const [touched, setTouched] = useState(false);

  // The list item has no description or gallery; load them when editing.
  if (!descriptionLoaded && details.data) {
    setForm((current) => ({ ...current, description: details.data.description, images: details.data.images.join("\n") }));
    setDescriptionLoaded(true);
  }

  const errors = touched ? validate(form, creating) : {};

  const save = useMutation({
    mutationFn: async () => {
      const input: ProductInput = {
        name: form.name.trim(),
        description: form.description.trim(),
        sku: form.sku.trim(),
        price: Math.round(Number(form.price) * 100) / 100,
        category_id: form.category_id,
        brand: form.brand.trim(),
        mrp: form.mrp ? Math.round(Number(form.mrp) * 100) / 100 : 0,
        image_url: form.image_url.trim() || imageLines(form.images)[0] || "",
        images: imageLines(form.images),
      };
      if (!creating) return productsApi.update(product.id, input);
      const created = await productsApi.create(input);
      // Products are only orderable once they have an inventory record.
      await inventoryApi.create(created.id, Number(form.initial_stock));
      return created;
    },
    onSuccess: async () => {
      toast.success(creating ? "Product created" : "Product updated");
      await queryClient.invalidateQueries();
      onClose();
    },
    onError: (error) => toast.error(creating ? "Could not create product" : "Could not update product", { description: errorMessage(error) }),
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    setTouched(true);
    if (Object.keys(validate(form, creating)).length === 0) save.mutate();
  }

  const set = (key: keyof FormState) => (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) =>
    setForm((current) => ({ ...current, [key]: event.target.value }));

  return (
    <Dialog
      open
      onOpenChange={(open) => !open && onClose()}
      title={creating ? "New product" : "Edit product"}
      description={creating ? "The product is added to the storefront with its starting stock." : product.sku}
      footer={
        <>
          <Button variant="secondary" onClick={onClose} disabled={save.isPending}>
            Cancel
          </Button>
          <Button type="submit" form="product-form" loading={save.isPending}>
            {creating ? "Create product" : "Save changes"}
          </Button>
        </>
      }
    >
      <form id="product-form" onSubmit={submit} className="space-y-4" noValidate>
        <Field label="Name" htmlFor="p-name" error={errors.name}>
          <Input id="p-name" value={form.name} onChange={set("name")} aria-invalid={!!errors.name} autoFocus />
        </Field>
        <Field label="Description" htmlFor="p-description" error={errors.description} hint={`${form.description.length}/1000`}>
          <Textarea id="p-description" value={form.description} onChange={set("description")} disabled={!descriptionLoaded} />
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="SKU" htmlFor="p-sku" error={errors.sku}>
            <Input id="p-sku" value={form.sku} onChange={set("sku")} aria-invalid={!!errors.sku} />
          </Field>
          <Field label="Brand" htmlFor="p-brand" error={errors.brand}>
            <Input id="p-brand" value={form.brand} onChange={set("brand")} aria-invalid={!!errors.brand} />
          </Field>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Price (₹)" htmlFor="p-price" error={errors.price}>
            <Input id="p-price" type="number" min="0.01" step="0.01" value={form.price} onChange={set("price")} aria-invalid={!!errors.price} />
          </Field>
          <Field label="MRP (₹)" htmlFor="p-mrp" error={errors.mrp} hint="List price, shown struck through. Optional.">
            <Input id="p-mrp" type="number" min="0" step="0.01" value={form.mrp} onChange={set("mrp")} aria-invalid={!!errors.mrp} />
          </Field>
        </div>
        <Field label="Main image URL" htmlFor="p-image" error={errors.image_url} hint="Defaults to the first gallery image.">
          <Input id="p-image" value={form.image_url} onChange={set("image_url")} placeholder="https://" />
        </Field>
        <Field label="Gallery image URLs" htmlFor="p-images" error={errors.images} hint="One URL per line, up to 10.">
          <Textarea id="p-images" value={form.images} onChange={set("images")} disabled={!descriptionLoaded} className="min-h-20 font-mono text-xs" />
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Category" htmlFor="p-category" error={errors.category_id}>
            <Select id="p-category" value={form.category_id} onChange={set("category_id")}>
              <option value="" disabled>
                Choose…
              </option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </Select>
          </Field>
          {creating && (
            <Field label="Starting stock" htmlFor="p-stock" error={errors.initial_stock}>
              <Input id="p-stock" type="number" min="0" step="1" value={form.initial_stock} onChange={set("initial_stock")} />
            </Field>
          )}
        </div>
      </form>
    </Dialog>
  );
}
