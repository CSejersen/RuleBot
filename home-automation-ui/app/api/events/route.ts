"use server";
import { NextResponse } from 'next/server';
import pool from '@/lib/db';
import { parseJsonOrFallback } from '@/lib/utils';

// Helper to convert MySQL DATETIME to ISO string (treats DATETIME as UTC)
function formatTimestamp(timestamp: any): string {
  if (!timestamp) return new Date().toISOString();
  
  // If it's already a Date object or ISO string with Z, use it directly
  if (timestamp instanceof Date) {
    return timestamp.toISOString();
  }
  
  // If it's already an ISO string with timezone info, use it
  if (typeof timestamp === 'string' && (timestamp.includes('T') || timestamp.includes('Z') || timestamp.includes('+'))) {
    return new Date(timestamp).toISOString();
  }
  
  // MySQL DATETIME format: "YYYY-MM-DD HH:mm:ss" - treat as UTC
  // Append 'Z' to indicate UTC, then convert to ISO
  const timestampStr = String(timestamp);
  const utcString = timestampStr.includes('T') ? timestampStr : timestampStr.replace(' ', 'T') + 'Z';
  return new Date(utcString).toISOString();
}

export async function GET(request: Request) {
  try {
    // Get hours parameter from query string (default to 24 hours)
    const { searchParams } = new URL(request.url)
    const hours = parseInt(searchParams.get("hours") || "24", 10)
    const validHours = isNaN(hours) || hours <= 0 || hours > 168 ? 24 : hours

    // LEFT JOIN to get context metadata too
    const [rows] = await pool.query(`
      SELECT 
        e.id,
        e.type,
        e.data,
        e.context_id,
        e.time_fired,
        e.created_at,
        e.updated_at,
        c.id AS context_id_full,
        c.parent_id AS context_parent_id,
        c.origin AS context_origin,
        c.created_at AS context_created_at
      FROM events e
      LEFT JOIN contexts c ON e.context_id = c.id
      WHERE e.time_fired >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)
      ORDER BY e.time_fired DESC
    `, [validHours]);
    // Only ensure output is JSON-serializable; no TS casting
    const events = (rows as any[]).map(row => {
      const parsedData = parseJsonOrFallback(row.data) || {};
      return {
        id: row.id, // ID is now a string (UUID), no need to convert
        type: row.type,
        data: parsedData,
        context_id: row.context_id,
        time_fired: formatTimestamp(row.time_fired),
        context: row.context_id_full
          ? {
            id: row.context_id_full,
            parent_id: row.context_parent_id || undefined,
            origin: row.context_origin || undefined,
            created_at: row.context_created_at
              ? formatTimestamp(row.context_created_at)
              : undefined,
          }
          : undefined,
        created_at: row.created_at
          ? formatTimestamp(row.created_at)
          : undefined,
        updated_at: row.updated_at
          ? formatTimestamp(row.updated_at)
          : undefined,
      };
    });
    return NextResponse.json(events);
  } catch (error) {
    console.error('Database error:', error);
    return NextResponse.json({ error: 'Failed to fetch events' }, { status: 500 });
  }
}
