import { describe, expect, test } from "bun:test";
import { applyToolExecutionResultToMessages } from "./toolExecutionUpdate";

describe("applyToolExecutionResultToMessages", () => {
  const assistantId = "assistant-1";
  const callId = "call-1";

  const createMessages = () => [
    {
      id: assistantId,
      role: "assistant",
      content: "",
      toolCalls: [
        {
          id: callId,
          name: "read_file",
          arguments: {},
          status: "running",
        },
      ],
      blocks: [
        { type: "tool_call", call: { id: callId, name: "read_file", arguments: {}, status: "running" } },
        { type: "tool_result", callId, status: "running" },
      ],
    },
  ];

  test("keeps toolCalls and tool_result block consistent in a single update", () => {
    const updated = applyToolExecutionResultToMessages(
      createMessages(),
      assistantId,
      callId,
      "success",
      "done"
    );

    const message = updated[0];
    expect(message.toolCalls?.[0].status).toBe("success");
    expect(message.toolCalls?.[0].result).toBe("done");

    const resultBlock = message.blocks?.find(
      (block) => block.type === "tool_result" && block.callId === callId
    );

    expect(resultBlock).toEqual({
      type: "tool_result",
      callId,
      status: "success",
      result: "done",
    });
  });

  test("applies same shape for error path", () => {
    const updated = applyToolExecutionResultToMessages(
      createMessages(),
      assistantId,
      callId,
      "error",
      "boom"
    );

    const message = updated[0];
    expect(message.toolCalls?.[0].status).toBe("error");
    expect(message.toolCalls?.[0].result).toBe("boom");

    const resultBlock = message.blocks?.find(
      (block) => block.type === "tool_result" && block.callId === callId
    );

    expect(resultBlock).toEqual({
      type: "tool_result",
      callId,
      status: "error",
      result: "boom",
    });
  });
});
