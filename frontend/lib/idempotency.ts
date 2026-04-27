/** Generate a UUID v4 idempotency key per form submission attempt */
export function generateIdempotencyKey(): string {
  return crypto.randomUUID();
}
