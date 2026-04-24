/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { ArrowLeft, ChevronRight } from "lucide-react";
import type { Provider } from "../../lib/providers";
import { useProviders } from "../../contexts/ProvidersContext";

type EditField = keyof Omit<Provider, "id"> | null;

interface EditProviderFlowProps {
  provider: Provider;
  onBack: () => void;
}

const FIELD_LABELS: Record<keyof Omit<Provider, "id">, string> = {
  name: "Name",
  baseUrl: "Base URL",
  model: "Model",
  apiKey: "API Key",
  contextWindow: "Context Window",
};

export function EditProviderFlow({ provider, onBack }: EditProviderFlowProps) {
  const { editProvider } = useProviders();
  const [editingField, setEditingField] = useState<EditField>(null);
  const [fieldValue, setFieldValue] = useState("");

  if (editingField !== null) {
    const label = FIELD_LABELS[editingField];
    const isPassword = editingField === "apiKey";
    const isNumber = editingField === "contextWindow";

    const handleSave = () => {
      if (!fieldValue.trim() && editingField !== "contextWindow") return;
      const val = isNumber ? Number(fieldValue) : fieldValue.trim();
      editProvider(provider.id, { [editingField]: val });
      setEditingField(null);
    };

    return (
      <div className="p-5 flex flex-col gap-4">
        <button
          onClick={() => setEditingField(null)}
          className="flex items-center gap-1.5 text-[0.72rem] text-white/30 hover:text-white/60 transition-colors w-fit"
        >
          <ArrowLeft size={13} />
          Back to fields
        </button>

        <div className="flex flex-col gap-2">
          <label className="text-[0.67rem] font-mono text-white/30 uppercase tracking-wider">
            {label}
          </label>
          <div className="text-[0.65rem] text-white/20 font-mono mb-1">
            Current:{" "}
            <span className="text-white/40">
              {isPassword
                ? "\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022"
                : String(provider[editingField] ?? "")}
            </span>
          </div>
          <input
            type={isPassword ? "password" : isNumber ? "number" : "text"}
            value={fieldValue}
            onChange={(e) => setFieldValue(e.target.value)}
            placeholder={`New ${label.toLowerCase()}...`}
            autoFocus
            className="
              w-full bg-[#0f0f0f] border border-white/[0.08]
              focus:border-white/[0.2] rounded-lg px-3 py-2
              text-[0.82rem] font-mono text-white/75
              placeholder-white/15 outline-none transition-all duration-150
            "
          />
        </div>

        <div className="flex gap-2">
          <button
            onClick={handleSave}
            disabled={!fieldValue.trim()}
            className="
              px-4 py-2 rounded-lg bg-white/[0.08] border border-white/[0.1]
              text-[0.75rem] text-white/80 hover:bg-white/[0.12] hover:text-white
              disabled:opacity-30 disabled:cursor-not-allowed
              transition-all duration-150 font-mono
            "
          >
            Apply
          </button>
          <button
            onClick={() => setEditingField(null)}
            className="
              px-4 py-2 rounded-lg text-[0.75rem] text-white/30
              hover:text-white/60 hover:bg-white/[0.05]
              transition-all duration-150 font-mono
            "
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="p-5 flex flex-col gap-4">
      <button
        onClick={onBack}
        className="flex items-center gap-1.5 text-[0.72rem] text-white/30 hover:text-white/60 transition-colors w-fit"
      >
        <ArrowLeft size={13} />
        Back
      </button>

      <div className="text-[0.78rem] text-white/50 font-medium">
        {provider.name}
      </div>

      <div className="flex flex-col gap-1">
        {(Object.keys(FIELD_LABELS) as (keyof Omit<Provider, "id">)[]).map(
          (field) => (
            <button
              key={field}
              onClick={() => {
                setEditingField(field);
                setFieldValue("");
              }}
              className="
                flex items-center justify-between
                px-3 py-2.5 rounded-xl
                border border-white/[0.05] hover:border-white/[0.1]
                hover:bg-white/[0.03] transition-all duration-150 text-left
              "
            >
              <div>
                <div className="text-[0.72rem] text-white/50">
                  {FIELD_LABELS[field]}
                </div>
                <div className="text-[0.67rem] text-white/25 font-mono">
                  {field === "apiKey"
                    ? "\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022"
                    : String(provider[field] ?? "")}
                </div>
              </div>
              <ChevronRight size={13} className="text-white/20" />
            </button>
          )
        )}
      </div>
    </div>
  );
}