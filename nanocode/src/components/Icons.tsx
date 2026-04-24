/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

interface IconProps {
  active?: boolean;
  className?: string;
}

export function FolderIcon() {
  return (
    <svg
      width="13"
      height="13"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M1.5 3.5A1 1 0 0 1 2.5 2.5H6L7.5 4H13.5A1 1 0 0 1 14.5 5V12.5A1 1 0 0 1 13.5 13.5H2.5A1 1 0 0 1 1.5 12.5V3.5Z"
        stroke="currentColor"
        strokeWidth="1.3"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function SendIcon() {
  return (
    <svg
      width="14"
      height="14"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M8 13V3M8 3L3.5 7.5M8 3L12.5 7.5"
        stroke="currentColor"
        strokeWidth="1.6"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function AskIcon({ active }: IconProps) {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <circle
        cx="8"
        cy="8"
        r="6.5"
        stroke={active ? "rgba(255,255,255,0.85)" : "rgba(255,255,255,0.35)"}
        strokeWidth="1.3"
      />
      <path
        d="M6.2 6.2C6.2 5.09 7.09 4.2 8 4.2C8.91 4.2 9.8 5.09 9.8 6.2C9.8 7.31 8.91 7.8 8 8V9"
        stroke={active ? "rgba(255,255,255,0.85)" : "rgba(255,255,255,0.35)"}
        strokeWidth="1.3"
        strokeLinecap="round"
      />
      <circle
        cx="8"
        cy="11.5"
        r="0.7"
        fill={active ? "rgba(255,255,255,0.85)" : "rgba(255,255,255,0.35)"}
      />
    </svg>
  );
}

export function CodeIcon({ active }: IconProps) {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 16 16"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M10.5 4.5L13.5 8L10.5 11.5"
        stroke={active ? "rgba(255,255,255,0.85)" : "rgba(255,255,255,0.35)"}
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M5.5 4.5L2.5 8L5.5 11.5"
        stroke={active ? "rgba(255,255,255,0.85)" : "rgba(255,255,255,0.35)"}
        strokeWidth="1.4"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
