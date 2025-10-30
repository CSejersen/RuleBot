import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { capitalize } from "@/lib/utils";

interface StateLike {
  last_updated?: string;
  attributes?: Record<string, any>;
}

interface EntityLike {
  name: string;
  entity_id: string;
  type: string;
  available: boolean;
  enabled: boolean;
  state?: StateLike;
}

interface EntityDetailsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  entity: EntityLike | null;
}

export function EntityDetailsDialog({ open, onOpenChange, entity }: EntityDetailsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{entity?.name ?? "Entity details"}</DialogTitle>
        </DialogHeader>
        {entity && (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant="secondary">{entity.entity_id}</Badge>
              <Badge variant="secondary">{capitalize(entity.type)}</Badge>
              <Badge variant={entity.available ? "default" : "destructive"}>{entity.available ? "Online" : "Offline"}</Badge>
              {!entity.enabled && <Badge variant="secondary">Disabled</Badge>}
            </div>
            <div className="bg-card/50 rounded-lg overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Attribute</TableHead>
                    <TableHead>Value</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {entity.state?.attributes && Object.keys(entity.state.attributes).length > 0 ? (
                    Object.entries(entity.state.attributes).map(([key, value]) => (
                      <TableRow key={key}>
                        <TableCell className="font-medium">{key}</TableCell>
                        <TableCell>{typeof value === 'object' ? JSON.stringify(value) : String(value)}</TableCell>
                      </TableRow>
                    ))
                  ) : (
                    <TableRow>
                      <TableCell colSpan={2} className="text-muted-foreground text-center py-4">No attributes</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
