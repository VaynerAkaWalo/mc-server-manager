export function sanitizeString(rawString: string): string {
  return rawString.replace(/[^a-zA-Z ]/g, '')
}
