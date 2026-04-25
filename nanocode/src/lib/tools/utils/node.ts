/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

type DynamicImport = <T>(specifier: string) => Promise<T>;

const dynamicImport = new Function(
  "specifier",
  "return import(specifier);"
) as DynamicImport;

export async function importNodeModule<T>(specifier: string): Promise<T> {
  return dynamicImport<T>(specifier);
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
