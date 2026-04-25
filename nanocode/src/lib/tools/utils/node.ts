/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export async function importNodeModule<T>(specifier: string): Promise<T> {
  if (!hasNodeRuntime()) {
    throw new Error(`Cannot import ${specifier}: Node.js runtime not available`);
  }

  return import(specifier) as Promise<T>;
}

export function hasNodeRuntime(): boolean {
  if (
    typeof window !== "undefined" &&
    typeof (window as Window & { electronAPI?: unknown }).electronAPI !==
      "undefined"
  ) {
    return true;
  }

  return typeof process !== "undefined" && !!process.versions?.node;
}
