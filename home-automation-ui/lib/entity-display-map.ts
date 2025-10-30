export const ENTITY_STATE_KEY_MAP: Record<string, { trueLabel: string; falseLabel: string; label?: string }> = {
  light: { trueLabel: "On", falseLabel: "Off", label: "Power" },
  grouped_light: { trueLabel: "On", falseLabel: "Off", label: "Power" },
  scene: { trueLabel: "Active", falseLabel: "Inactive", label: "Activation" },
}

export const ENTITY_ICON_MAP: Record<string, string> = {
  light: "💡",
  grouped_light: "💡💡",
  scene: "🎬",
}
