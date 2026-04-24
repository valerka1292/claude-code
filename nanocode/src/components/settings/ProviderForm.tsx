/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { ArrowLeft } from "lucide-react";
import type { Provider } from "../../lib/providers";

interface ProviderFormProps {
  initial?: Partial<Omit<Provider, "id">>;
  onSave: (data: Omit<Provider, "id">) => void;
  onCancel: () => void;
  saveLabel?: string;
}

const FIELDS: {
  key: keyof Omit<Provider, "id">;
  label: string;
  placeholder: string;
  type?: string;
}[] = [
  { key: "name", label: "Name", placeholder: "My Provider" },
  { key: "baseUrl", label: "Base URL", placeholder: "https://api.openai.com/v1" },
  { key: "model", label: "Model", placeholder: "gpt-4o" },
  { key: "apiKey", label: "API Key", placeholder: "sk-...", type: "password" },
  { key: "contextWindow", label: "Context Window", placeholder: "128000", type: "number" },
];

export function ProviderForm({
  initial = {},
  onSave,
  onCancel,
  saveLabel = "Save",
}: ProviderFormProps) {
  const [values, setValues] = useState<Record<string, string>>({
    name: initial.name ?? "",
    baseUrl: initial.baseUrl ?? "",
    model: initial.model ?? "",
    apiKey: initial.apiKey ?? "",
    contextWindow:
      initial.contextWindow != null ? String(initial.contextWindow) : "",
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const set = (key: string, val: string) => {
    setValues((v) => ({ ...v, [key]: val }));
    setErrors((e) => ({ ...e, [key]: "" }));
  };

  const validate = () => {
    const errs: Record<string, string> = {};
    if (!values.name.trim()) errs.name = "Required";
    if (!values.baseUrl.trim()) errs.baseUrl = "Required";
    if (!values.model.trim()) errs.model = "Required";
    if (!values.apiKey.trim()) errs.apiKey = "Required";
    const cw = Number(values.contextWindow);
    if (values.contextWindow && (isNaN(cw) || cw < 1))
      errs.contextWindow = "Must be a positive number";
    return errs;
  };

  const handleSave = () => {
    const errs = validate();
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    onSave({
      name: values.name.trim(),
      baseUrl: values.baseUrl.trim(),
      model: values.model.trim(),
      apiKey: values.apiKey.trim(),
      contextWindow: values.contextWindow ? Number(values.contextWindow) : 128000,
    });
  };

  return (
    <div className="p-5 flex flex-col gap-4">
      <button
        onClick={onCancel}
        className="flex items-center gap-1.5 text-[0.72rem] text-white/30 hover:text-white/60 transition-colors w-fit"
      >
        <ArrowLeft size={13} />
        Back
      </button>

      <div className="flex flex-col gap-3">
        {FIELDS.map(({ key, label, placeholder, type }) => (
          <div key={key} className="flex flex-col gap-1">
            <label className="text-[0.67rem] font-mono text-white/30 uppercase tracking-wider">
              {label}
            </label>
            <input
              type={type ?? "text"}
              value={values[key]}
              onChange={(e) => set(key, e.target.value)}
              placeholder={placeholder}
              className={`
                w-full bg-[#0f0f0f] border rounded-lg px-3 py-2
                text-[0.82rem] font-mono text-white/75
                placeholder-white/15 outline-none
                transition-all duration-150
                ${errors[key]
                  ? "border-red-500/50 focus:border-red-400/70"
                  : "border-white/[0.08] focus:border-white/[0.2]"
                }
              `}
            />
            {errors[key] && (
              <span className="text-[0.65rem] text-red-400/70">
                {errors[key]}
              </span>
            )}
          </div>
        ))}
      </div>

      <div className="flex gap-2 pt-1">
        <button
          onClick={handleSave}
          className="
            px-4 py-2 rounded-lg bg-white/[0.08] border border-white/[0.1]
            text-[0.75rem] text-white/80 hover:bg-white/[0.12] hover:text-white
            transition-all duration-150 font-mono
          "
        >
          {saveLabel}
        </button>
        <button
          onClick={onCancel}
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