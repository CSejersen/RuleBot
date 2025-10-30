import { Button } from "@/components/ui/button";

interface BulkActionBarProps {
  selectedCount: number;
  onEnable: () => void;
  onDisable: () => void;
  onDelete?: () => void;
  className?: string;
}

export function BulkActionBar({ selectedCount, onEnable, onDisable, onDelete, className }: BulkActionBarProps) {
  if (selectedCount <= 0) return null;
  return (
    <div className={"ml-auto flex items-center gap-2 " + (className ?? "") }>
      <span className="text-sm text-muted-foreground hidden md:inline">{selectedCount} selected</span>
      <Button variant="outline" size="sm" onClick={onEnable}>Enable</Button>
      <Button variant="outline" size="sm" onClick={onDisable}>Disable</Button>
      {onDelete && <Button variant="destructive" size="sm" onClick={onDelete}>Delete</Button>}
    </div>
  );
}
