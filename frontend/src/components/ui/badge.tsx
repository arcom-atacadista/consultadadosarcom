import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/cn";

const badgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-bold uppercase tracking-wider",
  {
    variants: {
      variant: {
        default: "bg-surface text-arcom-gray border border-surface-border",
        brand: "bg-verde-arcom text-white",
        accent: "bg-verde-lima text-verde-escuro",
        danger: "bg-danger text-white",
        outline: "border border-surface-border text-verde-escuro",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant, className }))} {...props} />;
}

export { Badge, badgeVariants };
