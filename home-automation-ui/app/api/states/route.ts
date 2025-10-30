"use server";
import { NextRequest, NextResponse } from 'next/server'
import type { State } from '@/types/state'

const ENGINE_BASE_URL = 'http://localhost:8080/api/states'

export async function GET(req: NextRequest) {
  const { searchParams } = req.nextUrl
  const entity_id = searchParams.get('entity_id')

  if (!entity_id) {
    return NextResponse.json(
      { error: 'Missing entity_id query parameter' },
      { status: 400 }
    )
  }

  try {
    const engineUrl = `${ENGINE_BASE_URL}?entity_id=${encodeURIComponent(entity_id)}`
    const engineRes = await fetch(engineUrl, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    })

    if (!engineRes.ok) {
      return NextResponse.json(
        { error: `Engine request failed with status ${engineRes.status}` },
        { status: engineRes.status }
      )
    }

    const data = await engineRes.json()
    const state: State | undefined = data.state

    if (!state || typeof state.entity_id !== 'string') {
      return NextResponse.json(
        { error: 'Invalid response from engine' },
        { status: 500 }
      )
    }

    return NextResponse.json(state)
  } catch (error: any) {
    console.error('Error fetching entity state:', error)
    return NextResponse.json(
      { error: 'Failed to fetch entity state', details: error.message },
      { status: 500 }
    )
  }
}
