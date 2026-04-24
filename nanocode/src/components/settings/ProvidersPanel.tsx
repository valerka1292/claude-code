/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { Plus, Pencil, Trash2 } from "lucide-react";
import { useProviders } from "../../contexts/ProvidersContext";
import type { Provider } from "../../lib/providers";
import { ProviderForm } from "./ProviderForm";
import { EditProviderFlow } from "./EditProviderFlow";

type PanelView =
  | { type: "list" }
  | { type: "create" }
  | { type: "edit"; providerId: string };

export function ProvidersPanel() {
  const { store, activeProvider, addProvider, removeProvider, switchProvider } =
    useProviders();
  const [view, setView] = useState<PanelView>({ type: "list" });

  if (view.type === "create") {
    return (
      <ProviderForm
        onCancel={() => setView({ type: "list" })}
        onSave={(data) => {
          addProvider(data);
          setView({ type: "list" });
        }}
      />
    );
  }

  if (view.type === "edit") {
    const p = store.providers[view.providerId];
    if (!p) {
      setView({ type: "list" });
      return null;
    }
    return (
      <EditProviderFlow
        provider={p}
        onBack={() => setView({ type: "list" })}
      />
    );
  }

  const providers = Object.values(store.providers);

  return (
    <div className="p-5 flex flex-col gap-4">
      {activeProvider && (
        <div className="
          rounded-xl bg-white/[0.03] border border-white/[0.07]
          px-4 py-3 flex items-center gap-3
        ">
          <div className="flex-1 min-w-0">
            <div className="text-[0.65rem] text-white/25 font-mono uppercase tracking-wider mb-0.5">
              Active
            </div>
            <div className="text-[0.82rem] text-white/80 font-medium truncate">
              {activeProvider.name}
            </div>
            <div className="text-[0.7rem] text-white/30 font-mono truncate">
              {activeProvider.model}
            </div>
          </div>
          <div className="w-2 h-2 rounded-full bg-emerald-400/70" />
        </div>
      )}

      <div className="flex flex-col gap-1.5">
        {providers.length === 0 && (
          <div className="text-center py-8 text-[0.75rem] text-white/20 font-mono">
            No providers configured
          </div>
        )}
        <AnimatePresence initial={false}>
          {providers.map((p) => (
            <ProviderRow
              key={p.id}
              provider={p}
              isActive={p.id === store.activeProviderId}
              onActivate={() => switchProvider(p.id)}
              onEdit={() => setView({ type: "edit", providerId: p.id })}
              onDelete={() => removeProvider(p.id)}
            />
          ))}
        </AnimatePresence>
      </div>

      <button
        onClick={() => setView({ type: "create" })}
        className="
          flex items-center gap-2 px-3 py-2.5 rounded-xl
          border border-dashed border-white/[0.1]
          text-[0.75rem] text-white/30 hover:text-white/60
          hover:border-white/[0.2] hover:bg-white/[0.03]
          transition-all duration-150 font-mono
        "
      >
        <Plus size={13} />
        Add provider
      </button>
    </div>
  );
}

interface ProviderRowProps {
  provider: Provider;
  isActive: boolean;
  onActivate: () => void;
  onEdit: () => void;
  onDelete: () => void;
}

function ProviderRow({
  provider,
  isActive,
  onActivate,
  onEdit,
  onDelete,
}: ProviderRowProps) {
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (
    <motion.div
      initial={{ opacity: 0, y: -4 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, y: -4 }}
      className={`
        flex items-center gap-3 px-3 py-2.5 rounded-xl
        border transition-all duration-150
        ${isActive
          ? "bg-white/[0.05] border-white/[0.1]"
          : "bg-transparent border-white/[0.05] hover:bg-white/[0.03]"
        }
      `}
    >
      <button
        onClick={onActivate}
        className="flex-shrink-0 flex items-center justify-center w-4 h-4"
      >
        {isActive ? (
          <div className="w-2 h-2 rounded-full bg-emerald-400/80" />
        ) : (
          <div className="w-2 h-2 rounded-full border border-white/20 hover:border-white/50 transition-colors" />
        )}
      </button>

      <div className="flex-1 min-w-0">
        <div className="text-[0.78rem] text-white/75 font-medium truncate">
          {provider.name}
        </div>
        <div className="text-[0.67rem] text-white/25 font-mono truncate">
          {provider.model}
        </div>
      </div>

      <div className="flex items-center gap-1 flex-shrink-0">
        <button
          onClick={onEdit}
          className="
            p-1.5 rounded-lg text-white/25 hover:text-white/60
            hover:bg-white/[0.06] transition-all duration-150
          "
        >
          <Pencil size={12} />
        </button>
        {confirmDelete ? (
          <div className="flex items-center gap-1">
            <button
              onClick={() => {
                onDelete();
                setConfirmDelete(false);
              }}
              className="px-2 py-1 rounded-lg text-[0.65rem] text-red-400/80 hover:bg-red-500/10 transition-all"
            >
              confirm
            </button>
            <button
              onClick={() => setConfirmDelete(false)}
              className="px-2 py-1 rounded-lg text-[0.65rem] text-white/30 hover:bg-white/[0.06] transition-all"
            >
              cancel
            </button>
          </div>
        ) : (
          <button
            onClick={() => setConfirmDelete(true)}
            className="
              p-1.5 rounded-lg text-white/25 hover:text-red-400/70
              hover:bg-red-500/[0.08] transition-all duration-150
            "
          >
            <Trash2 size={12} />
          </button>
        )}
      </div>
    </motion.div>
  );
}