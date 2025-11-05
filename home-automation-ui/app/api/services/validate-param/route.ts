"use server";

import { NextRequest, NextResponse } from "next/server";

const ENGINE_BASE_URL = "http://localhost:8080/api/services/validate-param";

export async function POST(req: NextRequest) {
  try {
    const body = await req.json();
    const { service, param_name, value } = body;

    if (!service || param_name === undefined || value === undefined) {
      return NextResponse.json(
        { error: "Missing required fields: service, param_name, value" },
        { status: 400 }
      );
    }

    const engineRes = await fetch(ENGINE_BASE_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ service, param_name, value }),
    });

    const data = await engineRes.json();

    if (!engineRes.ok) {
      return NextResponse.json(data, { status: engineRes.status });
    }

    return NextResponse.json(data);
  } catch (error: any) {
    console.error("Error validating param:", error);
    return NextResponse.json(
      { error: "Failed to validate param", details: error.message },
      { status: 500 }
    );
  }
}

