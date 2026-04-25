/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { motion } from "motion/react";
import { CheckCircle, User, Wrench, XCircle } from "lucide-react";
import { ReasoningBlock } from "./ReasoningBlock";
import type { Message } from "../types/message";
import { extractTag, isToolError } from "../lib/tools/utils/format";

interface MessageItemProps {
  message: Message;
}

export function MessageItem({ message }: MessageItemProps) {
  const isUser = message.role === "user";
  const isTool = message.role === "tool";

  if (isUser) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 4 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.15 }}
        className="flex w-full gap-3 py-3 px-5 group"
      >
        <div className="flex-shrink-0 mt-0.5">
          <div className="w-5 h-5 rounded-md bg-white/[0.07] border border-white/[0.1] flex items-center justify-center">
            <User size={11} className="text-white/40" />
          </div>
        </div>
        <div className="flex-1 min-w-0">
          <div className="text-[0.68rem] font-mono text-white/20 mb-1 uppercase tracking-wider">
            you
          </div>
          <div className="text-[0.88rem] text-white/75 leading-relaxed whitespace-pre-wrap break-words">
            {message.content}
          </div>
        </div>
      </motion.div>
    );
  }

  if (isTool) {
    return <ToolResultMessage message={message} />;
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.15 }}
      className="flex w-full gap-3 py-3 px-5 group"
    >
      <div className="flex-shrink-0 mt-0.5">
        <div className="w-5 h-5 rounded-md bg-white flex items-center justify-center shadow-[0_0_8px_rgba(255,255,255,0.15)]">
          <div className="w-2 h-2 bg-black rounded-full" />
        </div>
      </div>

      <div className="flex-1 min-w-0">
        <div className="text-[0.68rem] font-mono text-white/20 mb-1 uppercase tracking-wider">
          nanocode
        </div>

        {(message.reasoning || message.isReasoningStreaming) && (
          <ReasoningBlock
            content={message.reasoning ?? ""}
            isStreaming={message.isReasoningStreaming}
          />
        )}

        {message.toolCalls && message.toolCalls.length > 0 && (
          <ToolCallsBlock toolCalls={message.toolCalls} />
        )}

        {message.content ? (
          <div className="text-[0.88rem] text-white/70 leading-relaxed whitespace-pre-wrap break-words">
            {message.content}
            {message.isStreaming && (
              <span className="inline-block w-0.5 h-3.5 bg-white/40 ml-0.5 align-middle animate-pulse" />
            )}
          </div>
        ) : message.isStreaming && !message.isReasoningStreaming ? (
          <div className="flex gap-1 items-center h-5">
            {[0, 0.15, 0.3].map((delay, i) => (
              <span
                key={i}
                className="w-1 h-1 rounded-full bg-white/30 animate-pulse"
                style={{ animationDelay: `${delay}s` }}
              />
            ))}
          </div>
        ) : null}
      </div>
    </motion.div>
  );
}

function ToolCallsBlock({ toolCalls }: { toolCalls: Message["toolCalls"] }) {
  if (!toolCalls || toolCalls.length === 0) return null;

  return (
    <div className="mb-2 space-y-1">
      {toolCalls.map((tc) => (
        <div
          key={tc.id}
          className="
            flex items-start gap-2 px-3 py-2 rounded-lg
            bg-white/[0.03] border border-white/[0.06]
          "
        >
          <Wrench size={12} className="text-blue-400/60 mt-0.5 flex-shrink-0" />
          <div className="flex-1 min-w-0">
            <div className="text-[0.72rem] text-white/50 font-mono">
              {tc.name}
            </div>
            <div className="text-[0.67rem] text-white/25 font-mono mt-0.5 whitespace-pre-wrap">
              {JSON.stringify(tc.arguments, null, 2)}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

function ToolResultMessage({ message }: { message: Message }) {
  const isError = isToolError(message.content);
  const errorContent = isError
    ? extractTag(message.content, "tool_use_error")
    : null;

  return (
    <motion.div
      initial={{ opacity: 0, y: 4 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.15 }}
      className="flex w-full gap-3 py-2 px-5"
    >
      <div className="flex-shrink-0 mt-0.5">
        {isError ? (
          <XCircle size={14} className="text-red-400/60" />
        ) : (
          <CheckCircle size={14} className="text-emerald-400/60" />
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="text-[0.67rem] font-mono text-white/25 mb-1">
          {message.toolName || "tool"} result
        </div>
        <div
          className={`
            text-[0.75rem] font-mono leading-relaxed whitespace-pre-wrap break-words
            ${isError ? "text-red-400/70" : "text-white/40"}
          `}
        >
          {isError ? errorContent : message.content}
        </div>
      </div>
    </motion.div>
  );
}
