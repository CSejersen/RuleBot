"use server";
import { NextResponse } from "next/server";


interface IntegrationConfig {
  id: string;
  integration_name: string;
  display_name: string;
  enabled: boolean;
}

export async function GET() {
  try {
    const res = await fetch("http://localhost:8080/api/integrations");
    if (!res.ok) {
      return NextResponse.json({ error: "Failed to fetch integration" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching integration:", error);
    return NextResponse.json({ error: "Internal server error" }, { status: 500 });
  }
}
