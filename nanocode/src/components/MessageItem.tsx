/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { motion } from "motion/react";
import { User } from "lucide-react";
import { ReasoningBlock } from "./ReasoningBlock";

export interface Message {
  id: string;
  role: "user" | "assistant";
  content: string;
  reasoning?: string;
  isStreaming?: boolean;
  isReasoningStreaming?: boolean;
}

interface MessageItemProps {
  message: Message;
}

export function MessageItem({ message }: MessageItemProps) {
  const isUser = message.role === "user";

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