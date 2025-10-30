import { Card, CardContent } from "@/components/ui/card";

interface SummaryStatCardProps {
  label: string;
  value: string | number;
  className?: string;
}

export function SummaryStatCard({ label, value, className }: SummaryStatCardProps) {
  return (
    <Card className={className}>
      <CardContent className="p-4 flex flex-col items-center">
        <span className="text-xl font-bold">{value}</span>
        <span className="text-sm text-muted-foreground">{label}</span>
      </CardContent>
    </Card>
  );
}
