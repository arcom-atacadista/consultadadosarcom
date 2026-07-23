import * as React from "react";
import { cn } from "@/lib/cn";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: boolean;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, error, ...props }, ref) => (
    <input
      ref={ref}
      className={cn(
        "flex h-10 w-full rounded-md border bg-white px-3 py-2 text-sm text-verde-escuro placeholder:text-arcom-gray transition-colors duration-fast focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-verde-arcom disabled:cursor-not-allowed disabled:opacity-50",
        error ? "border-danger" : "border-surface-border",
        className,
      )}
      {...props}
    />
  ),
);
Input.displayName = "Input";

function FormField({
  label,
  hint,
  error,
  children,
}: {
  label?: string;
  hint?: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      {label && (
        <label className="text-xs font-bold uppercase tracking-wider text-arcom-gray">
          {label}
        </label>
      )}
      {children}
      {error ? (
        <p className="text-xs text-danger">{error}</p>
      ) : hint ? (
        <p className="text-xs text-arcom-gray">{hint}</p>
      ) : null}
    </div>
  );
}

export { Input, FormField };
