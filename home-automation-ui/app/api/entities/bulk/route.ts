"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import { z } from "zod";
import { validateRequest } from "@/lib/validate";

const EntityBulkActionSchema = z.object({
  action: z.enum(["enable", "disable"]),
  ids: z.array(z.string().min(1)).min(1),
});

export async function POST(req: Request) {
  try {
    const data = await req.json();
    const validation = validateRequest(EntityBulkActionSchema, data);
    if (!validation.success) {
      return NextResponse.json(validation.error, { status: 400 });
    }
    const { action, ids } = validation.data;
    const enabled = action === "enable" ? 1 : 0;
    const placeholders = ids.map(() => "?").join(",");
    await pool.query(
      `UPDATE entities SET enabled = ?, updated_at = NOW() WHERE external_id IN (${placeholders})`,
      [enabled, ...ids]
    );
    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Bulk entities action failed:", error);
    return NextResponse.json({ error: "Failed to process bulk entities action" }, { status: 500 });
  }
}
