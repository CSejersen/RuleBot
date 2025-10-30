"use server";

import { NextResponse } from "next/server";
import type {
  ServiceParam,
  TargetType,
  EntityType,
  TargetSpec,
  ServiceSpec,
  ServicesResponse,
} from "@/types/services/services";

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
