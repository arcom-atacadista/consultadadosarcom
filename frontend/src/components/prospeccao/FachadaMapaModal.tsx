import { useEffect, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Alert } from "@/components/ui/alert";
import { api } from "@/lib/api";
import type { Prospect } from "@/lib/prospeccao";

type GeoResultado = { lat: number; lng: number; preciso: boolean } | null;

function streetViewUrl(lat: number, lng: number) {
  return `https://maps.google.com/maps?q=&layer=c&cbll=${lat},${lng}&cbp=12,0,0,0,0&output=svembed`;
}

function mapaUrl(endereco: string) {
  return `https://maps.google.com/maps?q=${encodeURIComponent(endereco)}&t=&z=17&output=embed`;
}

export function FachadaMapaModal({ prospect, onClose }: { prospect: Prospect | null; onClose: () => void }) {
  const [aba, setAba] = useState<"fachada" | "mapa">("fachada");
  const [geo, setGeo] = useState<GeoResultado>(null);
  const [carregando, setCarregando] = useState(false);
  const [semFachada, setSemFachada] = useState(false);

  useEffect(() => {
    if (!prospect) return;
    setAba("fachada");
    setGeo(null);
    setSemFachada(false);
    if (!prospect.endereco || prospect.endereco === "—") {
      setSemFachada(true);
      return;
    }
    setCarregando(true);
    const query = `${prospect.endereco}, ${prospect.cep ?? ""}`.trim();
    api
      .get("/geo/geocode", { params: { q: query } })
      .then(({ data }) => {
        if (data) {
          setGeo(data);
        } else if (prospect.latitude != null && prospect.longitude != null) {
          setGeo({ lat: prospect.latitude, lng: prospect.longitude, preciso: false });
        } else {
          setSemFachada(true);
        }
      })
      .catch(() => {
        if (prospect.latitude != null && prospect.longitude != null) {
          setGeo({ lat: prospect.latitude, lng: prospect.longitude, preciso: false });
        } else {
          setSemFachada(true);
        }
      })
      .finally(() => setCarregando(false));
  }, [prospect]);

  if (!prospect) return null;

  return (
    <Dialog open={!!prospect} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{prospect.razao}</DialogTitle>
          <DialogDescription>{prospect.endereco || "Endereço não informado"}</DialogDescription>
        </DialogHeader>
        <Tabs value={aba} onValueChange={(v) => setAba(v as "fachada" | "mapa")}>
          <TabsList>
            <TabsTrigger value="fachada">Fachada</TabsTrigger>
            <TabsTrigger value="mapa">Mapa</TabsTrigger>
          </TabsList>
          <TabsContent value="fachada">
            {carregando && <p className="text-sm text-arcom-gray">Carregando a fachada…</p>}
            {!carregando && semFachada && (
              <Alert variant="info">Sem coordenada disponível para mostrar a fachada deste endereço.</Alert>
            )}
            {!carregando && geo && (
              <>
                {!geo.preciso && (
                  <Alert variant="warning" className="mb-2">
                    Coordenada aproximada — pode não bater exatamente com a fachada.
                  </Alert>
                )}
                <iframe
                  title="Street View"
                  src={streetViewUrl(geo.lat, geo.lng)}
                  className="w-full h-80 rounded-md border border-surface-border"
                />
              </>
            )}
          </TabsContent>
          <TabsContent value="mapa">
            <iframe
              title="Mapa"
              src={mapaUrl(`${prospect.endereco}, ${prospect.cidade} - ${prospect.uf}`)}
              className="w-full h-80 rounded-md border border-surface-border"
            />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}
