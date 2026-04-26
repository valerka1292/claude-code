import React from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { X, Plus, Trash2, Edit2, Key, Globe, Cpu } from 'lucide-react';
import { Provider } from '../types';
import { MOCK_PROVIDERS } from '../constants';
import { cn } from '../lib/utils';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export default function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  const [providers, setProviders] = React.useState<Provider[]>(MOCK_PROVIDERS);
  const [activeTab, setActiveTab] = React.useState<'providers' | 'general'>('providers');

  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
        <Dialog.Content 
          aria-describedby="settings-description"
          className="fixed left-1/2 top-1/2 z-50 w-full max-w-2xl max-h-[90vh] -translate-x-1/2 -translate-y-1/2 flex flex-col bg-bg-0 border border-border rounded-xl shadow-2xl overflow-hidden data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%]"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-border bg-bg-0 flex-shrink-0">
            <Dialog.Title className="text-lg font-medium text-text-primary">Settings</Dialog.Title>
            <Dialog.Close aria-label="Close" className="p-1.5 hover:bg-bg-2 text-text-secondary hover:text-text-primary rounded-md transition-colors outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
              <X size={18} />
            </Dialog.Close>
          </div>

          <Dialog.Description id="settings-description" className="sr-only">
            Configuration preferences for your agent workspace.
          </Dialog.Description>

          <div className="flex flex-1 overflow-hidden min-h-[400px]">
            {/* Sidebar Tabs */}
            <div className="w-[180px] border-r border-border bg-bg-1 p-4 flex flex-col gap-1 overflow-y-auto">
              <button
                onClick={() => setActiveTab('general')}
                className={cn(
                  "flex items-center justify-start gap-2 px-3 py-2 rounded-md text-sm transition-colors w-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring",
                  activeTab === 'general' ? 'bg-bg-3 text-text-primary font-medium' : 'hover:bg-bg-2 text-text-secondary text-text-primary hover:text-text-primary'
                )}
              >
                General
              </button>
              <button
                onClick={() => setActiveTab('providers')}
                className={cn(
                  "flex items-center justify-start gap-2 px-3 py-2 rounded-md text-sm transition-colors w-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring",
                  activeTab === 'providers' ? 'bg-bg-3 text-text-primary font-medium' : 'hover:bg-bg-2 text-text-secondary hover:text-text-primary'
                )}
              >
                Providers
              </button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto p-6 bg-bg-0">
              {activeTab === 'providers' && (
                <div className="flex flex-col gap-6">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-medium text-text-primary">Model Providers</h3>
                      <p className="text-sm text-text-secondary mt-1">Manage your AI backend connections.</p>
                    </div>
                    <button className="flex items-center gap-2 px-3 py-1.5 bg-accent text-accent-fg rounded-md text-sm font-medium hover:opacity-90 transition-opacity focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
                      <Plus size={16} />
                      Add connection
                    </button>
                  </div>

                  <div className="flex flex-col gap-3">
                    {providers.map((p) => (
                      <div key={p.id} className="p-4 border border-border rounded-lg bg-bg-1 hover:border-bg-3 transition-colors group">
                        <div className="flex items-start justify-between gap-3 mb-4">
                          <div className="flex items-center gap-3 min-w-0 flex-1">
                            <div className="w-10 h-10 bg-gradient-to-br from-bg-2 to-bg-3 border border-border rounded-lg flex items-center justify-center flex-shrink-0">
                              <Cpu size={18} className="text-text-primary" />
                            </div>
                            <div className="min-w-0">
                              <div className="text-sm font-medium text-text-primary mb-0.5 truncate">{p.name}</div>
                              <div className="text-xs text-text-secondary font-mono truncate">{p.model}</div>
                            </div>
                          </div>
                          
                          {/* Actions visible on mobile, highlighted on hover */}
                          <div className="flex items-center gap-1 text-text-secondary md:opacity-60 group-hover:opacity-100 transition-opacity">
                            <button 
                              aria-label="Edit provider"
                              className="p-1.5 hover:bg-bg-3 hover:text-text-primary rounded-md transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
                            >
                              <Edit2 size={14} />
                            </button>
                            <button 
                              aria-label="Delete provider"
                              className="p-1.5 hover:bg-red-500/10 hover:text-red-400 rounded-md transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
                            >
                              <Trash2 size={14} />
                            </button>
                          </div>
                        </div>

                        {/* Meta info */}
                        <div className="flex flex-col gap-2 text-xs">
                          <div className="flex items-center gap-2 text-text-secondary">
                            <Globe size={12} className="flex-shrink-0" />
                            <span className="truncate font-mono">{p.baseUrl}</span>
                          </div>
                          <div className="flex items-center gap-2 text-text-secondary">
                            <Key size={12} className="flex-shrink-0" />
                            <span className="font-mono">{p.apiKey ? '••••••••' : 'No key'}</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {activeTab === 'general' && (
                <div className="text-sm text-text-secondary">
                  General settings would go here...
                </div>
              )}
            </div>
          </div>

          {/* Footer */}
          <div className="px-6 py-4 border-t border-border bg-bg-1 flex justify-end gap-3 rounded-b-xl flex-shrink-0">
            <Dialog.Close className="px-4 py-2 text-sm font-medium text-text-primary hover:bg-bg-3 rounded-md transition-colors outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
              Cancel
            </Dialog.Close>
            <button className="px-4 py-2 text-sm font-medium bg-accent text-accent-fg rounded-md hover:opacity-90 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-focus-ring">
              Save changes
            </button>
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
