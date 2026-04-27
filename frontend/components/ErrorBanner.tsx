interface Props {
  message: string;
  onRetry?: () => void;
}

export function ErrorBanner({ message, onRetry }: Props) {
  return (
    <div className="rounded-md bg-red-50 border border-red-200 p-4 flex items-start gap-3">
      <span className="text-red-500 font-semibold shrink-0">Error:</span>
      <span className="text-red-700 flex-1 text-sm">{message}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          className="text-sm text-red-600 underline hover:text-red-800 shrink-0"
        >
          Retry
        </button>
      )}
    </div>
  );
}
