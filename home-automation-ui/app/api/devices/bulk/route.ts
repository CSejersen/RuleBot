"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import { validateRequest } from "@/lib/validate";
import { DeviceBulkActionSchema } from "@/types/device/device-bulk-schema";

export async function POST(req: Request) {
  try {
    const raw = await req.json();
    const validation = validateRequest(DeviceBulkActionSchema, raw);
    if (!validation.success) {
      return NextResponse.json(validation.error, { status: 400 });
    }
    const { action, ids } = validation.data;

    const placeholders = ids.map(() => "?").join(",");

    if (action === "enable" || action === "disable") {
      const enabled = action === "enable" ? 1 : 0;

      // For disable, also disable all underlying entities of those devices
      if (action === "disable") {
        const conn = await pool.getConnection();
        try {
          await conn.beginTransaction();
          await conn.query(`UPDATE devices SET enabled = ?, updated_at = NOW() WHERE id IN (${placeholders})`, [enabled, ...ids]);
          await conn.query(`UPDATE entities SET enabled = 0, updated_at = NOW() WHERE device_id IN (${placeholders})`, ids);
          await conn.commit();
          conn.release();
          return NextResponse.json({ success: true });
        } catch (err) {
          try { await conn.rollback(); } catch {}
          conn.release();
          throw err;
        }
      }

      // Default path for enable (only devices)
      await pool.query(`UPDATE devices SET enabled = ?, updated_at = NOW() WHERE id IN (${placeholders})`, [enabled, ...ids]);
      return NextResponse.json({ success: true });
    }

    if (action === "delete") {
      const conn = await pool.getConnection();
      try {
        await conn.beginTransaction();
        // Delete dependent entities first (to satisfy foreign key constraints)
        await conn.query(`DELETE FROM entities WHERE device_id IN (${placeholders})`, ids);
        // Finally delete devices
        await conn.query(`DELETE FROM devices WHERE id IN (${placeholders})`, ids);
        await conn.commit();
        conn.release();
        return NextResponse.json({ success: true });
      } catch (err) {
        try { await conn.rollback(); } catch {}
        conn.release();
        throw err;
      }
    }

    return NextResponse.json({ error: "Unsupported action" }, { status: 400 });
  } catch (error) {
    console.error("Bulk devices action failed:", error);
    return NextResponse.json({ error: "Failed to process bulk devices action" }, { status: 500 });
  }
}
