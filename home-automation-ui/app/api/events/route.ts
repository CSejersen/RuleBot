import { NextResponse } from 'next/server';
import pool from '@/lib/db';

export type Event = {
  id: string;
  source: string;
  type: string;
  entity: string;
  stateChange: string;
  payload: Record<string, any>;
  timestamp: string | null;
  triggeredRules: string[];
};

export async function GET() {
  try {
    const [rows] = await pool.query('SELECT * FROM events ORDER BY timestamp DESC LIMIT 100');

    const events: Event[] = (rows as any[]).map(row => ({
      id: row.id.toString(),
      source: row.source,
      type: row.type,
      entity: row.entity,
      stateChange: row.state_change,
      payload: typeof row.payload === 'string' ? JSON.parse(row.payload) : row.payload,
      timestamp: row.timestamp
        ? new Date(row.timestamp + 'Z').toISOString()
        : null,
      triggeredRules: (() => {
        try {
          const parsed = typeof row.triggered_rules === 'string'
            ? JSON.parse(row.triggered_rules)
            : row.triggered_rules;
          return Array.isArray(parsed) ? parsed.map(String) : [];
        } catch {
          return [];
        }
      })(),
    }));

    return NextResponse.json(events);
  } catch (error) {
    console.error('Database error:', error);
    return NextResponse.json({ error: 'Failed to fetch events' }, { status: 500 });
  }
}
