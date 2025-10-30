"use server";

import { NextResponse } from "next/server";

export async function GET(
  req: Request,
  { params }: { params: Promise<{ source: string }> }
) {
  try {
    const { source } = await params;

    const response = await fetch(
      `http://localhost:8080/api/integrations/${source}/event-types`,
      { cache: "no-store" }
    );

    if (!response.ok) {
      return NextResponse.json(
        { error: `Failed to fetch events for ${source}` },
        { status: response.status }
      );
    }

    const data = await response.json();

    const events =
      Array.isArray(data.events) && data.events.length
        ? data.events.map((event: any) => ({
          type: event.type,
          entities: event.entities || [],
          stateChanges: event.state_changes || [],
        }))
        : [];

    return NextResponse.json({ events });
  } catch (err) {
    console.error("Error fetching event types:", err);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}
