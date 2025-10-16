import { NextResponse } from "next/server";

export interface ServiceParam {
  DataType: string;
  Description: string;
}

export interface ServiceData {
  name: string;
  required_params: Record<string, ServiceParam>;
  requires_target_type: boolean;
  requires_target_id: boolean;
}

export interface ServicesResponse {
  services: ServiceData[];
  error?: string;
}

export async function GET(req: Request) {
  try {
    const res = await fetch("http://localhost:8080/api/services", { cache: "no-store" });

    if (!res.ok) {
      return NextResponse.json(
        { services: [], error: `Failed to fetch services: ${res.status}` },
        { status: res.status }
      );
    }

    const data: ServicesResponse = await res.json();
    return NextResponse.json(data);
  } catch (err) {
    console.error("Error fetching services:", err);
    return NextResponse.json(
      { services: [], error: "Internal server error" },
      { status: 500 }
    );
  }
}
