import { Card, CardContent } from "@/components/ui/card";

interface SummaryStatCardProps {
  label: string;
  value: string | number;
  className?: string;
  compact?: boolean;
}

export function SummaryStatCard({ label, value, className, compact = false }: SummaryStatCardProps) {
  if (compact) {
    return (
      <Card className={className}>
        <CardContent className="p-2 flex flex-row items-center justify-between gap-2">
          <div className="flex flex-col gap-0.5">
            <span className="text-base font-semibold leading-tight">{value}</span>
            <span className="text-xs text-muted-foreground leading-tight">{label}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className={className}>
      <CardContent className="p-4 flex flex-col items-center">
        <span className="text-xl font-bold">{value}</span>
        <span className="text-sm text-muted-foreground">{label}</span>
      </CardContent>
    </Card>
  );
}
