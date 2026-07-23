import * as React from "react";
import { cn } from "@/lib/cn";

export interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  error?: boolean;
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, error, ...props }, ref) => (
    <textarea
      ref={ref}
      className={cn(
        "flex min-h-24 w-full rounded-md border bg-white px-3 py-2 text-sm text-verde-escuro placeholder:text-arcom-gray transition-colors duration-fast focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-verde-arcom disabled:cursor-not-allowed disabled:opacity-50",
        error ? "border-danger" : "border-surface-border",
        className,
      )}
      {...props}
    />
  ),
);
Textarea.displayName = "Textarea";

export { Textarea };
