/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { motion } from "motion/react";
import { AlertTriangle, File, Loader2, Search, Wrench } from "lucide-react";
import type { ToolCallDisplay } from "../types/message";

interface ToolCallBlockProps {
  toolCall: ToolCallDisplay;
  statusOverride?: ToolCallDisplay["status"];
  resultOverride?: string;
}

export function ToolCallBlock({ toolCall, statusOverride, resultOverride }: ToolCallBlockProps) {
  const { name } = toolCall;

  if (name === "Glob") {
    return <GlobBlock toolCall={toolCall} statusOverride={statusOverride} resultOverride={resultOverride} />;
  }

  return <GenericToolBlock toolCall={toolCall} statusOverride={statusOverride} resultOverride={resultOverride} />;
}

function GlobBlock({
  toolCall,
  statusOverride,
  resultOverride,
}: {
  toolCall: ToolCallDisplay;
  statusOverride?: ToolCallDisplay["status"];
  resultOverride?: string;
}) {
  const { arguments: args } = toolCall;
  const pattern = args.pattern as string | undefined;
  const path = args.path as string | undefined;

  const effectiveStatus = statusOverride ?? toolCall.status;
  const effectiveResult = resultOverride ?? toolCall.result;

  const searching = effectiveStatus === "pending" || effectiveStatus === "running";

  const patternDisplay = pattern ?? "";
  const pathDisplay = path ? ` in ${path}/` : "";
  const verb = searching
    ? `Searching for ${patternDisplay}${pathDisplay}`
    : `Searched for ${patternDisplay}${pathDisplay}`;

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      className="my-2 rounded-lg bg-[#0f0f0f] border border-white/[0.06] overflow-hidden"
    >
      <div className="flex items-center gap-2 px-3 py-2 text-[0.7rem] text-white/40">
        {searching ? (
          <Loader2 size={12} className="animate-spin text-white/30" />
        ) : (
          <Search size={12} className="text-white/30" />
        )}
        <span>{verb}</span>
      </div>

      {effectiveStatus !== "pending" && effectiveStatus !== "running" && effectiveResult !== undefined && (
        <div className="border-t border-white/[0.06] mx-3" />
      )}

      {effectiveStatus === "success" && effectiveResult !== undefined && (
        <div className="px-3 pb-2 pt-1.5 text-[0.75rem] font-mono leading-relaxed text-white/60">
          {renderGlobResult(effectiveResult)}
        </div>
      )}

      {effectiveStatus === "error" && effectiveResult !== undefined && (
        <div className="px-3 pb-2 pt-1.5 text-[0.75rem] text-red-400/80 flex items-start gap-1.5">
          <AlertTriangle size={13} className="mt-0.5 flex-shrink-0" />
          <span>{formatGlobError(effectiveResult)}</span>
        </div>
      )}
    </motion.div>
  );
}

function renderGlobResult(rawOutput: string) {
  const lines = rawOutput.trim().split("\n").filter(Boolean);
  if (lines.length === 0 || rawOutput === "No files found") {
    return <div className="text-white/30 italic">No files found matching the pattern.</div>;
  }

  const lastLine = lines[lines.length - 1];
  const truncated = lastLine.startsWith("(Results are truncated");
  const fileLines = truncated ? lines.slice(0, -1) : lines;

  const SHOW_MAX = 10;
  const visible = fileLines.slice(0, SHOW_MAX);
  const hidden = fileLines.length - visible.length;

  return (
    <div className="space-y-0.5">
      {visible.map((file, i) => (
        <div key={i} className="flex items-center gap-1.5">
          <File size={11} className="text-white/20 flex-shrink-0" />
          <span className="truncate">{file}</span>
        </div>
      ))}
      {hidden > 0 && (
        <div className="text-white/25 pl-4 mt-1 text-[0.65rem]">
          ... and {hidden} more files
        </div>
      )}
      {truncated && (
        <div className="text-white/25 pl-4 mt-1 text-[0.65rem]">
          Results truncated, try narrowing the pattern or path.
        </div>
      )}
    </div>
  );
}

function formatGlobError(errorText: string): string {
  if (errorText.includes("<tool_use_error>")) {
    const match = errorText.match(/<tool_use_error>(.*?)<\/tool_use_error>/s);
    return match?.[1] ?? errorText;
  }
  return errorText;
}

function GenericToolBlock({
  toolCall,
  statusOverride,
  resultOverride,
}: {
  toolCall: ToolCallDisplay;
  statusOverride?: ToolCallDisplay["status"];
  resultOverride?: string;
}) {
  const { name } = toolCall;
  const effectiveStatus = statusOverride ?? toolCall.status;
  const effectiveResult = resultOverride ?? toolCall.result;
  const running = effectiveStatus === "running";

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      className="my-2 rounded-lg bg-[#0f0f0f] border border-white/[0.06] p-3"
    >
      <div className="flex items-center gap-2 text-[0.7rem] text-white/40 mb-1">
        <Wrench size={12} />
        <span>{name}</span>
        {running && <Loader2 size={12} className="animate-spin ml-auto" />}
      </div>
      {effectiveResult && (
        <pre className="text-[0.7rem] text-white/50 font-mono whitespace-pre-wrap break-all mt-1">
          {effectiveResult}
        </pre>
      )}
    </motion.div>
  );
}