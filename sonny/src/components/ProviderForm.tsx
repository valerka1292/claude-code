import React from 'react';
import type { Provider } from '../types';

export interface ProviderFormValues {
  visualName: string;
  baseUrl: string;
  apiKey: string;
  model: string;
  contextWindowSize: number;
}

interface ProviderFormProps {
  initialValues?: Provider;
  disabled?: boolean;
  onSubmit: (values: ProviderFormValues) => Promise<void>;
  onTest: (values: ProviderFormValues) => Promise<void>;
  submitLabel: string;
  testLabel?: string;
}

const DEFAULT_VALUES: ProviderFormValues = {
  visualName: '',
  baseUrl: '',
  apiKey: '',
  model: '',
  contextWindowSize: 128000,
};

export default function ProviderForm({
  initialValues,
  disabled = false,
  onSubmit,
  onTest,
  submitLabel,
  testLabel = 'Test',
}: ProviderFormProps) {
  const [values, setValues] = React.useState<ProviderFormValues>(DEFAULT_VALUES);

  React.useEffect(() => {
    if (!initialValues) {
      setValues(DEFAULT_VALUES);
      return;
    }

    setValues({
      visualName: initialValues.visualName,
      baseUrl: initialValues.baseUrl,
      apiKey: initialValues.apiKey,
      model: initialValues.model,
      contextWindowSize: initialValues.contextWindowSize,
    });
  }, [initialValues]);

  const onFieldChange =
    (field: keyof ProviderFormValues) => (event: React.ChangeEvent<HTMLInputElement>) => {
      const value = field === 'contextWindowSize' ? Number(event.target.value) : event.target.value;
      setValues((prev) => ({
        ...prev,
        [field]: value,
      }));
    };

  const isInvalid = !values.visualName || !values.baseUrl || !values.model || values.contextWindowSize <= 0;

  return (
    <form
      className="space-y-3 rounded-lg border border-border bg-bg-1 p-4"
      onSubmit={(event) => {
        event.preventDefault();
        void onSubmit(values);
      }}
    >
      <input
        className="w-full rounded-md border border-border bg-bg-0 px-3 py-2 text-sm text-text-primary outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        placeholder="Visual name"
        value={values.visualName}
        onChange={onFieldChange('visualName')}
      />
      <input
        className="w-full rounded-md border border-border bg-bg-0 px-3 py-2 text-sm text-text-primary outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        placeholder="Base URL"
        value={values.baseUrl}
        onChange={onFieldChange('baseUrl')}
      />
      <input
        className="w-full rounded-md border border-border bg-bg-0 px-3 py-2 text-sm text-text-primary outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        placeholder="Model"
        value={values.model}
        onChange={onFieldChange('model')}
      />
      <input
        className="w-full rounded-md border border-border bg-bg-0 px-3 py-2 text-sm text-text-primary outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        placeholder="API key"
        value={values.apiKey}
        onChange={onFieldChange('apiKey')}
      />
      <input
        className="w-full rounded-md border border-border bg-bg-0 px-3 py-2 text-sm text-text-primary outline-none focus-visible:ring-2 focus-visible:ring-focus-ring"
        placeholder="Context window size"
        type="number"
        min={1}
        value={values.contextWindowSize}
        onChange={onFieldChange('contextWindowSize')}
      />
      <div className="flex items-center justify-end gap-2 pt-2">
        <button
          type="button"
          onClick={() => void onTest(values)}
          disabled={disabled || isInvalid}
          className="rounded-md border border-border px-3 py-1.5 text-sm text-text-primary transition-colors hover:bg-bg-2 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {testLabel}
        </button>
        <button
          type="submit"
          disabled={disabled || isInvalid}
          className="rounded-md bg-accent px-3 py-1.5 text-sm font-medium text-accent-fg transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {submitLabel}
        </button>
      </div>
    </form>
  );
}
