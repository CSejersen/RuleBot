import { NextResponse } from 'next/server';
import pool from '@/lib/db';
import { Rule } from "@/app/rules/types/rule"

export async function GET() {
  try {
    const [rows] = await pool.query('SELECT * FROM rules');

    const rules: Rule[] = (rows as any[]).map(row => ({
      alias: row.alias,
      trigger: (() => {
        try {
          const parsed = typeof row.trigger_json === 'string' ? JSON.parse(row.trigger_json) : row.trigger_json;
          return {
            event: parsed.event,
            entityName: parsed.entity_name ? String(parsed.entity_name) : undefined,
            stateChange: parsed.state_change,
          };
        } catch {
          return { event: '', stateChange: '' };
        }
      })(),
      condition: (() => {
        try {
          const parsed = typeof row.condition_json === 'string' ? JSON.parse(row.condition_json) : row.condition_json;
          return Array.isArray(parsed) ? parsed : [];
        } catch {
          return [];
        }
      })(),
      action: (() => {
        try {
          const parsed = typeof row.action_json === 'string' ? JSON.parse(row.action_json) : row.action_json;
          return Array.isArray(parsed) ? parsed : [];
        } catch {
          return [];
        }
      })(),
      active: row.active,
      lastTriggered: row.last_triggered
        ? new Date(row.last_triggered).toISOString()
        : null,
    }));

    return NextResponse.json(rules);
  } catch (error) {
    console.error('Database error:', error);
    return NextResponse.json({ error: 'Failed to fetch rules' }, { status: 500 });
  }
}

export async function POST(req: Request) {
  try {
    // Parse the incoming JSON
    const body: Rule = await req.json();

    // Validate minimal required fields
    if (!body.alias || !body.trigger?.event) {
      return NextResponse.json({ error: "Missing required fields: alias or trigger.event" }, { status: 400 });
    }

    // Serialize nested objects
    const triggerJson = JSON.stringify({
      event: body.trigger.event,
      entity_name: body.trigger.entityName || null,
      state_change: body.trigger.stateChange,
    });

    const conditionJson = JSON.stringify(body.condition || []);
    const actionJson = JSON.stringify(body.action || []);

    // Insert into database
    await pool.query(
      'INSERT INTO rules (`alias`, `trigger_json`, `condition_json`, `action_json`, `active`, `last_triggered`) VALUES (?, ?, ?, ?, ?, ?)',
      [
        body.alias,
        triggerJson,
        conditionJson,
        actionJson,
        body.active ?? true,
        body.lastTriggered ? new Date(body.lastTriggered) : null,
      ]
    );

    // Return success with inserted rule
    return NextResponse.json({ success: true, rule: body });
  } catch (error) {
    console.error('Database insert error:', error);
    return NextResponse.json({ error: 'Failed to save rule' }, { status: 500 });
  }
}
