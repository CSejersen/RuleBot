import { useState } from "react";
import { Button } from "@/components/ui/button";
import { RefreshCw } from "lucide-react";

export function DiscoveryButton({ integrationName }: { integrationName: string }) {
  const [loading, setLoading] = useState(false);

  const handleScan = async () => {
    setLoading(true);
    try {
      const res = await fetch(`/api/integrations/configs/${integrationName}/discover`, {
        method: "POST",
      });

      if (!res.ok) {
        console.error("Failed to scan devices");
        return;
      }

      const data = await res.json();
      console.log("Discovery result:", data);
      // Optionally show a toast or notification here
    } catch (err) {
      console.error("Error scanning devices:", err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button size="sm" variant="outline" onClick={handleScan} disabled={loading}>
      <RefreshCw className="w-4 h-4 mr-1 animate-spin={loading}" />
      {loading ? "Scanning..." : "Scan For Devices"}
    </Button>
  );
}
