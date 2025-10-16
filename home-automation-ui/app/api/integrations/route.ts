import { NextResponse } from "next/server";

export async function GET() {
  try {
    const res = await fetch("http://localhost:8080/api/integrations");
    if (!res.ok) {
      return NextResponse.json({ error: "Failed to fetch integrations" }, { status: res.status });
    }

    const data = await res.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching integrations:", error);
    return NextResponse.json({ error: "Internal server error" }, { status: 500 });
  }
}
