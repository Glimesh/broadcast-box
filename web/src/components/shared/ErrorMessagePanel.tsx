
interface ErrorMessagePanelProps {
  title: string;
  message: string;
  className?: string;
}

export default function ErrorMessagePanel(props: ErrorMessagePanelProps) {
  const extraClassName = props.className ?? "";

  return (
    <div className={`error-panel ${extraClassName}`.trim()}>
      <div className="error-panel__head">
        <svg
          className="h-4 w-4 flex-shrink-0"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <span>{props.title}</span>
      </div>
      <p>{props.message}</p>
    </div>
  );
}
