"use server";
import { NextResponse } from "next/server";
import { validateRequest } from "@/lib/validate";
import { IntegrationDiscoverParamsSchema } from "@/types/integration/integration-discover-schema";

export async function POST(
  req: Request,
  { params }: { params: Promise<{ name: string }> }
) {
  // Validate params
  const resolvedParams = await params;
  const validation = validateRequest(IntegrationDiscoverParamsSchema, resolvedParams);
  if (!validation.success) {
    return NextResponse.json(validation.error, { status: 400 });
  }
  const { name } = validation.data;

  try {
    const res = await fetch(`http://localhost:8080/api/integrations/configs/${name}/discover`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
    });

    if (!res.ok) {
      const errorText = await res.text();
      console.error("Engine discovery error:", errorText);
      return NextResponse.json({ error: "Failed to trigger discovery" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (err) {
    console.error("Error proxying discovery request:", err);
    return NextResponse.json({ error: "Internal server error" }, { status: 500 });
  }
}
