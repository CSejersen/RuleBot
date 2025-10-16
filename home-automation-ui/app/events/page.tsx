"use client"

import { useEffect, useState } from "react"
import { columns } from "./columns"
import { Event } from "@/app/api/events/route"
import { DataTable } from "./data-table"

async function getData(): Promise<Event[]> {
  const res = await fetch("/api/events", { cache: "no-store" });

  if (!res.ok) {
    console.error("Failed to fetch events:", res.statusText);
    return [];
  }

  const data = (await res.json()) as Event[];
  return data;
}

function parseEventMessage(msg: any): Event {
  const raw = typeof msg === "string" ? JSON.parse(msg) : msg

  return {
    id: raw.Event.Id || "",
    source: raw.Event.Source,
    type: raw.Event.Type,
    entity: raw.Event.Entity,
    stateChange: raw.Event.StateChange,
    payload: raw.Event.Payload || {},
    timestamp: raw.Event.Time,
    triggeredRules: Array.isArray(raw.TriggeredRules) ? raw.TriggeredRules : [],
  }
}

export default function Dashboard() {
  const [events, setEvents] = useState<Event[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Fetch initial events from DB
    getData()
      .then((data) => setEvents(data))
      .catch((err) => {
        console.error(err);
        setError("Failed to load events");
      })
      .finally(() => setLoading(false));

    // Connect to WebSocket
    const ws = new WebSocket("ws://localhost:8080/ws");

    ws.onmessage = (message) => {
      try {
        const event = parseEventMessage(message.data);
        setEvents((prev) => [event, ...prev]);
      } catch (err) {
        console.error("Failed to parse WS message", err);
      }
    };

    return () => ws.close();
  }, []);

  if (loading) return <div>Loading events...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="container mx-auto py-10">
      <h2 className="text-2xl font-semibold tracking-tight mb-4">
        Processed Events
      </h2>

      <DataTable columns={columns} data={events} />
    </div >
  );
}
