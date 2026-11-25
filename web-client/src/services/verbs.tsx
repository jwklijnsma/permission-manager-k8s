/**
 * Helpers to summarise Kubernetes RBAC verbs as "read" / "write" access in the UI.
 *
 * Kubernetes has no "read" verb (issue #186): read access is granted by get/list/watch.
 * Custom templates usually list explicit verbs instead of "*", so checking only for "*"
 * or "create" made the access view look broken for them (issue #84).
 */

export const READ_VERBS = ['get', 'list', 'watch']

export const WRITE_VERBS = ['create', 'update', 'patch', 'delete', 'deletecollection']

function normalize(verbs: readonly string[] | undefined | null): string[] {
  return (verbs || []).map(v => v.toLowerCase())
}

/**
 * true when the verbs grant (at least partial) read access
 */
export function grantsRead(verbs: readonly string[] | undefined | null): boolean {
  const v = normalize(verbs)
  return v.includes('*') || READ_VERBS.some(r => v.includes(r))
}

/**
 * true when the verbs grant (at least partial) write access
 */
export function grantsWrite(verbs: readonly string[] | undefined | null): boolean {
  const v = normalize(verbs)
  return v.includes('*') || WRITE_VERBS.some(w => v.includes(w))
}
