import type { LucideIcon } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export function EmConstrucao({
  icon: Icon,
  titulo,
  descricao,
  fase,
}: {
  icon: LucideIcon;
  titulo: string;
  descricao: string;
  fase: string;
}) {
  return (
    <Card className="max-w-xl">
      <CardHeader>
        <div className="flex h-12 w-12 items-center justify-center rounded-md bg-surface">
          <Icon className="h-6 w-6 text-verde-arcom" strokeWidth={2} />
        </div>
        <CardTitle>{titulo}</CardTitle>
        <CardDescription>{descricao}</CardDescription>
        <Badge variant="outline" className="mt-2 w-fit">{fase}</Badge>
      </CardHeader>
    </Card>
  );
}
