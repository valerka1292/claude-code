export function pathToProjectKey(absPath: string): string {
  return absPath
    .replace(/\\/g, "/")
    .replace(/^\//, "")
    .replace(/\/$/, "")
    .replace(/\//g, "-");
}
