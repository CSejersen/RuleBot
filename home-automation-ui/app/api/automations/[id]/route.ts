"use server";
import { NextResponse } from 'next/server';
import pool from '@/lib/db';
import { validateRequest } from '@/lib/validate';
import { CreateAutomationSchema } from '@/types/automation/automation-schema';

export async function PUT(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!id) {
    return NextResponse.json({ error: 'Automation ID is required' }, { status: 400 });
  }
  try {
    const rawData = await req.json();
    const validation = validateRequest(CreateAutomationSchema, rawData);
    if (!validation.success) {
      return NextResponse.json(validation.error, { status: 400 });
    }
    const { alias, description, triggers, actions, conditions, enabled } = validation.data;
    const triggersStr = JSON.stringify(triggers);
    const actionsStr = JSON.stringify(actions);
    const conditionsStr = JSON.stringify(conditions || []);
    const [updateResult] = await pool.query(
      `
      UPDATE automations 
      SET alias = ?, description = ?, triggers = ?, conditions = ?, actions = ?, enabled = ?, updated_at = NOW()
      WHERE id = ?
      `,
      [alias, description, triggersStr, conditionsStr, actionsStr, enabled ? 1 : 0, id]
    );
    const affectedRows = (updateResult as any).affectedRows;
    if (affectedRows === 0) {
      return NextResponse.json({ error: 'Automation not found' }, { status: 404 });
    }
    const [rows] = await pool.query(
      `SELECT * FROM automations WHERE id = ?`,
      [id]
    );
    if ((rows as any[]).length === 0) {
      return NextResponse.json({ error: 'Automation not found after update' }, { status: 404 });
    }
    const row = (rows as any[])[0];
    const updatedAutomation = {
      id: row.id,
      alias: row.alias,
      description: row.description,
      triggers: row.triggers ? JSON.parse(row.triggers) : [],
      conditions: row.conditions ? JSON.parse(row.conditions) : [],
      actions: row.actions ? JSON.parse(row.actions) : [],
      enabled: row.enabled === 1,
      last_triggered: row.last_triggered,
      created_at: row.created_at,
      updated_at: row.updated_at,
    };
    return NextResponse.json({ automation: updatedAutomation }, { status: 200 });
  } catch (err: any) {
    console.error('Error updating automation:', err);
    return NextResponse.json(
      { error: 'Failed to update automation', details: err.message },
      { status: 500 }
    );
  }
}

export async function DELETE(req: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!id) {
    return NextResponse.json({ error: 'Automation ID is required' }, { status: 400 });
  }
  try {
    const [result] = await pool.query(
      `DELETE FROM automations WHERE id = ?`,
      [id]
    );
    const affectedRows = (result as any).affectedRows;
    if (affectedRows === 0) {
      return NextResponse.json({ error: 'Automation not found' }, { status: 404 });
    }
    return NextResponse.json({ message: 'Automation deleted successfully' }, { status: 200 });
  } catch (err: any) {
    console.error('Error deleting automation:', err);
    return NextResponse.json(
      { error: 'Failed to delete automation', details: err.message },
      { status: 500 }
    );
  }
}
