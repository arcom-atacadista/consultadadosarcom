import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { CheckCircle2, AlertTriangle, XCircle, Info } from "lucide-react";
import { cn } from "@/lib/cn";

const alertVariants = cva(
  "flex items-start gap-3 rounded-md border p-4 text-sm",
  {
    variants: {
      variant: {
        success: "bg-verde-arcom/10 border-verde-arcom/30 text-verde-escuro",
        danger: "bg-danger/10 border-danger/30 text-danger",
        warning: "bg-verde-lima/20 border-verde-lima/40 text-verde-escuro",
        info: "bg-surface border-surface-border text-arcom-gray",
        brand: "bg-verde-escuro text-white border-transparent",
      },
    },
    defaultVariants: { variant: "info" },
  },
);

const icons = {
  success: CheckCircle2,
  danger: XCircle,
  warning: AlertTriangle,
  info: Info,
  brand: Info,
} as const;

export interface AlertProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof alertVariants> {}

function Alert({ className, variant = "info", children, ...props }: AlertProps) {
  const Icon = icons[variant ?? "info"];
  return (
    <div role="alert" className={cn(alertVariants({ variant, className }))} {...props}>
      <Icon className="h-5 w-5 shrink-0 mt-0.5" strokeWidth={2} />
      <div>{children}</div>
    </div>
  );
}

export { Alert, alertVariants };
