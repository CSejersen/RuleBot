"use server";
import { NextResponse } from "next/server";
import type { State } from "@/types/state";

export async function GET(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;

  try {
    const engineRes = await fetch(`http://localhost:8080/api/devices/${id}/states`);
    if (!engineRes.ok) {
      throw new Error(`Engine API returned status ${engineRes.status}`);
    }

    const engineData = await engineRes.json();

    const states: State[] = (engineData.states ?? []).map((s: any) => ({
      entity_id: s.entity_id,
      state: s.state,
      attributes: s.attributes ?? {},
      last_changed: s.last_changed,
      last_updated: s.last_updated,
      context: s.context
        ? {
          id: s.context.id,
          parent_id: s.context.parent_id ?? undefined,
        }
        : undefined,
    }));

    return NextResponse.json({ states });
  } catch (error) {
    console.error(`Failed to fetch states for device ${id}:`, error);
    return NextResponse.json(
      { error: "Failed to fetch states from engine" },
      { status: 500 }
    );
  }
}
