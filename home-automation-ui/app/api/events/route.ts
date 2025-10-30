"use server";
import { NextResponse } from 'next/server';
import pool from '@/lib/db';

export async function GET() {
  try {
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
        c.created_at AS context_created_at
      FROM events e
      LEFT JOIN contexts c ON e.context_id = c.id
      ORDER BY e.time_fired DESC
      LIMIT 300
    `);
    // Only ensure output is JSON-serializable; no TS casting
    const events = (rows as any[]).map(row => {
      let parsedData = {};
      try {
        parsedData = typeof row.data === 'string' ? JSON.parse(row.data) : row.data;
      } catch (err) {
        console.warn(`Failed to parse event data for event ${row.id}:`, err);
      }
      return {
        id: row.id.toString(),
        type: row.type,
        data: parsedData,
        context_id: row.context_id,
        time_fired: new Date(row.time_fired + 'Z').toISOString(),
        context: row.context_id_full
          ? {
            id: row.context_id_full,
            parent_id: row.context_parent_id || undefined,
            created_at: row.context_created_at
              ? new Date(row.context_created_at + 'Z').toISOString()
              : undefined,
          }
          : undefined,
        created_at: row.created_at
          ? new Date(row.created_at + 'Z').toISOString()
          : undefined,
        updated_at: row.updated_at
          ? new Date(row.updated_at + 'Z').toISOString()
          : undefined,
      };
    });
    return NextResponse.json(events);
  } catch (error) {
    console.error('Database error:', error);
    return NextResponse.json({ error: 'Failed to fetch events' }, { status: 500 });
  }
}
