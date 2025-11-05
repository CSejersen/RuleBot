"use client"

import { useState, useEffect } from "react"
import { Event } from "@/types/events"
import { getServiceCallStatus } from "./causality-chain-utils"

type ServiceCallStatus = "pending" | "aggregated" | "has_descendants" | "not_service_call"

interface UseServiceCallStatusResult {
  status: ServiceCallStatus
  secondsUntilAggregated: number | null
}

/**
 * Hook that returns the current status of a service call event
 * Updates every second for pending events to enable smooth transitions
 * Automatically stops updating once event transitions to aggregated or has descendants
 */
export function useServiceCallStatus(
  event: Event,
  allEvents: Event[]
): UseServiceCallStatusResult {
  const [currentTime, setCurrentTime] = useState(Date.now())
  
  // Calculate event time once
  const eventTime = new Date(event.time_fired).getTime()
  
  // Recalculate status based on current time
  const currentStatus = getServiceCallStatus(event, allEvents)
  const currentAgeSeconds = (currentTime - eventTime) / 1000
  const secondsUntilAggregated = 
    currentStatus === "pending" && currentAgeSeconds >= 0 && currentAgeSeconds < 5
      ? Math.ceil(5 - currentAgeSeconds)
      : null
  
  // Set up interval to update time every second
  // Only run if event is pending (no need to update otherwise)
  useEffect(() => {
    // Recalculate status to check if we should continue updating
    const status = getServiceCallStatus(event, allEvents)
    if (status !== "pending") {
      return // Don't set up interval if not pending
    }
    
    const interval = setInterval(() => {
      setCurrentTime(Date.now())
    }, 1000)
    
    // Cleanup interval on unmount or when dependencies change
    return () => clearInterval(interval)
  }, [event.time_fired, allEvents]) // Only depend on event time and allEvents, not currentTime
  
  return {
    status: currentStatus,
    secondsUntilAggregated: secondsUntilAggregated,
  }
}

