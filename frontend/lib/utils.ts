import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Validate MongoDB ObjectId format
export function isValidObjectId(id: string): boolean {
  const objectIdPattern = /^[0-9a-fA-F]{24}$/;
  return objectIdPattern.test(id);
}

// Format MongoDB ObjectId for display
export function formatObjectId(id: string): string {
  return id.substring(0, 8); // Show first 8 characters
}

// Format timestamp for display
export function formatTimestamp(timestamp: string | Date): string {
  return new Date(timestamp).toLocaleString();
}