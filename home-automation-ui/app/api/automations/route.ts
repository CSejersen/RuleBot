"use server";

import { NextResponse } from 'next/server';
import pool from '@/lib/db';
import { parseJsonOrFallback } from '@/lib/utils';
import { validateRequest } from '@/lib/validate';
import { CreateAutomationSchema } from '@/types/automation/automation-schema';

export async function GET() {
  try {
    const [rows] = await pool.query(`
      SELECT 
        id,
        alias,
        description,
        triggers,
        conditions,
        actions,
        enabled,
        last_triggered,
        created_at,
        updated_at
      FROM automations
      ORDER BY id ASC
    `);

    const automations = (rows as any[]).map(row => ({
      id: row.id,
      alias: row.alias,
      description: row.description,
      triggers: parseJsonOrFallback(row.triggers) || [],
      conditions: parseJsonOrFallback(row.conditions) || [],
      actions: parseJsonOrFallback(row.actions) || [],
      enabled: row.enabled === 1,
      last_triggered: row.last_triggered,
      created_at: row.created_at,
      updated_at: row.updated_at,
    }));

    return NextResponse.json(automations, { status: 200 });
  } catch (err: any) {
    console.error('Error fetching automations:', err);
    return NextResponse.json(
      { error: 'Failed to fetch automations', details: err.message },
      { status: 500 }
    );
  }
}

export async function POST(req: Request) {
  try {
    const rawData = await req.json();
    const validation = validateRequest(CreateAutomationSchema, rawData);
    if (!validation.success) {
      return NextResponse.json(validation.error, { status: 400 });
    }
    // zod schema guarantees these properties:
    const { alias, description, triggers, actions, conditions, enabled } = validation.data;

    // Convert JSON fields to strings for MySQL storage
    const triggersStr = JSON.stringify(triggers);
    const actionsStr = JSON.stringify(actions);
    const conditionsStr = JSON.stringify(conditions || []);

    const [result] = await pool.query(
      `
      INSERT INTO automations 
        (alias, description, triggers, conditions, actions, enabled, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
      `,
      [alias, description, triggersStr, conditionsStr, actionsStr, enabled ? 1 : 0]
    );

    const insertId = (result as any).insertId;
    const now = new Date().toISOString();
    const newAutomation = {
      id: insertId,
      alias,
      description,
      triggers,
      conditions: conditions || [],
      actions,
      enabled: enabled ?? true,
      last_triggered: null,
      created_at: now,
      updated_at: now,
    };

    return NextResponse.json({ automation: newAutomation }, { status: 201 });
  } catch (err: any) {
    console.error('Error creating automation:', err);
    return NextResponse.json(
      { error: 'Failed to create automation', details: err.message },
      { status: 500 }
    );
  }
}
