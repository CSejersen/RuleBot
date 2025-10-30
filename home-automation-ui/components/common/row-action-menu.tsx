import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator } from "@/components/ui/dropdown-menu";
import { MoreVertical } from "lucide-react";

interface RowActionMenuProps {
  onEnable?: () => void;
  onDisable?: () => void;
  onDetails?: () => void;
  onEdit?: () => void;
  onDelete?: () => void;
}

export function RowActionMenu({ onEnable, onDisable, onDetails, onEdit, onDelete }: RowActionMenuProps) {
  const showTop = Boolean(onEnable || onDisable);
  const showBottom = Boolean(onDetails || onEdit || onDelete);
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Actions"
          onClick={(e) => {
            // Prevent table row onClick from firing when opening the menu
            e.stopPropagation();
          }}
        >
          <MoreVertical className="w-4 h-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {onEnable && <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onEnable(); }}>Enable</DropdownMenuItem>}
        {onDisable && <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onDisable(); }}>Disable</DropdownMenuItem>}
        {showTop && showBottom ? <DropdownMenuSeparator /> : null}
        {onDetails && <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onDetails(); }}>View details</DropdownMenuItem>}
        {onEdit && <DropdownMenuItem onClick={(e) => { e.stopPropagation(); onEdit(); }}>Edit</DropdownMenuItem>}
        {onDelete && <DropdownMenuItem data-variant="destructive" onClick={(e) => { e.stopPropagation(); onDelete(); }}>Delete</DropdownMenuItem>}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
