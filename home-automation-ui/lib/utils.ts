import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function toUTCTimestamp(date: string | Date | null) {
  if (!date) return null;
  const d = new Date(date);
  return d.toISOString().slice(0, 19).replace("T", " "); // 'YYYY-MM-DD HH:mm:ss'
}

export function parseJsonOrFallback(val: any): any {
  if (val === null || val === undefined) return null;
  if (typeof val !== "string") return val;

  try {
    return JSON.parse(val);
  } catch {
    return val;
  }
}
export function capitalize(s: string): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
} 
