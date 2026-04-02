const KEY = "accessibility_mode";

export function getAccessibilityMode(): boolean {
  if (typeof window === "undefined") return false;
  return localStorage.getItem(KEY) === "1";
}

export function applyAccessibilityMode(enabled: boolean): void {
  if (typeof document === "undefined") return;
  if (enabled) {
    document.documentElement.classList.add("a11y-mode");
  } else {
    document.documentElement.classList.remove("a11y-mode");
  }
  localStorage.setItem(KEY, enabled ? "1" : "0");
}
