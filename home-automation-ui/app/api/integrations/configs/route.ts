"use server";
import { NextResponse } from "next/server";
import pool from "@/lib/db";
import { IntegrationConfig } from "@/types/integration/integration-config";
import { validateRequest } from "@/lib/validate";
import { IntegrationConfigSchema } from "@/types/integration/integration-config-schema";

export async function GET() {
  try {
    const [rows] = await pool.query(
      `SELECT 
         id,
         integration_name,
         display_name,
         user_config,
         enabled,
         created_at,
         updated_at
       FROM integration_configs
       ORDER BY created_at DESC`
    );

    const configs = (rows as any[]).map((row) => ({
      id: row.id,
      integration_name: row.integration_name,
      display_name: row.display_name,
      user_config: typeof row.user_config === "string" ? JSON.parse(row.user_config) : row.user_config,
      enabled: !!row.enabled,
      created_at: row.created_at,
      updated_at: row.updated_at,
    })) as IntegrationConfig[];

    return NextResponse.json(configs);
  } catch (error) {
    console.error("Error fetching integration configs:", error);
    return NextResponse.json({ error: "Failed to fetch configs" }, { status: 500 });
  }
}

export async function POST(req: Request) {
  try {
    const rawData = await req.json();
    const validation = validateRequest(IntegrationConfigSchema, rawData);
    if (!validation.success) {
      return NextResponse.json(validation.error, { status: 400 });
    }
    const { integration_name, display_name, user_config, enabled = true } = validation.data;
    const now = new Date();
    const [result]: any = await pool.query(
      `INSERT INTO integration_configs 
         (integration_name, display_name, user_config, enabled, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [
        integration_name,
        display_name || integration_name,
        JSON.stringify(user_config || {}),
        enabled,
        now,
        now,
      ]
    );
    const insertedId = result.insertId;
    const [rows] = await pool.query(
      `SELECT 
         id,
         integration_name,
         display_name,
         user_config,
         enabled,
         created_at,
         updated_at
       FROM integration_configs
       WHERE id = ?`,
      [insertedId]
    );
    const row: any = (rows as any[])[0];
    const config: IntegrationConfig = {
      id: row.id,
      integration_name: row.integration_name,
      display_name: row.display_name,
      user_config:
        typeof row.user_config === "string"
          ? JSON.parse(row.user_config)
          : row.user_config,
      enabled: !!row.enabled,
      created_at: row.created_at,
      updated_at: row.updated_at,
    };
    return NextResponse.json(config, { status: 201 });
  } catch (error) {
    console.error("Error creating integration config:", error);
    return NextResponse.json({ error: "Failed to create config" }, { status: 500 });
  }
}
